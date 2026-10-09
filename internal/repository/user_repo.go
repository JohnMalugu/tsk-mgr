package repository

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var ErrDuplicateUser = errors.New("duplicate user")

type UserRepository interface {
	Create(user *model.User) error
	GetByUsername(username string) (*model.User, error)
	GetByID(id string) (*model.User, error)
	Update(user *model.User) error
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

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return ErrDuplicateUser
		}
	}

	return err
}

func (r *userRepositoryImpl) GetByUsername(username string) (*model.User, error) {
	query := `
	SELECT id, username, email, password_hash, created_at, updated_at
	FROM users WHERE username = ?`

	row := r.db.QueryRow(query, username)

	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // Return nil if not found
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepositoryImpl) GetByID(id string) (*model.User, error) {
	query := `
	SELECT id, username, email, password_hash, created_at, updated_at
	FROM users WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var user model.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepositoryImpl) Update(user *model.User) error {
	query := `
	UPDATE users 
	SET username = ?, email = ?, password_hash = ?, updated_at = ?
	WHERE id = ?`

	_, err := r.db.Exec(query, user.Username, user.Email, user.PasswordHash, user.UpdatedAt, user.ID)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") ||
			strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return ErrDuplicateUser
		}
	}

	return err
}
