package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrWebhookNotFound           = errors.New("webhook subscription not found")
	ErrInvalidTargetURL          = errors.New("invalid webhook target URL, must begin with http:// or https://")
	ErrNoEventsSpecified         = errors.New("at least one valid event type must be specified")
	ErrInvalidEventType          = errors.New("one or more specified event types are not supported")
	ErrDeliveryNotFound          = errors.New("webhook delivery record not found")
	ErrUnauthorizedWebhookAccess = errors.New("unauthorized access to webhook subscription")
)

// HTTPClient represents an interface for dispatching HTTP requests (allowing tests to mock outbound calls)
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// WebhookService defines the business logic for managing webhooks, dispatching events, and delivery retries
type WebhookService interface {
	CreateSubscription(userID string, req *model.CreateWebhookRequest) (*model.WebhookSubscriptionDB, error)
	GetSubscription(id int64) (*model.WebhookSubscriptionDB, error)
	ListSubscriptions(filter model.WebhookFilter) (*model.WebhookSummary, error)
	UpdateSubscription(id int64, userID string, req *model.UpdateWebhookRequest) (*model.WebhookSubscriptionDB, error)
	DeleteSubscription(id int64, userID string) error
	DispatchEvent(event *model.WebhookEvent) ([]model.WebhookDeliveryDB, error)
	PingWebhook(subscriptionID int64, userID string) (*model.WebhookDeliveryDB, error)
	ListDeliveries(filter model.DeliveryFilter) (*model.DeliverySummary, error)
	GetDelivery(id int64) (*model.WebhookDeliveryDB, error)
	RetryDelivery(deliveryID int64) (*model.WebhookDeliveryDB, error)
	GetWebhookStats(subscriptionID int64) (*model.WebhookStats, error)
}

type webhookService struct {
	repo       repository.WebhookRepository
	httpClient HTTPClient
}

// NewWebhookService creates a new WebhookService with the given repository and optional HTTP client
func NewWebhookService(repo repository.WebhookRepository, client HTTPClient) WebhookService {
	if client == nil {
		client = http.DefaultClient
	}
	return &webhookService{
		repo:       repo,
		httpClient: client,
	}
}

