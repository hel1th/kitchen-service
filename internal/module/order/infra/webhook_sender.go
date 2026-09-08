package infra

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/hel1th/kitchen-service/internal/module/order/domain"
	restUsecase "github.com/hel1th/kitchen-service/internal/module/restaurant/usecase"
)

type HTTPWebhookSender struct {
	client       *http.Client
	restProvider restUsecase.RestaurantProvider
}

func NewHTTPWebhookSender(restProvider restUsecase.RestaurantProvider) *HTTPWebhookSender {
	return &HTTPWebhookSender{
		client:       &http.Client{Timeout: 3 * time.Second},
		restProvider: restProvider,
	}
}

type webhookPayload struct {
	OrderID      uuid.UUID     `json:"orderId"`
	RestaurantID uuid.UUID     `json:"restaurantId"`
	Items        []webhookItem `json:"items"`
	CreatedAt    time.Time     `json:"createdAt"`
}

type webhookItem struct {
	DishID   uuid.UUID `json:"dishId"`
	Name     string    `json:"name"`
	Price    float64   `json:"price"`
	Quantity int       `json:"quantity"`
}

func (s *HTTPWebhookSender) SendOrderCreated(ctx context.Context, order *domain.Order) error {
	restaurant, err := s.restProvider.GetRestaurant(ctx, order.RestaurantID)
	if err != nil {
		return fmt.Errorf("failed to get restaurant for webhook: %w", err)
	}

	if restaurant.WebhookURL == "" {
		// No webhook configured, silently ignore
		return nil
	}

	payload := webhookPayload{
		OrderID:      order.ID,
		RestaurantID: order.RestaurantID,
		CreatedAt:    order.CreatedAt,
	}
	for _, item := range order.Items {
		payload.Items = append(payload.Items, webhookItem{
			DishID:   item.DishID,
			Name:     item.DishNameSnapshot,
			Price:    item.PriceSnapshot,
			Quantity: item.Quantity,
		})
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	return s.doRequestWithRetries(ctx, restaurant.WebhookURL, data)
}

func (s *HTTPWebhookSender) doRequestWithRetries(ctx context.Context, url string, data []byte) error {
	maxRetries := 2
	for i := 0; i <= maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to create webhook request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
		}

		if i < maxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(500 * time.Millisecond):
			}
		}
	}

	return fmt.Errorf("webhook delivery failed after retries")
}
