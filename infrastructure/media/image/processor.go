package image

import (
	"bytes"
	"context"
	"fmt"
	goimage "image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/config"
	"github.com/angelapytao/diffgram-go/domain/entity"
	domainservice "github.com/angelapytao/diffgram-go/domain/service"
)

type Processor struct {
	storage domainservice.StorageProvider
	db      *gorm.DB
	cfg     config.ProcessorConfig
	bucket  string
}

func NewProcessor(storage domainservice.StorageProvider, db *gorm.DB, cfg config.ProcessorConfig, bucket string) *Processor {
	return &Processor{storage: storage, db: db, cfg: cfg, bucket: bucket}
}

func (p *Processor) MediaType() string { return "image" }

func (p *Processor) Process(ctx context.Context, input *entity.Input) error {
	localPath := input.BlobPath
	if localPath == "" {
		localPath = filepath.Join(p.cfg.TempDir, fmt.Sprintf("input_%d", input.ID), input.OriginalFilename)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open image file: %w", err)
	}
	defer func() { _ = f.Close() }()

	img, _, err := goimage.Decode(f)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	imgRecord := &entity.Image{
		OriginalFilename: input.OriginalFilename,
		Width:            width,
		Height:           height,
	}
	if err := p.db.WithContext(ctx).Create(imgRecord).Error; err != nil {
		return fmt.Errorf("create image record: %w", err)
	}

	basePath := fmt.Sprintf("projects/%d/images/%d", input.ProjectID, imgRecord.ID)

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err := p.storage.Put(ctx, p.bucket, basePath+"/original", f); err != nil {
		return fmt.Errorf("upload original: %w", err)
	}
	imgRecord.URLSignedBlobPath = basePath + "/original"

	thumbLarge := resize.Thumbnail(uint(p.cfg.ThumbLargeSize), uint(p.cfg.ThumbLargeSize), img, resize.Lanczos3)
	var largeBuf bytes.Buffer
	if err := jpeg.Encode(&largeBuf, thumbLarge, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("encode thumb large: %w", err)
	}
	thumbLargePath := basePath + "/thumb_large"
	if err := p.storage.Put(ctx, p.bucket, thumbLargePath, &largeBuf); err != nil {
		return fmt.Errorf("upload thumb large: %w", err)
	}
	imgRecord.ThumbLargeBlobPath = thumbLargePath

	thumbSmall := resize.Thumbnail(uint(p.cfg.ThumbSmallSize), uint(p.cfg.ThumbSmallSize), img, resize.Lanczos3)
	var smallBuf bytes.Buffer
	if err := jpeg.Encode(&smallBuf, thumbSmall, &jpeg.Options{Quality: 75}); err != nil {
		return fmt.Errorf("encode thumb small: %w", err)
	}
	thumbSmallPath := basePath + "/thumb_small"
	if err := p.storage.Put(ctx, p.bucket, thumbSmallPath, &smallBuf); err != nil {
		return fmt.Errorf("upload thumb small: %w", err)
	}
	imgRecord.ThumbSmallBlobPath = thumbSmallPath

	if err := p.db.WithContext(ctx).Save(imgRecord).Error; err != nil {
		return fmt.Errorf("update image record: %w", err)
	}

	fileRecord := &entity.File{
		ProjectID:        input.ProjectID,
		Type:             "image",
		ImageID:          &imgRecord.ID,
		OriginalFilename: input.OriginalFilename,
		State:            "active",
		InputID:          &input.ID,
	}
	if err := p.db.WithContext(ctx).Create(fileRecord).Error; err != nil {
		return fmt.Errorf("create file record: %w", err)
	}

	input.FileID = &fileRecord.ID
	return nil
}
