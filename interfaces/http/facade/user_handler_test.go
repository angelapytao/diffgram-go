package facade_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	appservice "github.com/angelapytao/diffgram-go/application/service"
	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/entity"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/interfaces/http/facade"
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

func setupUserAppForHandler(t *testing.T) (repo *mockUserRepo) {
	t.Helper()
	repo = new(mockUserRepo)
	domainservice.GetUserService().Init(repo)
	tokenSvc := infratoken.NewJWTService(config.JWTConfig{Secret: "test-secret", Timeout: time.Hour})
	appservice.Init(nil, tokenSvc, nil)
	return repo
}

func newUserRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	facade.RegisterUserRoutes(r)
	return r
}

func TestRegisterHandler_Success(t *testing.T) {
	repo := setupUserAppForHandler(t)
	repo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, nil)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(u *entity.User) bool {
		return u.Email == "new@example.com"
	})).Return(nil)

	body, _ := json.Marshal(map[string]string{"email": "new@example.com", "password": "secret1234"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/user/new", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newUserRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), `"new@example.com"`)
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	repo := setupUserAppForHandler(t)
	repo.On("FindByEmail", mock.Anything, "dup@example.com").
		Return(&entity.User{Email: "dup@example.com"}, nil)

	body, _ := json.Marshal(map[string]string{"email": "dup@example.com", "password": "secret1234"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/user/new", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newUserRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), `"success":false`)
}

func TestLoginHandler_Success(t *testing.T) {
	repo := setupUserAppForHandler(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret1234"), bcrypt.DefaultCost)
	hashStr := string(hash)
	repo.On("FindByEmail", mock.Anything, "alice@example.com").
		Return(&entity.User{Email: "alice@example.com", PasswordHash: &hashStr}, nil)

	body, _ := json.Marshal(map[string]string{"email": "alice@example.com", "password": "secret1234", "mode": "password"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newUserRouter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	cookieHeader := w.Header().Get("Set-Cookie")
	assert.Contains(t, cookieHeader, "diffgram_jwt=")
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	repo := setupUserAppForHandler(t)
	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	hashStr := string(hash)
	repo.On("FindByEmail", mock.Anything, "alice@example.com").
		Return(&entity.User{Email: "alice@example.com", PasswordHash: &hashStr}, nil)

	body, _ := json.Marshal(map[string]string{"email": "alice@example.com", "password": "wrongpass", "mode": "password"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newUserRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginHandler_InvalidMode(t *testing.T) {
	setupUserAppForHandler(t)

	body, _ := json.Marshal(map[string]string{"email": "alice@example.com", "password": "secret1234", "mode": "magic"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newUserRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
