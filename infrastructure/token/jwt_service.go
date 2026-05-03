package token

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/angelapytao/diffgram-go/config"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
	"github.com/angelapytao/diffgram-go/util"
)

type jwtService struct {
	secret  []byte
	timeout time.Duration
}

func NewJWTService(cfg config.JWTConfig) domainservice.TokenService {
	return &jwtService{
		secret:  []byte(cfg.Secret),
		timeout: cfg.Timeout,
	}
}

type jwtClaims struct {
	jwt.RegisteredClaims
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

func (s *jwtService) Issue(_ context.Context, claims domainservice.Claims) (string, error) {
	now := time.Now()
	c := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.timeout)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID: claims.UserID,
		Email:  claims.Email,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(s.secret)
}

func (s *jwtService) Verify(_ context.Context, tokenStr string) (*domainservice.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, util.ErrUnauthorized
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, util.ErrUnauthorized
	}
	c, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, util.ErrUnauthorized
	}
	return &domainservice.Claims{UserID: c.UserID, Email: c.Email}, nil
}
