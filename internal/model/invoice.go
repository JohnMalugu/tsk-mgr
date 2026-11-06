package model

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrClientNameRequired = errors.New("client name is required")
	ErrInvalidClientEmail = errors.New("invalid client email format")
)

// Client represents an external organization or individual billed for workspace projects
type Client struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	CompanyName string    `json:"company_name,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	Address     string    `json:"address,omitempty"`
	TaxID       string    `json:"tax_id,omitempty"`
	Currency    string    `json:"currency"` // default: "USD"
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Enriched fields
	InvoicesCount int     `json:"invoices_count,omitempty"`
	TotalBilled   float64 `json:"total_billed,omitempty"`
	TotalPaid     float64 `json:"total_paid,omitempty"`
}

// ValidateClient performs basic validation on a client record
func ValidateClient(name, email string) error {
	if strings.TrimSpace(name) == "" {
		return ErrClientNameRequired
	}
	email = strings.TrimSpace(email)
	if email != "" && (!strings.Contains(email, "@") || !strings.Contains(email, ".")) {
		return ErrInvalidClientEmail
	}
	return nil
}
