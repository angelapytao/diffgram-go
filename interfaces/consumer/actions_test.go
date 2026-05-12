package consumer_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/angelapytao/diffgram-go/domain/action"
	"github.com/angelapytao/diffgram-go/domain/entity"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	tcsetup "github.com/angelapytao/diffgram-go/infrastructure/db"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
	infrarepo "github.com/angelapytao/diffgram-go/infrastructure/repository"
	"github.com/angelapytao/diffgram-go/interfaces/consumer"
)

type recordingRunner struct {
	mu      sync.Mutex
	called  []*entity.ActionRun
	failNow bool
}

func (r *recordingRunner) Name() string { return "test_runner" }
func (r *recordingRunner) Run(_ context.Context, run *entity.ActionRun) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.called = append(r.called, run)
	if r.failNow {
		return errors.New("intentional failure")
	}
	return nil
}

func TestActionsConsumer_RunsRunnerAndMarksComplete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test: requires Docker")
	}

	dsn, dbCleanup := startMySQLConsumer(t)
	t.Cleanup(dbCleanup)
	gormDB, err := tcsetup.NewConnection(dsn)
	require.NoError(t, err)
	runRepo := infrarepo.NewActionRunRepository(gormDB)
	runSvc := domainservice.NewActionRunService(runRepo)

	mqURL, mqCleanup := startRabbitMQ(t)
	t.Cleanup(mqCleanup)
	client, err := mq.NewClient(mqURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	registry := action.NewRegistry()
	rec := &recordingRunner{}
	registry.Register(rec)

	log := logrus.New()
	log.SetOutput(os.Stderr)
	c := consumer.NewActionsConsumer(client, 1, registry, runSvc, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = c.Start(ctx) }()
	time.Sleep(500 * time.Millisecond)

	run := &entity.ActionRun{RunnerName: "test_runner", Status: "pending"}
	require.NoError(t, runRepo.Create(context.Background(), run))

	pubCh, err := client.Channel()
	require.NoError(t, err)
	defer func() { _ = pubCh.Close() }()
	body, _ := json.Marshal(map[string]int64{"action_run_id": run.ID})
	require.NoError(t, pubCh.PublishWithContext(ctx, mq.ExchangeActions, mq.RoutingKeyActionsNewTrigger,
		false, false, amqp.Publishing{ContentType: "application/json", Body: body}))

	require.Eventually(t, func() bool {
		got, _ := runRepo.FindByID(context.Background(), run.ID)
		return got != nil && got.Status == "complete"
	}, 15*time.Second, 200*time.Millisecond, "ActionRun must reach status=complete")

	rec.mu.Lock()
	defer rec.mu.Unlock()
	assert.Len(t, rec.called, 1)
}

func TestActionsConsumer_MarksFailedOnRunnerError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test: requires Docker")
	}
	dsn, dbCleanup := startMySQLConsumer(t)
	t.Cleanup(dbCleanup)
	gormDB, err := tcsetup.NewConnection(dsn)
	require.NoError(t, err)
	runRepo := infrarepo.NewActionRunRepository(gormDB)
	runSvc := domainservice.NewActionRunService(runRepo)

	mqURL, mqCleanup := startRabbitMQ(t)
	t.Cleanup(mqCleanup)
	client, err := mq.NewClient(mqURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })

	registry := action.NewRegistry()
	registry.Register(&recordingRunner{failNow: true})

	c := consumer.NewActionsConsumer(client, 1, registry, runSvc, logrus.New())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = c.Start(ctx) }()
	time.Sleep(500 * time.Millisecond)

	run := &entity.ActionRun{RunnerName: "test_runner", Status: "pending"}
	require.NoError(t, runRepo.Create(context.Background(), run))

	pubCh, err := client.Channel()
	require.NoError(t, err)
	defer func() { _ = pubCh.Close() }()
	body, _ := json.Marshal(map[string]int64{"action_run_id": run.ID})
	require.NoError(t, pubCh.PublishWithContext(ctx, mq.ExchangeActions, mq.RoutingKeyActionsNewTrigger,
		false, false, amqp.Publishing{ContentType: "application/json", Body: body}))

	require.Eventually(t, func() bool {
		got, _ := runRepo.FindByID(context.Background(), run.ID)
		return got != nil && got.Status == "failed"
	}, 15*time.Second, 200*time.Millisecond)
}
