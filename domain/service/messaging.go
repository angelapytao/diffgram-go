package service

import "context"

type MessageBroker interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
	Subscribe(ctx context.Context, queue string, handler func([]byte) error) error
	Close() error
}
