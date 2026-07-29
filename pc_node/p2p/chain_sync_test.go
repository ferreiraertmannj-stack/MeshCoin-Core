package p2p

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// MockChainProvider implements ChainProvider for testing
type MockChainProvider struct {
	blocks map[string]uint64
	tip    string
	height uint64
	mu     sync.RWMutex
}

func NewMockChainProvider() *MockChainProvider {
	return &MockChainProvider{
		blocks: make(map[string]uint64),
	}
}

func (m *MockChainProvider) AddBlock(hash string, height uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blocks[hash] = height
	if height >= m.height {
		m.tip = hash
		m.height = height
	}
}

func (m *MockChainProvider) GetBlockHash(height uint64) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for hash, h := range m.blocks {
		if h == height {
			return hash, nil
		}
	}
	return "", fmt.Errorf("not found")
}

func (m *MockChainProvider) GetBlockHeight(hash string) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if h, ok := m.blocks[hash]; ok {
		return h, nil
	}
	return 0, fmt.Errorf("not found")
}

func (m *MockChainProvider) GetTip() (string, uint64) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tip, m.height
}

func (m *MockChainProvider) HasBlock(hash string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.blocks[hash]
	return ok
}

func TestPendingBlocks(t *testing.T) {
	pm := NewPendingBlocksManager()

	pm.AddBlock(PendingBlock{Hash: "B2", ParentHash: "B1", Height: 2})
	pm.AddBlock(PendingBlock{Hash: "B3", ParentHash: "B2", Height: 3})

	if !pm.HasBlock("B2") {
		t.Fatal("Expected B2 to be pending")
	}

	children := pm.GetChildren("B1")
	if len(children) != 1 || children[0].Hash != "B2" {
		t.Fatal("Expected B2 as child of B1")
	}

	pm.RemoveBlock("B2")
	if pm.HasBlock("B2") {
		t.Fatal("Expected B2 to be removed")
	}
	childrenAfter := pm.GetChildren("B1")
	if len(childrenAfter) != 0 {
		t.Fatal("Expected B1 to have no children waiting")
	}
}

func TestChainLocator(t *testing.T) {
	provider := NewMockChainProvider()
	for i := uint64(1); i <= 20; i++ {
		provider.AddBlock(fmt.Sprintf("H%d", i), i)
	}

	locator := NewChainLocator(provider)
	hashes := locator.BuildLocatorHashes()

	if len(hashes) == 0 {
		t.Fatal("Expected locators")
	}
	if hashes[0] != "H20" {
		t.Fatal("First locator should be tip")
	}
}

func TestForkDetection(t *testing.T) {
	provider := NewMockChainProvider()
	provider.AddBlock("B1", 1)
	provider.AddBlock("B2", 2)
	provider.AddBlock("B3", 3)

	var forkDetected int32

	events := BlockchainSyncEvents{
		OnForkDetected: func(hash string) {
			atomic.AddInt32(&forkDetected, 1)
		},
	}
	detector := NewForkDetector(provider, events)

	// Normal block
	detector.CheckForFork("B4", "B3", 4)

	// Fork block at B2
	detector.CheckForFork("B3-alt", "B2", 3)

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&forkDetected) != 1 {
		t.Fatalf("Expected 1 fork detected, got %d", forkDetected)
	}
}

func TestSchedulerAndTimeout(t *testing.T) {
	router, pm := setupTestRouter() // from router_test.go
	defer router.Stop()

	stats := &BlockchainSyncStatistics{}
	events := BlockchainSyncEvents{}

	// Small timeout to test retries
	scheduler := NewBlockRequestScheduler(router, pm, stats, events, 100*time.Millisecond, 2)

	peer := createTestPeer("test_peer")
	pm.AddPeer(peer)

	scheduler.Start()
	defer scheduler.Stop()

	scheduler.Schedule("B1", 1)

	time.Sleep(350 * time.Millisecond) // Let it timeout a few times

	if scheduler.Count() > 1 {
		t.Fatalf("Expected at most 1 pending, got %d", scheduler.Count())
	}
}

func TestBlockchainSyncManager_StressTest(t *testing.T) {
	router, pm := setupTestRouter()
	defer router.Stop()

	provider := NewMockChainProvider()
	provider.AddBlock("GENESIS", 0)

	var imported int32
	events := BlockchainSyncEvents{
		OnBlockImported: func(hash string, height uint64) {
			provider.AddBlock(hash, height)
			atomic.AddInt32(&imported, 1)
		},
	}

	manager := NewBlockchainSyncManager(router, pm, nil, provider, events, 500*time.Millisecond, 3)
	manager.Start()
	defer manager.Stop()

	peer := createTestPeer("peer_sync")
	pm.AddPeer(peer)

	// Simulate 1000 goroutines feeding out of order blocks
	var wg sync.WaitGroup
	for i := 1; i <= 1000; i++ {
		wg.Add(1)
		go func(height int) {
			defer wg.Done()
			hash := fmt.Sprintf("B%d", height)
			parentHash := fmt.Sprintf("B%d", height-1)
			if height == 1 {
				parentHash = "GENESIS"
			}
			manager.HandleBlockReceived(hash, parentHash, uint64(height), []byte{})
		}(i)
	}

	wg.Wait()

	// We need to wait for imports to finish. Since they can trigger cascading imports,
	// give it a short time.
	time.Sleep(500 * time.Millisecond)

	count := atomic.LoadInt32(&imported)
	if count != 1000 {
		t.Fatalf("Expected 1000 imported blocks, got %d", count)
	}
}
