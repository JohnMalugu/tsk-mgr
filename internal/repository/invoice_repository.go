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
	invoice.UpdatedAt = time.Now()
	invoice.BalanceDue = invoice.TotalAmount - invoice.AmountPaid
	if invoice.BalanceDue < 0 {
		invoice.BalanceDue = 0
	}

	query := `
	UPDATE invoices
	SET project_id = ?,
		timesheet_id = ?,
		invoice_number = ?,
		status = ?,
		issue_date = ?,
		due_date = ?,
		currency = ?,
		subtotal = ?,
		tax_rate = ?,
		tax_amount = ?,
		discount_amount = ?,
		total_amount = ?,
		amount_paid = ?,
		balance_due = ?,
		paid_at = ?,
		notes = ?,
		payment_terms = ?,
		updated_at = ?
	WHERE id = ?
	`
	res, err := r.db.Exec(
		query,
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
		invoice.UpdatedAt,
		invoice.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrInvoiceNotFound
	}
	return nil
}

func (r *invoiceRepositoryImpl) DeleteInvoice(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM invoice_payments WHERE invoice_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM invoice_items WHERE invoice_id = ?`, id); err != nil {
		return err
	}

	res, err := tx.Exec(`DELETE FROM invoices WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrInvoiceNotFound
	}

	return tx.Commit()
}

const invoiceItemBaseSelect = `
	SELECT 
		id, invoice_id, timesheet_line_id, description, item_type, quantity, unit_price, amount, created_at, updated_at
	FROM invoice_items
`

func scanInvoiceItem(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.InvoiceItem, error) {
	var item model.InvoiceItem
	var tsLineID sql.NullInt64

	err := scanner.Scan(
		&item.ID,
		&item.InvoiceID,
		&tsLineID,
		&item.Description,
		&item.ItemType,
		&item.Quantity,
		&item.UnitPrice,
		&item.Amount,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if tsLineID.Valid {
		item.TimesheetLineID = &tsLineID.Int64
	}
	return &item, nil
}

func (r *invoiceRepositoryImpl) AddItem(item *model.InvoiceItem) (int64, error) {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	if item.ItemType == "" {
		item.ItemType = model.InvoiceItemTypeService
	}
	item.Amount = item.Quantity * item.UnitPrice

	query := `
	INSERT INTO invoice_items (
		invoice_id, timesheet_line_id, description, item_type, quantity, unit_price, amount, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.Exec(
		query,
		item.InvoiceID,
		item.TimesheetLineID,
		item.Description,
		item.ItemType,
		item.Quantity,
		item.UnitPrice,
		item.Amount,
		item.CreatedAt,
		item.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	item.ID = id
	return id, nil
}

func (r *invoiceRepositoryImpl) GetItemByID(id int64) (*model.InvoiceItem, error) {
	query := invoiceItemBaseSelect + ` WHERE id = ?`
	row := r.db.QueryRow(query, id)
	item, err := scanInvoiceItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvoiceItemNotFound
		}
		return nil, err
	}
	return item, nil
}

func (r *invoiceRepositoryImpl) UpdateItem(item *model.InvoiceItem) error {
	item.UpdatedAt = time.Now()
	item.Amount = item.Quantity * item.UnitPrice

	query := `
	UPDATE invoice_items
	SET description = ?,
		item_type = ?,
		quantity = ?,
		unit_price = ?,
		amount = ?,
		timesheet_line_id = ?,
		updated_at = ?
	WHERE id = ?
	`
	res, err := r.db.Exec(
		query,
		item.Description,
		item.ItemType,
		item.Quantity,
		item.UnitPrice,
		item.Amount,
		item.TimesheetLineID,
		item.UpdatedAt,
		item.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrInvoiceItemNotFound
	}
	return nil
}

func (r *invoiceRepositoryImpl) DeleteItem(id int64) error {
	res, err := r.db.Exec(`DELETE FROM invoice_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrInvoiceItemNotFound
	}
	return nil
}

