package api

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"

	"AI-Note-Taker/notes"
	"AI-Note-Taker/storage"
	"AI-Note-Taker/transcription"
)

type Handler struct {
	DB *sql.DB
}

type RegenerateRequest struct {
	Prompt string `json:"prompt"`
}

func (h *Handler) DeleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")

	upload, err := GetUploadByID(h.DB, id)
	if err != nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	note, err := getNoteByUploadID(h.DB, id)
	if err != nil {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	history, err := GetNoteHistoryByUploadID(h.DB, id)
	if err != nil {
		http.Error(w, "failed to get note history", http.StatusInternalServerError)
		return
	}

	// Delete from DB (cascades to notes and note_history)
	err = deleteUpload(h.DB, id)
	if err != nil {
		http.Error(w, "failed to delete content", http.StatusInternalServerError)
		return
	}

	// Delete from R2 object storage
	_ = storage.DeleteTranscription(r.Context(), upload.StorageKey)
	_ = storage.DeleteTranscription(r.Context(), note.StorageKey)
	for _, hist := range history {
		_ = storage.DeleteTranscription(r.Context(), hist.StorageKey)
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func (h *Handler) GetUploadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uploads, err := getAllUploadIDs(h.DB)
	if err != nil {
		http.Error(w, "failed to upload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(uploads)

	return
}

func (h *Handler) GetNoteByUploadIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	format := r.URL.Query().Get("format")

	note, err := getNoteByUploadID(h.DB, id)
	if err != nil {
		http.Error(w, "failed to receive note", http.StatusInternalServerError)
		return
	}

	switch format {
	case "txt":
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="notes-%s.txt"`, id))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(note.Content))
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="notes-%s.pdf"`, id))
		w.WriteHeader(http.StatusOK)
		if err := generatePDF(w, note.Content); err != nil {
			http.Error(w, "failed to generate PDF", http.StatusInternalServerError)
			return
		}
	case "docx":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="notes-%s.docx"`, id))
		w.WriteHeader(http.StatusOK)
		if err := generateDOCX(w, note.Content); err != nil {
			http.Error(w, "failed to generate DOCX", http.StatusInternalServerError)
			return
		}
	default:
		// Return note with history
		noteWithHistory, err := GetNoteWithHistoryByUploadID(h.DB, id)
		if err != nil {
			http.Error(w, "failed to get note history", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(noteWithHistory)
	}
}

func (h *Handler) RegenerateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")

	// Decode request body
	var reqBody RegenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing note
	note, err := getNoteByUploadID(h.DB, id)
	if err != nil {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	// Regenerate study sheet with existing notes as context
	newContent, err := notes.RegenerateStudySheet(note.Content, reqBody.Prompt)
	if err != nil {
		http.Error(w, "failed to regenerate study sheet", http.StatusInternalServerError)
		return
	}

	// Upload new content to R2
	newStorageKey, err := storage.UploadTranscription(r.Context(), note.ID, newContent, os.Getenv("R2_BUCKET_NAME"))
	if err != nil {
		http.Error(w, "failed to store regenerated notes", http.StatusInternalServerError)
		return
	}

	// Insert into note_history
	historyID := uuid.New().String()
	_, err = InsertNoteHistory(h.DB, historyID, note.ID, id, reqBody.Prompt, newContent, newStorageKey)
	if err != nil {
		http.Error(w, "failed to store note history", http.StatusInternalServerError)
		return
	}

	// Update note in database
	updatedNote, err := UpdateNote(h.DB, note.ID, newContent, newStorageKey)
	if err != nil {
		http.Error(w, "failed to update note", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedNote)
}

func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if !validateUploadRequest(w, r) {
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file provided", http.StatusBadRequest)
		return
	}

	fileType, err := getFileType(file, header)
	if err != nil {
		log.Println("Invalid file type: ", err.Error())
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	var text string
	if isDocument(fileType) {
		text, err = transcription.ExtractText(file, fileType)
	} else {
		text, err = transcription.TranscribeAudio(file, header.Filename)
	}
	if err != nil {
		log.Println("Failed to extract text: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	studySheet, err := notes.GenerateStudySheet(text)
	if err != nil {
		log.Println("Failed to generate studysheet: ", err.Error())
		http.Error(w, "failed to generate study sheet", http.StatusInternalServerError)
		return
	}

	err = uploadToDB(h, w, r, text, header, studySheet)
	if err != nil {
		log.Println("Failed to upload to database: ", err.Error())
		return
	}

	writeSuccessResp(w)
}

func getFileType(file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Seek(0, io.SeekStart)

	// Try document validators first
	if err := validateDocument(file, header); err == nil {
		return header.Header.Get("Content-Type"), nil
	}
	file.Seek(0, io.SeekStart)

	if err := validateVideo(file, header); err == nil {
		return header.Header.Get("Content-Type"), nil
	}
	file.Seek(0, io.SeekStart)

	if err := validateAudio(file, header); err == nil {
		return header.Header.Get("Content-Type"), nil
	}
	file.Seek(0, io.SeekStart)

	return "", fmt.Errorf("unsupported file type")
}

func isDocument(fileType string) bool {
	docTypes := map[string]bool{
		"application/pdf": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"text/plain":                true,
		"text/plain; charset=utf-8": true,
	}
	return docTypes[fileType]
}

func validateUploadRequest(w http.ResponseWriter, r *http.Request) (ret bool) {
	// initialization of return value
	ret = false

	// Check method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// max upload size is set to 100MB
	const maxUploadSize = 100 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse the multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	ret = true

	return
}

func validateDocument(file multipart.File, header *multipart.FileHeader) error {
	allowedTypes := map[string]bool{
		"application/pdf": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"text/plain":                true,
		"text/plain; charset=utf-8": true,
	}
	return checkFileType(file, header, allowedTypes)
}

func validateVideo(file multipart.File, header *multipart.FileHeader) error {
	allowedTypes := map[string]bool{
		"video/mp4": true,
	}
	return checkFileType(file, header, allowedTypes)
}

func validateAudio(file multipart.File, header *multipart.FileHeader) error {
	allowedTypes := map[string]bool{
		"audio/mpeg": true,
	}
	return checkFileType(file, header, allowedTypes)
}

// shared logic used by all three
func checkFileType(file multipart.File, header *multipart.FileHeader, allowedTypes map[string]bool) error {
	declared := header.Header.Get("Content-Type")
	if !allowedTypes[declared] {
		return fmt.Errorf("unsupported file type: %s", declared)
	}

	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		return fmt.Errorf("could not read file")
	}
	defer file.Seek(0, io.SeekStart)

	detected := http.DetectContentType(buf)

	log.Println(detected)

	// normalize zip to docx if the declared type is docx
	if detected == "application/zip" && declared == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		detected = declared
	}

	if !allowedTypes[detected] {
		log.Printf("Error: expected:%s, found: %s\n", declared, detected)
		return fmt.Errorf("file content mismatch: declared %s but detected %s", declared, detected)
	}

	return nil
}

func writeSuccessResp(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "upload successful"})
}

func uploadToDB(h *Handler, w http.ResponseWriter, r *http.Request, text string, header *multipart.FileHeader, studysheet string) error {
	uploadID := uuid.New().String()
	noteID := uuid.New().String()

	fileType := header.Header.Get("Content-Type")

	// upload to r2
	storageKey, err := storage.UploadTranscription(r.Context(), uploadID, text, os.Getenv("R2_BUCKET_NAME"))
	if err != nil {
		http.Error(w, "failed to store transcription", http.StatusInternalServerError)
		return err
	}

	studySheetKey, err := storage.UploadTranscription(r.Context(), noteID, studysheet, os.Getenv("R2_BUCKET_NAME"))
	if err != nil {
		http.Error(w, "failed to store transcription", http.StatusInternalServerError)
		return err
	}

	_, err = insertUpload(h.DB, uploadID, header.Filename, fileType, header.Size, storageKey)
	if err != nil {
		http.Error(w, "failed to save upload", http.StatusInternalServerError)
		return err
	}

	_, err = insertNote(h.DB, noteID, uploadID, studysheet, studySheetKey)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "failed to save upload", http.StatusInternalServerError)
		return err
	}
	log.Println("note inserted into DB:", noteID)

	return nil
}

func generatePDF(w io.Writer, content string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "", 12)
	pdf.SetAutoPageBreak(true, 15)

	// Add title
	pdf.SetFont("Arial", "B", 16)
	pdf.MultiCell(0, 10, "Study Notes", "", "C", false)
	pdf.Ln(5)

	// Add content - split by lines and handle wrapped text
	pdf.SetFont("Arial", "", 11)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			pdf.Ln(4)
			continue
		}
		pdf.MultiCell(0, 5, line, "", "L", false)
	}

	return pdf.Output(w)
}

func generateDOCX(w io.Writer, content string) error {
	// Create a proper DOCX (ZIP) file
	zw := zip.NewWriter(w)
	defer zw.Close()

	// [Content_Types].xml
	ct, _ := zw.Create("[Content_Types].xml")
	ct.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`))

	// _rels/.rels
	rels, _ := zw.Create("_rels/.rels")
	rels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))

	// word/document.xml - the actual content
	doc, _ := zw.Create("word/document.xml")
	doc.Write([]byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="32"/></w:rPr><w:t>Study Notes</w:t></w:r></w:p>
%s
</w:body>
</w:document>`, contentToWordXML(content))))

	return nil
}

func contentToWordXML(content string) string {
	var sb strings.Builder
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			sb.WriteString(`<w:p/>`)
		} else {
			sb.WriteString(fmt.Sprintf(`<w:p><w:r><w:t>%s</w:t></w:r></w:p>`, escapeXML(line)))
		}
	}
	return sb.String()
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
