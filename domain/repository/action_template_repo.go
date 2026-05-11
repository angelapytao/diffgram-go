package repository

import (
	"context"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

type ActionTemplateRepository interface {
	Create(ctx context.Context, t *entity.ActionTemplate) error
	ListByEventType(ctx context.Context, eventType string) ([]*entity.ActionTemplate, error)
}
