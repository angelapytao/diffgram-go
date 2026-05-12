package action_runners

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

// VertexAITrainDatasetRunner imports data into a Vertex AI dataset.
type VertexAITrainDatasetRunner struct {
	client    *http.Client
	baseURL   string
	projectID string
	region    string
}

// NewVertexAITrainDatasetRunner creates a runner using the provided HTTP client,
// Vertex AI base URL, GCP project ID, and region.
func NewVertexAITrainDatasetRunner(c *http.Client, baseURL, projectID, region string) *VertexAITrainDatasetRunner {
	return &VertexAITrainDatasetRunner{
		client:    c,
		baseURL:   baseURL,
		projectID: projectID,
		region:    region,
	}
}

// Name satisfies the action.Runner interface.
func (v *VertexAITrainDatasetRunner) Name() string {
	return "vertex_ai_train_dataset"
}

// Run reads config_data for dataset_id and import_uri, then POSTs an import
// request to the Vertex AI dataset. Returns an error on missing config or
// HTTP status >= 400.
func (v *VertexAITrainDatasetRunner) Run(ctx context.Context, run *entity.ActionRun) error {
	var cfg struct {
		DatasetID string `json:"dataset_id"`
		ImportURI string `json:"import_uri"`
	}
	if err := json.Unmarshal(run.ConfigData, &cfg); err != nil {
		return fmt.Errorf("vertex_ai_train_dataset: failed to parse config_data: %w", err)
	}
	if cfg.DatasetID == "" {
		return fmt.Errorf("vertex_ai_train_dataset: config_data.dataset_id is required")
	}
	if cfg.ImportURI == "" {
		return fmt.Errorf("vertex_ai_train_dataset: config_data.import_uri is required")
	}

	payload := map[string]any{
		"importConfigs": []map[string]any{
			{"gcsSource": cfg.ImportURI},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("vertex_ai_train_dataset: failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/v1/projects/%s/locations/%s/datasets/%s:import",
		v.baseURL, v.projectID, v.region, cfg.DatasetID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("vertex_ai_train_dataset: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("vertex_ai_train_dataset: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("vertex_ai_train_dataset: endpoint returned status %d", resp.StatusCode)
	}
	return nil
}
