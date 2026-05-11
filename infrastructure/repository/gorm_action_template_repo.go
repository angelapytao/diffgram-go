package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainrepo "github.com/angelapytao/diffgram-go/domain/repository"
)

type gormActionTemplateRepo struct {
	db *gorm.DB
}

func NewActionTemplateRepository(db *gorm.DB) domainrepo.ActionTemplateRepository {
	return &gormActionTemplateRepo{db: db}
}

func (r *gormActionTemplateRepo) Create(ctx context.Context, t *entity.ActionTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *gormActionTemplateRepo) ListByEventType(ctx context.Context, eventType string) ([]*entity.ActionTemplate, error) {
	var list []*entity.ActionTemplate
	if err := r.db.WithContext(ctx).Where("event_type = ?", eventType).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
