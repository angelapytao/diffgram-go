package action_runners_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/infrastructure/action_runners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func makeTrainDatasetRun(configData map[string]any) *entity.ActionRun {
	cfgBytes, _ := json.Marshal(configData)
	return &entity.ActionRun{
		ConfigData:   datatypes.JSON(cfgBytes),
		EventPayload: datatypes.JSON([]byte(`{}`)),
	}
}

func TestVertexAITrainDatasetRunner_Name(t *testing.T) {
	r := action_runners.NewVertexAITrainDatasetRunner(http.DefaultClient, "http://localhost", "proj-1", "us-central1")
	assert.Equal(t, "vertex_ai_train_dataset", r.Name())
}

func TestVertexAITrainDatasetRunner_Success(t *testing.T) {
	var gotMethod string
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name": "projects/proj-1/locations/us-central1/operations/123"}`))
	}))
	defer srv.Close()

	run := makeTrainDatasetRun(map[string]any{
		"dataset_id": "ds-42",
		"import_uri": "gs://my-bucket/data.jsonl",
	})

	r := action_runners.NewVertexAITrainDatasetRunner(srv.Client(), srv.URL, "proj-1", "us-central1")
	err := r.Run(context.Background(), run)

	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/v1/projects/proj-1/locations/us-central1/datasets/ds-42:import", gotPath)
}
