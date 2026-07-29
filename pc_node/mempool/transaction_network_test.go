package mempool

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockGossip struct {
	published int32
}

func (m *mockGossip) Publish(topic string, data []byte) error {
	atomic.AddInt32(&m.published, 1)
	return nil
}

type mockInventory struct {
	announced int32
	requested int32
}

func (m *mockInventory) AnnounceObject(objectType string, hash string) error {
	atomic.AddInt32(&m.announced, 1)
	return nil
}

func (m *mockInventory) RequestObject(peerID string, objectType string, hash string) error {
	atomic.AddInt32(&m.requested, 1)
	return nil
}

type mockRouter struct {
	sent int32
}

func (m *mockRouter) SendToPeer(peerID string, msgType byte, payload interface{}) error {
	atomic.AddInt32(&m.sent, 1)
	return nil
}

func (m *mockRouter) RegisterHandler(msgType byte, handler func(peerID string, payload []byte) error) {
}

func TestNetworkBridge_Pipeline(t *testing.T) {
	policy := DefaultMempoolPolicy()
	poolEvents := MempoolEvents{}
	pool := NewTransactionPool(policy, poolEvents, 100, 2)
	pool.Start()
	defer pool.Stop()

	gossip := &mockGossip{}
	inv := &mockInventory{}
	router := &mockRouter{}

	events := TransactionNetworkEvents{}
	config := NetworkBridgeConfig{QueueSize: 100, Workers: 2, DedupTTL: time.Hour}

	bridge := NewNetworkBridge(pool, gossip, inv, router, router, events, config)
	bridge.Start()
	defer bridge.Stop()

	// 1. Submit tx
	tx := MsgTransaction{Hash: "tx1", Sender: "alice", Fee: 10, Timestamp: time.Now().Unix(), Payload: []byte("data")}
	raw, _ := json.Marshal(tx)

	err := bridge.handlers.HandleMsgTransaction("peer1", raw)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // async process

	stats := bridge.Statistics()
	if stats.Accepted != 1 {
		t.Fatalf("expected 1 accepted, got %d", stats.Accepted)
	}
	if atomic.LoadInt32(&inv.announced) != 1 {
		t.Fatalf("expected 1 announcement")
	}
	if atomic.LoadInt32(&gossip.published) != 1 {
		t.Fatalf("expected 1 gossip publish")
	}

	// 2. Submit duplicate
	err = bridge.handlers.HandleMsgTransaction("peer2", raw)
	if err != nil {
		t.Fatalf("queue should not fail")
	}

	time.Sleep(100 * time.Millisecond)

	stats = bridge.Statistics()
	if stats.Duplicates != 1 {
		t.Fatalf("expected 1 duplicate, got %d", stats.Duplicates)
	}

	// 3. Test Inventory Announcement
	annMsg := MsgTransactionAnnouncement{Hash: "tx2"}
	annRaw, _ := json.Marshal(annMsg)
	_ = bridge.handlers.HandleMsgInventory("peer3", annRaw)

	time.Sleep(50 * time.Millisecond)

	stats = bridge.Statistics()
	if stats.Announcements != 1 {
		t.Fatalf("expected 1 announcement received")
	}
	if stats.Downloads != 1 {
		t.Fatalf("expected 1 download requested")
	}
	if atomic.LoadInt32(&inv.requested) != 1 {
		t.Fatalf("expected 1 inventory request")
	}
}

func TestNetworkBridge_StressTest(t *testing.T) {
	policy := DefaultMempoolPolicy()
	policy.MaxTransactions = 50000
	poolEvents := MempoolEvents{}
	pool := NewTransactionPool(policy, poolEvents, 10000, 4)
	pool.Start()
	defer pool.Stop()

	gossip := &mockGossip{}
	inv := &mockInventory{}
	router := &mockRouter{}

	events := TransactionNetworkEvents{}
	config := NetworkBridgeConfig{QueueSize: 10000, Workers: 10, DedupTTL: time.Hour}

	bridge := NewNetworkBridge(pool, gossip, inv, router, router, events, config)
	bridge.Start()
	defer bridge.Stop()

	var wg sync.WaitGroup
	routines := 1000

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			tx := MsgTransaction{
				Hash:      fmt.Sprintf("st-tx-%d", id),
				Sender:    fmt.Sprintf("sender-%d", id%5),
				Fee:       10,
				Timestamp: time.Now().Unix(),
			}
			raw, _ := json.Marshal(tx)
			_ = bridge.handlers.HandleMsgTransaction("peer", raw)

			// Try to cause duplicates on 10%
			if id%10 == 0 {
				_ = bridge.handlers.HandleMsgTransaction("peer2", raw)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond) // Wait queue drain

	stats := bridge.Statistics()
	t.Logf("Stress Test Received: %d", stats.Received)
	t.Logf("Stress Test Accepted: %d", stats.Accepted)
	t.Logf("Stress Test Duplicates: %d", stats.Duplicates)

	if stats.Accepted != 1000 {
		t.Fatalf("expected 1000 accepted, got %d", stats.Accepted)
	}
}
