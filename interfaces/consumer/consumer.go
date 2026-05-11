// Package consumer holds RabbitMQ consumers for the worker process.
// Each consumer implements Consumer and is launched as a goroutine
// by cmd/worker/main.go using errgroup.
package consumer

import "context"

// Consumer is the contract for any MQ consumer the worker manages.
// Start MUST block until ctx is cancelled, then close its underlying
// Channel and return cleanly.
type Consumer interface {
	Name() string
	Start(ctx context.Context) error
}
