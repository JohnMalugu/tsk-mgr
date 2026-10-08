package service

import (
	"testing"
)

func TestAuthService_Register(t *testing.T) {
	svc := NewAuthService()
	
	// Test basic registration
	user, err := svc.Register("testuser", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", user.Username)
	}
	
	// Test duplicate registration
	_, err = svc.Register("testuser", "test2@example.com", "password456")
	if err != ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestAuthService_Login(t *testing.T) {
	svc := NewAuthService()
	svc.Register("loginuser", "login@example.com", "mypassword")
	
	// Test successful login
	token, err := svc.Login("loginuser", "mypassword")
	if err != nil {
		t.Fatalf("expected successful login, got %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
	
	// Test invalid password
	_, err = svc.Login("loginuser", "wrongpassword")
	if err != ErrInvalidCreds {
		t.Errorf("expected ErrInvalidCreds, got %v", err)
	}
	
	// Test unknown user
	_, err = svc.Login("unknownuser", "password")
	if err != ErrInvalidCreds {
		t.Errorf("expected ErrInvalidCreds, got %v", err)
	}
}
