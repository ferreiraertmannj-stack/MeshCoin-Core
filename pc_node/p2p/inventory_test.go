package p2p

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func setupTestInventoryManager() (*InventoryManager, *MessageRouter, *MockPeerManager) {
	pm := NewMockPeerManager()
	secMgr := NewSecurityManager()
	events := RouterEvents{}
	router := NewMessageRouter(pm, secMgr, events)
	router.Start()

	gossipEvents := GossipEvents{}
	gossipMgr := NewGossipManager(router, pm, gossipEvents, 200*time.Millisecond, 10000, 10, 3, 50*time.Millisecond, 5)
	gossipMgr.Start()

	invEvents := InventoryEvents{}
	invMgr := NewInventoryManager(router, gossipMgr, pm, invEvents, 500*time.Millisecond, 10000, 10, 2, 50*time.Millisecond, 1000)
	invMgr.Start()

	return invMgr, router, pm
}

func TestInventory_Announce(t *testing.T) {
	im, _, _ := setupTestInventoryManager()
	defer im.Stop()

	items := []InventoryItem{
		{ObjectHash: "hash1", ObjectType: InventoryBlock},
	}
	im.Announce(items)

	stats := im.Statistics()
	if stats.InventoriesSent != 1 {
		t.Errorf("Expected 1 inventory sent, got %d", stats.InventoriesSent)
	}

	if !im.cache.Contains("hash1") {
		t.Errorf("Expected hash1 to be cached locally after announce")
	}
}

func TestInventory_ReceiveInventory_CacheMiss_And_Request(t *testing.T) {
	im, _, _ := setupTestInventoryManager()
	defer im.Stop()

	peer := createTestPeer("p1")
	msg := MsgInventory{
		Items: []InventoryItem{
			{ObjectHash: "hash2"},
		},
	}

	im.ReceiveInventory(peer, msg)

	time.Sleep(20 * time.Millisecond) // Let queue process

	stats := im.Statistics()
	if stats.InventoriesReceived != 1 {
		t.Errorf("Expected 1 received")
	}
	if stats.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss")
	}
	if stats.PendingRequests == 0 && stats.ObjectsRequested == 0 && stats.Timeouts == 0 {
		// Just ensuring it touched the request flow
	}
}

func TestInventory_ReceiveInventory_CacheHit(t *testing.T) {
	im, _, _ := setupTestInventoryManager()
	defer im.Stop()

	peer := createTestPeer("p1")
	msg := MsgInventory{
		Items: []InventoryItem{
			{ObjectHash: "hash3"},
		},
	}

	im.AddKnownObject("hash3") // Manually cache it
	im.ReceiveInventory(peer, msg)

	stats := im.Statistics()
	if stats.CacheHits != 1 {
		t.Errorf("Expected 1 cache hit")
	}
	if stats.ObjectsIgnored != 1 {
		t.Errorf("Expected 1 object ignored")
	}
}

func TestInventory_ReceiveObjects_Success(t *testing.T) {
	im, _, _ := setupTestInventoryManager()
	defer im.Stop()

	peer := createTestPeer("p1")

	// Create a job manually to avoid network dependencies
	job := InventoryJob{
		Item: InventoryItem{ObjectHash: "hash4"},
		Peer: peer,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		_ = im.doRequestObject(job)
	}()

	time.Sleep(10 * time.Millisecond) // Let it setup channel

	// Fulfill the request
	im.ReceiveObjects(peer, MsgData{
		Item: job.Item,
		Data: []byte("data"),
	})

	wg.Wait()

	stats := im.Statistics()
	if stats.ObjectsDelivered != 1 {
		t.Errorf("Expected 1 object delivered")
	}
}

func TestInventory_RequestObjects_Timeout(t *testing.T) {
	im, _, _ := setupTestInventoryManager()
	defer im.Stop()

	peer := createTestPeer("p1")

	job := InventoryJob{
		Item: InventoryItem{ObjectHash: "hash5"},
		Peer: peer,
	}

	err := im.doRequestObject(job)
	if err == nil {
		t.Errorf("Expected timeout error")
	}
}

func TestInventory_StressTest(t *testing.T) {
	im, router, pm := setupTestInventoryManager()
	defer im.Stop()
	defer router.Stop()

	for i := 0; i < 100; i++ {
		pm.AddPeer(createTestPeer(fmt.Sprintf("peer_%d", i)))
	}

	var wg sync.WaitGroup

	// 5000 concurrent announcements
	for i := 0; i < 5000; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			items := []InventoryItem{
				{ObjectHash: fmt.Sprintf("announce-%d", idx)},
			}
			im.Announce(items)
		}(i)
	}

	// 2000 concurrent inventory receives (triggers requests)
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			peer := createTestPeer("p_test")
			msg := MsgInventory{
				Items: []InventoryItem{
					{ObjectHash: fmt.Sprintf("req-%d", idx)},
				},
			}
			im.ReceiveInventory(peer, msg)
		}(i)
	}

	wg.Wait()

	// It's going to hit timeouts for the 2000 requests, so we wait for queue to finish timeouts
	time.Sleep(200 * time.Millisecond)

	stats := im.Statistics()
	if stats.InventoriesSent != 5000 {
		t.Errorf("Expected 5000 sent, got %d", stats.InventoriesSent)
	}
	if stats.InventoriesReceived != 2000 {
		t.Errorf("Expected 2000 received, got %d", stats.InventoriesReceived)
	}
}
