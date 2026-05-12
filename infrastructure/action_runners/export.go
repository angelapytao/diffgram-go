package action_runners

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/infrastructure/mq"
)

// Publisher sends a message to a message broker exchange.
type Publisher interface {
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
}

// ExportRunner publishes an export notification to the exports exchange.
type ExportRunner struct {
	publisher Publisher
}

// NewExportRunner creates an ExportRunner using the provided Publisher.
func NewExportRunner(p Publisher) *ExportRunner {
	return &ExportRunner{publisher: p}
}

// Name satisfies the action.Runner interface.
func (e *ExportRunner) Name() string {
	return "export"
}

// Run marshals a notification from the ActionRun and publishes it to the
// exports exchange with routing key "exports.created".
func (e *ExportRunner) Run(ctx context.Context, run *entity.ActionRun) error {
	notification := map[string]json.RawMessage{}

	idBytes, err := json.Marshal(run.ID)
	if err != nil {
		return fmt.Errorf("export: failed to marshal action_run_id: %w", err)
	}
	notification["action_run_id"] = idBytes

	projectIDBytes, err := json.Marshal(run.ProjectID)
	if err != nil {
		return fmt.Errorf("export: failed to marshal project_id: %w", err)
	}
	notification["project_id"] = projectIDBytes

	notification["event_payload"] = json.RawMessage(run.EventPayload)
	notification["config_data"] = json.RawMessage(run.ConfigData)

	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("export: failed to marshal notification: %w", err)
	}

	if err := e.publisher.Publish(ctx, mq.ExchangeExports, "exports.created", body); err != nil {
		return fmt.Errorf("export: failed to publish: %w", err)
	}
	return nil
}
