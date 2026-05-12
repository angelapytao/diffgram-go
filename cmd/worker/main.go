package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/action"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	diffdb "github.com/angelapytao/diffgram-go/infrastructure/db"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
	infrarepo "github.com/angelapytao/diffgram-go/infrastructure/repository"
	"github.com/angelapytao/diffgram-go/interfaces/consumer"
	"github.com/angelapytao/diffgram-go/interfaces/http/health"

	"github.com/angelapytao/diffgram-go/infrastructure/action_runners"
)

type stubTaskCreator struct{ log *logrus.Logger }

func (s *stubTaskCreator) CreateTasks(_ context.Context, req action_runners.TaskCreateRequest) error {
	s.log.WithField("req", req).Info("task_template runner: would create tasks (P4 stub)")
	return nil
}

type rabbitPublisher struct{ client *mq.Client }

func (p *rabbitPublisher) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	ch, err := p.client.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	return ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func main() {
	cfg := config.Load()

	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetOutput(os.Stdout)
	if cfg.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(logrus.DebugLevel)
	}

	gormDB, err := diffdb.NewConnection(cfg.DBDsn)
	if err != nil {
		log.WithError(err).Fatal("worker: failed to connect to database")
	}

	mqClient, err := mq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		log.WithError(err).Fatal("worker: failed to connect to RabbitMQ")
	}
	defer mqClient.Close()

	// PLACEHOLDER_MAIN_BODY

	runRepo := infrarepo.NewActionRunRepository(gormDB)
	tplRepo := infrarepo.NewActionTemplateRepository(gormDB)
	runSvc := domainservice.NewActionRunService(runRepo)
	if n, err := runSvc.Recover(context.Background()); err != nil {
		log.WithError(err).Fatal("worker: recovery failed")
	} else if n > 0 {
		log.WithField("count", n).Info("worker: reset stale running ActionRuns to pending")
	}

	registry := action.NewRegistry()
	httpClient := &http.Client{Timeout: 30 * time.Second}
	publisher := &rabbitPublisher{client: mqClient}

	registry.Register(action_runners.NewWebhookRunner(httpClient))
	registry.Register(action_runners.NewExportRunner(publisher))
	registry.Register(action_runners.NewTaskTemplateRunner(&stubTaskCreator{log: log}))
	registry.Register(action_runners.NewVertexAIObjectDetectionRunner(httpClient, cfg.GCP.VertexBaseURL))
	registry.Register(action_runners.NewVertexAITrainDatasetRunner(httpClient, cfg.GCP.VertexBaseURL, cfg.GCP.ProjectID, cfg.GCP.Region))
	registry.Register(action_runners.NewMLRunnerProxy("deepcheck_image_property_outliers", cfg.MLRunner.HTTPAddr, httpClient))

	consumers := []consumer.Consumer{
		consumer.NewEventsConsumer(mqClient, cfg.Worker.EventsPrefetch, tplRepo, runRepo, runSvc, log),
		consumer.NewActionsConsumer(mqClient, cfg.Worker.ActionsPrefetch, registry, runSvc, log),
		consumer.NewJobsConsumer(mqClient, cfg.Worker.JobsPrefetch, log),
		consumer.NewSchedulerConsumer(mqClient, cfg.Worker.SchedulerPrefetch, log),
	}

	r := gin.New()
	r.Use(gin.Recovery())
	health.RegisterWorkerHealth(r, gormDB, mqClient)
	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Worker.HTTPPort), Handler: r}

	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	g, gCtx := errgroup.WithContext(rootCtx)
	for _, c := range consumers {
		c := c
		g.Go(func() error {
			log.WithField("consumer", c.Name()).Info("worker: consumer starting")
			return c.Start(gCtx)
		})
	}
	g.Go(func() error {
		log.WithField("addr", srv.Addr).Info("worker: http health server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-gCtx.Done()
		shutdownCtx, sc := context.WithTimeout(context.Background(), 10*time.Second)
		defer sc()
		return srv.Shutdown(shutdownCtx)
	})

	if err := g.Wait(); err != nil && err != context.Canceled {
		log.WithError(err).Error("worker: exited with error")
		os.Exit(1)
	}
	log.Info("worker: shutdown complete")
}
