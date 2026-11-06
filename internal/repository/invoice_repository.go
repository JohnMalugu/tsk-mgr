package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)


var (
	ErrClientNotFound      = errors.New("client not found")
	ErrInvoiceNotFound     = errors.New("invoice not found")
	ErrInvoiceItemNotFound = errors.New("invoice item not found")
)

const clientBaseSelect = `
	SELECT 
		c.id,
		c.workspace_id,
		c.name,
		c.email,
		c.company_name,
		c.phone,
		c.address,
		c.tax_id,
		c.currency,
		c.notes,
		c.created_at,
		c.updated_at,
		(SELECT COUNT(*) FROM invoices inv WHERE inv.client_id = c.id) AS invoices_count,
		(SELECT COALESCE(SUM(inv.total_amount), 0.0) FROM invoices inv WHERE inv.client_id = c.id) AS total_billed,
		(SELECT COALESCE(SUM(inv.amount_paid), 0.0) FROM invoices inv WHERE inv.client_id = c.id) AS total_paid
	FROM clients c
`

func scanClient(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.Client, error) {
	var c model.Client
	err := scanner.Scan(
		&c.ID,
		&c.WorkspaceID,
		&c.Name,
		&c.Email,
		&c.CompanyName,
		&c.Phone,
		&c.Address,
		&c.TaxID,
		&c.Currency,
		&c.Notes,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.InvoicesCount,
		&c.TotalBilled,
		&c.TotalPaid,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}


// InvoiceRepository manages database operations for clients, invoices, line items, and payments
type InvoiceRepository interface {
	// Client methods
	CreateClient(client *model.Client) (int64, error)
	GetClientByID(id int64) (*model.Client, error)
	ListClients(workspaceID int64) ([]*model.Client, error)
	UpdateClient(client *model.Client) error
	DeleteClient(id int64) error

	// Invoice methods
	GetNextInvoiceNumber(workspaceID int64) (string, error)
	CreateInvoice(invoice *model.Invoice) (int64, error)
	GetInvoiceByID(id int64) (*model.Invoice, error)
	GetInvoiceByNumber(workspaceID int64, invoiceNumber string) (*model.Invoice, error)
	ListInvoices(filter model.InvoiceFilter) ([]*model.Invoice, error)
	UpdateInvoice(invoice *model.Invoice) error
	DeleteInvoice(id int64) error

	// Line item methods
	AddItem(item *model.InvoiceItem) (int64, error)
	GetItemByID(id int64) (*model.InvoiceItem, error)
	UpdateItem(item *model.InvoiceItem) error
	DeleteItem(id int64) error
	ListItems(invoiceID int64) ([]model.InvoiceItem, error)
	RecalculateTotals(invoiceID int64) error

	// Payment methods
	RecordPayment(payment *model.InvoicePayment) (int64, error)
	ListPayments(invoiceID int64) ([]model.InvoicePayment, error)

	// Metrics
	GetInvoiceMetrics(workspaceID int64) (*model.InvoiceMetrics, error)
}

type invoiceRepositoryImpl struct {
	db *sql.DB
}

// NewInvoiceRepository creates a new instance of InvoiceRepository
func NewInvoiceRepository(db *sql.DB) InvoiceRepository {
	return &invoiceRepositoryImpl{db: db}
}

func (r *invoiceRepositoryImpl) CreateClient(client *model.Client) (int64, error) {
	now := time.Now()
	client.CreatedAt = now
	client.UpdatedAt = now
	if client.Currency == "" {
		client.Currency = "USD"
	}

	query := `
	INSERT INTO clients (
		workspace_id, name, email, company_name, phone, address, tax_id, currency, notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.Exec(
		query,
		client.WorkspaceID,
		client.Name,
		client.Email,
		client.CompanyName,
		client.Phone,
		client.Address,
		client.TaxID,
		client.Currency,
		client.Notes,
		client.CreatedAt,
		client.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	client.ID = id
	return id, nil
}

func (r *invoiceRepositoryImpl) GetClientByID(id int64) (*model.Client, error) {
	query := clientBaseSelect + ` WHERE c.id = ?`
	row := r.db.QueryRow(query, id)
	client, err := scanClient(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClientNotFound
		}
		return nil, err
	}
	return client, nil
}

func (r *invoiceRepositoryImpl) ListClients(workspaceID int64) ([]*model.Client, error) {
	query := clientBaseSelect + ` WHERE c.workspace_id = ? ORDER BY c.name ASC, c.id ASC`
	rows, err := r.db.Query(query, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*model.Client
	for rows.Next() {
		client, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}

func (r *invoiceRepositoryImpl) UpdateClient(client *model.Client) error {
	client.UpdatedAt = time.Now()
	query := `
	UPDATE clients
	SET name = ?,
		email = ?,
		company_name = ?,
		phone = ?,
		address = ?,
		tax_id = ?,
		currency = ?,
		notes = ?,
		updated_at = ?
	WHERE id = ?
	`
	res, err := r.db.Exec(
		query,
		client.Name,
		client.Email,
		client.CompanyName,
		client.Phone,
		client.Address,
		client.TaxID,
		client.Currency,
		client.Notes,
		client.UpdatedAt,
		client.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrClientNotFound
	}
	return nil
}

func (r *invoiceRepositoryImpl) DeleteClient(id int64) error {
	res, err := r.db.Exec(`DELETE FROM clients WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrClientNotFound
	}
	return nil
}


