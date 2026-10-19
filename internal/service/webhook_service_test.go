package service_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(`{"status":"ok"}`)),
	}, nil
}

func setupWebhookTestDB(t *testing.T) service.WebhookService {
	t.Helper()
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize in-memory database: %v", err)
	}

	repo := repository.NewWebhookRepository(sqliteDB)
	return service.NewWebhookService(repo, nil)
}

func TestWebhookSubscriptionLifecycle(t *testing.T) {
	svc := setupWebhookTestDB(t)
	user := "user_sub_test"

	// 1. Invalid target URL
	_, err := svc.CreateSubscription(user, &model.CreateWebhookRequest{
		WorkspaceID: 1,
		TargetURL:   "ftp://example.com/webhook",
		EventTypes:  []string{string(model.WebhookEventTaskCreated)},
	})
	if err != service.ErrInvalidTargetURL {
		t.Fatalf("expected ErrInvalidTargetURL, got: %v", err)
	}

	// 2. Empty event types
	_, err = svc.CreateSubscription(user, &model.CreateWebhookRequest{
		WorkspaceID: 1,
		TargetURL:   "https://example.com/webhook",
		EventTypes:  []string{},
	})
	if err != service.ErrNoEventsSpecified {
		t.Fatalf("expected ErrNoEventsSpecified, got: %v", err)
	}

	// 3. Invalid event type
	_, err = svc.CreateSubscription(user, &model.CreateWebhookRequest{
		WorkspaceID: 1,
		TargetURL:   "https://example.com/webhook",
		EventTypes:  []string{"invalid.event.name"},
	})
	if err != service.ErrInvalidEventType {
		t.Fatalf("expected ErrInvalidEventType, got: %v", err)
	}

	// 4. Create successful subscription with auto-generated secret
	created, err := svc.CreateSubscription(user, &model.CreateWebhookRequest{
		WorkspaceID: 1,
		TargetURL:   "https://api.example.com/webhook",
		EventTypes:  []string{string(model.WebhookEventTaskCreated), string(model.WebhookEventTaskCompleted)},
		Description: "Production task events",
	})
	if err != nil {
		t.Fatalf("failed to create valid subscription: %v", err)
	}
	if created.ID == 0 || created.SecretToken == "" {
		t.Fatalf("expected ID > 0 and non-empty secret token, got ID %d, token %s", created.ID, created.SecretToken)
	}

	// Verify HMAC-SHA256 signature utility
	payload := []byte(`{"test":"payload"}`)
	sig := service.ComputeWebhookSignature(payload, created.SecretToken)
	if !service.VerifyWebhookSignature(payload, created.SecretToken, sig) {
		t.Fatalf("signature verification failed")
	}
	if service.VerifyWebhookSignature(payload, "wrong_secret", sig) {
		t.Fatalf("signature verification should have failed with wrong secret")
	}

	// 5. Get subscription
	fetched, err := svc.GetSubscription(created.ID)
	if err != nil {
		t.Fatalf("failed to get subscription: %v", err)
	}
	if fetched.TargetURL != "https://api.example.com/webhook" || len(fetched.EventTypes) != 2 {
		t.Fatalf("unexpected fetched subscription: %+v", fetched)
	}

	// 6. List subscriptions
	summary, err := svc.ListSubscriptions(model.WebhookFilter{WorkspaceID: 1})
	if err != nil || summary.Total != 1 {
		t.Fatalf("expected 1 subscription, got total %d, err %v", summary.Total, err)
	}

	// 7. Update subscription
	newURL := "https://api.example.com/webhook/v2"
	fFalse := false
	updated, err := svc.UpdateSubscription(created.ID, user, &model.UpdateWebhookRequest{
		TargetURL: &newURL,
		IsActive:  &fFalse,
	})
	if err != nil {
		t.Fatalf("failed to update subscription: %v", err)
	}
	if updated.TargetURL != newURL || updated.IsActive {
		t.Fatalf("subscription updates not applied: %+v", updated)
	}

	// 8. Delete subscription
	if err := svc.DeleteSubscription(created.ID, user); err != nil {
		t.Fatalf("failed to delete subscription: %v", err)
	}

	_, err = svc.GetSubscription(created.ID)
	if err != service.ErrWebhookNotFound {
		t.Fatalf("expected ErrWebhookNotFound after deletion, got %v", err)
	}
}

