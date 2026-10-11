package repository

import (
	"database/sql"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type ProjectRepository interface {
	Create(project *model.Project) error
	GetByID(id int) (*model.Project, error)
	GetByUserID(userID string) ([]*model.Project, error)
	Update(project *model.Project) error
	Delete(id int) error
}

type projectRepositoryImpl struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepositoryImpl{db: db}
}

func (r *projectRepositoryImpl) Create(project *model.Project) error {
	query := `
	INSERT INTO projects (user_id, name, description, color, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query,
		project.UserID, project.Name, project.Description, project.Color,
		project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		project.ID = int(id)
	}

	return err
}

func (r *projectRepositoryImpl) GetByID(id int) (*model.Project, error) {
	query := `
	SELECT id, user_id, name, description, color, created_at, updated_at
	FROM projects WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var project model.Project

	err := row.Scan(&project.ID, &project.UserID, &project.Name, &project.Description,
		&project.Color, &project.CreatedAt, &project.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (r *projectRepositoryImpl) GetByUserID(userID string) ([]*model.Project, error) {
	query := `
	SELECT id, user_id, name, description, color, created_at, updated_at
	FROM projects WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*model.Project
	for rows.Next() {
		var p model.Project
		err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description,
			&p.Color, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, &p)
	}

	return projects, rows.Err()
}
