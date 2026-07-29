package mining

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type mockTemplateProvider struct {
	tmpl *BlockTemplate
}

func (m *mockTemplateProvider) GetLatestTemplate() (*BlockTemplate, error) {
	return m.tmpl, nil
}

type mockDifficultyProvider struct {
	target string
}

func (m *mockDifficultyProvider) GetCurrentDifficulty() uint64 { return 1000 }
func (m *mockDifficultyProvider) GetCurrentTarget() string     { return m.target }

func TestMiningEngine_BasicMining(t *testing.T) {
	tmpl := &BlockTemplate{
		Height:       1,
		PreviousHash: "prev",
		Coinbase:     &mockTx{hash: "cb"},
		Transactions: []Transaction{},
	}

	templateProvider := &mockTemplateProvider{tmpl: tmpl}
	diffProvider := &mockDifficultyProvider{
		// Easy target so we can find a block fast
		target: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	}
	network := &mockNetwork{}
	consensus := &mockConsensus{}
	pipeline := NewSHA256Pipeline()

	var blockFound *Block
	var wg sync.WaitGroup
	wg.Add(1)

	events := MiningEvents{
		OnBlockFound: func(block *Block) {
			blockFound = block
			wg.Done()
		},
	}

	policy := DefaultMiningPolicy()
	policy.MaxWorkers = 1

	engine := NewMiningEngine(
		templateProvider,
		consensus,
		diffProvider,
		network,
		policy,
		events,
		pipeline,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := engine.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Stop()

	// Wait for the block to be found
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatalf("Timeout waiting for block")
	case <-done:
	}

	if blockFound == nil {
		t.Fatalf("Block not found")
	}

	stats := engine.GetStatistics()
	if stats.BlocksFound != 1 {
		t.Fatalf("Expected 1 block found, got %d", stats.BlocksFound)
	}
	if stats.HashesComputed == 0 {
		t.Fatalf("Expected hashes computed > 0")
	}
}

func TestMiningEngine_CancellationAndRestart(t *testing.T) {
	tmpl1 := &BlockTemplate{Height: 1, PreviousHash: "prev1"}
	tmpl2 := &BlockTemplate{Height: 2, PreviousHash: "prev2"}

	templateProvider := &mockTemplateProvider{tmpl: tmpl1}
	diffProvider := &mockDifficultyProvider{
		target: "0000000000000000000000000000000000000000000000000000000000000000", // Impossible target so it loops
	}

	var cancelled int
	events := MiningEvents{
		OnJobCancelled: func(jobID string, reason string) {
			cancelled++
		},
	}

	policy := DefaultMiningPolicy()
	engine := NewMiningEngine(
		templateProvider,
		&mockConsensus{},
		diffProvider,
		&mockNetwork{},
		policy,
		events,
		NewSHA256Pipeline(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)
	defer engine.Stop()

	// Sleep to let worker start
	time.Sleep(100 * time.Millisecond)

	// Submit new template, should cancel old
	engine.SubmitTemplate(tmpl2)

	time.Sleep(100 * time.Millisecond)

	if cancelled != 1 {
		t.Fatalf("Expected 1 cancellation, got %d", cancelled)
	}

	stats := engine.GetStatistics()
	if stats.JobsCreated != 2 {
		t.Fatalf("Expected 2 jobs created, got %d", stats.JobsCreated)
	}
}

func TestMiningEngine_StressTest(t *testing.T) {
	diffProvider := &mockDifficultyProvider{target: "0000000000000000000000000000000000000000000000000000000000000000"}

	policy := DefaultMiningPolicy()
	policy.MaxWorkers = 10
	policy.QueueSize = 1000

	engine := NewMiningEngine(
		&mockTemplateProvider{},
		&mockConsensus{},
		diffProvider,
		&mockNetwork{},
		policy,
		MiningEvents{},
		NewSHA256Pipeline(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)
	defer engine.Stop()

	var wg sync.WaitGroup
	routines := 500

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tmpl := &BlockTemplate{Height: 1000, PreviousHash: fmt.Sprintf("st-%d", id)}
			engine.SubmitTemplate(tmpl)
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	stats := engine.GetStatistics()
	t.Logf("Stress Test Completed")
	t.Logf("Jobs Created: %d", stats.JobsCreated)
	t.Logf("Jobs Cancelled: %d", stats.JobsCancelled)

	if stats.JobsCreated != 500 {
		t.Fatalf("Expected 500 created, got %d", stats.JobsCreated)
	}
}
