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

	"AI-Note-Taker/middleware"
	"AI-Note-Taker/notes"
	"AI-Note-Taker/queue"
	"AI-Note-Taker/storage"
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

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	upload, err := GetUploadByIDAndUserID(h.DB, id, userID)
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

	// Delete from DB (cascades to notes, note_tags, and note_history)
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
}

func (h *Handler) GetUploadsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tagFilter := r.URL.Query().Get("tag")
	uploads, err := getAllUploadIDsWithTags(h.DB, userID, tagFilter)
	if err != nil {
		http.Error(w, "failed to upload", http.StatusInternalServerError)
		return
	}

	if uploads == nil {
		uploads = []UploadListItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(uploads)
}

func (h *Handler) GetNoteByUploadIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	if _, err := GetUploadByIDAndUserID(h.DB, id, userID); err != nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	format := r.URL.Query().Get("format")

	noteWithHistory, err := GetNoteWithHistoryByUploadID(h.DB, id)
	if err != nil {
		http.Error(w, "failed to receive note", http.StatusInternalServerError)
		return
	}

	switch format {
	case "txt":
		content := buildExportContent(noteWithHistory.History, noteWithHistory.Note.Content)
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="notes-%s.txt"`, id))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(content))
	case "pdf":
		content := buildExportContent(noteWithHistory.History, noteWithHistory.Note.Content)
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="notes-%s.pdf"`, id))
		w.WriteHeader(http.StatusOK)
		if err := generatePDF(w, content); err != nil {
			http.Error(w, "failed to generate PDF", http.StatusInternalServerError)
			return
		}
	case "docx":
		content := buildExportContent(noteWithHistory.History, noteWithHistory.Note.Content)
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="notes-%s.docx"`, id))
		w.WriteHeader(http.StatusOK)
		if err := generateDOCX(w, content); err != nil {
			http.Error(w, "failed to generate DOCX", http.StatusInternalServerError)
			return
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(noteWithHistory)
	}
}

// buildExportContent formats the full conversation history into a readable export document.
// The first user entry (raw uploaded transcription) is skipped. Each AI response is
// labelled with the prompt that produced it. fallback is used when history is empty.
//
// Handles two history formats:
//   - New format (role-split): user rows hold only the prompt; assistant rows hold only the content.
//   - Old format (single-row): a "user" row may hold both the prompt and the AI response in content
//     (written before the role column was added). Content from such rows is emitted as a response.
func buildExportContent(history []NoteHistory, fallback string) string {
	const divider = "------------------------------------------------------------"

	var sb strings.Builder
	firstAssistant := true

	writeAssistantContent := func(content string) {
		if firstAssistant {
			sb.WriteString("Study Sheet\n")
			sb.WriteString("============================================================\n\n")
			firstAssistant = false
		}
		sb.WriteString(content)
		sb.WriteString("\n")
	}

	for i, entry := range history {
		// The first user entry is the raw uploaded document text — skip it.
		// Exception: old-format rows stored the first follow-up response in content;
		// if content is non-empty the row is a real Q&A pair, not the transcription.
		if i == 0 && entry.Role == "user" && entry.Content == "" {
			continue
		}

		if entry.Role == "user" && entry.Prompt != "" {
			sb.WriteString("\n" + divider + "\n")
			sb.WriteString("Follow-up: " + entry.Prompt + "\n")
			sb.WriteString(divider + "\n\n")
			// Old format: AI response stored in the same row as the user prompt.
			if entry.Content != "" {
				writeAssistantContent(entry.Content)
			}
		} else if entry.Role == "assistant" && entry.Content != "" {
			writeAssistantContent(entry.Content)
		}
	}

	result := strings.TrimSpace(sb.String())
	if result == "" {
		return fallback
	}
	return result
}

func (h *Handler) RegenerateNoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")

	// Decode request body
	var reqBody RegenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Verify ownership before any further work
	if _, err := GetUploadByIDAndUserID(h.DB, id, userID); err != nil {
		http.Error(w, "upload not found", http.StatusNotFound)
		return
	}

	// Get existing note
	note, err := getNoteByUploadID(h.DB, id)
	if err != nil {
		http.Error(w, "note not found", http.StatusNotFound)
		return
	}

	// Get conversation history
	history, err := GetNoteHistoryByUploadID(h.DB, id)
	if err != nil {
		http.Error(w, "failed to get note history", http.StatusInternalServerError)
		return
	}

	// Build conversation history from database
	var conversationHistory []notes.Message
	for _, h := range history {
		// Add user message (prompt)
		if h.Prompt != "" {
			conversationHistory = append(conversationHistory, notes.Message{
				Role:    "user",
				Content: h.Prompt,
			})
		}
		// Add assistant response (content)
		if h.Content != "" {
			conversationHistory = append(conversationHistory, notes.Message{
				Role:    "assistant",
				Content: h.Content,
			})
		}
	}

	// If no history exists, start with the original transcription
	if len(conversationHistory) == 0 {
		conversationHistory = append(conversationHistory, notes.Message{
			Role:    "user",
			Content: note.Content,
		})
	}

	// Regenerate study sheet with full conversation history
	newContent, err := notes.RegenerateStudySheetWithHistory(conversationHistory, reqBody.Prompt)
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

	// Insert into note_history - store user prompt
	historyID := uuid.New().String()
	_, err = InsertNoteHistory(h.DB, historyID, note.ID, id, "user", reqBody.Prompt, "", "")
	if err != nil {
		http.Error(w, "failed to store note history", http.StatusInternalServerError)
		return
	}

	// Insert assistant response into note_history
	historyID = uuid.New().String()
	_, err = InsertNoteHistory(h.DB, historyID, note.ID, id, "assistant", "", newContent, newStorageKey)
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
	_ = json.NewEncoder(w).Encode(updatedNote)
}

func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if !validateUploadRequest(w, r) {
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file provided", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	fileType, err := getFileType(file, header)
	if err != nil {
		log.Println("Invalid file type:", err.Error())
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	// Read the entire file into memory so we can upload it to R2 for async processing.
	rawBytes, err := io.ReadAll(file)
	if err != nil {
		log.Println("Failed to read file:", err.Error())
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	uploadID := uuid.New().String()
	jobID := uuid.New().String()

	// Store the raw file in R2 so the worker can retrieve it.
	rawKey, err := storage.UploadRawFile(r.Context(), uploadID, rawBytes, fileType, os.Getenv("R2_BUCKET_NAME"))
	if err != nil {
		log.Println("Failed to upload raw file:", err.Error())
		http.Error(w, "failed to store file", http.StatusInternalServerError)
		return
	}

	// Create the upload row immediately with status "pending".
	if _, err = insertUploadPending(h.DB, uploadID, header.Filename, fileType, header.Size, rawKey, userID); err != nil {
		log.Println("Failed to insert upload:", err.Error())
		http.Error(w, "failed to save upload", http.StatusInternalServerError)
		return
	}

	// Record the job in Postgres so its status can be tracked.
	if err = queue.InsertJobRecord(h.DB, jobID, rawKey); err != nil {
		log.Println("Failed to insert job:", err.Error())
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	// Push the job onto the correct Redis queue.
	job := queue.Job{
		JobID:    jobID,
		UploadID: uploadID,
		FileType: fileType,
		FileKey:  rawKey,
		Filename: header.Filename,
		FileSize: header.Size,
		UserID:   userID,
	}

	if err = queue.EnqueueJob(job); err != nil {
		log.Println("Failed to enqueue job:", err.Error())
		http.Error(w, "failed to enqueue job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"job_id":    jobID,
		"upload_id": uploadID,
	})
}

func getFileType(file multipart.File, header *multipart.FileHeader) (string, error) {
	defer func() { _, _ = file.Seek(0, io.SeekStart) }()

	// Try document validators first
	if err := validateDocument(file, header); err == nil {
		return header.Header.Get("Content-Type"), nil
	}
	_, _ = file.Seek(0, io.SeekStart)

	if err := validateVideo(file, header); err == nil {
		return header.Header.Get("Content-Type"), nil
	}
	_, _ = file.Seek(0, io.SeekStart)

	if err := validateAudio(file, header); err == nil {
		return header.Header.Get("Content-Type"), nil
	}
	_, _ = file.Seek(0, io.SeekStart)

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
	defer func() { _, _ = file.Seek(0, io.SeekStart) }()

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
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "upload successful"})
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
	defer func() { _ = zw.Close() }()

	// [Content_Types].xml
	ct, _ := zw.Create("[Content_Types].xml")
	_, _ = ct.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`))

	// _rels/.rels
	rels, _ := zw.Create("_rels/.rels")
	_, _ = rels.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`))

	// word/document.xml - the actual content
	doc, _ := zw.Create("word/document.xml")
	_, _ = fmt.Fprintf(doc, `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body>
<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="32"/></w:rPr><w:t>Study Notes</w:t></w:r></w:p>
%s
</w:body>
</w:document>`, contentToWordXML(content))

	return nil
}

func contentToWordXML(content string) string {
	var sb strings.Builder
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			sb.WriteString(`<w:p/>`)
		} else {
			_, _ = fmt.Fprintf(&sb, `<w:p><w:r><w:t>%s</w:t></w:r></w:p>`, escapeXML(line))
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