const invoiceBaseSelect = `
	SELECT 
		inv.id,
		inv.workspace_id,
		inv.client_id,
		inv.project_id,
		inv.timesheet_id,
		inv.invoice_number,
		inv.status,
		inv.issue_date,
		inv.due_date,
		inv.currency,
		inv.subtotal,
		inv.tax_rate,
		inv.tax_amount,
		inv.discount_amount,
		inv.total_amount,
		inv.amount_paid,
		inv.balance_due,
		inv.paid_at,
		inv.notes,
		inv.payment_terms,
		inv.created_at,
		inv.updated_at,
		c.name AS client_name,
		c.email AS client_email,
		p.name AS project_name,
		(SELECT COUNT(*) FROM invoice_items ii WHERE ii.invoice_id = inv.id) AS items_count
	FROM invoices inv
	LEFT JOIN clients c ON inv.client_id = c.id
	LEFT JOIN projects p ON inv.project_id = p.id
`

func scanInvoice(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.Invoice, error) {
	var inv model.Invoice
	var projectID, timesheetID sql.NullInt64
	var paidAt sql.NullTime
	var clientName, clientEmail, projectName sql.NullString

	err := scanner.Scan(
		&inv.ID,
		&inv.WorkspaceID,
		&inv.ClientID,
		&projectID,
		&timesheetID,
		&inv.InvoiceNumber,
		&inv.Status,
		&inv.IssueDate,
		&inv.DueDate,
		&inv.Currency,
		&inv.Subtotal,
		&inv.TaxRate,
		&inv.TaxAmount,
		&inv.DiscountAmount,
		&inv.TotalAmount,
		&inv.AmountPaid,
		&inv.BalanceDue,
		&paidAt,
		&inv.Notes,
		&inv.PaymentTerms,
		&inv.CreatedAt,
		&inv.UpdatedAt,
		&clientName,
		&clientEmail,
		&projectName,
		&inv.ItemsCount,
	)
	if err != nil {
		return nil, err
	}
	if projectID.Valid {
		inv.ProjectID = &projectID.Int64
	}
	if timesheetID.Valid {
		inv.TimesheetID = &timesheetID.Int64
	}
	if paidAt.Valid {
		inv.PaidAt = &paidAt.Time
	}
	if clientName.Valid {
		inv.ClientName = clientName.String
	}
	if clientEmail.Valid {
		inv.ClientEmail = clientEmail.String
	}
	if projectName.Valid {
		inv.ProjectName = projectName.String
	}
	return &inv, nil
}

func (r *invoiceRepositoryImpl) GetNextInvoiceNumber(workspaceID int64) (string, error) {
	query := `SELECT COALESCE(MAX(id), 0) + 1 FROM invoices WHERE workspace_id = ?`
	var nextNum int
	if err := r.db.QueryRow(query, workspaceID).Scan(&nextNum); err != nil {
		return "INV-1", err
	}
	return fmt.Sprintf("INV-%d", nextNum), nil
}

