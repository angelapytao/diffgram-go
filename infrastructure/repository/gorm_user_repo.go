package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainrepo "github.com/angelapytao/diffgram-go/domain/repository"
)

type gormUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domainrepo.UserRepository {
	return &gormUserRepo{db: db}
}

func (r *gormUserRepo) FindByID(ctx context.Context, id int) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).First(&u, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *gormUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *gormUserRepo) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormUserRepo) Save(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}
