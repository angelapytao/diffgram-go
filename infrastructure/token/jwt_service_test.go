package token_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/service"
	infratoken "github.com/angelapytao/diffgram-go/infrastructure/token"
	"github.com/angelapytao/diffgram-go/util"
)

func newSvc(timeout time.Duration) service.TokenService {
	return infratoken.NewJWTService(config.JWTConfig{
		Secret:  "test-secret-key",
		Timeout: timeout,
	})
}

func TestJWTService_IssueAndVerify(t *testing.T) {
	svc := newSvc(time.Hour)
	ctx := context.Background()

	token, err := svc.Issue(ctx, service.Claims{UserID: 42, Email: "bob@example.com"})
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := svc.Verify(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, 42, claims.UserID)
	assert.Equal(t, "bob@example.com", claims.Email)
}

func TestJWTService_Expired(t *testing.T) {
	svc := newSvc(-time.Second)
	ctx := context.Background()

	token, err := svc.Issue(ctx, service.Claims{UserID: 1, Email: "x@y.com"})
	require.NoError(t, err)

	_, err = svc.Verify(ctx, token)
	assert.Equal(t, util.ErrUnauthorized, err)
}

func TestJWTService_InvalidToken(t *testing.T) {
	svc := newSvc(time.Hour)
	ctx := context.Background()

	_, err := svc.Verify(ctx, "not.a.valid.token")
	assert.Equal(t, util.ErrUnauthorized, err)
}

func TestJWTService_WrongSecret(t *testing.T) {
	svcA := newSvc(time.Hour)
	svcB := infratoken.NewJWTService(config.JWTConfig{Secret: "other-secret", Timeout: time.Hour})
	ctx := context.Background()

	token, err := svcA.Issue(ctx, service.Claims{UserID: 1})
	require.NoError(t, err)

	_, err = svcB.Verify(ctx, token)
	assert.Equal(t, util.ErrUnauthorized, err)
}
