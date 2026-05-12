package action_runners_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/angelapytao/diffgram-go/infrastructure/action_runners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

func makeRun(configData, eventPayload map[string]any) *entity.ActionRun {
	cfgBytes, _ := json.Marshal(configData)
	evtBytes, _ := json.Marshal(eventPayload)
	return &entity.ActionRun{
		ConfigData:   datatypes.JSON(cfgBytes),
		EventPayload: datatypes.JSON(evtBytes),
	}
}

func TestWebhookRunner_Name(t *testing.T) {
	r := action_runners.NewWebhookRunner(http.DefaultClient)
	assert.Equal(t, "webhook", r.Name())
}

func TestWebhookRunner_Success(t *testing.T) {
	var gotBody []byte
	var gotMethod string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	payload := map[string]any{"event": "label.created", "id": 42}
	run := makeRun(map[string]any{"url": srv.URL}, payload)

	r := action_runners.NewWebhookRunner(srv.Client())
	err := r.Run(context.Background(), run)

	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &decoded))
	assert.Equal(t, "label.created", decoded["event"])
}

func TestWebhookRunner_5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	run := makeRun(map[string]any{"url": srv.URL}, map[string]any{"x": 1})

	r := action_runners.NewWebhookRunner(srv.Client())
	err := r.Run(context.Background(), run)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestWebhookRunner_NoURL(t *testing.T) {
	run := makeRun(map[string]any{}, map[string]any{"x": 1})

	r := action_runners.NewWebhookRunner(http.DefaultClient)
	err := r.Run(context.Background(), run)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "url")
}
