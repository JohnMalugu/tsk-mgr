package repository

import (
	"database/sql"

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
	return 0, nil
}

func (r *sqlAttachmentRepository) GetAttachmentByID(id int64) (*model.AttachmentDB, error) {
	return nil, nil
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
