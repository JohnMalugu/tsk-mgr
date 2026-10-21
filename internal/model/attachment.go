package model

import (
	"time"
)

// AttachmentDB represents a file asset attached to a task
type AttachmentDB struct {
	ID             int64     `json:"id"`
	TaskID         int64     `json:"task_id"`
	UserID         string    `json:"user_id"`
	WorkspaceID    int64     `json:"workspace_id"`
	Filename       string    `json:"filename"`
	OriginalName   string    `json:"original_name"`
	FileSize       int64     `json:"file_size"`
	ContentType    string    `json:"content_type"`
	StoragePath    string    `json:"storage_path"`
	ChecksumSHA256 string    `json:"checksum_sha256"`
	DownloadCount  int64     `json:"download_count"`
	IsArchived     bool      `json:"is_archived"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
