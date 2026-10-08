package service

import (
	"errors"
	"sync"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
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
	mu    sync.RWMutex
	users map[string]*model.User
}

func NewAuthService() AuthService {
	return &authServiceImpl{
		users: make(map[string]*model.User),
	}
}

func (s *authServiceImpl) Register(username, email, password string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.users[username]; exists {
		return nil, ErrUserExists
	}
	
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	
	user := &model.User{
		ID:           "u-" + username, // Simple ID for now
		Username:     username,
		Email:        email,
		PasswordHash: string(hashed),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	
	s.users[username] = user
	return user, nil
}
