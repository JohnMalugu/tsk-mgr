package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// DefaultSecretByteLength specifies the default entropy length for auto-generated secrets
	DefaultSecretByteLength = 32
	// SignaturePrefix is the header value prefix for sha256 HMAC signatures
	SignaturePrefix = "sha256="
)

// GenerateRandomSecret produces a cryptographically secure random hex secret string
func GenerateRandomSecret(byteLength int) (string, error) {
	if byteLength <= 0 {
		byteLength = DefaultSecretByteLength
	}
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ComputeWebhookSignature calculates the HMAC-SHA256 signature for a payload and secret
func ComputeWebhookSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return SignaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

// VerifyWebhookSignature verifies an incoming signature against payload and secret in constant time
func VerifyWebhookSignature(payload []byte, secret string, signatureHeader string) bool {
	if !strings.HasPrefix(signatureHeader, SignaturePrefix) {
		return false
	}
	expected := ComputeWebhookSignature(payload, secret)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signatureHeader)) == 1
}
