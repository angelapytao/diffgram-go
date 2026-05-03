package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/angelapytao/diffgram-go/application/service"
	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/entity"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/interfaces/http/dto"
	infratoken "github.com/angelapytao/diffgram-go/infrastructure/token"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) FindByID(ctx context.Context, id int) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Save(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func setupUserApp(t *testing.T) {
	t.Helper()
	repo := new(mockUserRepo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	hashStr := string(hash)

	repo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, nil).Once()
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Email == "new@example.com"
	})).Return(nil).Once()

	repo.On("FindByEmail", mock.Anything, "alice@example.com").
		Return(&entity.User{Email: "alice@example.com", PasswordHash: &hashStr}, nil)

	userSvc := domainservice.GetUserService()
	userSvc.Init(repo)

	tokenSvc := infratoken.NewJWTService(config.JWTConfig{Secret: "test-secret", Timeout: time.Hour})
	service.Init(nil, tokenSvc, nil)
}

func TestUserApp_Register(t *testing.T) {
	setupUserApp(t)

	resp, err := service.GetUserApp().Register(context.Background(), &dto.RegisterReq{
		Email: "new@example.com", Password: "secret123",
	})
	require.NoError(t, err)
	assert.Equal(t, "new@example.com", resp.Email)
}

func TestUserApp_Login(t *testing.T) {
	setupUserApp(t)

	resp, err := service.GetUserApp().Login(context.Background(), &dto.LoginReq{
		Email: "alice@example.com", Password: "secret123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "alice@example.com", resp.Email)
}
