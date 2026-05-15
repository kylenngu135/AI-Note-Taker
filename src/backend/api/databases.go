package api

import (
	"database/sql"
	"encoding/json"
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

type NoteHistory struct {
	ID         string    `json:"id"`
	NoteID     string    `json:"note_id"`
	UploadID   string    `json:"upload_id"`
	Role       string    `json:"role"`
	Prompt     string    `json:"prompt"`
	Content    string    `json:"content"`
	StorageKey string    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
}

type Tag struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id,omitempty"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

type NoteWithHistory struct {
	Note    Note          `json:"note"`
	History []NoteHistory `json:"history"`
	Tags    []Tag         `json:"tags"`
}

type UploadListItem struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
	Tags      []Tag     `json:"tags"`
}

// insertUploadPending creates an upload row with status "pending" before async processing begins.
func insertUploadPending(database *sql.DB, uploadID, filename, fileType string, fileSize int64, storageKey, userID string) (Upload, error) {
	query := `
        INSERT INTO uploads (id, filename, file_type, file_size, storage_key, status, user_id)
        VALUES ($1, $2, $3, $4, $5, 'pending', $6)
        RETURNING id, filename, file_type, file_size, storage_key, status, created_at
    `
	var upload Upload
	err := database.QueryRow(query, uploadID, filename, fileType, fileSize, storageKey, userID).Scan(
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

func getAllUploadIDsWithTags(database *sql.DB, userID, tagFilter string) ([]UploadListItem, error) {
	baseCols := `
		SELECT
			u.id, u.filename, u.created_at,
			COALESCE(
				json_agg(
					json_build_object('id', t.id, 'name', t.name, 'type', t.type, 'color', COALESCE(t.color, ''))
				) FILTER (WHERE t.id IS NOT NULL),
				'[]'::json
			) AS tags
		FROM uploads u
		LEFT JOIN notes n ON n.upload_id = u.id
		LEFT JOIN note_tags nt ON nt.note_id = n.id
		LEFT JOIN tags t ON t.id = nt.tag_id
	`

	var (
		query string
		args  []interface{}
	)

	if tagFilter != "" {
		query = baseCols + `
		WHERE u.user_id = $1 AND u.id IN (
			SELECT DISTINCT u2.id FROM uploads u2
			JOIN notes n2 ON n2.upload_id = u2.id
			JOIN note_tags nt2 ON nt2.note_id = n2.id
			JOIN tags t2 ON t2.id = nt2.tag_id
			WHERE LOWER(t2.name) = LOWER($2)
		)
		GROUP BY u.id, u.filename, u.created_at, u.last_updated_at
		ORDER BY u.last_updated_at DESC
		`
		args = []interface{}{userID, tagFilter}
	} else {
		query = baseCols + `
		WHERE u.user_id = $1
		GROUP BY u.id, u.filename, u.created_at, u.last_updated_at
		ORDER BY u.last_updated_at DESC
		`
		args = []interface{}{userID}
	}

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get uploads: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var uploads []UploadListItem
	for rows.Next() {
		var upload UploadListItem
		var tagsJSON []byte
		if err := rows.Scan(&upload.ID, &upload.Filename, &upload.CreatedAt, &tagsJSON); err != nil {
			return nil, fmt.Errorf("failed to scan upload: %w", err)
		}
		if len(tagsJSON) > 0 {
			if jsonErr := json.Unmarshal(tagsJSON, &upload.Tags); jsonErr != nil {
				upload.Tags = []Tag{}
			}
		}
		if upload.Tags == nil {
			upload.Tags = []Tag{}
		}
		uploads = append(uploads, upload)
	}

	return uploads, nil
}

func deleteUpload(database *sql.DB, id string) error {
	// Delete in order: note_history → notes (cascades note_tags) → uploads (respecting FKs)
	if err := deleteNoteHistoryByUploadID(database, id); err != nil {
		return fmt.Errorf("failed to delete note history: %w", err)
	}
	if err := deleteNoteByUploadID(database, id); err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}
	query := `DELETE FROM uploads WHERE id = $1`
	_, err := database.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete upload: %w", err)
	}
	return nil
}

func deleteNoteByUploadID(database *sql.DB, uploadID string) error {
	query := `DELETE FROM notes WHERE upload_id = $1`
	_, err := database.Exec(query, uploadID)
	return err
}

