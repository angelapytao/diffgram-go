package action_runners

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

// MLRunnerProxy forwards ActionRun execution to an external ML service via HTTP.
// Multiple instances can be registered with different names to route to different
// ML endpoints while sharing the same proxy logic.
type MLRunnerProxy struct {
	name    string
	baseURL string
	client  *http.Client
}

// NewMLRunnerProxy creates an MLRunnerProxy that POSTs to {baseURL}/run/{name}.
func NewMLRunnerProxy(name, baseURL string, httpClient *http.Client) *MLRunnerProxy {
	return &MLRunnerProxy{name: name, baseURL: baseURL, client: httpClient}
}

// Name satisfies the action.Runner interface and returns the configured runner name.
func (m *MLRunnerProxy) Name() string {
	return m.name
}

// Run marshals the ActionRun fields and POSTs them to the remote ML service.
// Returns an error if the request fails or the response status is >= 400.
func (m *MLRunnerProxy) Run(ctx context.Context, run *entity.ActionRun) error {
	payload := struct {
		ActionRunID  int64           `json:"action_run_id"`
		ConfigData   json.RawMessage `json:"config_data"`
		EventPayload json.RawMessage `json:"event_payload"`
	}{
		ActionRunID:  run.ID,
		ConfigData:   json.RawMessage(run.ConfigData),
		EventPayload: json.RawMessage(run.EventPayload),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ml_runner_proxy: failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/run/%s", m.baseURL, m.name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ml_runner_proxy: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("ml_runner_proxy: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ml_runner_proxy: remote returned status %d", resp.StatusCode)
	}
	return nil
}
