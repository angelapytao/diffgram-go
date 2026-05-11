package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/domain/service"
)

type mockActionRunRepo struct{ mock.Mock }

func (m *mockActionRunRepo) Create(ctx context.Context, run *entity.ActionRun) error {
	return m.Called(ctx, run).Error(0)
}
func (m *mockActionRunRepo) FindByID(ctx context.Context, id int64) (*entity.ActionRun, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.ActionRun), args.Error(1)
}
func (m *mockActionRunRepo) UpdateStatus(ctx context.Context, id int64, status string, errMsg *string) error {
	return m.Called(ctx, id, status, errMsg).Error(0)
}
func (m *mockActionRunRepo) ResetRunningToPending(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func TestActionRunService_MarkRunning(t *testing.T) {
	repo := new(mockActionRunRepo)
	svc := service.NewActionRunService(repo)
	ctx := context.Background()

	repo.On("UpdateStatus", ctx, int64(7), "running", (*string)(nil)).Return(nil)
	require.NoError(t, svc.MarkRunning(ctx, 7))
	repo.AssertExpectations(t)
}

func TestActionRunService_MarkComplete(t *testing.T) {
	repo := new(mockActionRunRepo)
	svc := service.NewActionRunService(repo)
	ctx := context.Background()

	repo.On("UpdateStatus", ctx, int64(7), "complete", (*string)(nil)).Return(nil)
	require.NoError(t, svc.MarkComplete(ctx, 7))
}

func TestActionRunService_MarkFailed(t *testing.T) {
	repo := new(mockActionRunRepo)
	svc := service.NewActionRunService(repo)
	ctx := context.Background()

	expectedMsg := "timeout"
	repo.On("UpdateStatus", ctx, int64(7), "failed", &expectedMsg).Return(nil)
	require.NoError(t, svc.MarkFailed(ctx, 7, "timeout"))
}

func TestActionRunService_Recover(t *testing.T) {
	repo := new(mockActionRunRepo)
	svc := service.NewActionRunService(repo)
	ctx := context.Background()

	repo.On("ResetRunningToPending", ctx).Return(int64(3), nil)
	count, err := svc.Recover(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestActionRunService_LoadByID(t *testing.T) {
	repo := new(mockActionRunRepo)
	svc := service.NewActionRunService(repo)
	ctx := context.Background()

	expected := &entity.ActionRun{ID: 7, RunnerName: "webhook"}
	repo.On("FindByID", ctx, int64(7)).Return(expected, nil)
	got, err := svc.LoadByID(ctx, 7)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
