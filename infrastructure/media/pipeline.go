package media

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
)

type MediaProcessor interface {
	MediaType() string
	Process(ctx context.Context, input *entity.Input) error
}

type Pipeline struct {
	db         *gorm.DB
	publisher  mq.Publisher
	processors map[string]MediaProcessor
}

func NewPipeline(db *gorm.DB, pub mq.Publisher, procs map[string]MediaProcessor) *Pipeline {
	return &Pipeline{db: db, publisher: pub, processors: procs}
}

type MediaProcessedEvent struct {
	InputID   int64  `json:"input_id"`
	ProjectID int64  `json:"project_id"`
	FileID    int64  `json:"file_id"`
	MediaType string `json:"media_type"`
	Status    string `json:"status"`
}

func (p *Pipeline) Run(ctx context.Context, input *entity.Input) error {
	log := logrus.WithFields(logrus.Fields{
		"input_id":   input.ID,
		"project_id": input.ProjectID,
		"media_type": input.MediaType,
	})

	processor, ok := p.processors[input.MediaType]
	if !ok {
		return p.fail(input, fmt.Sprintf("unsupported media type: %s", input.MediaType))
	}

	input.Status = "processing"
	if err := p.db.WithContext(ctx).Save(input).Error; err != nil {
		return err
	}

	if err := processor.Process(ctx, input); err != nil {
		log.WithError(err).Error("processing failed")
		return p.fail(input, err.Error())
	}

	input.Status = "success"
	input.PercentComplete = 100
	if err := p.db.WithContext(ctx).Save(input).Error; err != nil {
		return err
	}

	fileID := int64(0)
	if input.FileID != nil {
		fileID = *input.FileID
	}
	event := MediaProcessedEvent{
		InputID:   input.ID,
		ProjectID: input.ProjectID,
		FileID:    fileID,
		MediaType: input.MediaType,
		Status:    "success",
	}
	if err := p.publisher.Publish(ctx, mq.ExchangeEvents, mq.RoutingKeyEventsNew, event); err != nil {
		log.WithError(err).Warn("failed to publish event")
	}

	return nil
}

func (p *Pipeline) fail(input *entity.Input, msg string) error {
	input.Status = "failed"
	input.StatusText = msg
	return p.db.Save(input).Error
}
