package queue

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"

	"AI-Note-Taker/notes"
	"AI-Note-Taker/storage"
	"AI-Note-Taker/transcription"
)

const brpopTimeoutSecs = 5

// bytesFile wraps *bytes.Reader so it satisfies mime/multipart.File (Reader + ReaderAt + Seeker + Closer).
type bytesFile struct{ *bytes.Reader }

func (bytesFile) Close() error { return nil }

// StartWorkers launches one goroutine per queue. Call once at startup.
func StartWorkers(db *sql.DB) {
	go runWorker(db, QueueDocuments, "document")
	go runWorker(db, QueueAudio, "audio")
}

func runWorker(db *sql.DB, queueName, label string) {
	client := newRedisClient(os.Getenv("REDIS_URL"), os.Getenv("REDIS_TOKEN"))
	log.Printf("[worker:%s] started, listening on %s", label, queueName)

	for {
		_, payload, err := client.brpop(brpopTimeoutSecs, queueName)
		if err != nil {
			log.Printf("[worker:%s] brpop error: %v", label, err)
			continue
		}
		if payload == "" {
			log.Printf("[worker:%s] idle", label)
			continue
		}

		var job Job
		if err := json.Unmarshal([]byte(payload), &job); err != nil {
			log.Printf("[worker:%s] malformed job payload: %v", label, err)
			continue
		}

		log.Printf("[worker:%s] processing job %s (%s)", label, job.JobID, job.Filename)
		if err := processJob(db, job); err != nil {
			log.Printf("[worker:%s] job %s failed: %v", label, job.JobID, err)
			_ = setJobFailed(db, job.JobID, err.Error())
		}
	}
}

func processJob(db *sql.DB, job Job) error {
	ctx := context.Background()

	if err := setJobStatus(db, job.JobID, "processing"); err != nil {
		return fmt.Errorf("set processing: %w", err)
	}

	// Download the raw uploaded file from R2
	rawData, err := storage.DownloadFile(ctx, job.FileKey, os.Getenv("R2_BUCKET_NAME"))
	if err != nil {
		return fmt.Errorf("download raw file: %w", err)
	}

	f := bytesFile{bytes.NewReader(rawData)}

	// Transcribe or extract text depending on file type
	var text string
	if isDocumentMIME(job.FileType) {
		text, err = transcription.ExtractText(f, job.FileType)
	} else {
		text, err = transcription.TranscribeAudio(f, job.Filename)
	}
	if err != nil {
		return fmt.Errorf("transcription: %w", err)
	}

	// Upload transcription text to R2
	transcriptionKey, err := storage.UploadTranscription(ctx, job.UploadID, text, os.Getenv("R2_BUCKET_NAME"))
	if err != nil {
		return fmt.Errorf("upload transcription: %w", err)
	}

	// Update upload row: transcription key is now known, mark complete
	if err := workerUpdateUpload(db, job.UploadID, transcriptionKey); err != nil {
		return fmt.Errorf("update upload: %w", err)
	}

	// Generate study sheet via OpenAI
	result, err := notes.GenerateStudySheet(text)
	if err != nil {
		return fmt.Errorf("generate study sheet: %w", err)
	}

	noteID := uuid.New().String()

	// Upload study sheet to R2
	noteKey, err := storage.UploadTranscription(ctx, noteID, result.Content, os.Getenv("R2_BUCKET_NAME"))
	if err != nil {
		return fmt.Errorf("upload study sheet: %w", err)
	}

	if err := workerInsertNote(db, noteID, job.UploadID, result.Content, noteKey); err != nil {
		return fmt.Errorf("insert note: %w", err)
	}

	// Store initial conversation: user transcription + AI study sheet
	userHistID := uuid.New().String()
	if err := workerInsertNoteHistory(db, userHistID, noteID, job.UploadID, "user", text, "", transcriptionKey); err != nil {
		return fmt.Errorf("insert user history: %w", err)
	}

	assistantHistID := uuid.New().String()
	if err := workerInsertNoteHistory(db, assistantHistID, noteID, job.UploadID, "assistant", "", result.Content, noteKey); err != nil {
		return fmt.Errorf("insert assistant history: %w", err)
	}

	// Best-effort tag creation
	if job.UserID != "" {
		for _, tagName := range result.Tags {
			tagID := uuid.New().String()
			tID, createErr := workerCreateOrGetTag(db, tagID, job.UserID, tagName, "auto", "#6b7fa3")
			if createErr != nil {
				log.Printf("[worker] create tag %q: %v", tagName, createErr)
				continue
			}
			if addErr := workerAddTagToNote(db, noteID, tID); addErr != nil {
				log.Printf("[worker] add tag %q to note: %v", tagName, addErr)
			}
		}

		ftTag, ftColor := filetypeTagInfo(job.FileType)
		if ftTag != "" {
			tagID := uuid.New().String()
			tID, createErr := workerCreateOrGetTag(db, tagID, job.UserID, ftTag, "filetype", ftColor)
			if createErr != nil {
				log.Printf("[worker] create filetype tag %q: %v", ftTag, createErr)
			} else if addErr := workerAddTagToNote(db, noteID, tID); addErr != nil {
				log.Printf("[worker] add filetype tag %q to note: %v", ftTag, addErr)
			}
		}
	}

	// Remove the temporary raw file from R2 now that transcription is done
	_ = storage.DeleteTranscription(ctx, job.FileKey)

	if err := setJobCompleted(db, job.JobID, noteID); err != nil {
		log.Printf("[worker] set job completed: %v", err)
	}

	log.Printf("[worker] job %s completed, note %s", job.JobID, noteID)
	return nil
}

