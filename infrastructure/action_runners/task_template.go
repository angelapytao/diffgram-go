package action_runners

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/angelapytao/diffgram-go/domain/entity"
)

// TaskCreateRequest holds the parameters for creating tasks from a template.
type TaskCreateRequest struct {
	TemplateID int
	Count      int
	ProjectID  int
}

// TaskCreator creates tasks from a template.
type TaskCreator interface {
	CreateTasks(ctx context.Context, req TaskCreateRequest) error
}

// TaskTemplateRunner creates tasks based on a template when an action fires.
type TaskTemplateRunner struct {
	creator TaskCreator
}

// NewTaskTemplateRunner creates a TaskTemplateRunner using the provided TaskCreator.
func NewTaskTemplateRunner(c TaskCreator) *TaskTemplateRunner {
	return &TaskTemplateRunner{creator: c}
}

// Name satisfies the action.Runner interface.
func (t *TaskTemplateRunner) Name() string {
	return "task_template"
}

// Run unmarshals config_data for template_id and count, validates count > 0,
// then delegates to TaskCreator.CreateTasks.
func (t *TaskTemplateRunner) Run(ctx context.Context, run *entity.ActionRun) error {
	var cfg struct {
		TemplateID int `json:"template_id"`
		Count      int `json:"count"`
	}
	if err := json.Unmarshal(run.ConfigData, &cfg); err != nil {
		return fmt.Errorf("task_template: failed to parse config_data: %w", err)
	}

	if cfg.Count <= 0 {
		return fmt.Errorf("task_template: count must be > 0, got %d", cfg.Count)
	}

	projectID := 0
	if run.ProjectID != nil {
		projectID = *run.ProjectID
	}

	return t.creator.CreateTasks(ctx, TaskCreateRequest{
		TemplateID: cfg.TemplateID,
		Count:      cfg.Count,
		ProjectID:  projectID,
	})
}
