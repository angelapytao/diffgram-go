package service

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/angelapytao/diffgram-go/domain/entity"
	domainrepo "github.com/angelapytao/diffgram-go/domain/repository"
	"github.com/angelapytao/diffgram-go/util"
)

type UserService struct {
	repo domainrepo.UserRepository
}

var userService UserService

func GetUserService() *UserService { return &userService }

func (s *UserService) Init(repo domainrepo.UserRepository) {
	s.repo = repo
}

func (s *UserService) Register(ctx context.Context, email, password string) (*entity.User, error) {
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, util.ErrInvalidInput
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, util.ErrInternal
	}
	hashStr := string(hash)
	user := &entity.User{Email: email, PasswordHash: &hashStr}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*entity.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil || user.PasswordHash == nil {
		return nil, util.ErrWrongPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return nil, util.ErrWrongPassword
	}
	return user, nil
}

func (s *UserService) GetByID(ctx context.Context, id int) (*entity.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *UserService) Save(ctx context.Context, user *entity.User) error {
	return s.repo.Save(ctx, user)
}
