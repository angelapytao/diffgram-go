package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainrepo "github.com/angelapytao/diffgram-go/domain/repository"
)

type gormActionRunRepo struct {
	db *gorm.DB
}

func NewActionRunRepository(db *gorm.DB) domainrepo.ActionRunRepository {
	return &gormActionRunRepo{db: db}
}

func (r *gormActionRunRepo) Create(ctx context.Context, run *entity.ActionRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *gormActionRunRepo) FindByID(ctx context.Context, id int64) (*entity.ActionRun, error) {
	var run entity.ActionRun
	if err := r.db.WithContext(ctx).First(&run, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &run, nil
}

func (r *gormActionRunRepo) UpdateStatus(ctx context.Context, id int64, status string, errMsg *string) error {
	updates := map[string]interface{}{"status": status}
	if errMsg != nil {
		updates["error_message"] = *errMsg
	}
	return r.db.WithContext(ctx).Model(&entity.ActionRun{}).Where("id = ?", id).Updates(updates).Error
}

func (r *gormActionRunRepo) ResetRunningToPending(ctx context.Context) (int64, error) {
	res := r.db.WithContext(ctx).Model(&entity.ActionRun{}).
		Where("status = ?", "running").
		Update("status", "pending")
	return res.RowsAffected, res.Error
}
