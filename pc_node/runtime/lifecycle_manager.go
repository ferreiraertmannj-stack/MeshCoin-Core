package runtime

import (
	"context"
	"fmt"
)

type LifecycleManager struct {
	registry *ModuleRegistry
	graph    *DependencyGraph
	policy   RuntimePolicy
	stats    *RuntimeStatistics
	events   RuntimeEvents
}

func NewLifecycleManager(
	registry *ModuleRegistry,
	graph *DependencyGraph,
	policy RuntimePolicy,
	stats *RuntimeStatistics,
	events RuntimeEvents,
) *LifecycleManager {
	return &LifecycleManager{
		registry: registry,
		graph:    graph,
		policy:   policy,
		stats:    stats,
		events:   events,
	}
}

func (m *LifecycleManager) Startup(ctx context.Context) error {
	order, err := m.graph.Build()
	if err != nil {
		return fmt.Errorf("failed to build startup order: %w", err)
	}

	startCtx, cancel := context.WithTimeout(ctx, m.policy.StartupTimeout)
	defer cancel()

	for _, name := range order {
		info, err := m.registry.Get(name)
		if err != nil {
			return err
		}

		m.registry.SetState(name, StateStarting)

		err = info.Instance.Start(startCtx)
		if err != nil {
			m.registry.SetState(name, StateFailed)
			m.stats.IncModulesFailed()
			if m.events.OnModuleFailed != nil {
				go m.events.OnModuleFailed(name, err)
			}
			return fmt.Errorf("failed to start module %s: %w", name, err)
		}

		m.registry.SetState(name, StateRunning)
		m.stats.IncModulesLoaded()
		if m.events.OnModuleLoaded != nil {
			go m.events.OnModuleLoaded(name)
		}
	}

	if m.events.OnNodeStarted != nil {
		go m.events.OnNodeStarted()
	}

	return nil
}

func (m *LifecycleManager) Shutdown(ctx context.Context) error {
	order, err := m.graph.Build()
	if err != nil {
		return err
	}

	// Reverse order for shutdown
	for i := len(order)/2 - 1; i >= 0; i-- {
		opp := len(order) - 1 - i
		order[i], order[opp] = order[opp], order[i]
	}

	stopCtx, cancel := context.WithTimeout(ctx, m.policy.ShutdownTimeout)
	defer cancel()

	var lastErr error
	for _, name := range order {
		info, err := m.registry.Get(name)
		if err != nil {
			continue
		}

		m.registry.SetState(name, StateStopping)
		err = info.Instance.Stop(stopCtx)
		if err != nil {
			m.registry.SetState(name, StateFailed)
			lastErr = fmt.Errorf("failed to stop module %s: %w", name, err)
		} else {
			m.registry.SetState(name, StateStopped)
		}
	}

	m.stats.IncShutdowns()
	if m.events.OnShutdown != nil {
		go m.events.OnShutdown()
	}

	return lastErr
}
