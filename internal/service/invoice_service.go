package service

import (
	"errors"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrClientNotFound       = errors.New("client not found")
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrInvoiceItemNotFound  = errors.New("invoice item not found")
	ErrInvoiceNotEditable   = errors.New("invoice cannot be edited in current status")
	ErrInvalidInvoiceDates  = errors.New("invalid invoice dates: due date must be on or after issue date")
	ErrInvalidPaymentAmount = errors.New("payment amount must be greater than zero and cannot exceed balance due")
	ErrInvalidItemQuantity  = errors.New("item quantity must be greater than zero")
	ErrInvalidItemUnitPrice = errors.New("unit price must be non-negative")
	ErrTimesheetNotApproved = errors.New("timesheet must be in approved status to generate an invoice")
	ErrUnauthorizedInvoice  = errors.New("user not authorized to perform action on invoice")
)

// InvoiceService manages enterprise client billing, invoices, line items, and payment settlements
type InvoiceService interface {
	// Client management
	CreateClient(userID string, req model.CreateClientRequest) (*model.Client, error)
	GetClient(id int64) (*model.Client, error)
	ListClients(workspaceID int64) ([]*model.Client, error)
	UpdateClient(id int64, userID string, req model.UpdateClientRequest) (*model.Client, error)
	DeleteClient(id int64, userID string) error

	// Invoice management
	CreateInvoice(userID string, req model.CreateInvoiceRequest) (*model.Invoice, error)
	GetInvoice(id int64) (*model.Invoice, error)
	ListInvoices(filter model.InvoiceFilter) ([]*model.Invoice, error)
	UpdateInvoice(id int64, userID string, req model.UpdateInvoiceRequest) (*model.Invoice, error)
	DeleteInvoice(id int64, userID string) error
	ChangeStatus(id int64, userID string, req model.ChangeInvoiceStatusRequest) (*model.Invoice, error)

	// Line items & timesheet generation
	AddItem(invoiceID int64, userID string, req model.InvoiceItemRequest) (*model.InvoiceItem, error)
	UpdateItem(invoiceID, itemID int64, userID string, req model.InvoiceItemRequest) (*model.InvoiceItem, error)
	DeleteItem(invoiceID, itemID int64, userID string) error
	ListItems(invoiceID int64) ([]model.InvoiceItem, error)
	GenerateFromTimesheet(userID string, req model.GenerateInvoiceFromTimesheetRequest) (*model.Invoice, error)

	// Payments & Metrics
	RecordPayment(invoiceID int64, userID string, req model.RecordPaymentRequest) (*model.InvoicePayment, error)
	ListPayments(invoiceID int64) ([]model.InvoicePayment, error)
	GetInvoiceMetrics(workspaceID int64) (*model.InvoiceMetrics, error)
}

type invoiceServiceImpl struct {
	repo          repository.InvoiceRepository
	workspaceRepo repository.WorkspaceRepository
	projectRepo   repository.ProjectRepository
	timesheetRepo repository.TimesheetRepository
}

// NewInvoiceService creates a new instance of InvoiceService
func NewInvoiceService(
	repo repository.InvoiceRepository,
	workspaceRepo repository.WorkspaceRepository,
	projectRepo repository.ProjectRepository,
	timesheetRepo repository.TimesheetRepository,
) InvoiceService {
	return &invoiceServiceImpl{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		projectRepo:   projectRepo,
		timesheetRepo: timesheetRepo,
	}
}

