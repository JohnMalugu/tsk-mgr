package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestInvoiceSvc(t *testing.T) (InvoiceService, repository.InvoiceRepository, repository.TimesheetRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_mgr', 'charlie', 'charlie@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Acme Studio WS', 'acme-ws-inv', 'user_mgr', datetime('now'), datetime('now')),
		       (2, 'Beta Tech WS', 'beta-ws-inv', 'user_mgr', datetime('now'), datetime('now'));

		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES (1, 'user_1', 'member', datetime('now')),
		       (1, 'user_mgr', 'admin', datetime('now')),
		       (2, 'user_mgr', 'admin', datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Client Portal', 'Frontend and Backend Billing', '#4f46e5', datetime('now'), datetime('now'));

		INSERT INTO tasks (id, user_id, project_id, title, description, status, completed, due_date, estimate_minutes, priority, created_at, updated_at)
		VALUES (501, 'user_1', 101, 'Invoice Architecture', 'Design invoice and ledger', 'done', 1, datetime('now'), 120, 'high', datetime('now'), datetime('now'));

		INSERT INTO timesheets (id, workspace_id, user_id, period_start_date, period_end_date, status, total_minutes, billable_minutes, submitted_at, reviewed_at, reviewer_id, notes, created_at, updated_at)
		VALUES (701, 1, 'user_1', '2026-11-01', '2026-11-07', 'approved', 480, 480, datetime('now'), datetime('now'), 'user_mgr', 'Sprint billing', datetime('now'), datetime('now'));

		INSERT INTO timesheet_lines (id, timesheet_id, task_id, project_id, work_date, duration_minutes, is_billable, description, created_at, updated_at)
		VALUES (801, 701, 501, 101, '2026-11-02', 240, 1, 'Initial backend setup', datetime('now'), datetime('now')),
		       (802, 701, 501, 101, '2026-11-03', 240, 1, 'Database indices & queries', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to seed test database: %v", err)
	}

	invoiceRepo := repository.NewInvoiceRepository(sqliteDB)
	timesheetRepo := repository.NewTimesheetRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	workspaceRepo := repository.NewWorkspaceRepository(sqliteDB)

	svc := NewInvoiceService(invoiceRepo, workspaceRepo, projectRepo, timesheetRepo)
	return svc, invoiceRepo, timesheetRepo, sqliteDB
}

func TestInvoiceService_ClientCRUD(t *testing.T) {
	svc, _, _, dbConn := setupTestInvoiceSvc(t)
	defer dbConn.Close()

	// Validation checks
	_, err := svc.CreateClient("user_1", model.CreateClientRequest{
		WorkspaceID: 0,
		Name:        "Stark Industries",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}

	_, err = svc.CreateClient("user_1", model.CreateClientRequest{
		WorkspaceID: 1,
		Name:        "",
	})
	if !errors.Is(err, model.ErrClientNameRequired) {
		t.Fatalf("expected ErrClientNameRequired, got %v", err)
	}

	_, err = svc.CreateClient("user_1", model.CreateClientRequest{
		WorkspaceID: 1,
		Name:        "Stark Industries",
		Email:       "invalid-email",
	})
	if !errors.Is(err, model.ErrInvalidClientEmail) {
		t.Fatalf("expected ErrInvalidClientEmail, got %v", err)
	}

	// Create valid client
	client, err := svc.CreateClient("user_1", model.CreateClientRequest{
		WorkspaceID: 1,
		Name:        "Stark Industries",
		Email:       "tony@stark.com",
		CompanyName: "Stark Enterprises Inc",
		Phone:       "+1-555-0199",
		Address:     "10880 Malibu Point, CA",
		TaxID:       "US-99887766",
		Currency:    "USD",
		Notes:       "High priority client",
	})
	if err != nil {
		t.Fatalf("CreateClient failed: %v", err)
	}
	if client.ID <= 0 || client.Name != "Stark Industries" || client.Currency != "USD" {
		t.Fatalf("unexpected client fields: %+v", client)
	}

	// Get client
	fetched, err := svc.GetClient(client.ID)
	if err != nil {
		t.Fatalf("GetClient failed: %v", err)
	}
	if fetched.Email != "tony@stark.com" || fetched.TaxID != "US-99887766" {
		t.Fatalf("unexpected fetched client: %+v", fetched)
	}

	// List clients
	clients, err := svc.ListClients(1)
	if err != nil {
		t.Fatalf("ListClients failed: %v", err)
	}
	if len(clients) != 1 || clients[0].ID != client.ID {
		t.Fatalf("expected 1 client in list, got %d", len(clients))
	}

	// Update client
	newName := "Stark Advanced Robotics"
	newTaxID := "US-11223344"
	updated, err := svc.UpdateClient(client.ID, "user_1", model.UpdateClientRequest{
		Name:  &newName,
		TaxID: &newTaxID,
	})
	if err != nil {
		t.Fatalf("UpdateClient failed: %v", err)
	}
	if updated.Name != newName || updated.TaxID != newTaxID {
		t.Fatalf("client was not updated properly: %+v", updated)
	}

	// Delete client
	if err := svc.DeleteClient(client.ID, "user_1"); err != nil {
		t.Fatalf("DeleteClient failed: %v", err)
	}
	_, err = svc.GetClient(client.ID)
	if !errors.Is(err, ErrClientNotFound) {
		t.Fatalf("expected ErrClientNotFound after delete, got %v", err)
	}
}

func TestInvoiceService_InvoiceCRUD(t *testing.T) {
	svc, _, _, dbConn := setupTestInvoiceSvc(t)
	defer dbConn.Close()

	// Setup client
	client, err := svc.CreateClient("user_1", model.CreateClientRequest{
		WorkspaceID: 1,
		Name:        "Wayne Enterprises",
		Email:       "bruce@wayne.com",
		Currency:    "USD",
	})
	if err != nil {
		t.Fatalf("CreateClient failed: %v", err)
	}

	projID := int64(101)
	// Create invoice with draft status & initial items
	inv, err := svc.CreateInvoice("user_1", model.CreateInvoiceRequest{
		WorkspaceID:    1,
		ClientID:       client.ID,
		ProjectID:      &projID,
		IssueDate:      "2026-11-01",
		DueDate:        "2026-12-01",
		Currency:       "USD",
		TaxRate:        10.0,
		DiscountAmount: 50.0,
		Notes:          "November consulting retainers",
		PaymentTerms:   "Net 30",
		Items: []model.InvoiceItemRequest{
			{
				Description: "Architecture & Cloud Security",
				ItemType:    model.InvoiceItemTypeService,
				Quantity:    10,
				UnitPrice:   150.0,
			},
			{
				Description: "Infra Provisioning",
				ItemType:    model.InvoiceItemTypeService,
				Quantity:    5,
				UnitPrice:   100.0,
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateInvoice failed: %v", err)
	}

	if inv.Status != model.InvoiceStatusDraft {
		t.Fatalf("expected status draft, got %s", inv.Status)
	}
	// Subtotal: 10*150 + 5*100 = 1500 + 500 = 2000
	// Tax: 2000 * 0.10 = 200
	// Total: 2000 + 200 - 50 = 2150
	// Balance: 2150
	if inv.Subtotal != 2000.0 || inv.TaxAmount != 200.0 || inv.TotalAmount != 2150.0 || inv.BalanceDue != 2150.0 {
		t.Fatalf("unexpected invoice totals: %+v", inv)
	}
	if inv.InvoiceNumber == "" {
		t.Fatalf("expected invoice number to be generated, got empty")
	}

	// Get invoice
	fetched, err := svc.GetInvoice(inv.ID)
	if err != nil {
		t.Fatalf("GetInvoice failed: %v", err)
	}
	if fetched.ClientName != "Wayne Enterprises" || fetched.ItemsCount != 2 {
		t.Fatalf("unexpected enriched invoice: %+v", fetched)
	}

	// List invoices
	list, err := svc.ListInvoices(model.InvoiceFilter{WorkspaceID: 1})
	if err != nil {
		t.Fatalf("ListInvoices failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 invoice, got %d", len(list))
	}

	// Update invoice
	newTaxRate := 5.0
	newNotes := "Updated discount agreement"
	updated, err := svc.UpdateInvoice(inv.ID, "user_1", model.UpdateInvoiceRequest{
		TaxRate: &newTaxRate,
		Notes:   &newNotes,
	})
	if err != nil {
		t.Fatalf("UpdateInvoice failed: %v", err)
	}
	// Subtotal: 2000, Tax: 2000*0.05=100, Total: 2000+100-50 = 2050
	if updated.TaxAmount != 100.0 || updated.TotalAmount != 2050.0 || updated.Notes != newNotes {
		t.Fatalf("unexpected updated totals: %+v", updated)
	}

	// Change status: Draft -> Sent
	sentInv, err := svc.ChangeStatus(inv.ID, "user_1", model.ChangeInvoiceStatusRequest{
		Status: model.InvoiceStatusSent,
	})
	if err != nil {
		t.Fatalf("ChangeStatus to sent failed: %v", err)
	}
	if sentInv.Status != model.InvoiceStatusSent {
		t.Fatalf("expected status sent, got %s", sentInv.Status)
	}

	// Sent invoice cannot be edited or deleted
	_, err = svc.UpdateInvoice(inv.ID, "user_1", model.UpdateInvoiceRequest{TaxRate: &newTaxRate})
	if !errors.Is(err, ErrInvoiceNotEditable) {
		t.Fatalf("expected ErrInvoiceNotEditable for sent invoice, got %v", err)
	}
	err = svc.DeleteInvoice(inv.ID, "user_1")
	if !errors.Is(err, ErrInvoiceNotEditable) {
		t.Fatalf("expected ErrInvoiceNotEditable on delete sent invoice, got %v", err)
	}

	// Void invoice
	voidInv, err := svc.ChangeStatus(inv.ID, "user_1", model.ChangeInvoiceStatusRequest{
		Status: model.InvoiceStatusVoid,
	})
	if err != nil {
		t.Fatalf("ChangeStatus to void failed: %v", err)
	}
	if voidInv.Status != model.InvoiceStatusVoid {
		t.Fatalf("expected void status, got %s", voidInv.Status)
	}

	// Void invoice can be deleted
	if err := svc.DeleteInvoice(inv.ID, "user_1"); err != nil {
		t.Fatalf("DeleteInvoice for void invoice failed: %v", err)
	}
	_, err = svc.GetInvoice(inv.ID)
	if !errors.Is(err, ErrInvoiceNotFound) {
		t.Fatalf("expected ErrInvoiceNotFound, got %v", err)
	}
}

func TestInvoiceService_LineItems(t *testing.T) {
	svc, _, _, dbConn := setupTestInvoiceSvc(t)
	defer dbConn.Close()

	client, _ := svc.CreateClient("user_1", model.CreateClientRequest{
		WorkspaceID: 1,
		Name:        "Cyberdyne Systems",
		Currency:    "USD",
	})

	inv, err := svc.CreateInvoice("user_1", model.CreateInvoiceRequest{
		WorkspaceID: 1,
		ClientID:    client.ID,
		IssueDate:   "2026-11-01",
		DueDate:     "2026-11-15",
	})
	if err != nil {
		t.Fatalf("CreateInvoice failed: %v", err)
	}

	// Add items
	item1, err := svc.AddItem(inv.ID, "user_1", model.InvoiceItemRequest{
		Description: "Neural Network Consultation",
		ItemType:    model.InvoiceItemTypeLabor,
		Quantity:    8,
		UnitPrice:   125.0,
	})
	if err != nil {
		t.Fatalf("AddItem 1 failed: %v", err)
	}
	if item1.Amount != 1000.0 {
		t.Fatalf("expected item amount 1000, got %f", item1.Amount)
	}

	item2, err := svc.AddItem(inv.ID, "user_1", model.InvoiceItemRequest{
		Description: "Hardware Server License",
		ItemType:    model.InvoiceItemTypeProduct,
		Quantity:    2,
		UnitPrice:   500.0,
	})
	if err != nil {
		t.Fatalf("AddItem 2 failed: %v", err)
	}
	if item2.Amount != 1000.0 {
		t.Fatalf("expected item amount 1000, got %f", item2.Amount)
	}

	// Verify invoice recalculation (1000 + 1000 = 2000)
	inv, _ = svc.GetInvoice(inv.ID)
	if inv.Subtotal != 2000.0 || inv.TotalAmount != 2000.0 {
		t.Fatalf("expected total 2000, got %f", inv.TotalAmount)
	}

	// Update item 1: change quantity from 8 to 4 -> amount becomes 500
	updatedItem, err := svc.UpdateItem(inv.ID, item1.ID, "user_1", model.InvoiceItemRequest{
		Description: item1.Description,
		Quantity:    4,
		UnitPrice:   125.0,
	})
	if err != nil {
		t.Fatalf("UpdateItem failed: %v", err)
	}
	if updatedItem.Amount != 500.0 {
		t.Fatalf("expected updated amount 500, got %f", updatedItem.Amount)
	}

	inv, _ = svc.GetInvoice(inv.ID)
	if inv.TotalAmount != 1500.0 {
		t.Fatalf("expected total 1500 after item update, got %f", inv.TotalAmount)
	}

	// List items
	items, err := svc.ListItems(inv.ID)
	if err != nil {
		t.Fatalf("ListItems failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	// Delete item 2 -> total becomes 500
	if err := svc.DeleteItem(inv.ID, item2.ID, "user_1"); err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}
	inv, _ = svc.GetInvoice(inv.ID)
	if inv.TotalAmount != 500.0 {
		t.Fatalf("expected total 500 after item deletion, got %f", inv.TotalAmount)
	}
}

func TestInvoiceService_GenerateFromTimesheet(t *testing.T) {
	svc, _, _, dbConn := setupTestInvoiceSvc(t)
	defer dbConn.Close()

	client, err := svc.CreateClient("user_1", model.CreateClientRequest{
		WorkspaceID: 1,
		Name:        "Oscorp Industries",
		Currency:    "USD",
	})
	if err != nil {
		t.Fatalf("CreateClient failed: %v", err)
	}

	// Generate invoice from seeded timesheet #701 (which has 2 lines x 240 mins = 8 hours total)
	inv, err := svc.GenerateFromTimesheet("user_1", model.GenerateInvoiceFromTimesheetRequest{
		WorkspaceID:  1,
		ClientID:     client.ID,
		TimesheetID:  701,
		HourlyRate:   150.0,
		IssueDate:    "2026-11-08",
		DueDate:      "2026-12-08",
		PaymentTerms: "Net 30",
	})
	if err != nil {
		t.Fatalf("GenerateFromTimesheet failed: %v", err)
	}

	if inv.TimesheetID == nil || *inv.TimesheetID != 701 {
		t.Fatalf("expected timesheet ID 701, got %v", inv.TimesheetID)
	}
	// 8 hours * 150.0 = 1200.0
	if inv.TotalAmount != 1200.0 || inv.BalanceDue != 1200.0 {
		t.Fatalf("expected total 1200, got %f", inv.TotalAmount)
	}

	items, err := svc.ListItems(inv.ID)
	if err != nil || len(items) != 2 {
		t.Fatalf("expected 2 line items from timesheet, got %d", len(items))
	}
	for _, it := range items {
		if it.Quantity != 4.0 || it.UnitPrice != 150.0 || it.Amount != 600.0 {
			t.Fatalf("unexpected timesheet line item: %+v", it)
		}
	}
}

func TestInvoiceService_PaymentsAndAutoStatus(t *testing.T) {
	svc, _, _, dbConn := setupTestInvoiceSvc(t)
	defer dbConn.Close()

	client, _ := svc.CreateClient("user_1", model.CreateClientRequest{
		WorkspaceID: 1,
		Name:        "Globex Corp",
		Currency:    "USD",
	})

	inv, _ := svc.CreateInvoice("user_1", model.CreateInvoiceRequest{
		WorkspaceID: 1,
		ClientID:    client.ID,
		IssueDate:   "2026-11-01",
		DueDate:     "2026-11-30",
		Items: []model.InvoiceItemRequest{
			{Description: "Cloud Architecture Service", Quantity: 5, UnitPrice: 100.0},
		},
	})

	// Total = 500.0, Status = draft
	// Can't pay draft invoice
	_, err := svc.RecordPayment(inv.ID, "user_1", model.RecordPaymentRequest{
		Amount: 200.0,
	})
	if !errors.Is(err, ErrInvoiceNotEditable) {
		t.Fatalf("expected ErrInvoiceNotEditable on draft invoice payment, got %v", err)
	}

	// Move to sent
	inv, _ = svc.ChangeStatus(inv.ID, "user_1", model.ChangeInvoiceStatusRequest{Status: model.InvoiceStatusSent})

	// Payment exceeding balance due should fail
	_, err = svc.RecordPayment(inv.ID, "user_1", model.RecordPaymentRequest{
		Amount: 600.0,
	})
	if !errors.Is(err, ErrInvalidPaymentAmount) {
		t.Fatalf("expected ErrInvalidPaymentAmount, got %v", err)
	}

	// Partial payment: $200
	p1, err := svc.RecordPayment(inv.ID, "user_1", model.RecordPaymentRequest{
		Amount:               200.0,
		PaymentDate:          "2026-11-05",
		PaymentMethod:        model.PaymentMethodBankTransfer,
		TransactionReference: "WIRE-00123",
		Notes:                "First deposit",
	})
	if err != nil {
		t.Fatalf("RecordPayment 1 failed: %v", err)
	}
	if p1.Amount != 200.0 {
		t.Fatalf("expected payment 200, got %f", p1.Amount)
	}

	// Verify invoice transitioned to partially_paid
	inv, _ = svc.GetInvoice(inv.ID)
	if inv.Status != model.InvoiceStatusPartiallyPaid || inv.AmountPaid != 200.0 || inv.BalanceDue != 300.0 {
		t.Fatalf("unexpected invoice state after partial payment: status=%s, paid=%f, balance=%f", inv.Status, inv.AmountPaid, inv.BalanceDue)
	}

	// Final settlement payment: $300
	p2, err := svc.RecordPayment(inv.ID, "user_1", model.RecordPaymentRequest{
		Amount:               300.0,
		PaymentDate:          "2026-11-10",
		PaymentMethod:        model.PaymentMethodCreditCard,
		TransactionReference: "TX-778899",
	})
	if err != nil {
		t.Fatalf("RecordPayment 2 failed: %v", err)
	}
	if p2.Amount != 300.0 {
		t.Fatalf("expected payment 300, got %f", p2.Amount)
	}

	// Verify invoice transitioned to paid
	inv, _ = svc.GetInvoice(inv.ID)
	if inv.Status != model.InvoiceStatusPaid || inv.AmountPaid != 500.0 || inv.BalanceDue != 0.0 || inv.PaidAt == nil {
		t.Fatalf("expected invoice status paid with zero balance: %+v", inv)
	}

	// List payments
	payments, err := svc.ListPayments(inv.ID)
	if err != nil {
		t.Fatalf("ListPayments failed: %v", err)
	}
	if len(payments) != 2 {
		t.Fatalf("expected 2 payments, got %d", len(payments))
	}
	if payments[0].RecordedByUsername != "alice" {
		t.Fatalf("expected username alice enriched on payment, got %s", payments[0].RecordedByUsername)
	}
}

func TestInvoiceService_MetricsAndAging(t *testing.T) {
	svc, _, _, dbConn := setupTestInvoiceSvc(t)
	defer dbConn.Close()

	client, _ := svc.CreateClient("user_1", model.CreateClientRequest{
		WorkspaceID: 1,
		Name:        "Initech LLC",
		Currency:    "USD",
	})

	today := time.Now().UTC()
	// Current invoice (0-30 days): $1000
	invCurrent, _ := svc.CreateInvoice("user_1", model.CreateInvoiceRequest{
		WorkspaceID: 1,
		ClientID:    client.ID,
		IssueDate:   today.Format("2006-01-02"),
		DueDate:     today.AddDate(0, 0, 15).Format("2006-01-02"),
		Items:       []model.InvoiceItemRequest{{Description: "Item A", Quantity: 1, UnitPrice: 1000.0}},
	})
	_, _ = svc.ChangeStatus(invCurrent.ID, "user_1", model.ChangeInvoiceStatusRequest{Status: model.InvoiceStatusSent})

	// Overdue invoice (due 45 days ago -> 30-60 aging bucket): $2000
	invOverdue45, _ := svc.CreateInvoice("user_1", model.CreateInvoiceRequest{
		WorkspaceID: 1,
		ClientID:    client.ID,
		IssueDate:   today.AddDate(0, 0, -60).Format("2006-01-02"),
		DueDate:     today.AddDate(0, 0, -45).Format("2006-01-02"),
		Items:       []model.InvoiceItemRequest{{Description: "Item B", Quantity: 2, UnitPrice: 1000.0}},
	})
	_, _ = svc.ChangeStatus(invOverdue45.ID, "user_1", model.ChangeInvoiceStatusRequest{Status: model.InvoiceStatusSent})

	// Paid invoice: $500
	invPaid, _ := svc.CreateInvoice("user_1", model.CreateInvoiceRequest{
		WorkspaceID: 1,
		ClientID:    client.ID,
		IssueDate:   today.AddDate(0, 0, -10).Format("2006-01-02"),
		DueDate:     today.AddDate(0, 0, 20).Format("2006-01-02"),
		Items:       []model.InvoiceItemRequest{{Description: "Item C", Quantity: 1, UnitPrice: 500.0}},
	})
	_, _ = svc.ChangeStatus(invPaid.ID, "user_1", model.ChangeInvoiceStatusRequest{Status: model.InvoiceStatusSent})
	_, _ = svc.RecordPayment(invPaid.ID, "user_1", model.RecordPaymentRequest{Amount: 500.0})

	// Fetch metrics
	metrics, err := svc.GetInvoiceMetrics(1)
	if err != nil {
		t.Fatalf("GetInvoiceMetrics failed: %v", err)
	}

	if metrics.TotalInvoices != 3 {
		t.Fatalf("expected 3 total invoices, got %d", metrics.TotalInvoices)
	}
	if metrics.PaidCount != 1 {
		t.Fatalf("expected 1 paid invoice, got %d", metrics.PaidCount)
	}
	if metrics.TotalBilledAmount != 3500.0 {
		t.Fatalf("expected total billed 3500, got %f", metrics.TotalBilledAmount)
	}
	if metrics.TotalPaidAmount != 500.0 {
		t.Fatalf("expected total paid 500, got %f", metrics.TotalPaidAmount)
	}
	if metrics.TotalOutstandingAmount != 3000.0 {
		t.Fatalf("expected total outstanding 3000, got %f", metrics.TotalOutstandingAmount)
	}
	if metrics.AgingCurrent != 1000.0 {
		t.Fatalf("expected aging current 1000, got %f", metrics.AgingCurrent)
	}
	if metrics.Aging30to60 != 2000.0 {
		t.Fatalf("expected aging 30-60 2000, got %f", metrics.Aging30to60)
	}
}
