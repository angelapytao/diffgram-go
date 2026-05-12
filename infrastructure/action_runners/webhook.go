package action_runners

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

// WebhookRunner posts the ActionRun's event_payload as JSON to the URL
// specified in config_data.url.
type WebhookRunner struct {
	client *http.Client
}

// NewWebhookRunner creates a WebhookRunner using the provided HTTP client.
func NewWebhookRunner(httpClient *http.Client) *WebhookRunner {
	return &WebhookRunner{client: httpClient}
}

// Name satisfies the action.Runner interface.
func (w *WebhookRunner) Name() string {
	return "webhook"
}

// Run reads config_data.url, then POSTs event_payload as JSON to that URL.
// Returns an error if the URL is missing, the request fails, or the response
// status code is >= 400.
func (w *WebhookRunner) Run(ctx context.Context, run *entity.ActionRun) error {
	// Unmarshal config_data to extract the target URL.
	var cfg struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(run.ConfigData, &cfg); err != nil {
		return fmt.Errorf("webhook: failed to parse config_data: %w", err)
	}
	if cfg.URL == "" {
		return fmt.Errorf("webhook: config_data.url is required but was not provided")
	}

	// Encode event_payload as the request body.
	body, err := json.Marshal(run.EventPayload)
	if err != nil {
		return fmt.Errorf("webhook: failed to marshal event_payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook: target returned status %d", resp.StatusCode)
	}
	return nil
}
