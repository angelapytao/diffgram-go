package action_runners_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/angelapytao/diffgram-go/infrastructure/action_runners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

func makeVertexRun(configData map[string]any) *entity.ActionRun {
	cfgBytes, _ := json.Marshal(configData)
	return &entity.ActionRun{
		ConfigData:   datatypes.JSON(cfgBytes),
		EventPayload: datatypes.JSON([]byte(`{}`)),
	}
}

func TestVertexAIObjectDetectionRunner_Name(t *testing.T) {
	r := action_runners.NewVertexAIObjectDetectionRunner(http.DefaultClient, "http://localhost")
	assert.Equal(t, "vertex_ai_object_detection", r.Name())
}

func TestVertexAIObjectDetectionRunner_Success(t *testing.T) {
	var gotMethod string
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"predictions": [{"label": "cat", "score": 0.95}]}`))
	}))
	defer srv.Close()

	run := makeVertexRun(map[string]any{
		"endpoint_id": "my-endpoint-123",
		"image_url":   "https://example.com/image.jpg",
	})

	r := action_runners.NewVertexAIObjectDetectionRunner(srv.Client(), srv.URL)
	err := r.Run(context.Background(), run)

	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/v1/endpoints/my-endpoint-123:predict", gotPath)
}

func TestVertexAIObjectDetectionRunner_5xxFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	run := makeVertexRun(map[string]any{
		"endpoint_id": "ep-1",
		"image_url":   "https://example.com/img.jpg",
	})

	r := action_runners.NewVertexAIObjectDetectionRunner(srv.Client(), srv.URL)
	err := r.Run(context.Background(), run)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}
