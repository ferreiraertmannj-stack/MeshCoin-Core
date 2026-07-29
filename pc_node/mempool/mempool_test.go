package mempool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockTx implements Transaction for testing
type mockTx struct {
	hash      string
	sender    string
	fee       uint64
	size      uint64
	timestamp int64
	nonce     uint64
}

func (m *mockTx) GetHash() string     { return m.hash }
func (m *mockTx) GetSender() string   { return m.sender }
func (m *mockTx) GetFee() uint64      { return m.fee }
func (m *mockTx) GetSize() uint64     { return m.size }
func (m *mockTx) GetTimestamp() int64 { return m.timestamp }
func (m *mockTx) GetNonce() uint64    { return m.nonce }

func TestMempool_BasicOperations(t *testing.T) {
	policy := DefaultMempoolPolicy()
	events := MempoolEvents{}
	pool := NewTransactionPool(policy, events, 100, 2)
	pool.Start()
	defer pool.Stop()

	tx1 := &mockTx{
		hash:      "tx1",
		sender:    "alice",
		fee:       10,
		size:      200,
		timestamp: time.Now().Unix(),
		nonce:     1,
	}

	err := pool.processTransaction(tx1)
	if err != nil {
		t.Fatalf("Failed to add valid tx: %v", err)
	}

	if !pool.Contains("tx1") {
		t.Fatal("Pool should contain tx1")
	}

	if pool.Count() != 1 {
		t.Fatalf("Expected pool size 1, got %d", pool.Count())
	}

	pool.RemoveTransaction("tx1")
	if pool.Contains("tx1") {
		t.Fatal("Pool should not contain tx1 after removal")
	}
}

func TestMempool_ValidationFailed(t *testing.T) {
	policy := DefaultMempoolPolicy()
	var failedCalled atomic.Bool
	events := MempoolEvents{
		OnValidationFailed: func(hash string, reason string) {
			failedCalled.Store(true)
		},
	}

	pool := NewTransactionPool(policy, events, 100, 2)
	pool.Start()
	defer pool.Stop()

	// Tx with 0 fee (below MinFee = 1)
	txInvalid := &mockTx{
		hash:      "tx-invalid",
		sender:    "bob",
		fee:       0,
		size:      100,
		timestamp: time.Now().Unix(),
		nonce:     1,
	}

	err := pool.processTransaction(txInvalid)
	if err == nil {
		t.Fatal("Expected validation error for low fee")
	}

	// Small sleep for async event
	time.Sleep(50 * time.Millisecond)
	if !failedCalled.Load() {
		t.Fatal("OnValidationFailed event not fired")
	}
}

func TestMempool_ReplaceByFee(t *testing.T) {
	policy := DefaultMempoolPolicy()
	policy.MaxTransactions = 1 // Force overflow immediately
	pool := NewTransactionPool(policy, MempoolEvents{}, 100, 1)
	pool.Start()
	defer pool.Stop()

	tx1 := &mockTx{hash: "tx1", sender: "charlie", fee: 10, size: 100, timestamp: time.Now().Unix()}
	tx2 := &mockTx{hash: "tx2", sender: "dave", fee: 20, size: 100, timestamp: time.Now().Unix()}

	_ = pool.processTransaction(tx1)
	err := pool.processTransaction(tx2) // Should evict tx1

	if err != nil {
		t.Fatalf("tx2 should be accepted, error: %v", err)
	}

	if pool.Contains("tx1") {
		t.Fatal("tx1 should have been evicted")
	}
	if !pool.Contains("tx2") {
		t.Fatal("tx2 should be in the pool")
	}
}

func TestMempool_StressTest(t *testing.T) {
	policy := DefaultMempoolPolicy()
	// Large limits to not fail by RBF randomly
	policy.MaxTransactions = 50000

	var (
		addedCount   int32
		removedCount int32
		dupCount     int32
	)

	events := MempoolEvents{
		OnTransactionAdded: func(hash string) {
			atomic.AddInt32(&addedCount, 1)
		},
		OnTransactionRemoved: func(hash string) {
			atomic.AddInt32(&removedCount, 1)
		},
		OnDuplicateTransaction: func(hash string) {
			atomic.AddInt32(&dupCount, 1)
		},
	}

	// Use a large queue and multiple workers for stress test
	pool := NewTransactionPool(policy, events, 10000, 10)
	pool.Start()
	defer pool.Stop()

	var wg sync.WaitGroup
	numRoutines := 1000

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Add Transaction
			txHash := fmt.Sprintf("tx-%d", id)
			tx := &mockTx{
				hash:      txHash,
				sender:    fmt.Sprintf("sender-%d", id%10),
				fee:       10,
				size:      200,
				timestamp: time.Now().Unix(),
				nonce:     uint64(id),
			}

			err := pool.AddTransaction(tx)
			if err != nil {
				return // Queue overflow or similar, though queue is 10k so it shouldn't
			}

			// Duplicate attempt
			_ = pool.AddTransaction(tx)

			// Remove half of them concurrently
			if id%2 == 0 {
				// Wait a tiny bit so they are processed from the queue
				time.Sleep(10 * time.Millisecond)
				pool.RemoveTransaction(txHash)
			}
		}(i)
	}

	wg.Wait()

	// Wait for queue to drain
	time.Sleep(500 * time.Millisecond)

	stats := pool.Statistics()
	t.Logf("Stress Test Completed:")
	t.Logf("Transactions added event count: %d", atomic.LoadInt32(&addedCount))
	t.Logf("Transactions removed event count: %d", atomic.LoadInt32(&removedCount))
	t.Logf("Duplicate events: %d", atomic.LoadInt32(&dupCount))
	t.Logf("Stats Transactions in pool: %d", stats.Transactions)
	t.Logf("Stats Hits: %d, Misses: %d", stats.Hits, stats.Misses)
}
