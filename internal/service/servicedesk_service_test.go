package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestServiceDeskSvc(t *testing.T) (ServiceDeskService, repository.ServiceDeskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}
	sqliteDB.SetMaxOpenConns(1)

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'ServiceDesk WS', 'sd-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewServiceDeskRepository(sqliteDB)
	svc := NewServiceDeskService(repo)
	return svc, repo, sqliteDB
}

func TestServiceDesk_QueueLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestServiceDeskSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	// Validation failure: missing name
	if _, err := svc.CreateQueue(ctx, "ws_1", model.CreateSDQueueRequest{Tier: model.SDTier1}); err == nil {
		t.Fatal("expected error for missing name, got nil")
	}

	// Create queue
	q, err := svc.CreateQueue(ctx, "ws_1", model.CreateSDQueueRequest{
		QueueCode:         "Q-IT-L1",
		Name:              "Tier 1 Global IT Desk",
		Description:       "Frontline omnichannel support desk",
		Tier:              model.SDTier1,
		DefaultSLAMinutes: 30,
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	if q.ID == "" || !q.IsActive {
		t.Fatalf("expected active queue with ID, got %+v", q)
	}

	// Duplicate code conflict
	if _, err := svc.CreateQueue(ctx, "ws_1", model.CreateSDQueueRequest{
		QueueCode: "Q-IT-L1",
		Name:      "Duplicate Code Queue",
	}); err != ErrSDQueueCodeConflict {
		t.Fatalf("expected ErrSDQueueCodeConflict, got %v", err)
	}

	// Get by ID and code
	gotByID, err := svc.GetQueueByID(ctx, q.ID)
	if err != nil || gotByID.Name != "Tier 1 Global IT Desk" {
		t.Fatalf("unexpected GetQueueByID result: %+v, err: %v", gotByID, err)
	}
	gotByCode, err := svc.GetQueueByCode(ctx, "ws_1", "Q-IT-L1")
	if err != nil || gotByCode.ID != q.ID {
		t.Fatalf("unexpected GetQueueByCode result: %+v, err: %v", gotByCode, err)
	}

	// Update queue
	newName := "Tier 1 Enterprise Service Desk"
	updated, err := svc.UpdateQueue(ctx, q.ID, model.UpdateSDQueueRequest{Name: &newName})
	if err != nil || updated.Name != newName {
		t.Fatalf("failed to update queue: %+v, err: %v", updated, err)
	}

	// List queues
	queues, total, err := svc.ListQueues(ctx, model.SDQueueFilter{WorkspaceID: "ws_1", Tier: model.SDTier1})
	if err != nil || total != 1 || len(queues) != 1 {
		t.Fatalf("expected 1 queue, got total %d, err %v", total, err)
	}

	// Delete queue
	if err := svc.DeleteQueue(ctx, q.ID); err != nil {
		t.Fatalf("failed to delete queue: %v", err)
	}
	if _, err := svc.GetQueueByID(ctx, q.ID); err == nil {
		t.Fatal("expected error getting deleted queue")
	}
}

func TestServiceDesk_TicketLifecycleAndFCR(t *testing.T) {
	svc, _, dbConn := setupTestServiceDeskSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	q, err := svc.CreateQueue(ctx, "ws_1", model.CreateSDQueueRequest{
		QueueCode: "Q-NET",
		Name:      "Network Support",
		Tier:      model.SDTier2,
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	// Create ticket via Slack
	ticket, err := svc.CreateTicket(ctx, "ws_1", model.CreateSDTicketRequest{
		TicketNumber: "SD-10001",
		QueueID:      q.ID,
		RequesterID:  "user_1",
		Channel:      model.SDChannelSlack,
		TicketType:   model.SDTicketTypeIncidentReport,
		Priority:     model.SDPriorityP2,
		Subject:      "VPN disconnects every 15 minutes",
		Description:  "Corporate WireGuard VPN drops on macOS 15",
	})
	if err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}
	if ticket.Status != model.SDTicketStatusTriaged {
		t.Fatalf("expected status 'triaged' when queue is set, got '%s'", ticket.Status)
	}

	// Assign agent
	assignee := "user_2"
	ticket, err = svc.UpdateTicket(ctx, ticket.ID, model.UpdateSDTicketRequest{AssigneeID: &assignee})
	if err != nil || ticket.AssigneeID != "user_2" {
		t.Fatalf("failed to assign ticket: %+v, err: %v", ticket, err)
	}

	// Agent outbound reply stamps FirstResponseAt and moves to in_progress
	interaction, err := svc.AddInteraction(ctx, "ws_1", model.AddSDInteractionRequest{
		TicketID:  ticket.ID,
		AuthorID:  "user_2",
		Channel:   model.SDChannelSlack,
		Direction: model.SDDirectionOutbound,
		Body:      "Please reset your VPN profile using KB-204.",
	})
	if err != nil || interaction.ID == "" {
		t.Fatalf("failed to add interaction: %v", err)
	}

	refreshed, err := svc.GetTicketByID(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("failed to reload ticket: %v", err)
	}
	if refreshed.FirstResponseAt == nil || refreshed.Status != model.SDTicketStatusInProgress {
		t.Fatalf("expected FirstResponseAt set and status 'in_progress', got %+v", refreshed)
	}

	// Resolve with FCR
	resolved, err := svc.ResolveTicket(ctx, ticket.ID, model.ResolveSDTicketRequest{
		FirstContactResolved: true,
		LinkedKBArticleID:    "KB-204",
		ResolutionNote:       "Profile reset resolved the drops immediately.",
	})
	if err != nil {
		t.Fatalf("failed to resolve ticket: %v", err)
	}
	if resolved.Status != model.SDTicketStatusResolved || !resolved.FirstContactResolved || resolved.ResolvedAt == nil {
		t.Fatalf("unexpected resolved ticket state: %+v", resolved)
	}

	// Verify interaction thread contains both replies
	interactions, totalInt, err := svc.ListInteractions(ctx, model.SDInteractionFilter{TicketID: ticket.ID})
	if err != nil || totalInt != 2 || len(interactions) != 2 {
		t.Fatalf("expected 2 interactions, got total %d, err %v", totalInt, err)
	}

	// Close ticket
	closed, err := svc.CloseTicket(ctx, ticket.ID)
	if err != nil || closed.Status != model.SDTicketStatusClosed {
		t.Fatalf("failed to close ticket: %+v, err: %v", closed, err)
	}
}

func TestServiceDesk_SurveysAndAnalytics(t *testing.T) {
	svc, _, dbConn := setupTestServiceDeskSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	_, _ = svc.CreateQueue(ctx, "ws_1", model.CreateSDQueueRequest{QueueCode: "Q-1", Name: "L1 Desk"})

	// Ticket 1: Resolved with FCR + CSAT 5, NPS 10
	t1, err := svc.CreateTicket(ctx, "ws_1", model.CreateSDTicketRequest{
		TicketNumber: "SD-20001",
		Channel:      model.SDChannelPortal,
		Priority:     model.SDPriorityP1,
		Subject:      "Password locked out",
	})
	if err != nil {
		t.Fatalf("failed to create t1: %v", err)
	}
	_, _ = svc.ResolveTicket(ctx, t1.ID, model.ResolveSDTicketRequest{FirstContactResolved: true})

	// Validation failure: CSAT out of range
	if _, err := svc.SubmitSurvey(ctx, "ws_1", model.SubmitCSATSurveyRequest{
		TicketID:  t1.ID,
		CSATScore: 6,
		NPSScore:  9,
	}); err == nil {
		t.Fatal("expected error for CSAT > 5")
	}

	s1, err := svc.SubmitSurvey(ctx, "ws_1", model.SubmitCSATSurveyRequest{
		TicketID:  t1.ID,
		CSATScore: 5,
		NPSScore:  10,
		Comment:   "Super fast unlock!",
	})
	if err != nil || s1.ID == "" {
		t.Fatalf("failed to submit survey 1: %v", err)
	}

	// Ticket 2: Resolved without FCR + CSAT 3, NPS 6
	t2, _ := svc.CreateTicket(ctx, "ws_1", model.CreateSDTicketRequest{
		TicketNumber: "SD-20002",
		Channel:      model.SDChannelEmail,
		Priority:     model.SDPriorityP3,
		Subject:      "Laptop provisioning request",
	})
	_, _ = svc.ResolveTicket(ctx, t2.ID, model.ResolveSDTicketRequest{FirstContactResolved: false})
	_, _ = svc.SubmitSurvey(ctx, "ws_1", model.SubmitCSATSurveyRequest{
		TicketID:  t2.ID,
		CSATScore: 3,
		NPSScore:  6,
	})

	// Ticket 3: Open ticket
	_, _ = svc.CreateTicket(ctx, "ws_1", model.CreateSDTicketRequest{
		TicketNumber: "SD-20003",
		Channel:      model.SDChannelPortal,
		Priority:     model.SDPriorityP3,
		Subject:      "Printer offline",
	})

	analytics, err := svc.GetServiceDeskAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to compute analytics: %v", err)
	}
	if analytics.TotalTickets != 3 || analytics.OpenTickets != 1 || analytics.ResolvedTickets != 2 {
		t.Fatalf("unexpected ticket counts: %+v", analytics)
	}
	if analytics.FCRRate != 50.0 {
		t.Fatalf("expected 50%% FCR rate, got %.2f%%", analytics.FCRRate)
	}
	if analytics.AvgCSATScore != 4.0 || analytics.AvgNPSScore != 8.0 || analytics.TotalSurveys != 2 {
		t.Fatalf("unexpected survey aggregates: CSAT=%.2f, NPS=%.2f, total=%d", analytics.AvgCSATScore, analytics.AvgNPSScore, analytics.TotalSurveys)
	}
	if analytics.ChannelDistribution[model.SDChannelPortal] != 2 || analytics.ChannelDistribution[model.SDChannelEmail] != 1 {
		t.Fatalf("unexpected channel distribution: %+v", analytics.ChannelDistribution)
	}
	if analytics.ServiceDeskHealthIndex <= 0 || analytics.ServiceDeskHealthIndex > 100 {
		t.Fatalf("unexpected health index: %.2f", analytics.ServiceDeskHealthIndex)
	}
}

func TestServiceDesk_Concurrency(t *testing.T) {
	svc, _, dbConn := setupTestServiceDeskSvc(t)
	defer dbConn.Close()
	ctx := context.Background()

	concurrency := 15
	var wg sync.WaitGroup
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ticket, err := svc.CreateTicket(ctx, "ws_1", model.CreateSDTicketRequest{
				TicketNumber: fmt.Sprintf("SD-CONC-%03d", idx),
				Channel:      model.SDChannelChat,
				Priority:     model.SDPriorityP3,
				Subject:      fmt.Sprintf("Concurrent support request %d", idx),
			})
			if err != nil {
				errCh <- err
				return
			}
			_, err = svc.AddInteraction(ctx, "ws_1", model.AddSDInteractionRequest{
				TicketID:  ticket.ID,
				AuthorID:  "user_2",
				Direction: model.SDDirectionOutbound,
				Body:      fmt.Sprintf("Acknowledged ticket %d", idx),
			})
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent service desk error: %v", err)
		}
	}

	tickets, total, err := svc.ListTickets(ctx, model.SDTicketFilter{WorkspaceID: "ws_1"})
	if err != nil || total != concurrency || len(tickets) != concurrency {
		t.Fatalf("expected %d tickets, got total %d, err %v", concurrency, total, err)
	}
}
