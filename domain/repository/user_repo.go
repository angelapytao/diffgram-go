package repository

import (
	"context"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

type UserRepository interface {
	FindByID(ctx context.Context, id int) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Save(ctx context.Context, user *entity.User) error
}
