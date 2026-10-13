package repository

import (
	"database/sql"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type CommentRepository interface {
	Create(comment *model.Comment) error
	GetByID(id int) (*model.Comment, error)
	GetByTaskID(taskID int) ([]*model.Comment, error)
	Update(comment *model.Comment) error
	Delete(id int) error
}

type commentRepositoryImpl struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) CommentRepository {
	return &commentRepositoryImpl{db: db}
}

func (r *commentRepositoryImpl) Create(comment *model.Comment) error {
	query := `
	INSERT INTO task_comments (task_id, user_id, content, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		comment.TaskID, comment.UserID, comment.Content, comment.CreatedAt, comment.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		comment.ID = int(id)
	}

	return err
}

func (r *commentRepositoryImpl) GetByID(id int) (*model.Comment, error) {
	query := `
	SELECT id, task_id, user_id, content, created_at, updated_at
	FROM task_comments WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var c model.Comment

	err := row.Scan(&c.ID, &c.TaskID, &c.UserID, &c.Content, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &c, nil
}

