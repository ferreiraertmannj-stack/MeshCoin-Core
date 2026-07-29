package runtime

import (
	"context"
	"sync"
	"testing"
)

type MockModule struct {
	name    string
	deps    []string
	started bool
	stopped bool
}

func (m *MockModule) Name() string                    { return m.name }
func (m *MockModule) Version() string                 { return "1.0" }
func (m *MockModule) Dependencies() []string          { return m.deps }
func (m *MockModule) Start(ctx context.Context) error { m.started = true; return nil }
func (m *MockModule) Stop(ctx context.Context) error  { m.stopped = true; return nil }
func (m *MockModule) Health() (HealthStatus, error)   { return HealthHealthy, nil }

func TestRuntimeEngine_StartupShutdown(t *testing.T) {
	policy := DefaultRuntimePolicy()
	engine := NewRuntimeEngine(policy, RuntimeEvents{})

	modA := &MockModule{name: "Storage", deps: []string{}}
	modB := &MockModule{name: "UTXO", deps: []string{"Storage"}}
	modC := &MockModule{name: "Mempool", deps: []string{"UTXO"}}

	engine.RegisterModule(modC, 1)
	engine.RegisterModule(modB, 2)
	engine.RegisterModule(modA, 3)

	err := engine.Start()
	if err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}

	if !modA.started || !modB.started || !modC.started {
		t.Fatalf("Modules did not start")
	}

	err = engine.Stop()
	if err != nil {
		t.Fatalf("Failed to stop engine: %v", err)
	}

	if !modA.stopped || !modB.stopped || !modC.stopped {
		t.Fatalf("Modules did not stop")
	}
}

func TestRuntimeEngine_CircularDependency(t *testing.T) {
	policy := DefaultRuntimePolicy()
	engine := NewRuntimeEngine(policy, RuntimeEvents{})

	modA := &MockModule{name: "A", deps: []string{"B"}}
	modB := &MockModule{name: "B", deps: []string{"A"}}

	engine.RegisterModule(modA, 1)
	engine.RegisterModule(modB, 1)

	err := engine.Start()
	if err == nil {
		t.Fatalf("Expected circular dependency error")
	}
}

func TestRuntimeEngine_EventBusConcurrency(t *testing.T) {
	policy := DefaultRuntimePolicy()
	engine := NewRuntimeEngine(policy, RuntimeEvents{})

	err := engine.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer engine.Stop()

	var wg sync.WaitGroup
	count := 0
	var mu sync.Mutex

	engine.SubscribeEvent("test_topic", func(e Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})

	routines := 1000
	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go engine.PublishEvent("test_topic", i)
	}

	wg.Wait()
	if count != routines {
		t.Fatalf("Expected %d events processed, got %d", routines, count)
	}
}

func TestRuntimeEngine_IntegrationScenario1(t *testing.T) {
	// Scenario 1: Initializing Storage, Consensus, UTXO, Mempool, Mining, P2P, Script
	policy := DefaultRuntimePolicy()
	engine := NewRuntimeEngine(policy, RuntimeEvents{})

	engine.RegisterModule(&MockModule{name: "Storage", deps: []string{}}, 1)
	engine.RegisterModule(&MockModule{name: "Script", deps: []string{}}, 1)
	engine.RegisterModule(&MockModule{name: "UTXO", deps: []string{"Storage", "Script"}}, 2)
	engine.RegisterModule(&MockModule{name: "Consensus", deps: []string{"Storage", "UTXO"}}, 3)
	engine.RegisterModule(&MockModule{name: "Mempool", deps: []string{"UTXO", "Script"}}, 4)
	engine.RegisterModule(&MockModule{name: "Mining", deps: []string{"Consensus", "Mempool"}}, 5)
	engine.RegisterModule(&MockModule{name: "P2P", deps: []string{"Mempool", "Consensus"}}, 6)

	err := engine.Start()
	if err != nil {
		t.Fatalf("Integration Scenario 1 Failed: %v", err)
	}

	status := engine.Status()
	if len(status) != 7 {
		t.Fatalf("Expected 7 modules running")
	}

	for k, v := range status {
		if v != string(StateRunning) {
			t.Fatalf("Module %s is not running: %s", k, v)
		}
	}

	engine.Stop()
}
