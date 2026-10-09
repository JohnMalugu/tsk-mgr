package service

import (
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/auth"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrInvalidCreds = errors.New("invalid credentials")
)

type AuthService interface {
	Register(username, email, password string) (*model.User, error)
	Login(username, password string) (string, error)
}

type authServiceImpl struct {
	repo repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) AuthService {
	return &authServiceImpl{
		repo: repo,
	}
}

func (s *authServiceImpl) Register(username, email, password string) (*model.User, error) {
	existingUser, err := s.repo.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserExists
	}
	
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	
	user := &model.User{
		ID:           "u-" + username, // Replace with UUID later
		Username:     username,
		Email:        email,
		PasswordHash: string(hashed),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	
	if err := s.repo.Create(user); err != nil {
		if err == repository.ErrDuplicateUser {
			return nil, ErrUserExists
		}
		return nil, err
	}
	
	return user, nil
}

func (s *authServiceImpl) Login(username, password string) (string, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCreds
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCreds
	}
	
	return auth.GenerateToken(user.ID, user.Username)
}
