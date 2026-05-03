package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/interfaces/http/middleware"
)

type mockTokenSvc struct{ mock.Mock }

func (m *mockTokenSvc) Issue(ctx context.Context, claims domainservice.Claims) (string, error) {
	args := m.Called(ctx, claims)
	return args.String(0), args.Error(1)
}

func (m *mockTokenSvc) Verify(ctx context.Context, token string) (*domainservice.Claims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainservice.Claims), args.Error(1)
}

func newRouter(svc domainservice.TokenService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Auth(svc))
	r.GET("/protected", func(c *gin.Context) {
		claims, _ := c.Get("claims")
		c.JSON(http.StatusOK, gin.H{"user_id": claims.(*domainservice.Claims).UserID})
	})
	return r
}

func TestAuth_MissingHeader(t *testing.T) {
	svc := new(mockTokenSvc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	svc := new(mockTokenSvc)
	svc.On("Verify", mock.Anything, "bad-token").Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuth_ValidToken(t *testing.T) {
	svc := new(mockTokenSvc)
	svc.On("Verify", mock.Anything, "valid-token").
		Return(&domainservice.Claims{UserID: 7, Email: "x@y.com"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":7`)
}

func TestAuth_ValidCookie(t *testing.T) {
	svc := new(mockTokenSvc)
	svc.On("Verify", mock.Anything, "cookie-token").
		Return(&domainservice.Claims{UserID: 9, Email: "c@d.com"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "diffgram_jwt", Value: "cookie-token"})
	newRouter(svc).ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":9`)
}
