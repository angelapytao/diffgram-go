package consumer

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/angelapytao/diffgram-go/infrastructure/mq"
)

type JobsConsumer struct {
	client   *mq.Client
	prefetch int
	log      *logrus.Logger
}

func NewJobsConsumer(client *mq.Client, prefetch int, log *logrus.Logger) *JobsConsumer {
	return &JobsConsumer{client: client, prefetch: prefetch, log: log}
}

func (c *JobsConsumer) Name() string { return "jobs" }

func (c *JobsConsumer) Start(ctx context.Context) error {
	ch, err := c.client.Channel()
	if err != nil {
		return err
	}
	defer func() { _ = ch.Close() }()

	if err := ch.ExchangeDeclare(mq.ExchangeJobs, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(mq.QueueJobTasks, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(mq.QueueJobTasks, mq.RoutingKeyJobsAddTask, mq.ExchangeJobs, false, nil); err != nil {
		return err
	}
	if err := ch.Qos(c.prefetch, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(mq.QueueJobTasks, c.Name(), false, false, false, false, nil)
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
			c.log.WithField("body", string(d.Body)).Info("jobs consumer received (skeleton)")
			_ = d.Ack(false)
		}
	}
}

var _ Consumer = (*JobsConsumer)(nil)
