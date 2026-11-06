package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrClientNotFound          = errors.New("client not found")
	ErrInvoiceNotFound         = errors.New("invoice not found")
	ErrInvoiceItemNotFound     = errors.New("invoice item not found")
	ErrInvoiceNotEditable      = errors.New("invoice cannot be edited in current status")
	ErrInvalidInvoiceDates     = errors.New("invalid invoice dates: due date must be on or after issue date")
	ErrInvalidPaymentAmount    = errors.New("payment amount must be greater than zero and cannot exceed balance due")
	ErrInvalidItemQuantity     = errors.New("item quantity must be greater than zero")
	ErrInvalidItemUnitPrice    = errors.New("unit price must be non-negative")
	ErrItemDescriptionRequired = errors.New("item description is required")
	ErrTimesheetNotApproved    = errors.New("timesheet must be in approved status to generate an invoice")
	ErrUnauthorizedInvoice     = errors.New("user not authorized to perform action on invoice")
	ErrInvalidInvoiceStatus    = errors.New("invalid invoice status")
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
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if req.ClientID <= 0 {
		return nil, ErrClientNotFound
	}

	ws, err := s.workspaceRepo.GetByID(int(req.WorkspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	client, err := s.repo.GetClientByID(req.ClientID)
	if err != nil || client == nil || client.WorkspaceID != req.WorkspaceID {
		return nil, ErrClientNotFound
	}

	if req.ProjectID != nil && *req.ProjectID > 0 {
		proj, err := s.projectRepo.GetByID(int(*req.ProjectID))
		if err != nil || proj == nil {
			return nil, ErrProjectNotFound
		}
	}

	issueStr := strings.TrimSpace(req.IssueDate)
	if issueStr == "" {
		issueStr = time.Now().UTC().Format("2006-01-02")
	}
	issueTime, err := time.Parse("2006-01-02", issueStr)
	if err != nil {
		return nil, ErrInvalidInvoiceDates
	}

	dueStr := strings.TrimSpace(req.DueDate)
	if dueStr == "" {
		dueStr = issueTime.AddDate(0, 0, 30).Format("2006-01-02")
	}
	dueTime, err := time.Parse("2006-01-02", dueStr)
	if err != nil {
		return nil, ErrInvalidInvoiceDates
	}

	if dueTime.Before(issueTime) {
		return nil, ErrInvalidInvoiceDates
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = client.Currency
	}
	if currency == "" {
		currency = "USD"
	}

	taxRate := req.TaxRate
	if taxRate < 0 {
		taxRate = 0
	}
	discount := req.DiscountAmount
	if discount < 0 {
		discount = 0
	}

	inv := &model.Invoice{
		WorkspaceID:    req.WorkspaceID,
		ClientID:       req.ClientID,
		ProjectID:      req.ProjectID,
		Status:         model.InvoiceStatusDraft,
		IssueDate:      issueStr,
		DueDate:        dueStr,
		Currency:       currency,
		TaxRate:        taxRate,
		DiscountAmount: discount,
		Notes:          strings.TrimSpace(req.Notes),
		PaymentTerms:   strings.TrimSpace(req.PaymentTerms),
	}

	id, err := s.repo.CreateInvoice(inv)
	if err != nil {
		return nil, err
	}

	for _, itemReq := range req.Items {
		desc := strings.TrimSpace(itemReq.Description)
		if desc == "" {
			continue
		}
		qty := itemReq.Quantity
		if qty <= 0 {
			qty = 1
		}
		unitPrice := itemReq.UnitPrice
		if unitPrice < 0 {
			unitPrice = 0
		}
		itemType := strings.TrimSpace(itemReq.ItemType)
		if itemType == "" {
			itemType = model.InvoiceItemTypeService
		}

		item := &model.InvoiceItem{
			InvoiceID:       id,
			Description:     desc,
			ItemType:        itemType,
			Quantity:        qty,
			UnitPrice:       unitPrice,
			Amount:          qty * unitPrice,
			TimesheetLineID: itemReq.TimesheetLineID,
		}
		_, _ = s.repo.AddItem(item)
	}

	if len(req.Items) > 0 {
		_ = s.repo.RecalculateTotals(id)
	}

	return s.GetInvoice(id)
}

func (s *invoiceServiceImpl) GetInvoice(id int64) (*model.Invoice, error) {
	if id <= 0 {
		return nil, ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(id)
	if err != nil || inv == nil {
		return nil, ErrInvoiceNotFound
	}
	return inv, nil
}

func (s *invoiceServiceImpl) ListInvoices(filter model.InvoiceFilter) ([]*model.Invoice, error) {
	if filter.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	invoices, err := s.repo.ListInvoices(filter)
	if err != nil {
		return nil, err
	}
	if invoices == nil {
		return []*model.Invoice{}, nil
	}
	return invoices, nil
}

func (s *invoiceServiceImpl) UpdateInvoice(id int64, userID string, req model.UpdateInvoiceRequest) (*model.Invoice, error) {
	if id <= 0 {
		return nil, ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(id)
	if err != nil || inv == nil {
		return nil, ErrInvoiceNotFound
	}

	if inv.Status != model.InvoiceStatusDraft {
		return nil, ErrInvoiceNotEditable
	}

	if req.ProjectID != nil {
		if *req.ProjectID > 0 {
			proj, err := s.projectRepo.GetByID(int(*req.ProjectID))
			if err != nil || proj == nil {
				return nil, ErrProjectNotFound
			}
			inv.ProjectID = req.ProjectID
		} else {
			inv.ProjectID = nil
		}
	}

	issueStr := inv.IssueDate
	if req.IssueDate != nil {
		issueStr = strings.TrimSpace(*req.IssueDate)
	}
	issueTime, err := time.Parse("2006-01-02", issueStr)
	if err != nil {
		return nil, ErrInvalidInvoiceDates
	}

	dueStr := inv.DueDate
	if req.DueDate != nil {
		dueStr = strings.TrimSpace(*req.DueDate)
	}
	dueTime, err := time.Parse("2006-01-02", dueStr)
	if err != nil {
		return nil, ErrInvalidInvoiceDates
	}

	if dueTime.Before(issueTime) {
		return nil, ErrInvalidInvoiceDates
	}

	inv.IssueDate = issueStr
	inv.DueDate = dueStr

	if req.TaxRate != nil {
		taxRate := *req.TaxRate
		if taxRate < 0 {
			taxRate = 0
		}
		inv.TaxRate = taxRate
	}

	if req.DiscountAmount != nil {
		discount := *req.DiscountAmount
		if discount < 0 {
			discount = 0
		}
		inv.DiscountAmount = discount
	}

	if req.Notes != nil {
		inv.Notes = strings.TrimSpace(*req.Notes)
	}

	if req.PaymentTerms != nil {
		inv.PaymentTerms = strings.TrimSpace(*req.PaymentTerms)
	}

	if err := s.repo.UpdateInvoice(inv); err != nil {
		return nil, err
	}
	_ = s.repo.RecalculateTotals(id)
	return s.GetInvoice(id)
}

func (s *invoiceServiceImpl) DeleteInvoice(id int64, userID string) error {
	if id <= 0 {
		return ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(id)
	if err != nil || inv == nil {
		return ErrInvoiceNotFound
	}

	if inv.Status != model.InvoiceStatusDraft && inv.Status != model.InvoiceStatusVoid {
		return ErrInvoiceNotEditable
	}

	return s.repo.DeleteInvoice(id)
}

func (s *invoiceServiceImpl) ChangeStatus(id int64, userID string, req model.ChangeInvoiceStatusRequest) (*model.Invoice, error) {
	if id <= 0 {
		return nil, ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(id)
	if err != nil || inv == nil {
		return nil, ErrInvoiceNotFound
	}

	target := strings.ToLower(strings.TrimSpace(req.Status))
	switch target {
	case model.InvoiceStatusDraft:
		if inv.AmountPaid > 0 {
			return nil, ErrInvoiceNotEditable
		}
		inv.Status = model.InvoiceStatusDraft
	case model.InvoiceStatusSent:
		if inv.Status == model.InvoiceStatusPaid {
			return nil, ErrInvoiceNotEditable
		}
		inv.Status = model.InvoiceStatusSent
	case model.InvoiceStatusVoid:
		if inv.AmountPaid > 0 {
			return nil, ErrInvoiceNotEditable
		}
		inv.Status = model.InvoiceStatusVoid
	default:
		return nil, ErrInvalidInvoiceStatus
	}

	if strings.TrimSpace(req.Notes) != "" {
		if inv.Notes != "" {
			inv.Notes += "\n" + strings.TrimSpace(req.Notes)
		} else {
			inv.Notes = strings.TrimSpace(req.Notes)
		}
	}

	if err := s.repo.UpdateInvoice(inv); err != nil {
		return nil, err
	}
	return s.GetInvoice(id)
}

func (s *invoiceServiceImpl) AddItem(invoiceID int64, userID string, req model.InvoiceItemRequest) (*model.InvoiceItem, error) {
	if invoiceID <= 0 {
		return nil, ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil || inv == nil {
		return nil, ErrInvoiceNotFound
	}
	if inv.Status != model.InvoiceStatusDraft {
		return nil, ErrInvoiceNotEditable
	}

	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		return nil, ErrItemDescriptionRequired
	}
	if req.Quantity <= 0 {
		return nil, ErrInvalidItemQuantity
	}
	if req.UnitPrice < 0 {
		return nil, ErrInvalidItemUnitPrice
	}

	itemType := strings.TrimSpace(req.ItemType)
	if itemType == "" {
		itemType = model.InvoiceItemTypeService
	}

	item := &model.InvoiceItem{
		InvoiceID:       invoiceID,
		Description:     desc,
		ItemType:        itemType,
		Quantity:        req.Quantity,
		UnitPrice:       req.UnitPrice,
		Amount:          req.Quantity * req.UnitPrice,
		TimesheetLineID: req.TimesheetLineID,
	}

	id, err := s.repo.AddItem(item)
	if err != nil {
		return nil, err
	}
	_ = s.repo.RecalculateTotals(invoiceID)
	return s.repo.GetItemByID(id)
}

func (s *invoiceServiceImpl) UpdateItem(invoiceID, itemID int64, userID string, req model.InvoiceItemRequest) (*model.InvoiceItem, error) {
	if invoiceID <= 0 {
		return nil, ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil || inv == nil {
		return nil, ErrInvoiceNotFound
	}
	if inv.Status != model.InvoiceStatusDraft {
		return nil, ErrInvoiceNotEditable
	}

	item, err := s.repo.GetItemByID(itemID)
	if err != nil || item == nil || item.InvoiceID != invoiceID {
		return nil, ErrInvoiceItemNotFound
	}

	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		return nil, ErrItemDescriptionRequired
	}
	if req.Quantity <= 0 {
		return nil, ErrInvalidItemQuantity
	}
	if req.UnitPrice < 0 {
		return nil, ErrInvalidItemUnitPrice
	}

	itemType := strings.TrimSpace(req.ItemType)
	if itemType == "" {
		itemType = item.ItemType
	}

	item.Description = desc
	item.ItemType = itemType
	item.Quantity = req.Quantity
	item.UnitPrice = req.UnitPrice
	item.Amount = req.Quantity * req.UnitPrice
	if req.TimesheetLineID != nil {
		item.TimesheetLineID = req.TimesheetLineID
	}

	if err := s.repo.UpdateItem(item); err != nil {
		return nil, err
	}
	_ = s.repo.RecalculateTotals(invoiceID)
	return s.repo.GetItemByID(itemID)
}

func (s *invoiceServiceImpl) DeleteItem(invoiceID, itemID int64, userID string) error {
	if invoiceID <= 0 {
		return ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil || inv == nil {
		return ErrInvoiceNotFound
	}
	if inv.Status != model.InvoiceStatusDraft {
		return ErrInvoiceNotEditable
	}

	item, err := s.repo.GetItemByID(itemID)
	if err != nil || item == nil || item.InvoiceID != invoiceID {
		return ErrInvoiceItemNotFound
	}

	if err := s.repo.DeleteItem(itemID); err != nil {
		return err
	}
	_ = s.repo.RecalculateTotals(invoiceID)
	return nil
}

func (s *invoiceServiceImpl) ListItems(invoiceID int64) ([]model.InvoiceItem, error) {
	if invoiceID <= 0 {
		return nil, ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil || inv == nil {
		return nil, ErrInvoiceNotFound
	}
	items, err := s.repo.ListItems(invoiceID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		return []model.InvoiceItem{}, nil
	}
	return items, nil
}

func (s *invoiceServiceImpl) GenerateFromTimesheet(userID string, req model.GenerateInvoiceFromTimesheetRequest) (*model.Invoice, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	if req.ClientID <= 0 {
		return nil, ErrClientNotFound
	}
	if req.TimesheetID <= 0 {
		return nil, ErrTimesheetNotFound
	}

	ws, err := s.workspaceRepo.GetByID(int(req.WorkspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	client, err := s.repo.GetClientByID(req.ClientID)
	if err != nil || client == nil || client.WorkspaceID != req.WorkspaceID {
		return nil, ErrClientNotFound
	}

	ts, err := s.timesheetRepo.GetTimesheetByID(req.TimesheetID)
	if err != nil || ts == nil || ts.WorkspaceID != req.WorkspaceID {
		return nil, ErrTimesheetNotFound
	}

	if ts.Status != model.TimesheetStatusApproved {
		return nil, ErrTimesheetNotApproved
	}

	rate := req.HourlyRate
	if rate <= 0 {
		rate = 100.0
	}

	issueStr := strings.TrimSpace(req.IssueDate)
	if issueStr == "" {
		issueStr = time.Now().UTC().Format("2006-01-02")
	}
	issueTime, err := time.Parse("2006-01-02", issueStr)
	if err != nil {
		return nil, ErrInvalidInvoiceDates
	}

	dueStr := strings.TrimSpace(req.DueDate)
	if dueStr == "" {
		dueStr = issueTime.AddDate(0, 0, 30).Format("2006-01-02")
	}
	dueTime, err := time.Parse("2006-01-02", dueStr)
	if err != nil {
		return nil, ErrInvalidInvoiceDates
	}
	if dueTime.Before(issueTime) {
		return nil, ErrInvalidInvoiceDates
	}

	currency := client.Currency
	if currency == "" {
		currency = "USD"
	}

	notes := strings.TrimSpace(req.Notes)
	if notes == "" {
		notes = fmt.Sprintf("Generated from Timesheet #%d (%s to %s)", ts.ID, ts.PeriodStartDate, ts.PeriodEndDate)
	}
	terms := strings.TrimSpace(req.PaymentTerms)
	if terms == "" {
		terms = "Net 30"
	}

	inv := &model.Invoice{
		WorkspaceID:  req.WorkspaceID,
		ClientID:     req.ClientID,
		TimesheetID:  &req.TimesheetID,
		Status:       model.InvoiceStatusDraft,
		IssueDate:    issueStr,
		DueDate:      dueStr,
		Currency:     currency,
		Notes:        notes,
		PaymentTerms: terms,
	}

	invID, err := s.repo.CreateInvoice(inv)
	if err != nil {
		return nil, err
	}

	lines, err := s.timesheetRepo.ListTimesheetLines(req.TimesheetID)
	if err == nil {
		for _, line := range lines {
			if !line.IsBillable || line.DurationMinutes <= 0 {
				continue
			}
			hours := float64(line.DurationMinutes) / 60.0
			desc := strings.TrimSpace(line.Description)
			if desc == "" {
				desc = fmt.Sprintf("Labor on %s", line.WorkDate)
			}
			if line.TaskTitle != "" {
				desc = fmt.Sprintf("%s (%s)", desc, line.TaskTitle)
			}

			lineID := line.ID
			item := &model.InvoiceItem{
				InvoiceID:       invID,
				Description:     desc,
				ItemType:        model.InvoiceItemTypeLabor,
				Quantity:        hours,
				UnitPrice:       rate,
				Amount:          hours * rate,
				TimesheetLineID: &lineID,
			}
			_, _ = s.repo.AddItem(item)
		}
	}

	_ = s.repo.RecalculateTotals(invID)
	return s.GetInvoice(invID)
}

func (s *invoiceServiceImpl) RecordPayment(invoiceID int64, userID string, req model.RecordPaymentRequest) (*model.InvoicePayment, error) {
	if invoiceID <= 0 {
		return nil, ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil || inv == nil {
		return nil, ErrInvoiceNotFound
	}

	if inv.Status == model.InvoiceStatusDraft || inv.Status == model.InvoiceStatusVoid {
		return nil, ErrInvoiceNotEditable
	}

	if req.Amount <= 0 || req.Amount > (inv.BalanceDue+0.0001) {
		return nil, ErrInvalidPaymentAmount
	}

	pDate := strings.TrimSpace(req.PaymentDate)
	if pDate == "" {
		pDate = time.Now().UTC().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", pDate); err != nil {
		return nil, errors.New("invalid payment date format (YYYY-MM-DD)")
	}

	method := strings.TrimSpace(req.PaymentMethod)
	if method == "" {
		method = model.PaymentMethodBankTransfer
	}

	payment := &model.InvoicePayment{
		InvoiceID:            invoiceID,
		Amount:               req.Amount,
		PaymentDate:          pDate,
		PaymentMethod:        method,
		TransactionReference: strings.TrimSpace(req.TransactionReference),
		Notes:                strings.TrimSpace(req.Notes),
		RecordedByUserID:     userID,
		CreatedAt:            time.Now(),
	}

	id, err := s.repo.RecordPayment(payment)
	if err != nil {
		return nil, err
	}

	payments, err := s.repo.ListPayments(invoiceID)
	if err == nil {
		for _, p := range payments {
			if p.ID == id {
				return &p, nil
			}
		}
	}

	payment.ID = id
	return payment, nil
}

func (s *invoiceServiceImpl) ListPayments(invoiceID int64) ([]model.InvoicePayment, error) {
	if invoiceID <= 0 {
		return nil, ErrInvoiceNotFound
	}
	inv, err := s.repo.GetInvoiceByID(invoiceID)
	if err != nil || inv == nil {
		return nil, ErrInvoiceNotFound
	}
	payments, err := s.repo.ListPayments(invoiceID)
	if err != nil {
		return nil, err
	}
	if payments == nil {
		return []model.InvoicePayment{}, nil
	}
	return payments, nil
}

func (s *invoiceServiceImpl) GetInvoiceMetrics(workspaceID int64) (*model.InvoiceMetrics, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	ws, err := s.workspaceRepo.GetByID(int(workspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}
	return s.repo.GetInvoiceMetrics(workspaceID)
}
