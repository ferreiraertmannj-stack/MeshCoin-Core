package mining

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockTx struct {
	hash string
	fee  uint64
	size uint64
}

func (m *mockTx) GetHash() string     { return m.hash }
func (m *mockTx) GetFee() uint64      { return m.fee }
func (m *mockTx) GetSize() uint64     { return m.size }
func (m *mockTx) GetSender() string   { return "sender" }
func (m *mockTx) GetTimestamp() int64 { return time.Now().Unix() }

type mockBlockchain struct{}

func (m *mockBlockchain) GetHighestBlockHeight() uint64 { return 100 }
func (m *mockBlockchain) GetHighestBlockHash() string   { return "prevhash" }
func (m *mockBlockchain) GetDifficultyTarget() string   { return "0000ffff" }

type mockMempool struct {
	txs []Transaction
	mu  sync.RWMutex
}

func (m *mockMempool) Snapshot() []Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]Transaction(nil), m.txs...)
}

func (m *mockMempool) add(tx Transaction) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txs = append(m.txs, tx)
}

func (m *mockMempool) clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txs = nil
}

type mockConsensus struct{}

func (m *mockConsensus) CalculateMerkleRoot(hashes []string) string {
	raw := ""
	for _, h := range hashes {
		raw += h
	}
	hashBytes := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hashBytes[:])
}

func (m *mockConsensus) ValidateBlockTemplate(template *BlockTemplate) error { return nil }

type mockNetwork struct{}

func (m *mockNetwork) GetNetworkTimestamp() int64 { return time.Now().Unix() }

func TestBlockTemplateEngine_EmptyMempool(t *testing.T) {
	engine := NewBlockTemplateEngine(
		&mockBlockchain{},
		&mockMempool{},
		&mockConsensus{},
		&mockNetwork{},
		DefaultTemplatePolicy(),
		TemplateEvents{},
	)

	tmpl, err := engine.BuildNewTemplate()
	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	if len(tmpl.Transactions) != 0 {
		t.Fatalf("Expected 0 transactions, got %d", len(tmpl.Transactions))
	}
	if tmpl.Coinbase == nil {
		t.Fatal("Missing coinbase")
	}
}

func TestBlockTemplateEngine_WeightLimitAndOrdering(t *testing.T) {
	mempool := &mockMempool{}

	// Add 3 transactions. Max weight allows 2.
	mempool.add(&mockTx{hash: "tx1", fee: 10, size: 500000}) // density 0.00002
	mempool.add(&mockTx{hash: "tx2", fee: 20, size: 500000}) // density 0.00004 (Best)
	mempool.add(&mockTx{hash: "tx3", fee: 15, size: 500000}) // density 0.00003

	policy := DefaultTemplatePolicy()
	policy.MaxBlockWeight = 1100000 // allows 2 txs (1M) + coinbase (100)

	engine := NewBlockTemplateEngine(
		&mockBlockchain{},
		mempool,
		&mockConsensus{},
		&mockNetwork{},
		policy,
		TemplateEvents{},
	)

	tmpl, err := engine.BuildNewTemplate()
	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	if len(tmpl.Transactions) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(tmpl.Transactions))
	}

	// Ordered by fee density: tx2, tx3
	if tmpl.Transactions[0].GetHash() != "tx2" {
		t.Fatalf("Expected tx2 first, got %s", tmpl.Transactions[0].GetHash())
	}
	if tmpl.Transactions[1].GetHash() != "tx3" {
		t.Fatalf("Expected tx3 second, got %s", tmpl.Transactions[1].GetHash())
	}
}

func TestBlockTemplateEngine_Duplicates(t *testing.T) {
	mempool := &mockMempool{}
	mempool.add(&mockTx{hash: "tx1", fee: 10, size: 1000})
	mempool.add(&mockTx{hash: "tx1", fee: 10, size: 1000}) // Duplicate

	engine := NewBlockTemplateEngine(
		&mockBlockchain{},
		mempool,
		&mockConsensus{},
		&mockNetwork{},
		DefaultTemplatePolicy(),
		TemplateEvents{},
	)

	tmpl, err := engine.BuildNewTemplate()
	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	if len(tmpl.Transactions) != 1 {
		t.Fatalf("Expected 1 transactions, got %d", len(tmpl.Transactions))
	}
}

func TestBlockTemplateEngine_Cache(t *testing.T) {
	mempool := &mockMempool{}
	engine := NewBlockTemplateEngine(
		&mockBlockchain{},
		mempool,
		&mockConsensus{},
		&mockNetwork{},
		DefaultTemplatePolicy(),
		TemplateEvents{},
	)

	tmpl1, _ := engine.GetLatestTemplate() // Miss, builds new
	tmpl2, _ := engine.GetLatestTemplate() // Hit

	if tmpl1.Coinbase.GetHash() != tmpl2.Coinbase.GetHash() {
		t.Fatalf("Expected cached template")
	}

	stats := engine.stats.Snapshot()
	if stats.CacheHits != 1 {
		t.Fatalf("Expected 1 cache hit, got %d", stats.CacheHits)
	}
	if stats.CacheMisses != 1 {
		t.Fatalf("Expected 1 cache miss, got %d", stats.CacheMisses)
	}
}

func TestBlockTemplateEngine_Scheduler(t *testing.T) {
	mempool := &mockMempool{}
	policy := DefaultTemplatePolicy()
	policy.RefreshInterval = 100 * time.Millisecond // Fast periodic

	var updated atomic.Int32
	events := TemplateEvents{
		OnTemplateUpdated: func(height uint64) {
			updated.Add(1)
		},
	}

	engine := NewBlockTemplateEngine(
		&mockBlockchain{},
		mempool,
		&mockConsensus{},
		&mockNetwork{},
		policy,
		events,
	)

	engine.Start()
	defer engine.Stop()

	// Wait for periodic updates
	time.Sleep(250 * time.Millisecond)

	count := updated.Load()
	if count < 2 {
		t.Fatalf("Expected at least 2 updates from scheduler, got %d", count)
	}
}

func TestBlockTemplateEngine_StressTest(t *testing.T) {
	mempool := &mockMempool{}
	policy := DefaultTemplatePolicy()
	policy.MaxTransactions = 100

	engine := NewBlockTemplateEngine(
		&mockBlockchain{},
		mempool,
		&mockConsensus{},
		&mockNetwork{},
		policy,
		TemplateEvents{},
	)

	var wg sync.WaitGroup
	routines := 500

	// Concurrent mempool additions + triggers
	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mempool.add(&mockTx{
				hash: fmt.Sprintf("st-tx-%d", id),
				fee:  uint64(id),
				size: 1000,
			})
			_, _ = engine.BuildNewTemplate()
		}(i)
	}

	wg.Wait()

	stats := engine.stats.Snapshot()
	t.Logf("Stress Test Completed")
	t.Logf("Templates Built: %d", stats.TemplatesCreated)
	if stats.TemplatesCreated != 500 {
		t.Fatalf("Expected 500 created, got %d", stats.TemplatesCreated)
	}
}
