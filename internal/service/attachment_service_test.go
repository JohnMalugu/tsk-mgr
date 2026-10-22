package service

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestAttachmentSvc(t *testing.T) (AttachmentService, *sql.DB, string) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}

	tempDir := t.TempDir()
	repo := repository.NewAttachmentRepository(sqliteDB)
	svc := NewAttachmentService(repo, tempDir)
	return svc, sqliteDB, tempDir
}

func TestAttachmentService_Validation(t *testing.T) {
	svc, _, _ := setupTestAttachmentSvc(t)

	// 1. Task ID invalid
	_, err := svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      0,
		WorkspaceID: 1,
		UserID:      "user-1",
		Filename:    "notes.txt",
		RawBytes:    []byte("hello"),
	})
	if !errors.Is(err, ErrInvalidTaskID) {
		t.Fatalf("expected ErrInvalidTaskID, got: %v", err)
	}

	// 2. Workspace ID invalid
	_, err = svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      1,
		WorkspaceID: 0,
		UserID:      "user-1",
		Filename:    "notes.txt",
		RawBytes:    []byte("hello"),
	})
	if !errors.Is(err, ErrInvalidWorkspaceID) {
		t.Fatalf("expected ErrInvalidWorkspaceID, got: %v", err)
	}

	// 3. User ID invalid
	_, err = svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      1,
		WorkspaceID: 1,
		UserID:      "   ",
		Filename:    "notes.txt",
		RawBytes:    []byte("hello"),
	})
	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got: %v", err)
	}

	// 4. Filename invalid
	_, err = svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      1,
		WorkspaceID: 1,
		UserID:      "user-1",
		Filename:    "   ",
		RawBytes:    []byte("hello"),
	})
	if !errors.Is(err, ErrInvalidFilename) {
		t.Fatalf("expected ErrInvalidFilename, got: %v", err)
	}

	// 5. Blocked extension
	for _, ext := range []string{"malware.exe", "script.sh", "installer.msi", "cmd.bat"} {
		_, err = svc.UploadAttachment(&model.UploadAttachmentRequest{
			TaskID:      1,
			WorkspaceID: 1,
			UserID:      "user-1",
			Filename:    ext,
			RawBytes:    []byte("payload"),
		})
		if !errors.Is(err, ErrBlockedFileType) {
			t.Fatalf("expected ErrBlockedFileType for %s, got: %v", ext, err)
		}
	}

	// 6. Empty file
	_, err = svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      1,
		WorkspaceID: 1,
		UserID:      "user-1",
		Filename:    "empty.txt",
		RawBytes:    []byte{},
	})
	if !errors.Is(err, ErrEmptyFile) {
		t.Fatalf("expected ErrEmptyFile, got: %v", err)
	}
}

func TestAttachmentService_UploadAndDeduplication(t *testing.T) {
	svc, _, _ := setupTestAttachmentSvc(t)

	content := []byte("Antigravity Task Manager Architecture Diagram Payload")

	// Upload first copy
	att1, err := svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      10,
		WorkspaceID: 1,
		UserID:      "usr-dev-1",
		Filename:    "architecture.png",
		RawBytes:    content,
	})
	if err != nil {
		t.Fatalf("upload att1 failed: %v", err)
	}
	if att1.ID == 0 {
		t.Fatalf("expected positive ID, got: %d", att1.ID)
	}
	if att1.FileSize != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), att1.FileSize)
	}

	// Verify file was written to disk
	if _, err := os.Stat(att1.StoragePath); os.IsNotExist(err) {
		t.Fatalf("storage file does not exist on disk: %s", att1.StoragePath)
	}

	// Upload duplicate file to a different task with base64 encoding
	encoded := base64.StdEncoding.EncodeToString(content)
	att2, err := svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:        20,
		WorkspaceID:   1,
		UserID:        "usr-dev-2",
		Filename:      "arch_copy.png",
		ContentBase64: encoded,
	})
	if err != nil {
		t.Fatalf("upload att2 failed: %v", err)
	}
	if att2.ID == att1.ID {
		t.Fatalf("att2 should have distinct ID from att1")
	}

	// Deduplication check: Both attachments must point to the identical physical storage path
	if att1.StoragePath != att2.StoragePath {
		t.Fatalf("expected storage path deduplication! att1: %s, att2: %s", att1.StoragePath, att2.StoragePath)
	}
	if att1.ChecksumSHA256 != att2.ChecksumSHA256 {
		t.Fatalf("expected identical checksums")
	}
}

