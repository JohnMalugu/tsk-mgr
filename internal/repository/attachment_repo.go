package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

// AttachmentRepository defines persistence operations for task file attachments and storage metrics
type AttachmentRepository interface {
	CreateAttachment(att *model.AttachmentDB) (int64, error)
	GetAttachmentByID(id int64) (*model.AttachmentDB, error)
	GetAttachmentByChecksum(workspaceID int64, checksum string) (*model.AttachmentDB, error)
	ListAttachments(filter model.AttachmentFilter) ([]model.AttachmentDB, int, error)
	UpdateAttachment(att *model.AttachmentDB) error
	DeleteAttachment(id int64) error
	IncrementDownloadCount(id int64) error
	GetStorageMetrics(workspaceID int64) (*model.StorageMetrics, error)
}

type sqlAttachmentRepository struct {
	db *sql.DB
}

// NewAttachmentRepository creates a new AttachmentRepository instance
func NewAttachmentRepository(db *sql.DB) AttachmentRepository {
	return &sqlAttachmentRepository{db: db}
}

func (r *sqlAttachmentRepository) CreateAttachment(att *model.AttachmentDB) (int64, error) {
	now := time.Now()
	if att.CreatedAt.IsZero() {
		att.CreatedAt = now
	}
	if att.UpdatedAt.IsZero() {
		att.UpdatedAt = now
	}

	query := `
	INSERT INTO attachments (
		task_id, user_id, workspace_id, filename, original_name, file_size, content_type,
		storage_path, checksum_sha256, download_count, is_archived, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(
		query,
		att.TaskID,
		att.UserID,
		att.WorkspaceID,
		att.Filename,
		att.OriginalName,
		att.FileSize,
		att.ContentType,
		att.StoragePath,
		att.ChecksumSHA256,
		att.DownloadCount,
		att.IsArchived,
		att.CreatedAt,
		att.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create attachment: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}
	att.ID = id
	return id, nil
}

func (r *sqlAttachmentRepository) GetAttachmentByID(id int64) (*model.AttachmentDB, error) {
	query := `
	SELECT id, task_id, user_id, workspace_id, filename, original_name, file_size, content_type,
		storage_path, checksum_sha256, download_count, is_archived, created_at, updated_at
	FROM attachments
	WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var att model.AttachmentDB
	err := row.Scan(
		&att.ID,
		&att.TaskID,
		&att.UserID,
		&att.WorkspaceID,
		&att.Filename,
		&att.OriginalName,
		&att.FileSize,
		&att.ContentType,
		&att.StoragePath,
		&att.ChecksumSHA256,
		&att.DownloadCount,
		&att.IsArchived,
		&att.CreatedAt,
		&att.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get attachment by id: %w", err)
	}
	return &att, nil
}

func (r *sqlAttachmentRepository) GetAttachmentByChecksum(workspaceID int64, checksum string) (*model.AttachmentDB, error) {
	return nil, nil
}

func (r *sqlAttachmentRepository) ListAttachments(filter model.AttachmentFilter) ([]model.AttachmentDB, int, error) {
	return nil, 0, nil
}

func (r *sqlAttachmentRepository) UpdateAttachment(att *model.AttachmentDB) error {
	return nil
}

func (r *sqlAttachmentRepository) DeleteAttachment(id int64) error {
	return nil
}

func (r *sqlAttachmentRepository) IncrementDownloadCount(id int64) error {
	return nil
}

func (r *sqlAttachmentRepository) GetStorageMetrics(workspaceID int64) (*model.StorageMetrics, error) {
	return nil, nil
}
