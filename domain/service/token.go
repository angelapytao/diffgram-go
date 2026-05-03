package service

import "context"

type TokenService interface {
	Issue(ctx context.Context, claims Claims) (string, error)
	Verify(ctx context.Context, token string) (*Claims, error)
}

type Claims struct {
	UserID int
	Email  string
}
