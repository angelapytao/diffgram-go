package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"

	"github.com/angelapytao/diffgram-go/domain/action"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
)

type ActionsConsumer struct {
	client   *mq.Client
	prefetch int
	registry *action.Registry
	runSvc   *domainservice.ActionRunService
	log      *logrus.Logger
}

func NewActionsConsumer(
	client *mq.Client,
	prefetch int,
	registry *action.Registry,
	runSvc *domainservice.ActionRunService,
	log *logrus.Logger,
) *ActionsConsumer {
	return &ActionsConsumer{
		client: client, prefetch: prefetch,
		registry: registry, runSvc: runSvc, log: log,
	}
}

func (c *ActionsConsumer) Name() string { return "actions" }

func (c *ActionsConsumer) Start(ctx context.Context) error {
	ch, err := c.client.Channel()
	if err != nil {
		return err
	}
	defer func() { _ = ch.Close() }()

	if err := ch.ExchangeDeclare(mq.ExchangeActions, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(mq.QueueActionsTriggers, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(mq.QueueActionsTriggers, mq.RoutingKeyActionsNewTrigger, mq.ExchangeActions, false, nil); err != nil {
		return err
	}
	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(mq.QueueActionsTriggers, c.Name(), false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return nil
			}
			c.dispatch(ctx, d)
		}
	}
}

type triggerEnvelope struct {
	ActionRunID int64 `json:"action_run_id"`
}

func (c *ActionsConsumer) dispatch(ctx context.Context, d amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			c.log.WithField("panic", fmt.Sprintf("%v", r)).Error("actions consumer panic recovered")
			_ = d.Nack(false, false)
		}
	}()

	var env triggerEnvelope
	if err := json.Unmarshal(d.Body, &env); err != nil {
		c.log.WithError(err).WithField("body", string(d.Body)).Error("invalid trigger body")
		_ = d.Nack(false, false)
		return
	}
	run, err := c.runSvc.LoadByID(ctx, env.ActionRunID)
	if err != nil || run == nil {
		c.log.WithField("action_run_id", env.ActionRunID).Error("action_run not found")
		_ = d.Nack(false, false)
		return
	}
	runner, err := c.registry.Get(run.RunnerName)
	if err != nil {
		c.log.WithError(err).WithField("runner", run.RunnerName).Error("runner not registered")
		_ = c.runSvc.MarkFailed(ctx, run.ID, err.Error())
		_ = d.Nack(false, false)
		return
	}

	if err := c.runSvc.MarkRunning(ctx, run.ID); err != nil {
		c.log.WithError(err).Error("mark running failed")
		_ = d.Nack(false, false)
		return
	}
	if err := runner.Run(ctx, run); err != nil {
		c.log.WithError(err).WithField("runner", runner.Name()).Warn("runner run failed")
		_ = c.runSvc.MarkFailed(ctx, run.ID, err.Error())
		_ = d.Nack(false, false)
		return
	}
	if err := c.runSvc.MarkComplete(ctx, run.ID); err != nil {
		c.log.WithError(err).Error("mark complete failed")
		_ = d.Nack(false, false)
		return
	}
	_ = d.Ack(false)
}

var _ Consumer = (*ActionsConsumer)(nil)
