package messaging_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/angelapytao/diffgram-go/infrastructure/messaging"
)

func startRabbitMQ(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:3-management",
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor:   wait.ForLog("Server startup complete").WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5672")
	require.NoError(t, err)
	return fmt.Sprintf("amqp://guest:guest@%s:%s/", host, port.Port())
}

func TestRabbitMQBroker_PublishSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test: requires Docker")
	}

	url := startRabbitMQ(t)
	ctx := context.Background()

	broker, err := messaging.NewRabbitMQBroker(url)
	require.NoError(t, err)
	defer func() { _ = broker.Close() }()

	const queue = "test.queue"

	received := make(chan []byte, 1)
	err = broker.Subscribe(ctx, queue, func(body []byte) error {
		received <- body
		return nil
	})
	require.NoError(t, err)

	payload := []byte(`{"event":"test"}`)
	err = broker.Publish(ctx, "", queue, payload)
	require.NoError(t, err)

	select {
	case got := <-received:
		assert.Equal(t, payload, got)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}
