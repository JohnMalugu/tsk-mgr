package service

import (
	"errors"
	"sync"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
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
