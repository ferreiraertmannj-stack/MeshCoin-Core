package p2p

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

func setupTestGossipManager() (*GossipManager, *MessageRouter, *MockPeerManager) {
	pm := NewMockPeerManager()
	secMgr := NewSecurityManager()
	events := RouterEvents{}
	router := NewMessageRouter(pm, secMgr, events)
	router.Start()

	gossipEvents := GossipEvents{}
	// fast expiration for tests
	gossipMgr := NewGossipManager(router, pm, gossipEvents, 200*time.Millisecond, 10000, 10, 3, 50*time.Millisecond, 5)
	gossipMgr.Start()

	return gossipMgr, router, pm
}

func TestGossip_Publish(t *testing.T) {
	gm, _, _ := setupTestGossipManager()
	defer gm.Stop()

	err := gm.Publish("TEST_MSG", MsgBlock{Data: []byte("block1")}, "msg-1")
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	stats := gm.Statistics()
	if stats.PublishedMessages != 1 {
		t.Errorf("Expected 1 published message, got %d", stats.PublishedMessages)
	}
}

func TestGossip_Receive_And_Deduplicate(t *testing.T) {
	gm, _, _ := setupTestGossipManager()
	defer gm.Stop()

	peer := createTestPeer("p1")

	env := GossipEnvelope{
		MessageID: "msg-1",
		TTL:       5,
		Payload:   MsgBlock{Data: []byte("block1")},
	}

	// First receive
	_, err := gm.Receive(env, peer, "GOSSIP")
	if err != nil {
		t.Fatalf("Expected receive to succeed, got %v", err)
	}

	// Duplicate receive
	_, err = gm.Receive(env, peer, "GOSSIP")
	if err == nil {
		t.Fatalf("Expected duplicate error, got nil")
	}

	stats := gm.Statistics()
	if stats.ReceivedMessages != 2 {
		t.Errorf("Expected 2 received messages")
	}
	if stats.DuplicateMessages != 1 {
		t.Errorf("Expected 1 duplicate message")
	}
}

func TestGossip_TTL(t *testing.T) {
	gm, _, _ := setupTestGossipManager()
	defer gm.Stop()

	peer := createTestPeer("p1")

	env := GossipEnvelope{
		MessageID: "msg-1",
		TTL:       0, // Expired
		Payload:   MsgBlock{Data: []byte("block1")},
	}

	_, err := gm.Receive(env, peer, "GOSSIP")
	if err == nil {
		t.Fatalf("Expected TTL expired error")
	}

	stats := gm.Statistics()
	if stats.TTLExpired != 1 {
		t.Errorf("Expected 1 TTL expired")
	}
}

func TestGossip_CacheCleanup(t *testing.T) {
	gm, _, _ := setupTestGossipManager()
	defer gm.Stop()

	env := GossipEnvelope{
		MessageID: "msg-1",
		TTL:       5,
		Payload:   MsgBlock{Data: []byte("block1")},
	}

	_, _ = gm.Receive(env, nil, "GOSSIP")

	if !gm.cache.Contains("msg-1") {
		t.Fatalf("Cache should contain message")
	}

	// Wait for expiration
	time.Sleep(300 * time.Millisecond)

	if gm.cache.Contains("msg-1") {
		t.Fatalf("Cache should have expired the message")
	}
}

func TestGossip_StressTest_Concurrency(t *testing.T) {
	gm, router, pm := setupTestGossipManager()
	defer gm.Stop()
	defer router.Stop()

	// Add a few test peers
	for i := 0; i < 10; i++ {
		pm.AddPeer(createTestPeer(fmt.Sprintf("peer_%d", i)))
	}

	var wg sync.WaitGroup

	// 1000 concurrent receives
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			env := GossipEnvelope{
				MessageID: fmt.Sprintf("msg-%d", idx),
				TTL:       5,
				Payload:   MsgBlock{Data: []byte("data")},
			}
			_, _ = gm.Receive(env, nil, "GOSSIP")
		}(i)
	}

	wg.Wait()

	// Wait for queue to process
	time.Sleep(300 * time.Millisecond)

	stats := gm.Statistics()
	if stats.ReceivedMessages != 1000 {
		t.Errorf("Expected 1000 receives, got %d", stats.ReceivedMessages)
	}
	if stats.ForwardedMessages != 1000 {
		t.Errorf("Expected 1000 forwards, got %d", stats.ForwardedMessages)
	}
}

func TestGossip_ExtractPayload(t *testing.T) {
	raw := map[string]interface{}{
		"data": "aGVsbG8=", // base64 "hello" if bytes
	}

	var block MsgBlock
	// Simulate interface{} conversion inside JSON
	b, _ := json.Marshal(raw)
	var iface interface{}
	json.Unmarshal(b, &iface)

	err := ExtractGossipPayload(iface, &block)
	if err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	if string(block.Data) != "hello" {
		t.Errorf("Expected data 'hello'")
	}
}
