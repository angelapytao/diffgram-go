package geotiff

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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

func (p *Processor) MediaType() string { return "geo_tiff" }

type GDALInfo struct {
	DriverShortName string `json:"driverShortName"`
	Size            []int  `json:"size"`
}

func (p *Processor) Process(ctx context.Context, input *entity.Input) error {
	localPath := input.BlobPath
	if localPath == "" {
		localPath = filepath.Join(p.cfg.TempDir, fmt.Sprintf("input_%d", input.ID), input.OriginalFilename)
	}

	info, err := p.gdalinfo(ctx, localPath)
	if err != nil {
		return fmt.Errorf("gdalinfo: %w", err)
	}
	_ = info // metadata available for future use

	baseBlobPath := fmt.Sprintf("projects/%d/geotiff/%d", input.ProjectID, input.ID)

	// Upload original
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open geotiff: %w", err)
	}
	if err := p.storage.Put(ctx, p.bucket, baseBlobPath+"/original", f); err != nil {
		_ = f.Close()
		return fmt.Errorf("upload original: %w", err)
	}
	_ = f.Close()

	// Generate preview PNG
	previewPath := filepath.Join(p.cfg.TempDir, fmt.Sprintf("geotiff_preview_%d.png", input.ID))
	defer func() { _ = os.Remove(previewPath) }()

	if err := p.generatePreview(ctx, localPath, previewPath); err == nil {
		pf, err := os.Open(previewPath)
		if err == nil {
			_ = p.storage.Put(ctx, p.bucket, baseBlobPath+"/preview.png", pf)
			_ = pf.Close()
		}
	}

	fileRecord := &entity.File{
		ProjectID:        input.ProjectID,
		Type:             "geo_tiff",
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

func (p *Processor) gdalinfo(ctx context.Context, path string) (*GDALInfo, error) {
	cmd := exec.CommandContext(ctx, p.cfg.GDALInfoPath, "-json", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gdalinfo exec: %w", err)
	}
	var info GDALInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("gdalinfo parse: %w", err)
	}
	return &info, nil
}

func (p *Processor) generatePreview(ctx context.Context, inputPath, outputPath string) error {
	cmd := exec.CommandContext(ctx, p.cfg.GDALTranslatePath,
		"-of", "PNG",
		"-outsize", "800", "0",
		inputPath,
		outputPath,
	)
	return cmd.Run()
}
