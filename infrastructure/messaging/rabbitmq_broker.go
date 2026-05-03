package messaging

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	domainservice "github.com/angelapytao/diffgram-go/domain/service"
)

type rabbitBroker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	mu   sync.Mutex
}

func NewRabbitMQBroker(url string) (domainservice.MessageBroker, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &rabbitBroker{conn: conn, ch: ch}, nil
}

func (b *rabbitBroker) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (b *rabbitBroker) Subscribe(ctx context.Context, queue string, handler func([]byte) error) error {
	_, err := b.ch.QueueDeclare(queue, false, false, false, false, nil)
	if err != nil {
		return err
	}
	msgs, err := b.ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				if err := handler(msg.Body); err != nil {
					_ = msg.Nack(false, false)
				} else {
					_ = msg.Ack(false)
				}
			}
		}
	}()
	return nil
}

func (b *rabbitBroker) Close() error {
	if err := b.ch.Close(); err != nil {
		return err
	}
	return b.conn.Close()
}
