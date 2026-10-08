package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	secretKey = []byte("my-secret-key-for-auth") // In production, load from env
	ErrInvalidToken = errors.New("invalid token")
)
