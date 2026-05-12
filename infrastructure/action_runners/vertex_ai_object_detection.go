package action_runners

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

// VertexAIObjectDetectionRunner calls a Vertex AI online prediction endpoint
// with an image URL and returns any error from the response.
type VertexAIObjectDetectionRunner struct {
	client  *http.Client
	baseURL string
}

// NewVertexAIObjectDetectionRunner creates a runner using the provided HTTP
// client and Vertex AI base URL (e.g. "https://us-central1-aiplatform.googleapis.com").
func NewVertexAIObjectDetectionRunner(c *http.Client, baseURL string) *VertexAIObjectDetectionRunner {
	return &VertexAIObjectDetectionRunner{client: c, baseURL: baseURL}
}

// Name satisfies the action.Runner interface.
func (v *VertexAIObjectDetectionRunner) Name() string {
	return "vertex_ai_object_detection"
}

// Run reads config_data for endpoint_id and image_url, then POSTs a predict
// request to the Vertex AI endpoint. Returns an error on missing config or
// HTTP status >= 400.
func (v *VertexAIObjectDetectionRunner) Run(ctx context.Context, run *entity.ActionRun) error {
	var cfg struct {
		EndpointID string `json:"endpoint_id"`
		ImageURL   string `json:"image_url"`
	}
	if err := json.Unmarshal(run.ConfigData, &cfg); err != nil {
		return fmt.Errorf("vertex_ai_object_detection: failed to parse config_data: %w", err)
	}
	if cfg.EndpointID == "" {
		return fmt.Errorf("vertex_ai_object_detection: config_data.endpoint_id is required")
	}
	if cfg.ImageURL == "" {
		return fmt.Errorf("vertex_ai_object_detection: config_data.image_url is required")
	}

	payload := map[string]any{
		"instances": []map[string]any{
			{"image_url": cfg.ImageURL},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("vertex_ai_object_detection: failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/v1/endpoints/%s:predict", v.baseURL, cfg.EndpointID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("vertex_ai_object_detection: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("vertex_ai_object_detection: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("vertex_ai_object_detection: endpoint returned status %d", resp.StatusCode)
	}
	return nil
}