func (r *invoiceRepositoryImpl) CreateInvoice(invoice *model.Invoice) (int64, error) {
	now := time.Now()
	invoice.CreatedAt = now
	invoice.UpdatedAt = now
	if invoice.Status == "" {
		invoice.Status = model.InvoiceStatusDraft
	}
	if invoice.Currency == "" {
		invoice.Currency = "USD"
	}
	if invoice.PaymentTerms == "" {
		invoice.PaymentTerms = "Net 30"
	}
	if invoice.InvoiceNumber == "" {
		num, err := r.GetNextInvoiceNumber(invoice.WorkspaceID)
		if err == nil {
			invoice.InvoiceNumber = num
		} else {
			invoice.InvoiceNumber = "INV-1"
		}
	}
	invoice.BalanceDue = invoice.TotalAmount - invoice.AmountPaid
	if invoice.BalanceDue < 0 {
		invoice.BalanceDue = 0
	}

	query := `
	INSERT INTO invoices (
		workspace_id, client_id, project_id, timesheet_id, invoice_number, status,
		issue_date, due_date, currency, subtotal, tax_rate, tax_amount, discount_amount,
		total_amount, amount_paid, balance_due, paid_at, notes, payment_terms, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.Exec(
		query,
		invoice.WorkspaceID,
		invoice.ClientID,
		invoice.ProjectID,
		invoice.TimesheetID,
		invoice.InvoiceNumber,
		invoice.Status,
		invoice.IssueDate,
		invoice.DueDate,
		invoice.Currency,
		invoice.Subtotal,
		invoice.TaxRate,
		invoice.TaxAmount,
		invoice.DiscountAmount,
		invoice.TotalAmount,
		invoice.AmountPaid,
		invoice.BalanceDue,
		invoice.PaidAt,
		invoice.Notes,
		invoice.PaymentTerms,
		invoice.CreatedAt,
		invoice.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	invoice.ID = id
	return id, nil
}

func (r *invoiceRepositoryImpl) GetInvoiceByID(id int64) (*model.Invoice, error) {
	query := invoiceBaseSelect + ` WHERE inv.id = ?`
	row := r.db.QueryRow(query, id)
	inv, err := scanInvoice(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}
	return inv, nil
}

func (r *invoiceRepositoryImpl) GetInvoiceByNumber(workspaceID int64, invoiceNumber string) (*model.Invoice, error) {
	query := invoiceBaseSelect + ` WHERE inv.workspace_id = ? AND inv.invoice_number = ?`
	row := r.db.QueryRow(query, workspaceID, invoiceNumber)
	inv, err := scanInvoice(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvoiceNotFound
		}
		return nil, err
	}
	return inv, nil
}


func (r *invoiceRepositoryImpl) ListInvoices(filter model.InvoiceFilter) ([]*model.Invoice, error) {
	query := invoiceBaseSelect + ` WHERE 1=1`
	var args []interface{}

	if filter.WorkspaceID > 0 {
		query += ` AND inv.workspace_id = ?`
		args = append(args, filter.WorkspaceID)
	}
	if filter.ClientID != nil && *filter.ClientID > 0 {
		query += ` AND inv.client_id = ?`
		args = append(args, *filter.ClientID)
	}
	if filter.ProjectID != nil && *filter.ProjectID > 0 {
		query += ` AND inv.project_id = ?`
		args = append(args, *filter.ProjectID)
	}
	if filter.Status != nil && *filter.Status != "" {
		query += ` AND inv.status = ?`
		args = append(args, *filter.Status)
	}
	if filter.StartDate != nil && *filter.StartDate != "" {
		query += ` AND inv.issue_date >= ?`
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil && *filter.EndDate != "" {
		query += ` AND inv.issue_date <= ?`
		args = append(args, *filter.EndDate)
	}

	query += ` ORDER BY inv.issue_date DESC, inv.created_at DESC, inv.id DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += ` OFFSET ?`
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*model.Invoice
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if invoices == nil {
		invoices = []*model.Invoice{}
	}
	return invoices, nil
}


func (r *invoiceRepositoryImpl) UpdateInvoice(invoice *model.Invoice) error {
	return nil
}

func (r *invoiceRepositoryImpl) DeleteInvoice(id int64) error {
	return nil
}

func (r *invoiceRepositoryImpl) AddItem(item *model.InvoiceItem) (int64, error) {
	return 0, nil
}

func (r *invoiceRepositoryImpl) GetItemByID(id int64) (*model.InvoiceItem, error) {
	return nil, nil
}

func (r *invoiceRepositoryImpl) UpdateItem(item *model.InvoiceItem) error {
	return nil
}

func (r *invoiceRepositoryImpl) DeleteItem(id int64) error {
	return nil
}

func (r *invoiceRepositoryImpl) ListItems(invoiceID int64) ([]model.InvoiceItem, error) {
	return nil, nil
}

func (r *invoiceRepositoryImpl) RecalculateTotals(invoiceID int64) error {
	return nil
}

func (r *invoiceRepositoryImpl) RecordPayment(payment *model.InvoicePayment) (int64, error) {
	return 0, nil
}

func (r *invoiceRepositoryImpl) ListPayments(invoiceID int64) ([]model.InvoicePayment, error) {
	return nil, nil
}

func (r *invoiceRepositoryImpl) GetInvoiceMetrics(workspaceID int64) (*model.InvoiceMetrics, error) {
	return nil, nil
}
