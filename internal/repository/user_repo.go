package repository

import (
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type UserRepository interface {
	Create(user *model.User) error
	GetByUsername(username string) (*model.User, error)
}
