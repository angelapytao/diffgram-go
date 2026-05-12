package media

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"

	"github.com/sirupsen/logrus"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

var ErrQueueFull = errors.New("worker pool queue is full")

type WorkerPool struct {
	queue    chan *entity.Input
	pipeline *Pipeline
	size     int
	pending  atomic.Int64
}

func NewWorkerPool(size int, pipeline *Pipeline) *WorkerPool {
	if size <= 0 {
		size = runtime.NumCPU()
	}
	return &WorkerPool{
		queue:    make(chan *entity.Input, size*4),
		pipeline: pipeline,
		size:     size,
	}
}

func (wp *WorkerPool) Submit(input *entity.Input) error {
	select {
	case wp.queue <- input:
		wp.pending.Add(1)
		return nil
	default:
		return ErrQueueFull
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.size; i++ {
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case input := <-wp.queue:
			if err := wp.pipeline.Run(ctx, input); err != nil {
				logrus.WithError(err).WithField("input_id", input.ID).Error("pipeline run failed")
			}
			wp.pending.Add(-1)
		}
	}
}

func (wp *WorkerPool) Pending() int64 {
	return wp.pending.Load()
}

func (wp *WorkerPool) QueueLen() int {
	return len(wp.queue)
}

func (wp *WorkerPool) QueueCap() int {
	return cap(wp.queue)
}
