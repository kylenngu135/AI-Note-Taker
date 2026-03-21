package api 

import (
	"database/sql"
	"fmt"
	"time"
)

type Upload struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	FileType   string    `json:"file_type"`
	FileSize   int64     `json:"file_size"`
	StorageKey string    `json:"-"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

func insertUpload(database *sql.DB, uploadID, filename, fileType string, fileSize int64, storageKey string) (Upload, error) {
	query := `
        INSERT INTO uploads (id, filename, file_type, file_size, storage_key, status)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, filename, file_type, file_size, storage_key, status, created_at
    `

	var upload Upload
	err := database.QueryRow(query, uploadID, filename, fileType, fileSize, storageKey, "complete").Scan(
		&upload.ID,
		&upload.Filename,
		&upload.FileType,
		&upload.FileSize,
		&upload.StorageKey,
		&upload.Status,
		&upload.CreatedAt,
	)
	if err != nil {
		return Upload{}, fmt.Errorf("failed to insert upload: %w", err)
	}

	return upload, nil
}

func getAllUploads(database *sql.DB) (uploads []Upload, err error) {
	query := `SELECT * FROM uploads;`

	rows, err := database.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get uploads: %w", err)
	}
	defer rows.Close()

    for rows.Next() {
        var upload Upload
        err := rows.Scan(
            &upload.ID,
            &upload.Filename,
            &upload.FileType,
            &upload.FileSize,
            &upload.StorageKey,
            &upload.Status,
            &upload.CreatedAt,
            &upload.LastUpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan upload: %w", err)
        }
        uploads = append(uploads, upload)
    }

	return
}

func DeleteUpload(database *sql.DB, id string) (err error) {
	query := `DELETE FROM uploads WHERE id = $1`

	_, err = database.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete upload: %w", err)
	}

	return nil
}

func GetUploadByID(database *sql.DB, id string) (Upload, error) {
    query := `SELECT id, filename, file_type, file_size, storage_key, status, created_at, last_updated_at FROM uploads WHERE id = $1`

    var upload Upload
    err := database.QueryRow(query, id).Scan(
        &upload.ID,
        &upload.Filename,
        &upload.FileType,
        &upload.FileSize,
        &upload.StorageKey,
        &upload.Status,
        &upload.CreatedAt,
        &upload.LastUpdatedAt,
    )
    if err == sql.ErrNoRows {
        return Upload{}, fmt.Errorf("upload not found")
    }
    if err != nil {
        return Upload{}, fmt.Errorf("failed to get upload: %w", err)
    }

    return upload, nil
}
