package mq_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	"github.com/angelapytao/diffgram-go/infrastructure/mq"
)

func TestClient_DialAndChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test: requires Docker")
	}
	ctx := context.Background()

	container, err := rabbitmq.Run(ctx, "rabbitmq:3.13-management")
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	url, err := container.AmqpURL(ctx)
	require.NoError(t, err)

	client, err := mq.NewClient(url)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	ch, err := client.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()
	assert.NotNil(t, ch)
}