// --- job status helpers ---

func setJobStatus(db *sql.DB, jobID, status string) error {
	_, err := db.Exec(
		`UPDATE jobs SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, jobID,
	)
	return err
}

func setJobFailed(db *sql.DB, jobID, errMsg string) error {
	_, err := db.Exec(
		`UPDATE jobs SET status = 'failed', error = $1, updated_at = NOW() WHERE id = $2`,
		errMsg, jobID,
	)
	return err
}

func setJobCompleted(db *sql.DB, jobID, result string) error {
	_, err := db.Exec(
		`UPDATE jobs SET status = 'completed', result = $1, updated_at = NOW() WHERE id = $2`,
		result, jobID,
	)
	return err
}

// --- upload / note / history / tag DB helpers ---

func workerUpdateUpload(db *sql.DB, uploadID, storageKey string) error {
	_, err := db.Exec(
		`UPDATE uploads SET storage_key = $1, status = 'complete', last_updated_at = NOW() WHERE id = $2`,
		storageKey, uploadID,
	)
	return err
}

func workerInsertNote(db *sql.DB, noteID, uploadID, content, storageKey string) error {
	_, err := db.Exec(
		`INSERT INTO notes (id, upload_id, content, storage_key) VALUES ($1, $2, $3, $4)`,
		noteID, uploadID, content, storageKey,
	)
	return err
}

func workerInsertNoteHistory(db *sql.DB, historyID, noteID, uploadID, role, prompt, content, storageKey string) error {
	_, err := db.Exec(
		`INSERT INTO note_history (id, note_id, upload_id, role, prompt, content, storage_key) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		historyID, noteID, uploadID, role, prompt, content, storageKey,
	)
	return err
}

func workerCreateOrGetTag(db *sql.DB, id, userID, name, tagType, color string) (string, error) {
	var tagID string
	err := db.QueryRow(
		`INSERT INTO tags (id, user_id, name, type, color) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
		 RETURNING id`,
		id, userID, name, tagType, color,
	).Scan(&tagID)
	if err != nil {
		return "", err
	}
	return tagID, nil
}

func workerAddTagToNote(db *sql.DB, noteID, tagID string) error {
	_, err := db.Exec(
		`INSERT INTO note_tags (note_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		noteID, tagID,
	)
	return err
}

func filetypeTagInfo(fileType string) (string, string) {
	switch fileType {
	case "application/pdf":
		return "pdf", "#ef4444"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "docx", "#3b82f6"
	case "text/plain", "text/plain; charset=utf-8":
		return "txt", "#22c55e"
	case "video/mp4":
		return "video", "#f97316"
	case "audio/mpeg":
		return "audio", "#a855f7"
	}
	return "", ""
}