func (r *invoiceRepositoryImpl) ListItems(invoiceID int64) ([]model.InvoiceItem, error) {
	query := invoiceItemBaseSelect + ` WHERE invoice_id = ? ORDER BY id ASC`
	rows, err := r.db.Query(query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.InvoiceItem
	for rows.Next() {
		item, err := scanInvoiceItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if items == nil {
		items = []model.InvoiceItem{}
	}
	return items, nil
}

func (r *invoiceRepositoryImpl) RecalculateTotals(invoiceID int64) error {
	var subtotal float64
	err := r.db.QueryRow(`SELECT COALESCE(SUM(amount), 0.0) FROM invoice_items WHERE invoice_id = ?`, invoiceID).Scan(&subtotal)
	if err != nil {
		return err
	}

	var taxRate, discountAmount, amountPaid float64
	err = r.db.QueryRow(`SELECT tax_rate, discount_amount, amount_paid FROM invoices WHERE id = ?`, invoiceID).Scan(
		&taxRate, &discountAmount, &amountPaid,
	)
	if err != nil {
		return err
	}

	taxAmount := subtotal * (taxRate / 100.0)
	totalAmount := subtotal + taxAmount - discountAmount
	if totalAmount < 0 {
		totalAmount = 0
	}
	balanceDue := totalAmount - amountPaid
	if balanceDue < 0 {
		balanceDue = 0
	}

	updateQuery := `
	UPDATE invoices
	SET subtotal = ?,
		tax_amount = ?,
		total_amount = ?,
		balance_due = ?,
		updated_at = ?
	WHERE id = ?
	`
	_, err = r.db.Exec(updateQuery, subtotal, taxAmount, totalAmount, balanceDue, time.Now(), invoiceID)
	return err
}

func (r *invoiceRepositoryImpl) RecordPayment(payment *model.InvoicePayment) (int64, error) {
	now := time.Now()
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = now
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO invoice_payments (
		invoice_id, amount, payment_date, payment_method, transaction_reference, notes, recorded_by_user_id, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := tx.Exec(
		query,
		payment.InvoiceID,
		payment.Amount,
		payment.PaymentDate,
		payment.PaymentMethod,
		payment.TransactionReference,
		payment.Notes,
		payment.RecordedByUserID,
		payment.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	payment.ID = id

	// Recompute invoice payment totals and status
	var amountPaid float64
	err = tx.QueryRow(`SELECT COALESCE(SUM(amount), 0.0) FROM invoice_payments WHERE invoice_id = ?`, payment.InvoiceID).Scan(&amountPaid)
	if err != nil {
		return 0, err
	}

	var totalAmount float64
	var currentStatus string
	err = tx.QueryRow(`SELECT total_amount, status FROM invoices WHERE id = ?`, payment.InvoiceID).Scan(&totalAmount, &currentStatus)
	if err != nil {
		return 0, err
	}

	balanceDue := totalAmount - amountPaid
	var paidAt *time.Time
	newStatus := currentStatus

	if balanceDue <= 0.0001 {
		balanceDue = 0
		newStatus = model.InvoiceStatusPaid
		paidAt = &now
	} else if amountPaid > 0 {
		newStatus = model.InvoiceStatusPartiallyPaid
	}

	updateQuery := `
	UPDATE invoices
	SET amount_paid = ?,
		balance_due = ?,
		status = ?,
		paid_at = ?,
		updated_at = ?
	WHERE id = ?
	`
	_, err = tx.Exec(updateQuery, amountPaid, balanceDue, newStatus, paidAt, now, payment.InvoiceID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (r *invoiceRepositoryImpl) ListPayments(invoiceID int64) ([]model.InvoicePayment, error) {
	query := `
	SELECT 
		ip.id,
		ip.invoice_id,
		ip.amount,
		ip.payment_date,
		ip.payment_method,
		ip.transaction_reference,
		ip.notes,
		ip.recorded_by_user_id,
		ip.created_at,
		u.username AS recorded_by_username
	FROM invoice_payments ip
	LEFT JOIN users u ON ip.recorded_by_user_id = u.id
	WHERE ip.invoice_id = ?
	ORDER BY ip.payment_date ASC, ip.id ASC
	`
	rows, err := r.db.Query(query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []model.InvoicePayment
	for rows.Next() {
		var p model.InvoicePayment
		var username sql.NullString
		err := rows.Scan(
			&p.ID,
			&p.InvoiceID,
			&p.Amount,
			&p.PaymentDate,
			&p.PaymentMethod,
			&p.TransactionReference,
			&p.Notes,
			&p.RecordedByUserID,
			&p.CreatedAt,
			&username,
		)
		if err != nil {
			return nil, err
		}
		if username.Valid {
			p.RecordedByUsername = username.String
		}
		payments = append(payments, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if payments == nil {
		payments = []model.InvoicePayment{}
	}
	return payments, nil
}

func (r *invoiceRepositoryImpl) GetInvoiceMetrics(workspaceID int64) (*model.InvoiceMetrics, error) {
	metrics := &model.InvoiceMetrics{
		WorkspaceID: workspaceID,
	}

	summaryQuery := `
	SELECT 
		COUNT(*),
		COALESCE(SUM(CASE WHEN status = 'draft' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'paid' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'partially_paid' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'overdue' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(total_amount), 0.0),
		COALESCE(SUM(amount_paid), 0.0),
		COALESCE(SUM(balance_due), 0.0)
	FROM invoices
	WHERE workspace_id = ? AND status != 'void'
	`
	err := r.db.QueryRow(summaryQuery, workspaceID).Scan(
		&metrics.TotalInvoices,
		&metrics.DraftCount,
		&metrics.SentCount,
		&metrics.PaidCount,
		&metrics.PartiallyPaidCount,
		&metrics.OverdueCount,
		&metrics.TotalBilledAmount,
		&metrics.TotalPaidAmount,
		&metrics.TotalOutstandingAmount,
	)
	if err != nil {
		return nil, err
	}

	// Calculate Aging and Overdue amounts
	agingQuery := `
	SELECT due_date, balance_due
	FROM invoices
	WHERE workspace_id = ? AND balance_due > 0 AND status != 'void'
	`
	rows, err := r.db.Query(agingQuery, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now().Truncate(24 * time.Hour)
	for rows.Next() {
		var dueDateStr string
		var balanceDue float64
		if err := rows.Scan(&dueDateStr, &balanceDue); err != nil {
			return nil, err
		}

		dueDate, err := time.Parse("2006-01-02", dueDateStr)
		if err != nil {
			metrics.AgingCurrent += balanceDue
			continue
		}

		daysPastDue := int(now.Sub(dueDate).Hours() / 24)
		if daysPastDue > 0 {
			metrics.TotalOverdueAmount += balanceDue
		}

		switch {
		case daysPastDue <= 30:
			metrics.AgingCurrent += balanceDue
		case daysPastDue <= 60:
			metrics.Aging30to60 += balanceDue
		case daysPastDue <= 90:
			metrics.Aging60to90 += balanceDue
		default:
			metrics.AgingOver90 += balanceDue
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}
