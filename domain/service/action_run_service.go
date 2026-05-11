package service

import (
	"context"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainrepo "github.com/angelapytao/diffgram-go/domain/repository"
)

type ActionRunService struct {
	repo domainrepo.ActionRunRepository
}

func NewActionRunService(repo domainrepo.ActionRunRepository) *ActionRunService {
	return &ActionRunService{repo: repo}
}

func (s *ActionRunService) LoadByID(ctx context.Context, id int64) (*entity.ActionRun, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ActionRunService) MarkRunning(ctx context.Context, id int64) error {
	return s.repo.UpdateStatus(ctx, id, "running", nil)
}

func (s *ActionRunService) MarkComplete(ctx context.Context, id int64) error {
	return s.repo.UpdateStatus(ctx, id, "complete", nil)
}

func (s *ActionRunService) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	msg := errMsg
	return s.repo.UpdateStatus(ctx, id, "failed", &msg)
}

func (s *ActionRunService) Recover(ctx context.Context) (int64, error) {
	return s.repo.ResetRunningToPending(ctx)
}
