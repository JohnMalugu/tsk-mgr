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

// AttachmentFilter represents criteria to filter and paginate attachments
type AttachmentFilter struct {
	TaskID      *int64  `json:"task_id,omitempty"`
	WorkspaceID *int64  `json:"workspace_id,omitempty"`
	UserID      *string `json:"user_id,omitempty"`
	ContentType *string `json:"content_type,omitempty"`
	IsArchived  *bool   `json:"is_archived,omitempty"`
	Offset      int     `json:"offset"`
	Limit       int     `json:"limit"`
}

// AttachmentSummary represents a paginated list of attachments with total metadata
type AttachmentSummary struct {
	Attachments []AttachmentDB `json:"attachments"`
	Total       int            `json:"total"`
	TotalBytes  int64          `json:"total_bytes"`
	Offset      int            `json:"offset"`
	Limit       int            `json:"limit"`
}

// StorageMetrics summarizes workspace file asset storage consumption and MIME distribution
type StorageMetrics struct {
	WorkspaceID    int64            `json:"workspace_id"`
	TotalFiles     int              `json:"total_files"`
	TotalBytes     int64            `json:"total_bytes"`
	TotalDownloads int64            `json:"total_downloads"`
	MimeBreakdown  map[string]int64 `json:"mime_breakdown"`
	LargestFiles   []AttachmentDB   `json:"largest_files"`
}
