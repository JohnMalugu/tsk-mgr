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

func GenerateToken(userID, username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"usr": username,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