func (s *invoiceServiceImpl) CreateClient(userID string, req model.CreateClientRequest) (*model.Client, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if err := model.ValidateClient(req.Name, req.Email); err != nil {
		return nil, err
	}

	ws, err := s.workspaceRepo.GetByID(int(req.WorkspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	client := &model.Client{
		WorkspaceID: req.WorkspaceID,
		Name:        strings.TrimSpace(req.Name),
		Email:       strings.TrimSpace(req.Email),
		CompanyName: strings.TrimSpace(req.CompanyName),
		Phone:       strings.TrimSpace(req.Phone),
		Address:     strings.TrimSpace(req.Address),
		TaxID:       strings.TrimSpace(req.TaxID),
		Currency:    currency,
		Notes:       strings.TrimSpace(req.Notes),
	}

	id, err := s.repo.CreateClient(client)
	if err != nil {
		return nil, err
	}
	return s.repo.GetClientByID(id)
}

func (s *invoiceServiceImpl) GetClient(id int64) (*model.Client, error) {
	if id <= 0 {
		return nil, ErrClientNotFound
	}
	client, err := s.repo.GetClientByID(id)
	if err != nil || client == nil {
		return nil, ErrClientNotFound
	}
	return client, nil
}

func (s *invoiceServiceImpl) ListClients(workspaceID int64) ([]*model.Client, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	clients, err := s.repo.ListClients(workspaceID)
	if err != nil {
		return nil, err
	}
	if clients == nil {
		return []*model.Client{}, nil
	}
	return clients, nil
}

func (s *invoiceServiceImpl) UpdateClient(id int64, userID string, req model.UpdateClientRequest) (*model.Client, error) {
	if id <= 0 {
		return nil, ErrClientNotFound
	}
	client, err := s.repo.GetClientByID(id)
	if err != nil || client == nil {
		return nil, ErrClientNotFound
	}

	name := client.Name
	if req.Name != nil {
		name = *req.Name
	}
	email := client.Email
	if req.Email != nil {
		email = *req.Email
	}
	if err := model.ValidateClient(name, email); err != nil {
		return nil, err
	}

	client.Name = strings.TrimSpace(name)
	client.Email = strings.TrimSpace(email)
	if req.CompanyName != nil {
		client.CompanyName = strings.TrimSpace(*req.CompanyName)
	}
	if req.Phone != nil {
		client.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Address != nil {
		client.Address = strings.TrimSpace(*req.Address)
	}
	if req.TaxID != nil {
		client.TaxID = strings.TrimSpace(*req.TaxID)
	}
	if req.Currency != nil {
		curr := strings.ToUpper(strings.TrimSpace(*req.Currency))
		if curr != "" {
			client.Currency = curr
		}
	}
	if req.Notes != nil {
		client.Notes = strings.TrimSpace(*req.Notes)
	}

	if err := s.repo.UpdateClient(client); err != nil {
		return nil, err
	}
	return s.repo.GetClientByID(id)
}

func (s *invoiceServiceImpl) DeleteClient(id int64, userID string) error {
	if id <= 0 {
		return ErrClientNotFound
	}
	client, err := s.repo.GetClientByID(id)
	if err != nil || client == nil {
		return ErrClientNotFound
	}
	return s.repo.DeleteClient(id)
}

func (s *invoiceServiceImpl) CreateInvoice(userID string, req model.CreateInvoiceRequest) (*model.Invoice, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) GetInvoice(id int64) (*model.Invoice, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) ListInvoices(filter model.InvoiceFilter) ([]*model.Invoice, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) UpdateInvoice(id int64, userID string, req model.UpdateInvoiceRequest) (*model.Invoice, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) DeleteInvoice(id int64, userID string) error {
	return nil
}

func (s *invoiceServiceImpl) ChangeStatus(id int64, userID string, req model.ChangeInvoiceStatusRequest) (*model.Invoice, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) AddItem(invoiceID int64, userID string, req model.InvoiceItemRequest) (*model.InvoiceItem, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) UpdateItem(invoiceID, itemID int64, userID string, req model.InvoiceItemRequest) (*model.InvoiceItem, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) DeleteItem(invoiceID, itemID int64, userID string) error {
	return nil
}

func (s *invoiceServiceImpl) ListItems(invoiceID int64) ([]model.InvoiceItem, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) GenerateFromTimesheet(userID string, req model.GenerateInvoiceFromTimesheetRequest) (*model.Invoice, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) RecordPayment(invoiceID int64, userID string, req model.RecordPaymentRequest) (*model.InvoicePayment, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) ListPayments(invoiceID int64) ([]model.InvoicePayment, error) {
	return nil, nil
}

func (s *invoiceServiceImpl) GetInvoiceMetrics(workspaceID int64) (*model.InvoiceMetrics, error) {
	return nil, nil
}
