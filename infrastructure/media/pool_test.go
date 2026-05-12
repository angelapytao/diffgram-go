package media

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

func TestWorkerPool_SubmitAndProcess(t *testing.T) {
	db := setupTestDB(t)
	pub := &mockPublisher{}
	proc := &mockProcessor{}
	pipeline := NewPipeline(db, pub, map[string]MediaProcessor{"image": proc})
	pool := NewWorkerPool(2, pipeline)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	input := &entity.Input{ProjectID: 1, MediaType: "image", Status: "pending"}
	require.NoError(t, db.Create(input).Error)

	err := pool.Submit(input)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
	require.True(t, proc.called)
}

func TestWorkerPool_QueueFull(t *testing.T) {
	db := setupTestDB(t)
	pub := &mockPublisher{}
	pipeline := NewPipeline(db, pub, map[string]MediaProcessor{})
	pool := NewWorkerPool(1, pipeline)

	for i := 0; i < 4; i++ {
		err := pool.Submit(&entity.Input{ID: int64(i + 1), ProjectID: 1, MediaType: "image"})
		require.NoError(t, err)
	}
	err := pool.Submit(&entity.Input{ID: 5, ProjectID: 1, MediaType: "image"})
	require.ErrorIs(t, err, ErrQueueFull)
}
