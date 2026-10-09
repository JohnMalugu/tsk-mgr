package repository

import (
	"database/sql"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type UserRepository interface {
	Create(user *model.User) error
	GetByUsername(username string) (*model.User, error)
}

type userRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) Create(user *model.User) error {
	query := `
	INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`
	
	_, err := r.db.Exec(query, user.ID, user.Username, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	return err
}
