package text

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (p *Processor) MediaType() string { return "text" }

func (p *Processor) Process(ctx context.Context, input *entity.Input) error {
	localPath := input.BlobPath
	if localPath == "" {
		localPath = filepath.Join(p.cfg.TempDir, fmt.Sprintf("input_%d", input.ID), input.OriginalFilename)
	}

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open text file: %w", err)
	}
	defer func() { _ = f.Close() }()

	tokenCount := countTokens(f)

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	textRecord := &entity.TextFile{
		OriginalFilename: input.OriginalFilename,
		TokenCount:       tokenCount,
	}
	if err := p.db.WithContext(ctx).Create(textRecord).Error; err != nil {
		return fmt.Errorf("create text_file record: %w", err)
	}

	blobPath := fmt.Sprintf("projects/%d/text/%d", input.ProjectID, textRecord.ID)
	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if err := p.storage.Put(ctx, p.bucket, blobPath, f); err != nil {
		return fmt.Errorf("upload text: %w", err)
	}

	textRecord.BlobPath = blobPath
	if err := p.db.WithContext(ctx).Save(textRecord).Error; err != nil {
		return err
	}

	fileRecord := &entity.File{
		ProjectID:        input.ProjectID,
		Type:             "text",
		TextFileID:       &textRecord.ID,
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

func countTokens(f *os.File) int {
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanWords)
	count := 0
	for scanner.Scan() {
		word := scanner.Text()
		if strings.TrimSpace(word) != "" {
			count++
		}
	}
	return count
}