func (s *webhookService) CreateSubscription(userID string, req *model.CreateWebhookRequest) (*model.WebhookSubscriptionDB, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	targetURL := strings.TrimSpace(req.TargetURL)
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		return nil, ErrInvalidTargetURL
	}

	if len(req.EventTypes) == 0 {
		return nil, ErrNoEventsSpecified
	}

	var validEvents []string
	for _, et := range req.EventTypes {
		trimmed := strings.TrimSpace(et)
		if trimmed == "" {
			continue
		}
		if !model.IsValidWebhookEventType(trimmed) {
			return nil, ErrInvalidEventType
		}
		validEvents = append(validEvents, trimmed)
	}

	if len(validEvents) == 0 {
		return nil, ErrNoEventsSpecified
	}

	secret := strings.TrimSpace(req.SecretToken)
	if secret == "" {
		generated, err := GenerateRandomSecret(DefaultSecretByteLength)
		if err != nil {
			return nil, err
		}
		secret = generated
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	sub := &model.WebhookSubscriptionDB{
		WorkspaceID: req.WorkspaceID,
		UserID:      userID,
		TargetURL:   targetURL,
		SecretToken: secret,
		EventTypes:  validEvents,
		IsActive:    isActive,
		Description: strings.TrimSpace(req.Description),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	id, err := s.repo.CreateSubscription(sub)
	if err != nil {
		return nil, err
	}
	sub.ID = id
	return sub, nil
}

func (s *webhookService) GetSubscription(id int64) (*model.WebhookSubscriptionDB, error) {
	sub, err := s.repo.GetSubscriptionByID(id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrWebhookNotFound
	}
	return sub, nil
}

func (s *webhookService) ListSubscriptions(filter model.WebhookFilter) (*model.WebhookSummary, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	subs, total, err := s.repo.ListSubscriptions(filter)
	if err != nil {
		return nil, err
	}

	return &model.WebhookSummary{
		Subscriptions: subs,
		Total:         total,
		Offset:        filter.Offset,
		Limit:         filter.Limit,
	}, nil
}

func (s *webhookService) UpdateSubscription(id int64, userID string, req *model.UpdateWebhookRequest) (*model.WebhookSubscriptionDB, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	sub, err := s.repo.GetSubscriptionByID(id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrWebhookNotFound
	}

	if userID != "" && sub.UserID != userID {
		return nil, ErrUnauthorizedWebhookAccess
	}

	if req.TargetURL != nil {
		url := strings.TrimSpace(*req.TargetURL)
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			return nil, ErrInvalidTargetURL
		}
		sub.TargetURL = url
	}

	if req.EventTypes != nil {
		if len(req.EventTypes) == 0 {
			return nil, ErrNoEventsSpecified
		}
		var validEvents []string
		for _, et := range req.EventTypes {
			trimmed := strings.TrimSpace(et)
			if trimmed == "" {
				continue
			}
			if !model.IsValidWebhookEventType(trimmed) {
				return nil, ErrInvalidEventType
			}
			validEvents = append(validEvents, trimmed)
		}
		if len(validEvents) == 0 {
			return nil, ErrNoEventsSpecified
		}
		sub.EventTypes = validEvents
	}

	if req.SecretToken != nil {
		secret := strings.TrimSpace(*req.SecretToken)
		if secret != "" {
			sub.SecretToken = secret
		}
	}

	if req.Description != nil {
		sub.Description = strings.TrimSpace(*req.Description)
	}

	if req.IsActive != nil {
		sub.IsActive = *req.IsActive
	}

	sub.UpdatedAt = time.Now()
	if err := s.repo.UpdateSubscription(sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *webhookService) DeleteSubscription(id int64, userID string) error {
	sub, err := s.repo.GetSubscriptionByID(id)
	if err != nil {
		return err
	}
	if sub == nil {
		return ErrWebhookNotFound
	}

	if userID != "" && sub.UserID != userID {
		return ErrUnauthorizedWebhookAccess
	}

	return s.repo.DeleteSubscription(id)
}

func (s *webhookService) DispatchEvent(event *model.WebhookEvent) ([]model.WebhookDeliveryDB, error) {
	if event == nil {
		return nil, errors.New("event cannot be nil")
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.EventID == "" {
		event.EventID = fmt.Sprintf("evt_%d_%s", event.Timestamp.UnixNano(), event.EventType)
	}

	tTrue := true
	filter := model.WebhookFilter{
		WorkspaceID: event.WorkspaceID,
		IsActive:    &tTrue,
		Limit:       1000,
	}

	subs, _, err := s.repo.ListSubscriptions(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions for dispatch: %w", err)
	}

	payloadBytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	var deliveries []model.WebhookDeliveryDB
	for _, sub := range subs {
		matched := false
		for _, pattern := range sub.EventTypes {
			if model.MatchesWebhookEventType(pattern, event.EventType) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		delivery := s.executeDelivery(&sub, event.EventType, payloadBytes, 1)
		deliveries = append(deliveries, *delivery)
	}

	return deliveries, nil
}

func (s *webhookService) executeDelivery(sub *model.WebhookSubscriptionDB, eventType string, payloadBytes []byte, attemptNumber int) *model.WebhookDeliveryDB {
	delivery := &model.WebhookDeliveryDB{
		SubscriptionID: sub.ID,
		EventType:      eventType,
		PayloadJSON:    string(payloadBytes),
		AttemptNumber:  attemptNumber,
		CreatedAt:      time.Now(),
	}

	req, err := http.NewRequest(http.MethodPost, sub.TargetURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		delivery.Status = model.WebhookDeliveryStatusFailed
		delivery.ErrorMessage = fmt.Sprintf("failed to create request: %v", err)
		_, _ = s.repo.CreateDelivery(delivery)
		return delivery
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "tsk-mgr-webhook-dispatcher/1.0")
	req.Header.Set("X-Webhook-Event", eventType)
	req.Header.Set("X-Webhook-Delivery", fmt.Sprintf("del_%d_%d", sub.ID, time.Now().UnixNano()))
	req.Header.Set("X-Webhook-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))

	if sub.SecretToken != "" {
		signature := ComputeWebhookSignature(payloadBytes, sub.SecretToken)
		req.Header.Set("X-Webhook-Signature-256", signature)
	}

	start := time.Now()
	resp, err := s.httpClient.Do(req)
	duration := time.Since(start).Milliseconds()
	delivery.DurationMs = duration

	if err != nil {
		delivery.Status = model.WebhookDeliveryStatusFailed
		delivery.ErrorMessage = err.Error()
		_, _ = s.repo.CreateDelivery(delivery)
		return delivery
	}
	defer resp.Body.Close()

	delivery.StatusCode = resp.StatusCode

	var headerPairs []string
	for k, v := range resp.Header {
		headerPairs = append(headerPairs, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
	}
	delivery.ResponseHeaders = strings.Join(headerPairs, "\n")

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
	delivery.ResponseBody = string(bodyBytes)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Status = model.WebhookDeliveryStatusSuccess
	} else {
		delivery.Status = model.WebhookDeliveryStatusFailed
		delivery.ErrorMessage = fmt.Sprintf("HTTP status code: %d", resp.StatusCode)
	}

	id, _ := s.repo.CreateDelivery(delivery)
	delivery.ID = id
	return delivery
}

func (s *webhookService) PingWebhook(subscriptionID int64, userID string) (*model.WebhookDeliveryDB, error) {
	sub, err := s.repo.GetSubscriptionByID(subscriptionID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrWebhookNotFound
	}

	if userID != "" && sub.UserID != userID {
		return nil, ErrUnauthorizedWebhookAccess
	}

	pingEvent := &model.WebhookEvent{
		EventID:      fmt.Sprintf("evt_ping_%d_%d", sub.ID, time.Now().UnixNano()),
		EventType:    string(model.WebhookEventPing),
		Timestamp:    time.Now(),
		WorkspaceID:  sub.WorkspaceID,
		ActorUserID:  userID,
		ResourceType: "webhook_subscription",
		ResourceID:   strconv.FormatInt(sub.ID, 10),
		Data: map[string]interface{}{
			"message":         "Ping event triggered from tsk-mgr",
			"subscription_id": sub.ID,
			"target_url":      sub.TargetURL,
			"event_types":     sub.EventTypes,
		},
	}

	payloadBytes, err := json.Marshal(pingEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ping event: %w", err)
	}

	delivery := s.executeDelivery(sub, string(model.WebhookEventPing), payloadBytes, 1)
	return delivery, nil
}

func (s *webhookService) ListDeliveries(filter model.DeliveryFilter) (*model.DeliverySummary, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	deliveries, total, err := s.repo.ListDeliveries(filter)
	if err != nil {
		return nil, err
	}

	return &model.DeliverySummary{
		Deliveries: deliveries,
		Total:      total,
		Offset:     filter.Offset,
		Limit:      filter.Limit,
	}, nil
}

func (s *webhookService) GetDelivery(id int64) (*model.WebhookDeliveryDB, error) {
	delivery, err := s.repo.GetDeliveryByID(id)
	if err != nil {
		return nil, err
	}
	if delivery == nil {
		return nil, ErrDeliveryNotFound
	}
	return delivery, nil
}

func (s *webhookService) RetryDelivery(deliveryID int64) (*model.WebhookDeliveryDB, error) {
	prevDelivery, err := s.repo.GetDeliveryByID(deliveryID)
	if err != nil {
		return nil, err
	}
	if prevDelivery == nil {
		return nil, ErrDeliveryNotFound
	}

	sub, err := s.repo.GetSubscriptionByID(prevDelivery.SubscriptionID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrWebhookNotFound
	}

	retryAttempt := prevDelivery.AttemptNumber + 1
	delivery := s.executeDelivery(sub, prevDelivery.EventType, []byte(prevDelivery.PayloadJSON), retryAttempt)
	return delivery, nil
}

func (s *webhookService) GetWebhookStats(subscriptionID int64) (*model.WebhookStats, error) {
	sub, err := s.repo.GetSubscriptionByID(subscriptionID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrWebhookNotFound
	}

	return s.repo.GetWebhookStats(subscriptionID)
}
