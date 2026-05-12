package action_runners_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/angelapytao/diffgram-go/domain/entity"
	"github.com/angelapytao/diffgram-go/infrastructure/action_runners"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

type stubTaskCreator struct {
	called bool
	req    action_runners.TaskCreateRequest
}

func (s *stubTaskCreator) CreateTasks(_ context.Context, req action_runners.TaskCreateRequest) error {
	s.called = true
	s.req = req
	return nil
}

func TestTaskTemplateRunner_Name(t *testing.T) {
	r := action_runners.NewTaskTemplateRunner(&stubTaskCreator{})
	assert.Equal(t, "task_template", r.Name())
}

func TestTaskTemplateRunner_CallsCreatorWithCount(t *testing.T) {
	stub := &stubTaskCreator{}
	r := action_runners.NewTaskTemplateRunner(stub)

	cfgBytes, _ := json.Marshal(map[string]any{"template_id": 5, "count": 3})
	run := &entity.ActionRun{
		ID:         1,
		ProjectID:  ptr(42),
		ConfigData: datatypes.JSON(cfgBytes),
	}

	err := r.Run(context.Background(), run)
	require.NoError(t, err)
	assert.True(t, stub.called)
	assert.Equal(t, 5, stub.req.TemplateID)
	assert.Equal(t, 3, stub.req.Count)
	assert.Equal(t, 42, stub.req.ProjectID)
}

func TestTaskTemplateRunner_RejectsZeroCount(t *testing.T) {
	stub := &stubTaskCreator{}
	r := action_runners.NewTaskTemplateRunner(stub)

	cfgBytes, _ := json.Marshal(map[string]any{"template_id": 1, "count": 0})
	run := &entity.ActionRun{
		ID:         2,
		ProjectID:  ptr(10),
		ConfigData: datatypes.JSON(cfgBytes),
	}

	err := r.Run(context.Background(), run)
	assert.Error(t, err)
	assert.False(t, stub.called)
}
