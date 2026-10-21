package service

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func (s *attachmentService) saveFile(workspaceID, taskID int64, filename, checksum string, data []byte) (string, error) {
	dir := filepath.Join(s.storageDir, fmt.Sprintf("ws_%d", workspaceID), fmt.Sprintf("task_%d", taskID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	shortChecksum := checksum
	if len(shortChecksum) > 12 {
		shortChecksum = shortChecksum[:12]
	}
	storedName := fmt.Sprintf("%s_%s", shortChecksum, filename)
	fullPath := filepath.Join(dir, storedName)

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write attachment file to disk: %w", err)
	}

	return fullPath, nil
}

func (s *attachmentService) readFile(storagePath string) (io.ReadCloser, error) {
	file, err := os.Open(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open attachment file: %w", err)
	}
	return file, nil
}

func (s *attachmentService) removeFile(storagePath string) error {
	if storagePath == "" {
		return nil
	}
	if err := os.Remove(storagePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file on disk: %w", err)
	}
	return nil
}

func (s *attachmentService) UploadAttachment(req *model.UploadAttachmentRequest) (*model.AttachmentDB, error) {
	data, err := ValidateUpload(req)
	if err != nil {
		return nil, err
	}

	checksum := CalculateSHA256(data)
	sanitized := SanitizeFilename(req.Filename)

	var storagePath string
	existing, err := s.repo.GetAttachmentByChecksum(req.WorkspaceID, checksum)
	if err == nil && existing != nil && existing.StoragePath != "" {
		// Deduplication hit: reuse existing physical file
		storagePath = existing.StoragePath
	} else {
		savedPath, err := s.saveFile(req.WorkspaceID, req.TaskID, sanitized, checksum, data)
		if err != nil {
			return nil, err
		}
		storagePath = savedPath
	}

	now := time.Now()
	att := &model.AttachmentDB{
		TaskID:         req.TaskID,
		UserID:         req.UserID,
		WorkspaceID:    req.WorkspaceID,
		Filename:       sanitized,
		OriginalName:   req.Filename,
		FileSize:       int64(len(data)),
		ContentType:    req.ContentType,
		StoragePath:    storagePath,
		ChecksumSHA256: checksum,
		DownloadCount:  0,
		IsArchived:     false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	id, err := s.repo.CreateAttachment(att)
	if err != nil {
		if existing == nil {
			_ = s.removeFile(storagePath)
		}
		return nil, fmt.Errorf("failed to save attachment metadata: %w", err)
	}

	att.ID = id
	return att, nil
}

func (s *attachmentService) GetAttachment(id int64) (*model.AttachmentDB, error) {
	if id <= 0 {
		return nil, ErrAttachmentNotFound
	}

	att, err := s.repo.GetAttachmentByID(id)
	if err != nil {
		return nil, err
	}
	if att == nil {
		return nil, ErrAttachmentNotFound
	}

	return att, nil
}

func (s *attachmentService) ListAttachments(filter model.AttachmentFilter) (*model.AttachmentSummary, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	atts, total, err := s.repo.ListAttachments(filter)
	if err != nil {
		return nil, err
	}

	var totalBytes int64
	for _, a := range atts {
		totalBytes += a.FileSize
	}

	return &model.AttachmentSummary{
		Attachments: atts,
		Total:       total,
		TotalBytes:  totalBytes,
		Offset:      filter.Offset,
		Limit:       filter.Limit,
	}, nil
}

func (s *attachmentService) OpenAttachmentFile(id int64) (io.ReadCloser, *model.AttachmentDB, string, error) {
	att, err := s.GetAttachment(id)
	if err != nil {
		return nil, nil, "", err
	}
	if att.IsArchived {
		return nil, nil, "", errors.New("cannot download archived attachment")
	}

	reader, err := s.readFile(att.StoragePath)
	if err != nil {
		return nil, nil, "", fmt.Errorf("file not found on disk: %w", err)
	}

	etag := fmt.Sprintf("\"%s\"", att.ChecksumSHA256)

	// Increment download counter
	_ = s.repo.IncrementDownloadCount(id)
	att.DownloadCount++

	return reader, att, etag, nil
}

func (s *attachmentService) UpdateAttachment(id int64, req *model.UpdateAttachmentRequest) (*model.AttachmentDB, error) {
	att, err := s.GetAttachment(id)
	if err != nil {
		return nil, err
	}

	if req.Filename != nil {
		sanitized := SanitizeFilename(*req.Filename)
		if model.IsBlockedExtension(sanitized) {
			return nil, ErrBlockedFileType
		}
		att.Filename = sanitized
	}

	if req.IsArchived != nil {
		att.IsArchived = *req.IsArchived
	}

	if err := s.repo.UpdateAttachment(att); err != nil {
		return nil, err
	}

	return att, nil
}

func (s *attachmentService) DeleteAttachment(id int64) error {
	att, err := s.GetAttachment(id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteAttachment(id); err != nil {
		return err
	}

	// If no other record in the workspace uses this file path, clean up physical disk storage
	other, _ := s.repo.GetAttachmentByChecksum(att.WorkspaceID, att.ChecksumSHA256)
	if other == nil {
		_ = s.removeFile(att.StoragePath)
	}

	return nil
}

func (s *attachmentService) GetStorageMetrics(workspaceID int64) (*model.StorageMetrics, error) {
	if workspaceID <= 0 {
		return nil, ErrInvalidWorkspaceID
	}
	return s.repo.GetStorageMetrics(workspaceID)
}
