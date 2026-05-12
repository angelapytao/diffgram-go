package media

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

type mockPublisher struct{ published []any }

func (m *mockPublisher) Publish(_ context.Context, _, _ string, body any) error {
	m.published = append(m.published, body)
	return nil
}

type mockProcessor struct{ called bool }

func (m *mockProcessor) MediaType() string                                    { return "image" }
func (m *mockProcessor) Process(_ context.Context, _ *entity.Input) error { m.called = true; return nil }

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Input{}))
	return db
}

func TestPipelineRun_Success(t *testing.T) {
	db := setupTestDB(t)
	pub := &mockPublisher{}
	proc := &mockProcessor{}
	pipeline := NewPipeline(db, pub, map[string]MediaProcessor{"image": proc})

	input := &entity.Input{ProjectID: 1, MediaType: "image", Status: "pending"}
	require.NoError(t, db.Create(input).Error)

	err := pipeline.Run(context.Background(), input)
	require.NoError(t, err)
	require.True(t, proc.called)
	require.Equal(t, "success", input.Status)
	require.Len(t, pub.published, 1)
}

func TestPipelineRun_UnsupportedType(t *testing.T) {
	db := setupTestDB(t)
	pub := &mockPublisher{}
	pipeline := NewPipeline(db, pub, map[string]MediaProcessor{})

	input := &entity.Input{ProjectID: 1, MediaType: "unknown", Status: "pending"}
	require.NoError(t, db.Create(input).Error)

	err := pipeline.Run(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, "failed", input.Status)
}
