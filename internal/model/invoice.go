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

// Invoice status constants
const (
	InvoiceStatusDraft         = "draft"
	InvoiceStatusSent          = "sent"
	InvoiceStatusPaid          = "paid"
	InvoiceStatusPartiallyPaid = "partially_paid"
	InvoiceStatusOverdue       = "overdue"
	InvoiceStatusVoid          = "void"
)

// Invoice represents a billing document sent to a client for work completed
type Invoice struct {
	ID             int64      `json:"id"`
	WorkspaceID    int64      `json:"workspace_id"`
	ClientID       int64      `json:"client_id"`
	ProjectID      *int64     `json:"project_id,omitempty"`
	TimesheetID    *int64     `json:"timesheet_id,omitempty"`
	InvoiceNumber  string     `json:"invoice_number"`
	Status         string     `json:"status"`
	IssueDate      string     `json:"issue_date"` // YYYY-MM-DD
	DueDate        string     `json:"due_date"`   // YYYY-MM-DD
	Currency       string     `json:"currency"`
	Subtotal       float64    `json:"subtotal"`
	TaxRate        float64    `json:"tax_rate"`
	TaxAmount      float64    `json:"tax_amount"`
	DiscountAmount float64    `json:"discount_amount"`
	TotalAmount    float64    `json:"total_amount"`
	AmountPaid     float64    `json:"amount_paid"`
	BalanceDue     float64    `json:"balance_due"`
	PaidAt         *time.Time `json:"paid_at,omitempty"`
	Notes          string     `json:"notes,omitempty"`
	PaymentTerms   string     `json:"payment_terms,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Enriched fields
	ClientName  string `json:"client_name,omitempty"`
	ClientEmail string `json:"client_email,omitempty"`
	ProjectName string `json:"project_name,omitempty"`
	ItemsCount  int    `json:"items_count,omitempty"`
}

// Invoice item type constants
const (
	InvoiceItemTypeLabor   = "labor"
	InvoiceItemTypeService = "service"
	InvoiceItemTypeExpense = "expense"
	InvoiceItemTypeProduct = "product"
)

// InvoiceItem represents a discrete line item billed within an invoice
type InvoiceItem struct {
	ID              int64     `json:"id"`
	InvoiceID       int64     `json:"invoice_id"`
	TimesheetLineID *int64    `json:"timesheet_line_id,omitempty"`
	Description     string    `json:"description"`
	ItemType        string    `json:"item_type"` // labor, service, expense, product
	Quantity        float64   `json:"quantity"`
	UnitPrice       float64   `json:"unit_price"`
	Amount          float64   `json:"amount"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Payment method constants
const (
	PaymentMethodCreditCard   = "credit_card"
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodCheck        = "check"
	PaymentMethodCash         = "cash"
	PaymentMethodStripe       = "stripe"
	PaymentMethodOther        = "other"
)

// InvoicePayment records a settlement payment against an invoice
type InvoicePayment struct {
	ID                   int64     `json:"id"`
	InvoiceID            int64     `json:"invoice_id"`
	Amount               float64   `json:"amount"`
	PaymentDate          string    `json:"payment_date"` // YYYY-MM-DD
	PaymentMethod        string    `json:"payment_method"`
	TransactionReference string    `json:"transaction_reference,omitempty"`
	Notes                string    `json:"notes,omitempty"`
	RecordedByUserID     string    `json:"recorded_by_user_id"`
	CreatedAt            time.Time `json:"created_at"`

	// Enriched fields
	RecordedByUsername string `json:"recorded_by_username,omitempty"`
}


