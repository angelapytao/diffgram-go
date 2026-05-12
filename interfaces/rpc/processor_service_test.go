package rpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/infrastructure/media"
)

type mockPublisher struct{}

func (m *mockPublisher) Publish(_ context.Context, _, _ string, _ any) error { return nil }

func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Input{}))
	return db
}

func TestProcessMedia(t *testing.T) {
	db := setupDB(t)
	pub := &mockPublisher{}
	pipeline := media.NewPipeline(db, pub, map[string]media.MediaProcessor{})
	pool := media.NewWorkerPool(2, pipeline)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pool.Start(ctx)

	handler := NewProcessorServiceHandler(db, pool)

	input := &entity.Input{ProjectID: 1, MediaType: "image", Status: "pending"}
	require.NoError(t, db.Create(input).Error)

	resp, err := handler.ProcessMedia(ctx, &ProcessMediaRequest{
		ProjectID: 1,
		InputID:   input.ID,
		MediaType: "image",
	})
	require.NoError(t, err)
	require.Equal(t, "accepted", resp.Status)
}

func TestGetInput(t *testing.T) {
	db := setupDB(t)
	pool := media.NewWorkerPool(1, nil)
	handler := NewProcessorServiceHandler(db, pool)

	input := &entity.Input{ProjectID: 1, MediaType: "video", Status: "success", OriginalFilename: "test.mp4"}
	require.NoError(t, db.Create(input).Error)

	info, err := handler.GetInput(context.Background(), &GetInputRequest{InputID: input.ID})
	require.NoError(t, err)
	require.Equal(t, "video", info.MediaType)
	require.Equal(t, "success", info.Status)
	require.Equal(t, "test.mp4", info.OriginalFilename)
}

func TestGetInput_NotFound(t *testing.T) {
	db := setupDB(t)
	pool := media.NewWorkerPool(1, nil)
	handler := NewProcessorServiceHandler(db, pool)

	_, err := handler.GetInput(context.Background(), &GetInputRequest{InputID: 999})
	require.Error(t, err)
	require.Contains(t, err.Error(), "input not found")
}

func TestListInputs(t *testing.T) {
	db := setupDB(t)
	pool := media.NewWorkerPool(1, nil)
	handler := NewProcessorServiceHandler(db, pool)

	for i := 0; i < 5; i++ {
		db.Create(&entity.Input{ProjectID: 1, MediaType: "image", Status: "success"})
	}
	db.Create(&entity.Input{ProjectID: 1, MediaType: "image", Status: "failed"})

	resp, err := handler.ListInputs(context.Background(), &ListInputsRequest{
		ProjectID: 1,
		Page:      1,
		PageSize:  10,
	})
	require.NoError(t, err)
	require.Equal(t, int64(6), resp.Total)
	require.Len(t, resp.Inputs, 6)

	// With status filter
	resp, err = handler.ListInputs(context.Background(), &ListInputsRequest{
		ProjectID:    1,
		Page:         1,
		PageSize:     10,
		StatusFilter: "failed",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.Total)
	require.Len(t, resp.Inputs, 1)
}

func TestListInputs_Pagination(t *testing.T) {
	db := setupDB(t)
	pool := media.NewWorkerPool(1, nil)
	handler := NewProcessorServiceHandler(db, pool)

	for i := 0; i < 10; i++ {
		db.Create(&entity.Input{ProjectID: 2, MediaType: "text", Status: "success"})
	}

	resp, err := handler.ListInputs(context.Background(), &ListInputsRequest{
		ProjectID: 2,
		Page:      1,
		PageSize:  3,
	})
	require.NoError(t, err)
	require.Equal(t, int64(10), resp.Total)
	require.Len(t, resp.Inputs, 3)
	require.Equal(t, 1, resp.Page)
	require.Equal(t, 3, resp.PageSize)
}

func TestListInputs_DefaultPageSize(t *testing.T) {
	db := setupDB(t)
	pool := media.NewWorkerPool(1, nil)
	handler := NewProcessorServiceHandler(db, pool)

	resp, err := handler.ListInputs(context.Background(), &ListInputsRequest{
		ProjectID: 1,
		Page:      0,
		PageSize:  0,
	})
	require.NoError(t, err)
	require.Equal(t, 1, resp.Page)
	require.Equal(t, 25, resp.PageSize)
}
