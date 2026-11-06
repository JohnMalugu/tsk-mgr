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

// CreateClientRequest payload for adding a new client to a workspace
type CreateClientRequest struct {
	WorkspaceID int64  `json:"workspace_id"`
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Address     string `json:"address,omitempty"`
	TaxID       string `json:"tax_id,omitempty"`
	Currency    string `json:"currency,omitempty"` // defaults to "USD" if blank
	Notes       string `json:"notes,omitempty"`
}

// UpdateClientRequest payload for modifying client attributes
type UpdateClientRequest struct {
	Name        *string `json:"name,omitempty"`
	Email       *string `json:"email,omitempty"`
	CompanyName *string `json:"company_name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Address     *string `json:"address,omitempty"`
	TaxID       *string `json:"tax_id,omitempty"`
	Currency    *string `json:"currency,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// CreateInvoiceRequest payload for drafting a new invoice
type CreateInvoiceRequest struct {
	WorkspaceID    int64   `json:"workspace_id"`
	ClientID       int64   `json:"client_id"`
	ProjectID      *int64  `json:"project_id,omitempty"`
	IssueDate      string  `json:"issue_date"` // YYYY-MM-DD
	DueDate        string  `json:"due_date"`   // YYYY-MM-DD
	Currency       string  `json:"currency,omitempty"`
	TaxRate        float64 `json:"tax_rate,omitempty"`
	DiscountAmount float64 `json:"discount_amount,omitempty"`
	Notes          string  `json:"notes,omitempty"`
	PaymentTerms   string  `json:"payment_terms,omitempty"`
	Items          []InvoiceItemRequest `json:"items,omitempty"`
}

// UpdateInvoiceRequest payload for modifying draft invoice attributes
type UpdateInvoiceRequest struct {
	ProjectID      *int64   `json:"project_id,omitempty"`
	IssueDate      *string  `json:"issue_date,omitempty"`
	DueDate        *string  `json:"due_date,omitempty"`
	TaxRate        *float64 `json:"tax_rate,omitempty"`
	DiscountAmount *float64 `json:"discount_amount,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
	PaymentTerms   *string  `json:"payment_terms,omitempty"`
}

// ChangeInvoiceStatusRequest payload for manually transitioning invoice status
type ChangeInvoiceStatusRequest struct {
	Status string `json:"status"` // draft, sent, void
	Notes  string `json:"notes,omitempty"`
}

// InvoiceItemRequest payload for creating or updating an individual invoice line item
type InvoiceItemRequest struct {
	Description     string  `json:"description"`
	ItemType        string  `json:"item_type,omitempty"` // labor, service, expense, product
	Quantity        float64 `json:"quantity"`
	UnitPrice       float64 `json:"unit_price"`
	TimesheetLineID *int64  `json:"timesheet_line_id,omitempty"`
}

// RecordPaymentRequest payload for logging a settlement payment against an invoice
type RecordPaymentRequest struct {
	Amount               float64 `json:"amount"`
	PaymentDate          string  `json:"payment_date"` // YYYY-MM-DD
	PaymentMethod        string  `json:"payment_method"`
	TransactionReference string  `json:"transaction_reference,omitempty"`
	Notes                string  `json:"notes,omitempty"`
}

// GenerateInvoiceFromTimesheetRequest payload for creating an invoice directly from an approved timesheet
type GenerateInvoiceFromTimesheetRequest struct {
	WorkspaceID  int64   `json:"workspace_id"`
	ClientID     int64   `json:"client_id"`
	TimesheetID  int64   `json:"timesheet_id"`
	HourlyRate   float64 `json:"hourly_rate,omitempty"`
	IssueDate    string  `json:"issue_date,omitempty"` // YYYY-MM-DD
	DueDate      string  `json:"due_date,omitempty"`   // YYYY-MM-DD
	PaymentTerms string  `json:"payment_terms,omitempty"`
	Notes        string  `json:"notes,omitempty"`
}

// InvoiceFilter provides querying and filtering options across invoices
type InvoiceFilter struct {
	WorkspaceID int64   `json:"workspace_id"`
	ClientID    *int64  `json:"client_id,omitempty"`
	ProjectID   *int64  `json:"project_id,omitempty"`
	Status      *string `json:"status,omitempty"`
	StartDate   *string `json:"start_date,omitempty"` // YYYY-MM-DD
	EndDate     *string `json:"end_date,omitempty"`   // YYYY-MM-DD
	Limit       int     `json:"limit,omitempty"`
	Offset      int     `json:"offset,omitempty"`
}

// InvoiceMetrics aggregates financial billing, payment status, and AR aging buckets
type InvoiceMetrics struct {
	WorkspaceID            int64   `json:"workspace_id"`
	TotalInvoices          int     `json:"total_invoices"`
	DraftCount             int     `json:"draft_count"`
	SentCount              int     `json:"sent_count"`
	PaidCount              int     `json:"paid_count"`
	PartiallyPaidCount     int     `json:"partially_paid_count"`
	OverdueCount           int     `json:"overdue_count"`
	TotalBilledAmount      float64 `json:"total_billed_amount"`
	TotalPaidAmount        float64 `json:"total_paid_amount"`
	TotalOutstandingAmount float64 `json:"total_outstanding_amount"`
	TotalOverdueAmount     float64 `json:"total_overdue_amount"`
	AgingCurrent           float64 `json:"aging_current"` // 0-30 days
	Aging30to60            float64 `json:"aging_30_to_60"`
	Aging60to90            float64 `json:"aging_60_to_90"`
	AgingOver90            float64 `json:"aging_over_90"`
}






