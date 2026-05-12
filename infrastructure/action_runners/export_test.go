package action_runners_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/angelapytao/diffgram-go/infrastructure/action_runners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

type stubPublisher struct {
	exchange   string
	routingKey string
	body       []byte
}

func (s *stubPublisher) Publish(_ context.Context, exchange, routingKey string, body []byte) error {
	s.exchange = exchange
	s.routingKey = routingKey
	s.body = body
	return nil
}

func ptr(i int) *int { return &i }

func TestExportRunner_Name(t *testing.T) {
	r := action_runners.NewExportRunner(&stubPublisher{})
	assert.Equal(t, "export", r.Name())
}

func TestExportRunner_PublishesNotification(t *testing.T) {
	pub := &stubPublisher{}
	r := action_runners.NewExportRunner(pub)

	cfgBytes, _ := json.Marshal(map[string]any{"format": "coco"})
	evtBytes, _ := json.Marshal(map[string]any{"file_id": 99})

	run := &entity.ActionRun{
		ID:           42,
		ProjectID:    ptr(7),
		ConfigData:   datatypes.JSON(cfgBytes),
		EventPayload: datatypes.JSON(evtBytes),
	}

	err := r.Run(context.Background(), run)
	require.NoError(t, err)

	assert.Equal(t, "exports", pub.exchange)
	assert.Equal(t, "exports.created", pub.routingKey)

	var notification map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(pub.body, &notification))

	var actionRunID int64
	require.NoError(t, json.Unmarshal(notification["action_run_id"], &actionRunID))
	assert.Equal(t, int64(42), actionRunID)

	var projectID int
	require.NoError(t, json.Unmarshal(notification["project_id"], &projectID))
	assert.Equal(t, 7, projectID)
}
