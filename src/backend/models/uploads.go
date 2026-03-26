package models

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

type UploadListItem struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	CreatedAt time.Time `json:"created_at"`
}
