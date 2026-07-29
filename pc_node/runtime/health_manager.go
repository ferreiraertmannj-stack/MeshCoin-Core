package runtime

import (
	"context"
	"sync"
	"time"
)

type HealthStatus string

const (
	HealthHealthy  HealthStatus = "HEALTHY"
	HealthDegraded HealthStatus = "DEGRADED"
	HealthStopped  HealthStatus = "STOPPED"
	HealthFailed   HealthStatus = "FAILED"
)

type HealthCheckable interface {
	Health() (HealthStatus, error)
}

type HealthManager struct {
	registry *ModuleRegistry
	stats    *RuntimeStatistics
	events   RuntimeEvents
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	statuses map[string]HealthStatus
	mu       sync.RWMutex
}

func NewHealthManager(
	registry *ModuleRegistry,
	stats *RuntimeStatistics,
	events RuntimeEvents,
	interval time.Duration,
) *HealthManager {
	return &HealthManager{
		registry: registry,
		stats:    stats,
		events:   events,
		interval: interval,
		statuses: make(map[string]HealthStatus),
	}
}

func (h *HealthManager) Start(ctx context.Context) {
	h.ctx, h.cancel = context.WithCancel(ctx)
	h.wg.Add(1)
	go h.monitor()
}

func (h *HealthManager) Stop() {
	if h.cancel != nil {
		h.cancel()
	}
	h.wg.Wait()
}

func (h *HealthManager) monitor() {
	defer h.wg.Done()
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.checkAll()
		}
	}
}

func (h *HealthManager) checkAll() {
	modules := h.registry.All()
	for _, info := range modules {
		if checkable, ok := info.Instance.(HealthCheckable); ok {
			status, _ := checkable.Health()

			h.mu.Lock()
			prev := h.statuses[info.Instance.Name()]
			h.statuses[info.Instance.Name()] = status
			h.mu.Unlock()

			if prev != status && h.events.OnHealthChanged != nil {
				go h.events.OnHealthChanged(info.Instance.Name(), string(status))
			}
		}
	}
	h.stats.IncHealthChecks()
}

func (h *HealthManager) GetStatus(name string) (HealthStatus, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	s, ok := h.statuses[name]
	return s, ok
}
