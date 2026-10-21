package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

// ErrAttachmentNotFound is returned when an attachment is not found
var ErrAttachmentNotFound = errors.New("attachment not found")

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
	query := `
	SELECT id, task_id, user_id, workspace_id, filename, original_name, file_size, content_type,
		storage_path, checksum_sha256, download_count, is_archived, created_at, updated_at
	FROM attachments
	WHERE workspace_id = ? AND checksum_sha256 = ? AND is_archived = 0
	LIMIT 1`

	row := r.db.QueryRow(query, workspaceID, checksum)
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
		return nil, fmt.Errorf("failed to get attachment by checksum: %w", err)
	}
	return &att, nil
}

func (r *sqlAttachmentRepository) ListAttachments(filter model.AttachmentFilter) ([]model.AttachmentDB, int, error) {
	var conditions []string
	var args []interface{}

	if filter.TaskID != nil {
		conditions = append(conditions, "task_id = ?")
		args = append(args, *filter.TaskID)
	}
	if filter.WorkspaceID != nil {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, *filter.WorkspaceID)
	}
	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *filter.UserID)
	}
	if filter.ContentType != nil {
		conditions = append(conditions, "content_type LIKE ?")
		args = append(args, "%"+*filter.ContentType+"%")
	}
	if filter.IsArchived != nil {
		conditions = append(conditions, "is_archived = ?")
		args = append(args, *filter.IsArchived)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM attachments %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count attachments: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, task_id, user_id, workspace_id, filename, original_name, file_size, content_type,
		storage_path, checksum_sha256, download_count, is_archived, created_at, updated_at
	FROM attachments
	%s
	ORDER BY id DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query attachments: %w", err)
	}
	defer rows.Close()

	var atts []model.AttachmentDB
	for rows.Next() {
		var att model.AttachmentDB
		err := rows.Scan(
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
			return nil, 0, fmt.Errorf("failed to scan attachment: %w", err)
		}
		atts = append(atts, att)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row error listing attachments: %w", err)
	}

	return atts, total, nil
}

func (r *sqlAttachmentRepository) UpdateAttachment(att *model.AttachmentDB) error {
	query := `
	UPDATE attachments
	SET filename = ?, content_type = ?, is_archived = ?, updated_at = ?
	WHERE id = ?`

	att.UpdatedAt = time.Now()
	res, err := r.db.Exec(query,
		att.Filename,
		att.ContentType,
		att.IsArchived,
		att.UpdatedAt,
		att.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update attachment %d: %w", att.ID, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrAttachmentNotFound
	}

	return nil
}

func (r *sqlAttachmentRepository) DeleteAttachment(id int64) error {
	query := `DELETE FROM attachments WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete attachment %d: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrAttachmentNotFound
	}

	return nil
}

func (r *sqlAttachmentRepository) IncrementDownloadCount(id int64) error {
	return nil
}

func (r *sqlAttachmentRepository) GetStorageMetrics(workspaceID int64) (*model.StorageMetrics, error) {
	return nil, nil
}
