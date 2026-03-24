package api

import (
    "database/sql"
	"encoding/json"
	"fmt"
	"os"
	"io"
	"mime/multipart"
	"net/http"
	"log"

	"github.com/google/uuid"

	"AI-Note-Taker/transcription"
	"AI-Note-Taker/storage"
	"AI-Note-Taker/notes"
)

type Handler struct {
	DB *sql.DB
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

    // get note for study sheet storage key
    note, err := getNoteByUploadID(h.DB, id)
    if err != nil {
        http.Error(w, "note not found", http.StatusNotFound)
        return
    }

	// Delete from r2 object storage
	if err = storage.DeleteTranscription(r.Context(), upload.StorageKey); err != nil {
        http.Error(w, "failed to delete from storage", http.StatusInternalServerError)
        return
    }

	if err = storage.DeleteTranscription(r.Context(), note.StorageKey); err != nil {
        http.Error(w, "failed to delete from storage", http.StatusInternalServerError)
        return
    }

	err = deleteUpload(h.DB, id);
	if err != nil {
        http.Error(w, "failed to delete content", http.StatusInternalServerError)
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

	note, err := getNoteByUploadID(h.DB, id)
	if err != nil {
        http.Error(w, "failed to revieve note", http.StatusInternalServerError)
		return
	}

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(note)
	return
}


func (h *Handler) DocumentUploadHandler(w http.ResponseWriter, r *http.Request) {
	if !validateUploadRequest(w, r) {
		return
	}

	// get file contents
    file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file provided", http.StatusBadRequest) 
	}

	if err := validateDocument(file, header); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	fileType := header.Header.Get("Content-Type")

	text, err := transcription.ExtractText(file, fileType)
	if err != nil {
		http.Error(w, "failed to extract text", http.StatusInternalServerError)
		return
	}

	studySheet, err := notes.GenerateStudySheet(text)
	if err != nil {
		http.Error(w, "failed to generate study sheet", http.StatusInternalServerError)
		return
	}

	err = uploadToDB(h, w, r, text, header, studySheet)
	if err != nil {
		return
	}

	writeSuccessResp(w)
}

func (h *Handler) VideoUploadHandler(w http.ResponseWriter, r *http.Request) {
	if !validateUploadRequest(w, r) {
		return
	}

	// get file contents
    file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file provided", http.StatusBadRequest) 
	}

	if err := validateVideo(file, header); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType) // 415
		return
	}

	text, err := transcription.TranscribeAudio(file, header.Filename)
	if err != nil {
		http.Error(w, "failed to transcribe video", http.StatusInternalServerError)
		return
	}

	fmt.Println(text)

	/*
	err = uploadToDB(h, w, r, text, header)
	if err != nil {
		return
	}
	*/

	writeSuccessResp(w)
}

func (h *Handler) AudioUploadHandler(w http.ResponseWriter, r *http.Request) {
	if !validateUploadRequest(w, r) {
		return
	}

	// get file contents
    file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "no file provided", http.StatusBadRequest) 
	}

	if err := validateAudio(file, header); err != nil {
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType) // 415
		return
	}

	text, err := transcription.TranscribeAudio(file, header.Filename)

	fmt.Println(text)
	
	/*
	err = uploadToDB(h, w, r, text, header)
	if err != nil {
		return
	}
	*/


	writeSuccessResp(w)
}

func validateUploadRequest(w http.ResponseWriter, r *http.Request) (ret bool) {
	// initialization of return value 
	ret = false

    // Check method is POST
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }

	// max upload size is set to 50MB
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
        "text/plain": true,
        "text/plain; charset=utf-8": true,
    }
    return checkFileType(file, header, allowedTypes)
}

func validateVideo(file multipart.File, header *multipart.FileHeader) error {
    allowedTypes := map[string]bool{
        "video/mp4":       true,
    }
    return checkFileType(file, header, allowedTypes)
}

func validateAudio(file multipart.File, header *multipart.FileHeader) error {
    allowedTypes := map[string]bool{
        "audio/mpeg":  true,
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

    // normalize zip to docx if the declared type is docx
    if detected == "application/zip" && declared == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
        detected = declared
    }

    if !allowedTypes[detected] {
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

	_, err = insertNote(h.DB, noteID, uploadID, studysheet, studySheetKey);
    if err != nil {
		fmt.Println(err)
        http.Error(w, "failed to save upload", http.StatusInternalServerError)
        return err
    }
	log.Println("note inserted into DB:", noteID)

	return nil
}

