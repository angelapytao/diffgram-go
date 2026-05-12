package audio

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

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

func (p *Processor) MediaType() string { return "audio" }

type probeResult struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		SampleRate string `json:"sample_rate"`
		Channels   int    `json:"channels"`
	} `json:"streams"`
}

func (p *Processor) Process(ctx context.Context, input *entity.Input) error {
	localPath := input.BlobPath
	if localPath == "" {
		localPath = filepath.Join(p.cfg.TempDir, fmt.Sprintf("input_%d", input.ID), input.OriginalFilename)
	}

	probe, err := p.ffprobe(ctx, localPath)
	if err != nil {
		return fmt.Errorf("ffprobe: %w", err)
	}

	duration, _ := strconv.ParseFloat(probe.Format.Duration, 64)
	var sampleRate, channels int
	if len(probe.Streams) > 0 {
		sampleRate, _ = strconv.Atoi(probe.Streams[0].SampleRate)
		channels = probe.Streams[0].Channels
	}

	audioRecord := &entity.AudioFile{
		OriginalFilename: input.OriginalFilename,
		Duration:         duration,
		SampleRate:       sampleRate,
		Channels:         channels,
	}
	if err := p.db.WithContext(ctx).Create(audioRecord).Error; err != nil {
		return fmt.Errorf("create audio_file record: %w", err)
	}

	blobPath := fmt.Sprintf("projects/%d/audio/%d", input.ProjectID, audioRecord.ID)

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open audio file: %w", err)
	}
	if err := p.storage.Put(ctx, p.bucket, blobPath, f); err != nil {
		_ = f.Close()
		return fmt.Errorf("upload audio: %w", err)
	}
	_ = f.Close()

	audioRecord.BlobPath = blobPath
	if err := p.db.WithContext(ctx).Save(audioRecord).Error; err != nil {
		return err
	}

	fileRecord := &entity.File{
		ProjectID:        input.ProjectID,
		Type:             "audio",
		AudioFileID:      &audioRecord.ID,
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

func (p *Processor) ffprobe(ctx context.Context, path string) (*probeResult, error) {
	cmd := exec.CommandContext(ctx, p.cfg.FFprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe exec: %w", err)
	}
	var result probeResult
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("ffprobe parse: %w", err)
	}
	return &result, nil
}
