package p2p

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

// MockPeerManager for testing
type MockPeerManager struct {
	peers map[string]*Peer
	mu    sync.RWMutex
}

func NewMockPeerManager() *MockPeerManager {
	return &MockPeerManager{peers: make(map[string]*Peer)}
}

func (m *MockPeerManager) AddPeer(p *Peer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers[p.Info().NodeID] = p
}

func (m *MockPeerManager) GetConnectedPeers() []*Peer {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var list []*Peer
	for _, p := range m.peers {
		list = append(list, p)
	}
	return list
}

func (m *MockPeerManager) GetPeer(nodeID string) (*Peer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.peers[nodeID]
	return p, ok
}

func (m *MockPeerManager) DisconnectPeer(nodeID string, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.peers[nodeID]
	if ok {
		p.Disconnect(reason)
		delete(m.peers, nodeID)
	}
}

func setupTestRouter() (*MessageRouter, *MockPeerManager) {
	pm := NewMockPeerManager()
	secMgr := NewSecurityManager()
	events := RouterEvents{}
	router := NewMessageRouter(pm, secMgr, events)
	router.Start()
	return router, pm
}

func createTestPeer(nodeID string) *Peer {
	c1, c2 := net.Pipe()

	go func() {
		buf := make([]byte, 1024)
		for {
			_, err := c2.Read(buf)
			if err != nil {
				return
			}
		}
	}()

	config := Config{
		NodeID:            nodeID,
		HeartbeatTimeout:  1 * time.Second,
		HeartbeatInterval: 1 * time.Second,
	}
	peer := NewPeer(c1, config, NewSecurityManager(), PeerEventHandlers{})

	peer.mu.Lock()
	peer.info = PeerInfo{
		NodeID:  nodeID,
		Address: "127.0.0.1",
	}
	peer.mu.Unlock()

	return peer
}

func TestRouter_RegisterRemoveHandler(t *testing.T) {
	router, _ := setupTestRouter()
	defer router.Stop()

	router.RegisterHandler("TEST_MSG", func(peer *Peer, msg P2PMessage) error {
		return nil
	})

	if router.HandlerCount() != 1 {
		t.Errorf("Expected 1 handler")
	}

	router.RemoveHandler("TEST_MSG")
	if router.HandlerCount() != 0 {
		t.Errorf("Expected 0 handlers")
	}
}

func TestRouter_Dispatch_And_Unknown(t *testing.T) {
	router, _ := setupTestRouter()
	defer router.Stop()

	called := false
	router.RegisterHandler("VALID_MSG", func(peer *Peer, msg P2PMessage) error {
		called = true
		return nil
	})

	peer := createTestPeer("node1")

	err := router.Dispatch(peer, P2PMessage{Type: "VALID_MSG"})
	if err != nil || !called {
		t.Errorf("Expected dispatch to succeed")
	}

	err = router.Dispatch(peer, P2PMessage{Type: "UNKNOWN_MSG"})
	if err == nil {
		t.Errorf("Expected error for unknown message")
	}

	stats := router.GetStatistics()
	if stats.MessagesDispatched != 1 || stats.UnknownMessages != 1 {
		t.Errorf("Invalid stats")
	}
}

func TestRouter_PanicRecovery(t *testing.T) {
	router, _ := setupTestRouter()
	defer router.Stop()

	router.RegisterHandler("PANIC_MSG", func(peer *Peer, msg P2PMessage) error {
		panic("test panic")
	})

	peer := createTestPeer("node1")
	err := router.Dispatch(peer, P2PMessage{Type: "PANIC_MSG"})

	if err == nil {
		t.Errorf("Expected error from panic recovery")
	}

	stats := router.GetStatistics()
	if stats.DispatchErrors != 1 {
		t.Errorf("Expected 1 dispatch error")
	}
}

func TestRouter_Broadcast_And_DirectSend(t *testing.T) {
	router, pm := setupTestRouter()
	defer router.Stop()

	p1 := createTestPeer("node1")
	p2 := createTestPeer("node2")
	pm.AddPeer(p1)
	pm.AddPeer(p2)

	router.Broadcast(P2PMessage{Type: "BCAST"})

	time.Sleep(50 * time.Millisecond)

	stats := router.GetStatistics()
	if stats.Broadcasts != 1 {
		t.Errorf("Expected 1 broadcast")
	}

	err := router.SendToPeer(p1, P2PMessage{Type: "DIRECT"})
	if err != nil {
		t.Errorf("Direct send failed")
	}
}

func TestRouter_StressTest_Concurrency(t *testing.T) {
	router, _ := setupTestRouter()
	defer router.Stop()

	router.RegisterHandler("STRESS", func(peer *Peer, msg P2PMessage) error {
		time.Sleep(2 * time.Millisecond) // Slow handler
		return nil
	})

	peer := createTestPeer("stress_node")

	var wg sync.WaitGroup
	// 500 goroutines, each sends 2 messages (1000 messages total)
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = router.Dispatch(peer, P2PMessage{Type: "STRESS", Payload: idx})
			_ = router.Dispatch(peer, P2PMessage{Type: "STRESS", Payload: idx})
		}(i)
	}

	// Concurrent registration and removal
	go func() {
		for i := 0; i < 100; i++ {
			router.RegisterHandler(fmt.Sprintf("TEMP_%d", i), func(peer *Peer, msg P2PMessage) error { return nil })
			router.RemoveHandler(fmt.Sprintf("TEMP_%d", i))
		}
	}()

	wg.Wait()

	stats := router.GetStatistics()
	if stats.MessagesDispatched != 1000 {
		t.Errorf("Expected 1000 dispatched, got %d", stats.MessagesDispatched)
	}
}

func TestRouter_MessageLoop_Shutdown(t *testing.T) {
	router, _ := setupTestRouter()
	peer := createTestPeer("loop_node")

	go MessageLoop(peer, router)

	router.Stop()
	time.Sleep(50 * time.Millisecond) // Give time to exit

	if router.Running() {
		t.Errorf("Expected router to be stopped")
	}
	// The message loop should have exited due to router.ctx.Done()

	// Close peer as well
	peer.Disconnect("test")
}

func TestRouter_HandlerError(t *testing.T) {
	router, _ := setupTestRouter()
	defer router.Stop()

	router.RegisterHandler("ERR_MSG", func(peer *Peer, msg P2PMessage) error {
		return errors.New("handler specific error")
	})

	peer := createTestPeer("node1")
	err := router.Dispatch(peer, P2PMessage{Type: "ERR_MSG"})

	if err == nil {
		t.Errorf("Expected error from handler")
	}

	stats := router.GetStatistics()
	if stats.DispatchErrors != 1 {
		t.Errorf("Expected 1 dispatch error")
	}
}
