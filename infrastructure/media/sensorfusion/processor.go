package sensorfusion

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

func (p *Processor) MediaType() string { return "sensor_fusion" }

type Manifest struct {
	Files []ManifestFile `json:"files"`
}

type ManifestFile struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
}

func (p *Processor) Process(ctx context.Context, input *entity.Input) error {
	localPath := input.BlobPath
	if localPath == "" {
		localPath = filepath.Join(p.cfg.TempDir, fmt.Sprintf("input_%d", input.ID), input.OriginalFilename)
	}

	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse manifest: %w", err)
	}

	baseDir := filepath.Dir(localPath)
	baseBlobPath := fmt.Sprintf("projects/%d/sensor_fusion/%d", input.ProjectID, input.ID)

	var firstFileID *int64
	for _, mf := range manifest.Files {
		filePath := filepath.Join(baseDir, mf.Filename)
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("open %s: %w", mf.Filename, err)
		}

		blobKey := fmt.Sprintf("%s/%s", baseBlobPath, mf.Filename)
		if err := p.storage.Put(ctx, p.bucket, blobKey, f); err != nil {
			_ = f.Close()
			return fmt.Errorf("upload %s: %w", mf.Filename, err)
		}
		_ = f.Close()

		fileRecord := &entity.File{
			ProjectID:        input.ProjectID,
			Type:             mf.Type,
			OriginalFilename: mf.Filename,
			State:            "active",
			InputID:          &input.ID,
		}
		if err := p.db.WithContext(ctx).Create(fileRecord).Error; err != nil {
			return fmt.Errorf("create file record for %s: %w", mf.Filename, err)
		}

		if firstFileID == nil {
			firstFileID = &fileRecord.ID
		}
	}

	if firstFileID != nil {
		input.FileID = firstFileID
	}
	return nil
}
