package consumer

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/angelapytao/diffgram-go/infrastructure/mq"
)

type SchedulerConsumer struct {
	client   *mq.Client
	prefetch int
	log      *logrus.Logger
}

func NewSchedulerConsumer(client *mq.Client, prefetch int, log *logrus.Logger) *SchedulerConsumer {
	return &SchedulerConsumer{client: client, prefetch: prefetch, log: log}
}

func (c *SchedulerConsumer) Name() string { return "scheduler" }

func (c *SchedulerConsumer) Start(ctx context.Context) error {
	ch, err := c.client.Channel()
	if err != nil {
		return err
	}
	defer func() { _ = ch.Close() }()

	if err := ch.ExchangeDeclare(mq.ExchangeScheduler, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(mq.QueueSchedulerTasks, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(mq.QueueSchedulerTasks, mq.RoutingKeySchedulerAll, mq.ExchangeScheduler, false, nil); err != nil {
		return err
	}
	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(mq.QueueSchedulerTasks, c.Name(), false, false, false, false, nil)
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
			c.log.WithField("body", string(d.Body)).Info("scheduler consumer received (skeleton)")
			_ = d.Ack(false)
		}
	}
}

var _ Consumer = (*SchedulerConsumer)(nil)
