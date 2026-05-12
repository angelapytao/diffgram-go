package video

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

// Processor handles video file ingestion: probes metadata via ffprobe,
// persists a Video record, uploads the original blob, and links a File record.
type Processor struct {
	storage domainservice.StorageProvider
	db      *gorm.DB
	cfg     config.ProcessorConfig
	bucket  string
}

func NewProcessor(storage domainservice.StorageProvider, db *gorm.DB, cfg config.ProcessorConfig, bucket string) *Processor {
	return &Processor{storage: storage, db: db, cfg: cfg, bucket: bucket}
}

func (p *Processor) MediaType() string { return "video" }

type probeResult struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
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
	var width, height int
	var fps float64
	if len(probe.Streams) > 0 {
		width = probe.Streams[0].Width
		height = probe.Streams[0].Height
		fps = parseFPS(probe.Streams[0].RFrameRate)
	}
	if fps == 0 {
		fps = float64(p.cfg.VideoFPSDefault)
	}

	frameCount := int(duration * fps)

	videoRecord := &entity.Video{
		OriginalFilename: input.OriginalFilename,
		FPS:              fps,
		FrameCount:       frameCount,
		Duration:         duration,
		Width:            width,
		Height:           height,
		Status:           "processing",
	}
	if err := p.db.WithContext(ctx).Create(videoRecord).Error; err != nil {
		return fmt.Errorf("create video record: %w", err)
	}

	basePath := fmt.Sprintf("projects/%d/videos/%d", input.ProjectID, videoRecord.ID)

	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open video file: %w", err)
	}
	if err := p.storage.Put(ctx, p.bucket, basePath+"/original", f); err != nil {
		_ = f.Close()
		return fmt.Errorf("upload video: %w", err)
	}
	_ = f.Close()

	videoRecord.FileBlobPath = basePath + "/original"
	videoRecord.Status = "complete"
	if err := p.db.WithContext(ctx).Save(videoRecord).Error; err != nil {
		return err
	}

	fileRecord := &entity.File{
		ProjectID:        input.ProjectID,
		Type:             "video",
		VideoID:          &videoRecord.ID,
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

func parseFPS(rFrameRate string) float64 {
	// r_frame_rate is typically "30/1" or "30000/1001"
	var num, den int
	if _, err := fmt.Sscanf(rFrameRate, "%d/%d", &num, &den); err == nil && den > 0 {
		return float64(num) / float64(den)
	}
	f, _ := strconv.ParseFloat(rFrameRate, 64)
	return f
}
