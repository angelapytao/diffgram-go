package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/util"
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

func TestUserService_Register_Success(t *testing.T) {
	repo := new(mockUserRepo)
	svc := &service.UserService{}
	svc.Init(repo)

	repo.On("FindByEmail", mock.Anything, "alice@example.com").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Email == "alice@example.com" && u.PasswordHash != nil
	})).Return(nil)

	user, err := svc.Register(context.Background(), "alice@example.com", "secret123")
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.NotNil(t, user.PasswordHash)
	repo.AssertExpectations(t)
}

func TestUserService_Register_DuplicateEmail(t *testing.T) {
	repo := new(mockUserRepo)
	svc := &service.UserService{}
	svc.Init(repo)

	existing := &entity.User{Email: "alice@example.com"}
	repo.On("FindByEmail", mock.Anything, "alice@example.com").Return(existing, nil)

	_, err := svc.Register(context.Background(), "alice@example.com", "secret123")
	assert.Equal(t, util.ErrInvalidInput, err)
	repo.AssertExpectations(t)
}

func TestUserService_Login_Success(t *testing.T) {
	repo := new(mockUserRepo)
	svc := &service.UserService{}
	svc.Init(repo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	hashStr := string(hash)
	repo.On("FindByEmail", mock.Anything, "alice@example.com").
		Return(&entity.User{Email: "alice@example.com", PasswordHash: &hashStr}, nil)

	user, err := svc.Login(context.Background(), "alice@example.com", "secret123")
	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", user.Email)
	repo.AssertExpectations(t)
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	repo := new(mockUserRepo)
	svc := &service.UserService{}
	svc.Init(repo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	hashStr := string(hash)
	repo.On("FindByEmail", mock.Anything, "alice@example.com").
		Return(&entity.User{Email: "alice@example.com", PasswordHash: &hashStr}, nil)

	_, err := svc.Login(context.Background(), "alice@example.com", "wrongpassword")
	assert.Equal(t, util.ErrWrongPassword, err)
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := new(mockUserRepo)
	svc := &service.UserService{}
	svc.Init(repo)

	repo.On("FindByEmail", mock.Anything, "ghost@example.com").Return(nil, nil)

	_, err := svc.Login(context.Background(), "ghost@example.com", "pass")
	assert.Equal(t, util.ErrWrongPassword, err)
}
