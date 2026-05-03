package service

import (
	"context"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainrepo "github.com/angelapytao/diffgram-go/domain/repository"
)

type ProjectService struct {
	repo domainrepo.ProjectRepository
}

var projectService ProjectService

func GetProjectService() *ProjectService { return &projectService }

func (s *ProjectService) Init(repo domainrepo.ProjectRepository) {
	s.repo = repo
}

func (s *ProjectService) Create(ctx context.Context, project *entity.Project) error {
	return s.repo.Create(ctx, project)
}

func (s *ProjectService) GetByID(ctx context.Context, id int) (*entity.Project, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ProjectService) GetByStringID(ctx context.Context, stringID string) (*entity.Project, error) {
	return s.repo.FindByStringID(ctx, stringID)
}

func (s *ProjectService) ListByOrgID(ctx context.Context, orgID int) ([]*entity.Project, error) {
	return s.repo.ListByOrgID(ctx, orgID)
}

func (s *ProjectService) Save(ctx context.Context, project *entity.Project) error {
	return s.repo.Save(ctx, project)
}
