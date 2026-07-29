package runtime

import (
	"context"
	"fmt"
	"sync"
)

type ModuleState string

const (
	StateRegistered ModuleState = "REGISTERED"
	StateStarting   ModuleState = "STARTING"
	StateRunning    ModuleState = "RUNNING"
	StateStopping   ModuleState = "STOPPING"
	StateStopped    ModuleState = "STOPPED"
	StateFailed     ModuleState = "FAILED"
)

type Module interface {
	Name() string
	Version() string
	Dependencies() []string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type ModuleInfo struct {
	Instance Module
	State    ModuleState
	Priority int
}

type ModuleRegistry struct {
	mu      sync.RWMutex
	modules map[string]*ModuleInfo
}

func NewModuleRegistry() *ModuleRegistry {
	return &ModuleRegistry{
		modules: make(map[string]*ModuleInfo),
	}
}

func (r *ModuleRegistry) Register(m Module, priority int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := m.Name()
	if _, exists := r.modules[name]; exists {
		return fmt.Errorf("module %s already registered", name)
	}

	r.modules[name] = &ModuleInfo{
		Instance: m,
		State:    StateRegistered,
		Priority: priority,
	}

	return nil
}

func (r *ModuleRegistry) Get(name string) (*ModuleInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, exists := r.modules[name]
	if !exists {
		return nil, fmt.Errorf("module %s not found", name)
	}
	return info, nil
}

func (r *ModuleRegistry) SetState(name string, state ModuleState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, exists := r.modules[name]
	if !exists {
		return fmt.Errorf("module %s not found", name)
	}
	info.State = state
	return nil
}

func (r *ModuleRegistry) All() []*ModuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*ModuleInfo, 0, len(r.modules))
	for _, info := range r.modules {
		list = append(list, info)
	}
	return list
}
