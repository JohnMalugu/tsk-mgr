package repository

import (
	"database/sql"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type WorkspaceRepository interface {
	Create(w *model.WorkspaceDB) error
	GetByID(id int) (*model.WorkspaceDB, error)
	GetBySlug(slug string) (*model.WorkspaceDB, error)
	ListUserWorkspaces(userID string) ([]*model.WorkspaceDB, error)
	Update(w *model.WorkspaceDB) error
	Delete(id int) error

	AddMember(m *model.WorkspaceMemberDB) error
	GetMember(workspaceID int, userID string) (*model.WorkspaceMemberDB, error)
	ListMembers(workspaceID int) ([]*model.WorkspaceMemberInfo, error)
	UpdateMemberRole(workspaceID int, userID string, role model.WorkspaceRole) error
	RemoveMember(workspaceID int, userID string) error

	CreateInvite(inv *model.WorkspaceInviteDB) error
	GetInviteByToken(token string) (*model.WorkspaceInviteDB, error)
	UpdateInviteStatus(id int, status string) error
}

type workspaceRepositoryImpl struct {
	db *sql.DB
}

func NewWorkspaceRepository(db *sql.DB) WorkspaceRepository {
	return &workspaceRepositoryImpl{db: db}
}

func (r *workspaceRepositoryImpl) Create(w *model.WorkspaceDB) error {
	now := time.Now().UTC()
	if w.CreatedAt.IsZero() {
		w.CreatedAt = now
	}
	w.UpdatedAt = now

	query := `
	INSERT INTO workspaces (name, slug, owner_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query, w.Name, w.Slug, w.OwnerID, w.CreatedAt, w.UpdatedAt)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	w.ID = int(id)
	return nil
}

func (r *workspaceRepositoryImpl) GetByID(id int) (*model.WorkspaceDB, error) {
	query := `SELECT id, name, slug, owner_id, created_at, updated_at FROM workspaces WHERE id = ?`
	var w model.WorkspaceDB
	err := r.db.QueryRow(query, id).Scan(&w.ID, &w.Name, &w.Slug, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func (r *workspaceRepositoryImpl) GetBySlug(slug string) (*model.WorkspaceDB, error) {
	query := `SELECT id, name, slug, owner_id, created_at, updated_at FROM workspaces WHERE slug = ?`
	var w model.WorkspaceDB
	err := r.db.QueryRow(query, slug).Scan(&w.ID, &w.Name, &w.Slug, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &w, nil
}

func (r *workspaceRepositoryImpl) ListUserWorkspaces(userID string) ([]*model.WorkspaceDB, error) {
	query := `
	SELECT DISTINCT w.id, w.name, w.slug, w.owner_id, w.created_at, w.updated_at
	FROM workspaces w
	LEFT JOIN workspace_members wm ON w.id = wm.workspace_id
	WHERE w.owner_id = ? OR wm.user_id = ?
	ORDER BY w.created_at DESC`

	rows, err := r.db.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.WorkspaceDB
	for rows.Next() {
		var w model.WorkspaceDB
		if err := rows.Scan(&w.ID, &w.Name, &w.Slug, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *workspaceRepositoryImpl) Update(w *model.WorkspaceDB) error {
	w.UpdatedAt = time.Now().UTC()
	query := `UPDATE workspaces SET name = ?, slug = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, w.Name, w.Slug, w.UpdatedAt, w.ID)
	return err
}

func (r *workspaceRepositoryImpl) Delete(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM workspace_invites WHERE workspace_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM workspace_members WHERE workspace_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM workspaces WHERE id = ?`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *workspaceRepositoryImpl) AddMember(m *model.WorkspaceMemberDB) error {
	return nil
}

func (r *workspaceRepositoryImpl) GetMember(workspaceID int, userID string) (*model.WorkspaceMemberDB, error) {
	return nil, nil
}

func (r *workspaceRepositoryImpl) ListMembers(workspaceID int) ([]*model.WorkspaceMemberInfo, error) {
	return nil, nil
}

func (r *workspaceRepositoryImpl) UpdateMemberRole(workspaceID int, userID string, role model.WorkspaceRole) error {
	return nil
}

func (r *workspaceRepositoryImpl) RemoveMember(workspaceID int, userID string) error {
	return nil
}

func (r *workspaceRepositoryImpl) CreateInvite(inv *model.WorkspaceInviteDB) error {
	return nil
}

func (r *workspaceRepositoryImpl) GetInviteByToken(token string) (*model.WorkspaceInviteDB, error) {
	return nil, nil
}

func (r *workspaceRepositoryImpl) UpdateInviteStatus(id int, status string) error {
	return nil
}

