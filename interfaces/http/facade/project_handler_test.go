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

	appservice "github.com/angelapytao/diffgram-go/application/service"
	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/entity"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	infratoken "github.com/angelapytao/diffgram-go/infrastructure/token"
	"github.com/angelapytao/diffgram-go/interfaces/http/facade"
	"github.com/angelapytao/diffgram-go/interfaces/http/middleware"
)

type mockProjectRepo struct{ mock.Mock }

func (m *mockProjectRepo) FindByID(ctx context.Context, id int) (*entity.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}
func (m *mockProjectRepo) FindByStringID(ctx context.Context, sid string) (*entity.Project, error) {
	args := m.Called(ctx, sid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Project), args.Error(1)
}
func (m *mockProjectRepo) ListByOrgID(ctx context.Context, orgID int) ([]*entity.Project, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]*entity.Project), args.Error(1)
}
func (m *mockProjectRepo) ListByUserPrimaryID(ctx context.Context, userID int) ([]*entity.Project, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*entity.Project), args.Error(1)
}
func (m *mockProjectRepo) Create(ctx context.Context, project *entity.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}
func (m *mockProjectRepo) Save(ctx context.Context, project *entity.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

var testTokenSvc domainservice.TokenService

func setupProjectAppForHandler(t *testing.T) *mockProjectRepo {
	t.Helper()
	repo := new(mockProjectRepo)
	domainservice.GetProjectService().Init(repo)
	testTokenSvc = infratoken.NewJWTService(config.JWTConfig{Secret: "test-secret", Timeout: time.Hour})
	appservice.Init(nil, testTokenSvc, nil)
	return repo
}

func bearerToken(t *testing.T, userID int) string {
	t.Helper()
	tok, err := testTokenSvc.Issue(context.Background(), domainservice.Claims{UserID: userID, Email: "u@u.com"})
	require.NoError(t, err)
	return "Bearer " + tok
}

func newProjectRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	authed := r.Group("/", middleware.Auth(testTokenSvc))
	facade.RegisterProjectRoutes(authed)
	return r
}

func TestCreateProjectHandler_Success(t *testing.T) {
	repo := setupProjectAppForHandler(t)
	name := "My Project"
	sid := "my-project"
	repo.On("Create", mock.Anything, mock.MatchedBy(func(p *entity.Project) bool {
		return p.ProjectStringID != nil && *p.ProjectStringID == sid
	})).Return(nil)

	body, _ := json.Marshal(map[string]string{"project_name": name, "project_string_id": sid})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/project/new", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(t, 1))
	newProjectRouter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":true`)
	assert.Contains(t, w.Body.String(), sid)
}

func TestCreateProjectHandler_NoAuth(t *testing.T) {
	setupProjectAppForHandler(t)

	body, _ := json.Marshal(map[string]string{"project_name": "X", "project_string_id": "x"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/project/new", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newProjectRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListProjectsHandler_Success(t *testing.T) {
	repo := setupProjectAppForHandler(t)
	sid := "listed-project"
	repo.On("ListByUserPrimaryID", mock.Anything, 42).
		Return([]*entity.Project{{ProjectStringID: &sid}}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/project/list", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearerToken(t, 42))
	newProjectRouter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), sid)
}

func TestViewProjectHandler_Found(t *testing.T) {
	repo := setupProjectAppForHandler(t)
	sid := "view-project"
	name := "View Project"
	repo.On("FindByStringID", mock.Anything, sid).
		Return(&entity.Project{ProjectStringID: &sid, Name: &name}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/project/"+sid+"/view", nil)
	req.Header.Set("Authorization", bearerToken(t, 1))
	newProjectRouter().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), sid)
}

func TestViewProjectHandler_NotFound(t *testing.T) {
	repo := setupProjectAppForHandler(t)
	repo.On("FindByStringID", mock.Anything, "ghost").Return(nil, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/project/ghost/view", nil)
	req.Header.Set("Authorization", bearerToken(t, 1))
	newProjectRouter().ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