func TestAttachmentService_DownloadAndETag(t *testing.T) {
	svc, _, _ := setupTestAttachmentSvc(t)

	content := []byte("Confidential Project Specifications v2.0")
	att, err := svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      5,
		WorkspaceID: 1,
		UserID:      "usr-lead",
		Filename:    "specs.pdf",
		RawBytes:    content,
	})
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	// Open file
	reader, meta, etag, err := svc.OpenAttachmentFile(att.ID)
	if err != nil {
		t.Fatalf("open attachment file failed: %v", err)
	}
	defer reader.Close()

	if etag != `"`+att.ChecksumSHA256+`"` {
		t.Fatalf("expected etag %q, got %q", `"`+att.ChecksumSHA256+`"`, etag)
	}
	if meta.DownloadCount != 1 {
		t.Fatalf("expected download count 1, got %d", meta.DownloadCount)
	}

	readBytes, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if !bytes.Equal(readBytes, content) {
		t.Fatalf("content mismatch: got %q, want %q", string(readBytes), string(content))
	}
}

func TestAttachmentService_ListUpdateDelete(t *testing.T) {
	svc, _, _ := setupTestAttachmentSvc(t)

	contentA := []byte("File A Content")
	contentB := []byte("File B Content")

	attA, err := svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      100,
		WorkspaceID: 2,
		UserID:      "usr-qa",
		Filename:    "test_plan.md",
		RawBytes:    contentA,
	})
	if err != nil {
		t.Fatalf("upload A failed: %v", err)
	}

	_, err = svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      100,
		WorkspaceID: 2,
		UserID:      "usr-qa",
		Filename:    "results.json",
		RawBytes:    contentB,
	})
	if err != nil {
		t.Fatalf("upload B failed: %v", err)
	}

	// List for Task 100
	taskID := int64(100)
	summary, err := svc.ListAttachments(model.AttachmentFilter{
		TaskID: &taskID,
	})
	if err != nil {
		t.Fatalf("list attachments failed: %v", err)
	}
	if summary.Total != 2 {
		t.Fatalf("expected 2 attachments, got: %d", summary.Total)
	}

	// Update metadata
	newFilename := "renamed_plan.md"
	archived := true
	updated, err := svc.UpdateAttachment(attA.ID, &model.UpdateAttachmentRequest{
		Filename:   &newFilename,
		IsArchived: &archived,
	})
	if err != nil {
		t.Fatalf("update attachment failed: %v", err)
	}
	if updated.Filename != "renamed_plan.md" {
		t.Fatalf("expected updated filename, got %s", updated.Filename)
	}
	if !updated.IsArchived {
		t.Fatalf("expected archived true")
	}

	// Try opening archived file -> should fail
	_, _, _, err = svc.OpenAttachmentFile(attA.ID)
	if err == nil {
		t.Fatalf("expected error opening archived file, got nil")
	}

	// Delete attachment
	err = svc.DeleteAttachment(attA.ID)
	if err != nil {
		t.Fatalf("delete attachment failed: %v", err)
	}

	_, err = svc.GetAttachment(attA.ID)
	if !errors.Is(err, ErrAttachmentNotFound) {
		t.Fatalf("expected ErrAttachmentNotFound after delete, got: %v", err)
	}
}

func TestAttachmentService_StorageMetrics(t *testing.T) {
	svc, _, _ := setupTestAttachmentSvc(t)

	wsID := int64(99)
	_, _ = svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      1,
		WorkspaceID: wsID,
		UserID:      "u1",
		Filename:    "doc1.txt",
		RawBytes:    []byte("hello world 1"),
	})
	_, _ = svc.UploadAttachment(&model.UploadAttachmentRequest{
		TaskID:      2,
		WorkspaceID: wsID,
		UserID:      "u1",
		Filename:    "doc2.txt",
		RawBytes:    []byte("hello world 222"),
	})

	metrics, err := svc.GetStorageMetrics(wsID)
	if err != nil {
		t.Fatalf("get storage metrics failed: %v", err)
	}
	if metrics.TotalFiles != 2 {
		t.Fatalf("expected 2 files, got %d", metrics.TotalFiles)
	}
	if metrics.TotalBytes <= 0 {
		t.Fatalf("expected positive total bytes, got %d", metrics.TotalBytes)
	}
	if len(metrics.LargestFiles) != 2 {
		t.Fatalf("expected 2 largest files, got %d", len(metrics.LargestFiles))
	}
}
