package runtime

import (
	"context"
	"fmt"
)

type RuntimeEngine struct {
	policy    RuntimePolicy
	events    RuntimeEvents
	stats     *RuntimeStatistics
	registry  *ModuleRegistry
	graph     *DependencyGraph
	lifecycle *LifecycleManager
	container *ServiceContainer
	health    *HealthManager
	bus       *EventBus
	scheduler *RuntimeScheduler

	ctx    context.Context
	cancel context.CancelFunc
}

func NewRuntimeEngine(policy RuntimePolicy, events RuntimeEvents) *RuntimeEngine {
	stats := &RuntimeStatistics{}
	registry := NewModuleRegistry()
	graph := NewDependencyGraph(registry)
	lifecycle := NewLifecycleManager(registry, graph, policy, stats, events)
	container := NewServiceContainer()
	health := NewHealthManager(registry, stats, events, policy.HealthInterval)
	bus := NewEventBus(policy.EventQueueSize)
	scheduler := NewRuntimeScheduler()

	return &RuntimeEngine{
		policy:    policy,
		events:    events,
		stats:     stats,
		registry:  registry,
		graph:     graph,
		lifecycle: lifecycle,
		container: container,
		health:    health,
		bus:       bus,
		scheduler: scheduler,
	}
}

func (e *RuntimeEngine) RegisterModule(m Module, priority int) error {
	return e.registry.Register(m, priority)
}

func (e *RuntimeEngine) RegisterProvider(name string, provider interface{}) error {
	return e.container.RegisterProvider(name, provider)
}

func (e *RuntimeEngine) ResolveProvider(name string) (interface{}, error) {
	return e.container.Resolve(name)
}

func (e *RuntimeEngine) Start() error {
	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.bus.Start()
	e.health.Start(e.ctx)

	err := e.lifecycle.Startup(e.ctx)
	if err != nil {
		e.cancel()
		return err
	}

	e.stats.SetRunningServices(uint64(len(e.registry.All())))
	return nil
}

func (e *RuntimeEngine) Stop() error {
	if e.cancel != nil {
		e.cancel()
	}

	e.scheduler.StopAll()
	e.health.Stop()

	err := e.lifecycle.Shutdown(context.Background())
	e.bus.Stop()

	e.stats.SetRunningServices(0)

	return err
}

func (e *RuntimeEngine) Restart() error {
	e.stats.IncRestarts()
	if e.events.OnRestart != nil {
		go e.events.OnRestart()
	}

	err := e.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop during restart: %w", err)
	}

	return e.Start()
}

func (e *RuntimeEngine) Wait() {
	<-e.ctx.Done()
}

func (e *RuntimeEngine) Shutdown() {
	e.Stop()
}

func (e *RuntimeEngine) Status() map[string]string {
	status := make(map[string]string)
	for _, m := range e.registry.All() {
		status[m.Instance.Name()] = string(m.State)
	}
	return status
}

func (e *RuntimeEngine) Health() map[string]string {
	status := make(map[string]string)
	for _, m := range e.registry.All() {
		h, ok := e.health.GetStatus(m.Instance.Name())
		if ok {
			status[m.Instance.Name()] = string(h)
		} else {
			status[m.Instance.Name()] = "UNKNOWN"
		}
	}
	return status
}

func (e *RuntimeEngine) GetStatistics() RuntimeStatistics {
	return e.stats.Snapshot()
}

func (e *RuntimeEngine) PublishEvent(topic string, data interface{}) error {
	return e.bus.Publish(topic, data)
}

func (e *RuntimeEngine) SubscribeEvent(topic string, handler EventHandler) {
	e.bus.Subscribe(topic, handler)
}
