package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainrepo "github.com/angelapytao/diffgram-go/domain/repository"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
)

type EventsConsumer struct {
	client   *mq.Client
	prefetch int
	tplRepo  domainrepo.ActionTemplateRepository
	runRepo  domainrepo.ActionRunRepository
	runSvc   *domainservice.ActionRunService
	log      *logrus.Logger
}

func NewEventsConsumer(
	client *mq.Client,
	prefetch int,
	tplRepo domainrepo.ActionTemplateRepository,
	runRepo domainrepo.ActionRunRepository,
	runSvc *domainservice.ActionRunService,
	log *logrus.Logger,
) *EventsConsumer {
	return &EventsConsumer{
		client: client, prefetch: prefetch,
		tplRepo: tplRepo, runRepo: runRepo, runSvc: runSvc,
		log: log,
	}
}

func (c *EventsConsumer) Name() string { return "events" }

func (c *EventsConsumer) Start(ctx context.Context) error {
	ch, err := c.client.Channel()
	if err != nil {
		return err
	}
	defer func() { _ = ch.Close() }()

	if err := ch.ExchangeDeclare(mq.ExchangeEvents, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(mq.ExchangeActions, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(mq.QueueEventNew, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(mq.QueueEventNew, mq.RoutingKeyEventsNew, mq.ExchangeEvents, false, nil); err != nil {
		return err
	}
	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(mq.QueueEventNew, c.Name(), false, false, false, false, nil)
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
			if err := c.handle(ctx, ch, d); err != nil {
				c.log.WithError(err).WithField("body", string(d.Body)).Error("events consumer handle failed")
				_ = d.Nack(false, false)
				continue
			}
			_ = d.Ack(false)
		}
	}
}

type eventEnvelope struct {
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
}

func (c *EventsConsumer) handle(ctx context.Context, ch *amqp.Channel, d amqp.Delivery) error {
	var env eventEnvelope
	if err := json.Unmarshal(d.Body, &env); err != nil {
		return fmt.Errorf("invalid event body: %w", err)
	}
	templates, err := c.tplRepo.ListByEventType(ctx, env.EventType)
	if err != nil {
		return fmt.Errorf("list templates: %w", err)
	}
	if len(templates) == 0 {
		c.log.WithField("event_type", env.EventType).Debug("no matching templates")
		return nil
	}

	payloadJSON, _ := json.Marshal(env.Payload)

	for _, tpl := range templates {
		runnerName := ""
		if tpl.RunnerName != nil {
			runnerName = *tpl.RunnerName
		}
		tplID := tpl.ID
		run := &entity.ActionRun{
			ActionTemplateID: &tplID,
			ProjectID:        tpl.ProjectID,
			RunnerName:       runnerName,
			Status:           "pending",
			ConfigData:       tpl.ConfigData,
			EventPayload:     payloadJSON,
		}
		if err := c.runRepo.Create(ctx, run); err != nil {
			return fmt.Errorf("create action_run: %w", err)
		}
		body, _ := json.Marshal(map[string]int64{"action_run_id": run.ID})
		if err := ch.PublishWithContext(ctx, mq.ExchangeActions, mq.RoutingKeyActionsNewTrigger,
			false, false, amqp.Publishing{ContentType: "application/json", Body: body}); err != nil {
			return fmt.Errorf("publish actions: %w", err)
		}
	}
	return nil
}

var _ Consumer = (*EventsConsumer)(nil)
