package service

import (
	"errors"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrUserExists   = errors.New("user already exists")
	ErrInvalidCreds = errors.New("invalid credentials")
)
