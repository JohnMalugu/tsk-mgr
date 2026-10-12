package repository

import (
	"database/sql"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type TagRepository interface {
	Create(tag *model.Tag) error
	GetByID(id int) (*model.Tag, error)
	GetByUserID(userID string) ([]*model.Tag, error)
	Update(tag *model.Tag) error
	Delete(id int) error
	AttachTagToTask(taskID, tagID int) error
	DetachTagFromTask(taskID, tagID int) error
	GetTagsForTask(taskID int) ([]*model.Tag, error)
	GetTasksForTag(tagID int) ([]*model.Task, error)
}

type tagRepositoryImpl struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) TagRepository {
	return &tagRepositoryImpl{db: db}
}

func (r *tagRepositoryImpl) Create(tag *model.Tag) error {
	query := `
	INSERT INTO tags (user_id, name, color, created_at)
	VALUES (?, ?, ?, ?)`

	result, err := r.db.Exec(query, tag.UserID, tag.Name, tag.Color, tag.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		tag.ID = int(id)
	}

	return err
}

func (r *tagRepositoryImpl) GetByID(id int) (*model.Tag, error) {
	query := `SELECT id, user_id, name, color, created_at FROM tags WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var tag model.Tag
	err := row.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &tag, nil
}

func (r *tagRepositoryImpl) GetByUserID(userID string) ([]*model.Tag, error) {
	query := `SELECT id, user_id, name, color, created_at FROM tags WHERE user_id = ? ORDER BY name ASC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []*model.Tag
	for rows.Next() {
		var tag model.Tag
		err := rows.Scan(&tag.ID, &tag.UserID, &tag.Name, &tag.Color, &tag.CreatedAt)
		if err != nil {
			return nil, err
		}
		tags = append(tags, &tag)
	}

	return tags, rows.Err()
}

func (r *tagRepositoryImpl) Update(tag *model.Tag) error {
	query := `UPDATE tags SET name = ?, color = ? WHERE id = ? AND user_id = ?`
	_, err := r.db.Exec(query, tag.Name, tag.Color, tag.ID, tag.UserID)
	return err
}
