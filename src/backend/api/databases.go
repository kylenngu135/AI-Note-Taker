package api

import (
	"database/sql"
	"fmt"
	"time"
)

type Upload struct {
	ID            string    `json:"id"`
	Filename      string    `json:"filename"`
	FileType      string    `json:"file_type"`
	FileSize      int64     `json:"file_size"`
	StorageKey    string    `json:"-"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

type Note struct {
	ID            string    `json:"id"`
	UploadID      string    `json:"upload_id"`
	Content       string    `json:"content"`
	StorageKey    string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

type UploadListItem struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
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

func getAllUploadIDs(database *sql.DB) (uploads []UploadListItem, err error) {
	query := `
		SELECT id, filename, created_at 
		FROM uploads
		ORDER BY last_updated_at DESC;
	`

	rows, err := database.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get uploads: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var upload UploadListItem
		err := rows.Scan(
			&upload.ID,
			&upload.Filename,
			&upload.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan upload: %w", err)
		}
		uploads = append(uploads, upload)
	}

	return
}

func deleteUpload(database *sql.DB, id string) (err error) {
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

func getNoteByUploadID(database *sql.DB, uploadID string) (Note, error) {
	query := `SELECT id, upload_id, content, storage_key, created_at, last_updated_at FROM notes WHERE upload_id = $1`

	var note Note
	err := database.QueryRow(query, uploadID).Scan(
		&note.ID,
		&note.UploadID,
		&note.Content,
		&note.StorageKey,
		&note.CreatedAt,
		&note.LastUpdatedAt,
	)
	if err == sql.ErrNoRows {
		return Note{}, fmt.Errorf("note not found")
	}
	if err != nil {
		return Note{}, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}

func insertNote(database *sql.DB, noteID, uploadID, content, storageKey string) (Note, error) {
	query := `
        INSERT INTO notes (id, upload_id, content, storage_key)
        VALUES ($1, $2, $3, $4)
        RETURNING id, upload_id, content, storage_key, created_at, last_updated_at
    `
	var note Note
	err := database.QueryRow(query, noteID, uploadID, content, storageKey).Scan(
		&note.ID,
		&note.UploadID,
		&note.Content,
		&note.StorageKey,
		&note.CreatedAt,
		&note.LastUpdatedAt,
	)
	if err != nil {
		return Note{}, fmt.Errorf("failed to insert note: %w", err)
	}

	return note, nil
}

func CheckUserExists(database *sql.DB, email string) (bool, error) {
	query := ` SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := database.QueryRow(query, email).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to query database: %w", err)
	}

	return exists, nil
}

func InsertUser(database *sql.DB, id, email, passwordHash string) error {
	query := `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`

	_, err := database.Exec(query, id, email, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func GetHashedPasswordByEmail(database *sql.DB, email string) (string, error) {
	query := `SELECT password_hash FROM users WHERE email = ($1)`

	var password string
	err := database.QueryRow(query, email).Scan(&password)

	if err == sql.ErrNoRows {
		return "", err
	}

	if err != nil {
		return "", fmt.Errorf("failed to query database: %w", err)
	}

	return password, nil
}

func GetUserIDByEmail(database *sql.DB, email string) (string, error) {
	query := `SELECT id FROM users WHERE email = ($1)`

	var id string
	err := database.QueryRow(query, email).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to query database: %w", err)
	}

	return id, nil
}
