package repository

import (
	"context"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

type ProjectRepository interface {
	FindByID(ctx context.Context, id int) (*entity.Project, error)
	FindByStringID(ctx context.Context, stringID string) (*entity.Project, error)
	ListByOrgID(ctx context.Context, orgID int) ([]*entity.Project, error)
	ListByUserPrimaryID(ctx context.Context, userID int) ([]*entity.Project, error)
	Create(ctx context.Context, project *entity.Project) error
	Save(ctx context.Context, project *entity.Project) error
}