func deleteNoteHistoryByUploadID(database *sql.DB, uploadID string) error {
	query := `DELETE FROM note_history WHERE upload_id = $1`
	_, err := database.Exec(query, uploadID)
	return err
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

// GetUploadByIDAndUserID returns the upload only when it exists AND belongs to
// userID. A missing upload and a wrong-owner upload both return an error so
// callers can safely map the result to a 404 without revealing existence.
func GetUploadByIDAndUserID(database *sql.DB, id, userID string) (Upload, error) {
	query := `SELECT id, filename, file_type, file_size, storage_key, status, created_at, last_updated_at
	          FROM uploads WHERE id = $1 AND user_id = $2`
	var upload Upload
	err := database.QueryRow(query, id, userID).Scan(
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

func UpdateNote(database *sql.DB, noteID, content, storageKey string) (Note, error) {
	query := `UPDATE notes SET content = $1, storage_key = $2 WHERE id = $3 RETURNING id, upload_id, content, storage_key, created_at, last_updated_at`

	var note Note
	err := database.QueryRow(query, content, storageKey, noteID).Scan(
		&note.ID,
		&note.UploadID,
		&note.Content,
		&note.StorageKey,
		&note.CreatedAt,
		&note.LastUpdatedAt,
	)
	if err != nil {
		return Note{}, fmt.Errorf("failed to update note: %w", err)
	}

	return note, nil
}

func InsertNoteHistory(database *sql.DB, historyID, noteID, uploadID, role, prompt, content, storageKey string) (NoteHistory, error) {
	query := `
		INSERT INTO note_history (id, note_id, upload_id, role, prompt, content, storage_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, note_id, upload_id, role, prompt, content, storage_key, created_at
	`

	var history NoteHistory
	err := database.QueryRow(query, historyID, noteID, uploadID, role, prompt, content, storageKey).Scan(
		&history.ID,
		&history.NoteID,
		&history.UploadID,
		&history.Role,
		&history.Prompt,
		&history.Content,
		&history.StorageKey,
		&history.CreatedAt,
	)
	if err != nil {
		return NoteHistory{}, fmt.Errorf("failed to insert note history: %w", err)
	}

	return history, nil
}

func GetNoteHistoryByUploadID(database *sql.DB, uploadID string) ([]NoteHistory, error) {
	query := `
		SELECT id, note_id, upload_id, role, prompt, content, storage_key, created_at
		FROM note_history
		WHERE upload_id = $1
		ORDER BY created_at ASC
	`

	rows, err := database.Query(query, uploadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var history []NoteHistory
	for rows.Next() {
		var h NoteHistory
		err := rows.Scan(
			&h.ID,
			&h.NoteID,
			&h.UploadID,
			&h.Role,
			&h.Prompt,
			&h.Content,
			&h.StorageKey,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan note history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

func GetNoteWithHistoryByUploadID(database *sql.DB, uploadID string) (NoteWithHistory, error) {
	note, err := getNoteByUploadID(database, uploadID)
	if err != nil {
		return NoteWithHistory{}, err
	}

	history, err := GetNoteHistoryByUploadID(database, uploadID)
	if err != nil {
		return NoteWithHistory{}, err
	}

	// best-effort: tags are non-critical
	tags, _ := getTagsByNoteID(database, note.ID)
	if tags == nil {
		tags = []Tag{}
	}

	return NoteWithHistory{
		Note:    note,
		History: history,
		Tags:    tags,
	}, nil
}

// --- Tag DB functions ---

func createOrGetTag(database *sql.DB, id, userID, name, tagType, color string) (Tag, error) {
	query := `
		INSERT INTO tags (id, user_id, name, type, color)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, user_id, name, type, COALESCE(color, ''), created_at
	`
	var tag Tag
	err := database.QueryRow(query, id, userID, name, tagType, color).Scan(
		&tag.ID, &tag.UserID, &tag.Name, &tag.Type, &tag.Color, &tag.CreatedAt,
	)
	if err != nil {
		return Tag{}, fmt.Errorf("failed to create/get tag: %w", err)
	}
	return tag, nil
}

func addTagToNote(database *sql.DB, noteID, tagID string) error {
	query := `INSERT INTO note_tags (note_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := database.Exec(query, noteID, tagID)
	return err
}

func removeTagFromNote(database *sql.DB, noteID, tagID string) error {
	query := `DELETE FROM note_tags WHERE note_id = $1 AND tag_id = $2`
	_, err := database.Exec(query, noteID, tagID)
	return err
}

func getTagsByNoteID(database *sql.DB, noteID string) ([]Tag, error) {
	query := `
		SELECT t.id, t.user_id, t.name, t.type, COALESCE(t.color, ''), t.created_at
		FROM tags t
		JOIN note_tags nt ON nt.tag_id = t.id
		WHERE nt.note_id = $1
		ORDER BY t.type, t.name
	`
	rows, err := database.Query(query, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Type, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func getAllTagsForUser(database *sql.DB, userID string) ([]Tag, error) {
	query := `
		SELECT id, user_id, name, type, COALESCE(color, ''), created_at
		FROM tags
		WHERE user_id = $1
		ORDER BY name ASC
	`
	rows, err := database.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Type, &tag.Color, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func deleteTagForUser(database *sql.DB, tagID, userID string) error {
	query := `DELETE FROM tags WHERE id = $1 AND user_id = $2`
	result, err := database.Exec(query, tagID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}

func getTagByID(database *sql.DB, tagID string) (Tag, error) {
	query := `SELECT id, user_id, name, type, COALESCE(color, ''), created_at FROM tags WHERE id = $1`
	var tag Tag
	err := database.QueryRow(query, tagID).Scan(
		&tag.ID, &tag.UserID, &tag.Name, &tag.Type, &tag.Color, &tag.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return Tag{}, fmt.Errorf("tag not found")
	}
	if err != nil {
		return Tag{}, fmt.Errorf("failed to get tag: %w", err)
	}
	return tag, nil
}

// --- User DB functions ---

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
