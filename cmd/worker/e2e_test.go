package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	"golang.org/x/sync/errgroup"

	"github.com/angelapytao/diffgram-go/domain/action"
	"github.com/angelapytao/diffgram-go/domain/entity"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	diffdb "github.com/angelapytao/diffgram-go/infrastructure/db"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
	infrarepo "github.com/angelapytao/diffgram-go/infrastructure/repository"
	"github.com/angelapytao/diffgram-go/infrastructure/action_runners"
	"github.com/angelapytao/diffgram-go/interfaces/consumer"
)

func TestE2E_WorkerEventToCompleteRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e: requires Docker")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)

	// 1. Start MySQL container and run migrations.
	mysqlC, err := tcmysql.Run(ctx, "mysql:8.0",
		tcmysql.WithDatabase("diffgram_test"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("testpass"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = mysqlC.Terminate(context.Background()) })

	dsn, err := mysqlC.ConnectionString(ctx, "charset=utf8mb4&parseTime=True&loc=Local")
	require.NoError(t, err)

	sqlDB, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer sqlDB.Close()

	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	require.NoError(t, diffdb.RunMigrations(sqlDB, migrationsDir))

	gormDB, err := diffdb.NewConnection(dsn)
	require.NoError(t, err)

	// 2. Start RabbitMQ container.
	rmqC, err := rabbitmq.Run(ctx, "rabbitmq:3.13-management")
	require.NoError(t, err)
	t.Cleanup(func() { _ = rmqC.Terminate(context.Background()) })

	rmqURL, err := rmqC.AmqpURL(ctx)
	require.NoError(t, err)

	mqClient, err := mq.NewClient(rmqURL)
	require.NoError(t, err)
	defer mqClient.Close()

	// 3. Seed ActionTemplate.
	runRepo := infrarepo.NewActionRunRepository(gormDB)
	tplRepo := infrarepo.NewActionTemplateRepository(gormDB)

	webhookCalled := make(chan struct{}, 1)
	webhookSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case webhookCalled <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer webhookSrv.Close()

	eventType := "annotation_created"
	runnerName := "webhook"
	configJSON, _ := json.Marshal(map[string]string{"url": webhookSrv.URL})
	tpl := &entity.ActionTemplate{
		EventType:  &eventType,
		RunnerName: &runnerName,
		ConfigData: configJSON,
	}
	require.NoError(t, gormDB.Create(tpl).Error)

	// 4. Build worker assembly.
	runSvc := domainservice.NewActionRunService(runRepo)
	registry := action.NewRegistry()
	httpClient := &http.Client{Timeout: 10 * time.Second}
	registry.Register(action_runners.NewWebhookRunner(httpClient))

	eventsConsumer := consumer.NewEventsConsumer(mqClient, 1, tplRepo, runRepo, runSvc, log)
	actionsConsumer := consumer.NewActionsConsumer(mqClient, 1, registry, runSvc, log)

	// 5. Start consumers.
	consumerCtx, consumerCancel := context.WithCancel(ctx)
	defer consumerCancel()

	g, gCtx := errgroup.WithContext(consumerCtx)
	_ = gCtx
	g.Go(func() error { return eventsConsumer.Start(consumerCtx) })
	g.Go(func() error { return actionsConsumer.Start(consumerCtx) })

	// Give consumers time to bind queues.
	time.Sleep(500 * time.Millisecond)

	// 6. Publish event message.
	pubCh, err := mqClient.Channel()
	require.NoError(t, err)
	defer pubCh.Close()

	err = pubCh.ExchangeDeclare(mq.ExchangeEvents, "topic", true, false, false, false, nil)
	require.NoError(t, err)

	eventBody, _ := json.Marshal(map[string]interface{}{
		"event_type": "annotation_created",
		"payload":    map[string]int{"file_id": 7},
	})
	err = pubCh.PublishWithContext(ctx, mq.ExchangeEvents, mq.RoutingKeyEventsNew,
		false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        eventBody,
		})
	require.NoError(t, err)

	// 7. Assert webhook is called.
	select {
	case <-webhookCalled:
		// success
	case <-time.After(15 * time.Second):
		t.Fatal("webhook was not called within 15s")
	}

	// 8. Assert ActionRun reaches status="complete".
	require.Eventually(t, func() bool {
		var run entity.ActionRun
		if err := gormDB.Where("runner_name = ?", "webhook").First(&run).Error; err != nil {
			return false
		}
		return run.Status == "complete"
	}, 10*time.Second, 200*time.Millisecond, "ActionRun did not reach complete status")

	// 9. Shutdown consumers.
	consumerCancel()
	_ = g.Wait()
}
