package queue

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
)

const (
	QueueDocuments = "queue:documents"
	QueueAudio     = "queue:audio"
)

// Job is the message pushed onto a processing queue.
type Job struct {
	JobID    string `json:"job_id"`
	UploadID string `json:"upload_id"`
	FileType string `json:"file_type"`
	FileKey  string `json:"file_key"`
	Filename string `json:"filename"`
	FileSize int64  `json:"file_size"`
	UserID   string `json:"user_id"`
}

// EnqueueJob serializes job as JSON and pushes it to the appropriate queue.
// Documents (pdf, docx, txt) go to QueueDocuments; audio/video go to QueueAudio.
func EnqueueJob(job Job) error {
	client := newRedisClient(os.Getenv("REDIS_URL"), os.Getenv("REDIS_TOKEN"))

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	queueName := QueueAudio
	if isDocumentMIME(job.FileType) {
		queueName = QueueDocuments
	}

	if err := client.lpush(queueName, string(payload)); err != nil {
		return fmt.Errorf("lpush to %s: %w", queueName, err)
	}
	return nil
}

// InsertJobRecord creates a jobs row with status "pending".
func InsertJobRecord(db *sql.DB, jobID, fileKey string) error {
	_, err := db.Exec(
		`INSERT INTO jobs (id, status, file_key) VALUES ($1, 'pending', $2)`,
		jobID, fileKey,
	)
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	return nil
}

func isDocumentMIME(mime string) bool {
	switch mime {
	case "application/pdf",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
		"text/plain; charset=utf-8":
		return true
	}
	return false
}
