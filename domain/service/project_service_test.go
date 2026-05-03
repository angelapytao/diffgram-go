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

type mockProjectRepo struct{ mock.Mock }

func (m *mockProjectRepo) FindByID(ctx context.Context, id int) (*entity.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}

func (m *mockProjectRepo) FindByStringID(ctx context.Context, sid string) (*entity.Project, error) {
	args := m.Called(ctx, sid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}

func (m *mockProjectRepo) ListByOrgID(ctx context.Context, orgID int) ([]*entity.Project, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]*entity.Project), args.Error(1)
}

func (m *mockProjectRepo) Create(ctx context.Context, project *entity.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *mockProjectRepo) Save(ctx context.Context, project *entity.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *mockProjectRepo) ListByUserPrimaryID(ctx context.Context, userID int) ([]*entity.Project, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entity.Project), args.Error(1)
}

func TestProjectService_Create(t *testing.T) {
	repo := new(mockProjectRepo)
	svc := &service.ProjectService{}
	svc.Init(repo)

	sid := "my-project"
	proj := &entity.Project{ProjectStringID: &sid}
	repo.On("Create", mock.Anything, proj).Return(nil)

	err := svc.Create(context.Background(), proj)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestProjectService_GetByStringID(t *testing.T) {
	repo := new(mockProjectRepo)
	svc := &service.ProjectService{}
	svc.Init(repo)

	sid := "my-project"
	repo.On("FindByStringID", mock.Anything, sid).
		Return(&entity.Project{ProjectStringID: &sid}, nil)

	p, err := svc.GetByStringID(context.Background(), sid)
	require.NoError(t, err)
	assert.Equal(t, sid, *p.ProjectStringID)
}

func TestProjectService_ListByOrgID_Empty(t *testing.T) {
	repo := new(mockProjectRepo)
	svc := &service.ProjectService{}
	svc.Init(repo)

	repo.On("ListByOrgID", mock.Anything, 42).Return([]*entity.Project{}, nil)

	list, err := svc.ListByOrgID(context.Background(), 42)
	require.NoError(t, err)
	assert.Empty(t, list)
}
