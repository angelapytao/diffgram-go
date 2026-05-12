package mq

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublisherMarshal(t *testing.T) {
	type event struct {
		InputID int64  `json:"input_id"`
		Status  string `json:"status"`
	}
	e := event{InputID: 42, Status: "success"}
	data, err := json.Marshal(e)
	require.NoError(t, err)
	require.Contains(t, string(data), `"input_id":42`)
}
