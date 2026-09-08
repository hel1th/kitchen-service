package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:       100,
				IdleConnTimeout:    90 * time.Second,
				DisableCompression: true,
			},
		},
	}
}

func (c *Client) PatchStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	url := fmt.Sprintf("%s/api/v1/orders/%s/status", c.baseURL, orderID.String())

	payload := map[string]string{"status": status}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // not critical to check close on read

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

type Simulator struct {
	cfg    *Config
	store  *Store
	client *Client
}

func NewSimulator(cfg *Config, store *Store, client *Client) *Simulator {
	return &Simulator{
		cfg:    cfg,
		store:  store,
		client: client,
	}
}

func (s *Simulator) Run(orderID uuid.UUID) {
	time.Sleep(s.cfg.CookingDelay)
	s.notify(orderID, "cooking")

	time.Sleep(s.cfg.ReadyDelay)
	s.notify(orderID, "ready")
}

func (s *Simulator) notify(orderID uuid.UUID, status string) {
	for attempt := 0; attempt < 2; attempt++ { // 1 retry
		ctx, cancel := context.WithTimeout(context.Background(), s.cfg.CallbackTimeout)
		err := s.client.PatchStatus(ctx, orderID, status)
		cancel()
		if err == nil {
			s.store.UpdateStatus(orderID, status)
			return
		}
		log.Printf("callback failed (attempt %d): order=%s status=%s err=%v", attempt+1, orderID, status, err)
	}
}
