package storage

import (
	"context"
	"io"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"

	"github.com/angelapytao/diffgram-go/config"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
)

type tosProvider struct {
	client *tos.ClientV2
}

func NewTOSProvider(cfg config.TOSConfig) (domainservice.StorageProvider, error) {
	credential := tos.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey)
	client, err := tos.NewClientV2(cfg.Host,
		tos.WithCredentials(credential),
		tos.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, err
	}
	return &tosProvider{client: client}, nil
}

func (p *tosProvider) Put(ctx context.Context, bucket, key string, r io.Reader) error {
	_, err := p.client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             r,
	})
	return err
}

func (p *tosProvider) Get(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	output, err := p.client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	if err != nil {
		return nil, err
	}
	return output.Content, nil
}

func (p *tosProvider) Delete(ctx context.Context, bucket, key string) error {
	_, err := p.client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	return err
}

func (p *tosProvider) PresignGet(_ context.Context, bucket, key string, ttl time.Duration) (string, error) {
	output, err := p.client.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: enum.HttpMethodGet,
		Bucket:     bucket,
		Key:        key,
		Expires:    int64(ttl.Seconds()),
	})
	if err != nil {
		return "", err
	}
	return output.SignedUrl, nil
}

func (p *tosProvider) PresignPut(_ context.Context, bucket, key string, ttl time.Duration) (string, error) {
	output, err := p.client.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: enum.HttpMethodPut,
		Bucket:     bucket,
		Key:        key,
		Expires:    int64(ttl.Seconds()),
	})
	if err != nil {
		return "", err
	}
	return output.SignedUrl, nil
}
