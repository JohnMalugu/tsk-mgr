package service

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrWebhookNotFound          = errors.New("webhook subscription not found")
	ErrInvalidTargetURL         = errors.New("invalid webhook target URL, must begin with http:// or https://")
	ErrNoEventsSpecified        = errors.New("at least one valid event type must be specified")
	ErrInvalidEventType         = errors.New("one or more specified event types are not supported")
	ErrDeliveryNotFound         = errors.New("webhook delivery record not found")
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
	return nil, nil
}

func (s *webhookService) ListSubscriptions(filter model.WebhookFilter) (*model.WebhookSummary, error) {
	return nil, nil
}

func (s *webhookService) UpdateSubscription(id int64, userID string, req *model.UpdateWebhookRequest) (*model.WebhookSubscriptionDB, error) {
	return nil, nil
}

func (s *webhookService) DeleteSubscription(id int64, userID string) error {
	return nil
}

func (s *webhookService) DispatchEvent(event *model.WebhookEvent) ([]model.WebhookDeliveryDB, error) {
	return nil, nil
}

func (s *webhookService) PingWebhook(subscriptionID int64, userID string) (*model.WebhookDeliveryDB, error) {
	return nil, nil
}

func (s *webhookService) ListDeliveries(filter model.DeliveryFilter) (*model.DeliverySummary, error) {
	return nil, nil
}

func (s *webhookService) GetDelivery(id int64) (*model.WebhookDeliveryDB, error) {
	return nil, nil
}

func (s *webhookService) RetryDelivery(deliveryID int64) (*model.WebhookDeliveryDB, error) {
	return nil, nil
}

func (s *webhookService) GetWebhookStats(subscriptionID int64) (*model.WebhookStats, error) {
	return nil, nil
}
