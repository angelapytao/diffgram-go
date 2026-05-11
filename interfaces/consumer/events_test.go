package consumer_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	tcsetup "github.com/angelapytao/diffgram-go/infrastructure/db"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
	infrarepo "github.com/angelapytao/diffgram-go/infrastructure/repository"
	"github.com/angelapytao/diffgram-go/interfaces/consumer"
)

func startMySQLConsumer(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	container, err := tcmysql.Run(ctx, "mysql:8.0",
		tcmysql.WithDatabase("diffgram_test"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("testpass"),
	)
	require.NoError(t, err)
	dsn, err := container.ConnectionString(ctx, "charset=utf8mb4&parseTime=True&loc=Local")
	require.NoError(t, err)

	sqlDB, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	require.NoError(t, tcsetup.RunMigrations(sqlDB, migrationsDir))
	_ = sqlDB.Close()

	return dsn, func() { _ = container.Terminate(ctx) }
}

func startRabbitMQ(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	c, err := rabbitmq.Run(ctx, "rabbitmq:3.13-management")
	require.NoError(t, err)
	url, err := c.AmqpURL(ctx)
	require.NoError(t, err)
	return url, func() { _ = c.Terminate(ctx) }
}

func TestEventsConsumer_CreatesActionRunAndPublishes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test: requires Docker")
	}

	dsn, dbCleanup := startMySQLConsumer(t)
	t.Cleanup(dbCleanup)

	gormDB, err := tcsetup.NewConnection(dsn)
	require.NoError(t, err)

	tplRepo := infrarepo.NewActionTemplateRepository(gormDB)
	runRepo := infrarepo.NewActionRunRepository(gormDB)
	runSvc := domainservice.NewActionRunService(runRepo)

	et := "annotation_created"
	rn := "webhook"
	require.NoError(t, tplRepo.Create(context.Background(), &entity.ActionTemplate{EventType: &et, RunnerName: &rn}))

	mqURL, mqCleanup := startRabbitMQ(t)
	t.Cleanup(mqCleanup)

	client, err := mq.NewClient(mqURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	log := logrus.New()
	log.SetOutput(os.Stderr)
	c := consumer.NewEventsConsumer(client, 1, tplRepo, runRepo, runSvc, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = c.Start(ctx) }()
	time.Sleep(500 * time.Millisecond)

	verifyCh, err := client.Channel()
	require.NoError(t, err)
	defer verifyCh.Close()
	require.NoError(t, verifyCh.ExchangeDeclare(mq.ExchangeActions, "topic", true, false, false, false, nil))
	_, err = verifyCh.QueueDeclare(mq.QueueActionsTriggers, true, false, false, false, nil)
	require.NoError(t, err)
	require.NoError(t, verifyCh.QueueBind(mq.QueueActionsTriggers, mq.RoutingKeyActionsNewTrigger, mq.ExchangeActions, false, nil))

	deliveries, err := verifyCh.Consume(mq.QueueActionsTriggers, "verifier", false, false, false, false, nil)
	require.NoError(t, err)

	pubCh, err := client.Channel()
	require.NoError(t, err)
	defer pubCh.Close()
	body, _ := json.Marshal(map[string]interface{}{
		"event_type": "annotation_created",
		"payload":    map[string]any{"file_id": 42},
	})
	require.NoError(t, pubCh.PublishWithContext(ctx, mq.ExchangeEvents, mq.RoutingKeyEventsNew, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}))

	select {
	case msg := <-deliveries:
		var got map[string]int64
		require.NoError(t, json.Unmarshal(msg.Body, &got))
		runID := got["action_run_id"]
		assert.Greater(t, runID, int64(0))

		run, err := runRepo.FindByID(context.Background(), runID)
		require.NoError(t, err)
		require.NotNil(t, run)
		assert.Equal(t, "pending", run.Status)
		assert.Equal(t, "webhook", run.RunnerName)
		_ = msg.Ack(false)
	case <-time.After(15 * time.Second):
		t.Fatal("no actions message received within 15s")
	}
}
