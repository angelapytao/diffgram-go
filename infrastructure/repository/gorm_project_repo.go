package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainrepo "github.com/angelapytao/diffgram-go/domain/repository"
)

type gormProjectRepo struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) domainrepo.ProjectRepository {
	return &gormProjectRepo{db: db}
}

func (r *gormProjectRepo) FindByID(ctx context.Context, id int) (*entity.Project, error) {
	var p entity.Project
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *gormProjectRepo) FindByStringID(ctx context.Context, stringID string) (*entity.Project, error) {
	var p entity.Project
	if err := r.db.WithContext(ctx).
		Where("project_string_id = ?", stringID).
		First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *gormProjectRepo) ListByOrgID(ctx context.Context, orgID int) ([]*entity.Project, error) {
	var projects []*entity.Project
	if err := r.db.WithContext(ctx).Where("org_id = ?", orgID).Find(&projects).Error; err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *gormProjectRepo) Create(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *gormProjectRepo) Save(ctx context.Context, project *entity.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}
