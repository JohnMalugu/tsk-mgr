package repository

import (
	"database/sql"
	"errors"
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


func (r *invoiceRepositoryImpl) GetNextInvoiceNumber(workspaceID int64) (string, error) {
	return "", nil
}

func (r *invoiceRepositoryImpl) CreateInvoice(invoice *model.Invoice) (int64, error) {
	return 0, nil
}

func (r *invoiceRepositoryImpl) GetInvoiceByID(id int64) (*model.Invoice, error) {
	return nil, nil
}

func (r *invoiceRepositoryImpl) GetInvoiceByNumber(workspaceID int64, invoiceNumber string) (*model.Invoice, error) {
	return nil, nil
}

func (r *invoiceRepositoryImpl) ListInvoices(filter model.InvoiceFilter) ([]*model.Invoice, error) {
	return nil, nil
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
