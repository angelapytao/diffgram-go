package action_runners_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/infrastructure/action_runners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func makeMLRun(id int64) *entity.ActionRun {
	cfg, _ := json.Marshal(map[string]any{"model": "resnet50"})
	evt, _ := json.Marshal(map[string]any{"file_id": 99})
	return &entity.ActionRun{
		ID:           id,
		ConfigData:   datatypes.JSON(cfg),
		EventPayload: datatypes.JSON(evt),
	}
}

func TestMLRunnerProxy_Name(t *testing.T) {
	p := action_runners.NewMLRunnerProxy("custom_runner", "http://localhost:9000", http.DefaultClient)
	assert.Equal(t, "custom_runner", p.Name())
}

func TestMLRunnerProxy_Success(t *testing.T) {
	var gotPath string
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	p := action_runners.NewMLRunnerProxy("ml_proxy", srv.URL, srv.Client())
	err := p.Run(context.Background(), makeMLRun(42))

	require.NoError(t, err)
	assert.Equal(t, "/run/ml_proxy", gotPath)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &decoded))
	assert.Equal(t, float64(42), decoded["action_run_id"])
}

func TestMLRunnerProxy_5xxFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	p := action_runners.NewMLRunnerProxy("ml_proxy", srv.URL, srv.Client())
	err := p.Run(context.Background(), makeMLRun(7))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "502")
}