func TestWebhookDispatchAndDeliveryTracking(t *testing.T) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize in-memory database: %v", err)
	}
	repo := repository.NewWebhookRepository(sqliteDB)

	var lastHeaders http.Header
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			lastHeaders = req.Header.Clone()
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewBufferString(`{"received":true}`)),
			}, nil
		},
	}

	svc := service.NewWebhookService(repo, mockClient)
	user := "user_dispatch"

	// Sub 1: task.created
	sub1, err := svc.CreateSubscription(user, &model.CreateWebhookRequest{
		WorkspaceID: 10,
		TargetURL:   "https://receiver1.test/webhook",
		EventTypes:  []string{string(model.WebhookEventTaskCreated)},
	})
	if err != nil {
		t.Fatalf("failed to create sub 1: %v", err)
	}

	// Sub 2: wildcard
	sub2, err := svc.CreateSubscription(user, &model.CreateWebhookRequest{
		WorkspaceID: 10,
		TargetURL:   "https://receiver2.test/webhook",
		EventTypes:  []string{string(model.WebhookEventWildcard)},
	})
	if err != nil {
		t.Fatalf("failed to create sub 2: %v", err)
	}
	if sub2.ID == 0 {
		t.Fatalf("expected sub2 ID > 0")
	}

	// Sub 3: task.completed (should not match task.created)
	_, err = svc.CreateSubscription(user, &model.CreateWebhookRequest{
		WorkspaceID: 10,
		TargetURL:   "https://receiver3.test/webhook",
		EventTypes:  []string{string(model.WebhookEventTaskCompleted)},
	})
	if err != nil {
		t.Fatalf("failed to create sub 3: %v", err)
	}

	// Dispatch task.created
	deliveries, err := svc.DispatchEvent(&model.WebhookEvent{
		EventType:   string(model.WebhookEventTaskCreated),
		WorkspaceID: 10,
		Data:        map[string]string{"title": "Deploy v2"},
	})
	if err != nil {
		t.Fatalf("failed to dispatch event: %v", err)
	}

	if len(deliveries) != 2 {
		t.Fatalf("expected exactly 2 matching deliveries (sub1 and sub2), got %d", len(deliveries))
	}

	// Verify headers from dispatch
	if lastHeaders.Get("X-Webhook-Event") != string(model.WebhookEventTaskCreated) {
		t.Fatalf("expected X-Webhook-Event header to be task.created, got %s", lastHeaders.Get("X-Webhook-Event"))
	}
	if lastHeaders.Get("X-Webhook-Signature-256") == "" {
		t.Fatalf("expected X-Webhook-Signature-256 header to be present")
	}

	// Test Ping
	pingDelivery, err := svc.PingWebhook(sub1.ID, user)
	if err != nil {
		t.Fatalf("ping failed: %v", err)
	}
	if pingDelivery.Status != model.WebhookDeliveryStatusSuccess || pingDelivery.EventType != string(model.WebhookEventPing) {
		t.Fatalf("unexpected ping delivery: %+v", pingDelivery)
	}

	// List deliveries for sub1
	delivSummary, err := svc.ListDeliveries(model.DeliveryFilter{SubscriptionID: sub1.ID})
	if err != nil || delivSummary.Total != 2 {
		t.Fatalf("expected 2 deliveries for sub1 (1 task.created + 1 ping), got %d", delivSummary.Total)
	}

	// Retry delivery
	retried, err := svc.RetryDelivery(delivSummary.Deliveries[0].ID)
	if err != nil {
		t.Fatalf("retry failed: %v", err)
	}
	if retried.AttemptNumber != 2 {
		t.Fatalf("expected attempt number 2 on retry, got %d", retried.AttemptNumber)
	}

	// Check Stats
	stats, err := svc.GetWebhookStats(sub1.ID)
	if err != nil {
		t.Fatalf("get stats failed: %v", err)
	}
	if stats.TotalDeliveries < 2 || stats.SuccessfulDeliveries < 2 || stats.SuccessRate != 100.0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
