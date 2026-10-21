package service

import (
	"errors"
	"io"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrFileTooLarge       = errors.New("file exceeds maximum allowed size of 50MB")
	ErrBlockedFileType    = errors.New("upload of executable or dangerous file types is blocked")
	ErrEmptyFile          = errors.New("cannot upload empty file")
	ErrInvalidTaskID      = errors.New("task ID must be positive")
	ErrInvalidWorkspaceID = errors.New("workspace ID must be positive")
	ErrInvalidUserID      = errors.New("user ID is required")
	ErrInvalidFilename    = errors.New("filename cannot be empty")
)

// AttachmentService defines business operations for managing task attachments and files
type AttachmentService interface {
	UploadAttachment(req *model.UploadAttachmentRequest) (*model.AttachmentDB, error)
	GetAttachment(id int64) (*model.AttachmentDB, error)
	ListAttachments(filter model.AttachmentFilter) (*model.AttachmentSummary, error)
	OpenAttachmentFile(id int64) (io.ReadCloser, *model.AttachmentDB, string, error)
	UpdateAttachment(id int64, req *model.UpdateAttachmentRequest) (*model.AttachmentDB, error)
	DeleteAttachment(id int64) error
	GetStorageMetrics(workspaceID int64) (*model.StorageMetrics, error)
}

type attachmentService struct {
	repo       repository.AttachmentRepository
	storageDir string
}

// NewAttachmentService creates a new AttachmentService instance
func NewAttachmentService(repo repository.AttachmentRepository, storageDir string) AttachmentService {
	if storageDir == "" {
		storageDir = "data/uploads"
	}
	return &attachmentService{
		repo:       repo,
		storageDir: storageDir,
	}
}

func (s *attachmentService) UploadAttachment(req *model.UploadAttachmentRequest) (*model.AttachmentDB, error) {
	return nil, nil
}

func (s *attachmentService) GetAttachment(id int64) (*model.AttachmentDB, error) {
	return nil, nil
}

func (s *attachmentService) ListAttachments(filter model.AttachmentFilter) (*model.AttachmentSummary, error) {
	return nil, nil
}

func (s *attachmentService) OpenAttachmentFile(id int64) (io.ReadCloser, *model.AttachmentDB, string, error) {
	return nil, nil, "", nil
}

func (s *attachmentService) UpdateAttachment(id int64, req *model.UpdateAttachmentRequest) (*model.AttachmentDB, error) {
	return nil, nil
}

func (s *attachmentService) DeleteAttachment(id int64) error {
	return nil
}

func (s *attachmentService) GetStorageMetrics(workspaceID int64) (*model.StorageMetrics, error) {
	return nil, nil
}
