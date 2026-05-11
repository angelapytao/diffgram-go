package repository

import (
	"context"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

type ActionRunRepository interface {
	Create(ctx context.Context, run *entity.ActionRun) error
	FindByID(ctx context.Context, id int64) (*entity.ActionRun, error)
	UpdateStatus(ctx context.Context, id int64, status string, errMsg *string) error
	ResetRunningToPending(ctx context.Context) (int64, error)
}
