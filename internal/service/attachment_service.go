package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

// SanitizeFilename cleans the filename to prevent directory traversal and remove problematic characters
func SanitizeFilename(name string) string {
	clean := filepath.Base(filepath.Clean(name))
	clean = strings.ReplaceAll(clean, "..", "")
	clean = strings.ReplaceAll(clean, "/", "")
	clean = strings.ReplaceAll(clean, "\\", "")

	var b strings.Builder
	for _, r := range clean {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) {
			b.WriteRune('_')
		}
	}
	res := strings.Trim(b.String(), "._-")
	if res == "" {
		return "unnamed_file"
	}
	return res
}

// CalculateSHA256 computes the hex-encoded SHA-256 checksum of the given byte payload
func CalculateSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// ValidateUpload checks common constraints for file uploads
func ValidateUpload(req *model.UploadAttachmentRequest) ([]byte, error) {
	if req.TaskID <= 0 {
		return nil, ErrInvalidTaskID
	}
	if req.WorkspaceID <= 0 {
		return nil, ErrInvalidWorkspaceID
	}
	if strings.TrimSpace(req.UserID) == "" {
		return nil, ErrInvalidUserID
	}
	if strings.TrimSpace(req.Filename) == "" {
		return nil, ErrInvalidFilename
	}

	if model.IsBlockedExtension(req.Filename) {
		return nil, ErrBlockedFileType
	}

	var data []byte
	if len(req.RawBytes) > 0 {
		data = req.RawBytes
	} else if req.ContentBase64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(req.ContentBase64)
		if err != nil {
			return nil, errors.New("invalid base64 content")
		}
		data = decoded
	}

	if len(data) == 0 {
		return nil, ErrEmptyFile
	}
	if len(data) > model.MaxAttachmentSizeBytes {
		return nil, ErrFileTooLarge
	}

	if req.ContentType == "" || req.ContentType == "application/octet-stream" {
		req.ContentType = http.DetectContentType(data)
	}

	return data, nil
}

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
