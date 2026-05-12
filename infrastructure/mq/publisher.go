package mq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Publisher abstracts event publishing so callers can be tested with a mock.
type Publisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body any) error
}

// AMQPPublisher implements Publisher over a single AMQP channel.
type AMQPPublisher struct {
	ch *amqp.Channel
}

func NewPublisher(client *Client) (*AMQPPublisher, error) {
	ch, err := client.Channel()
	if err != nil {
		return nil, err
	}
	return &AMQPPublisher{ch: ch}, nil
}

func (p *AMQPPublisher) Publish(ctx context.Context, exchange, routingKey string, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
}

func (p *AMQPPublisher) Close() error {
	if p.ch == nil {
		return nil
	}
	return p.ch.Close()
}
