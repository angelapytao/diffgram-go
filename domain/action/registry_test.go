package action_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/angelapytao/diffgram-go/domain/action"
	"github.com/angelapytao/diffgram-go/domain/entity"
)

type stubRunner struct {
	name string
	ran  bool
}

func (s *stubRunner) Name() string { return s.name }
func (s *stubRunner) Run(_ context.Context, _ *entity.ActionRun) error {
	s.ran = true
	return nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := action.NewRegistry()
	w := &stubRunner{name: "webhook"}
	r.Register(w)

	got, err := r.Get("webhook")
	require.NoError(t, err)
	assert.Same(t, w, got)
}

func TestRegistry_GetUnknown(t *testing.T) {
	r := action.NewRegistry()
	_, err := r.Get("nope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nope")
}
