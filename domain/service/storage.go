package service

import (
	"context"
	"io"
	"time"
)

type StorageProvider interface {
	Put(ctx context.Context, bucket, key string, r io.Reader) error
	Get(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, key string) error
	PresignGet(ctx context.Context, bucket, key string, ttl time.Duration) (string, error)
	PresignPut(ctx context.Context, bucket, key string, ttl time.Duration) (string, error)
}
