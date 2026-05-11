package action

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu      sync.RWMutex
	runners map[string]Runner
}

func NewRegistry() *Registry {
	return &Registry{runners: make(map[string]Runner)}
}

func (r *Registry) Register(runner Runner) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runners[runner.Name()] = runner
}

func (r *Registry) Get(name string) (Runner, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	runner, ok := r.runners[name]
	if !ok {
		return nil, fmt.Errorf("runner not found: %s", name)
	}
	return runner, nil
}
