package p2p

import (
	"net"
	"sync"
	"testing"
	"time"
)

func setupTestNodes() (*Peer, *Peer, net.Conn, net.Conn, *SecurityManager) {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	var s net.Conn

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s, _ = listener.Accept()
		wg.Done()
	}()

	c, _ := net.Dial("tcp", listener.Addr().String())
	wg.Wait()
	listener.Close()

	secManager := NewSecurityManager()

	config1 := Config{
		NodeID:            "node1",
		Version:           "1.0.0",
		Agent:             "NebulaCore",
		ProtocolVersion:   1,
		NetworkID:         "nebula-testnet",
		GenesisHash:       "gen123",
		HeartbeatInterval: 50 * time.Millisecond,
		HeartbeatTimeout:  200 * time.Millisecond,
	}

	config2 := Config{
		NodeID:            "node2",
		Version:           "1.0.0",
		Agent:             "NebulaCore",
		ProtocolVersion:   1,
		NetworkID:         "nebula-testnet",
		GenesisHash:       "gen123",
		HeartbeatInterval: 50 * time.Millisecond,
		HeartbeatTimeout:  200 * time.Millisecond,
	}

	peer1 := NewPeer(s, config1, secManager, PeerEventHandlers{})
	peer2 := NewPeer(c, config2, secManager, PeerEventHandlers{})

	return peer1, peer2, s, c, secManager
}

func TestHandshake_Valid(t *testing.T) {
	p1, p2, _, _, _ := setupTestNodes()

	var wg sync.WaitGroup
	wg.Add(2)

	opts := HandshakeValidationOptions{
		ExpectedNetworkID:      "nebula-testnet",
		ExpectedGenesisHash:    "gen123",
		MinimumProtocolVersion: 1,
	}

	go func() {
		defer wg.Done()
		err := p1.DoHandshake(opts)
		if err != nil {
			t.Errorf("Server handshake failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		err := p2.DoClientHandshake(opts)
		if err != nil {
			t.Errorf("Client handshake failed: %v", err)
		}
	}()

	wg.Wait()

	if p1.Info().NodeID != "node2" {
		t.Errorf("p1 didn't get node2 ID")
	}
	if p2.Info().NodeID != "node1" {
		t.Errorf("p2 didn't get node1 ID")
	}
}

func TestHandshake_DifferentNetwork(t *testing.T) {
	p1, p2, _, _, _ := setupTestNodes()
	p2.config.NetworkID = "wrong-network"

	opts := HandshakeValidationOptions{
		ExpectedNetworkID:      "nebula-testnet",
		ExpectedGenesisHash:    "gen123",
		MinimumProtocolVersion: 1,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = p1.DoHandshake(opts) // Should fail
	}()

	go func() {
		defer wg.Done()
		err := p2.DoClientHandshake(opts)
		if err == nil {
			t.Errorf("Client handshake should have failed due to wrong network")
		}
	}()

	wg.Wait()
}

func TestHandshake_DifferentGenesis(t *testing.T) {
	p1, p2, _, _, _ := setupTestNodes()
	p2.config.GenesisHash = "wrong-genesis"

	opts := HandshakeValidationOptions{
		ExpectedNetworkID:      "nebula-testnet",
		ExpectedGenesisHash:    "gen123",
		MinimumProtocolVersion: 1,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = p1.DoHandshake(opts)
	}()

	go func() {
		defer wg.Done()
		err := p2.DoClientHandshake(opts)
		if err == nil {
			t.Errorf("Client handshake should have failed due to wrong genesis")
		}
	}()

	wg.Wait()
}

func TestHandshake_IncompatibleVersion(t *testing.T) {
	p1, p2, _, _, _ := setupTestNodes()
	p2.config.ProtocolVersion = 0

	opts := HandshakeValidationOptions{
		ExpectedNetworkID:      "nebula-testnet",
		ExpectedGenesisHash:    "gen123",
		MinimumProtocolVersion: 1,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = p1.DoHandshake(opts)
	}()

	go func() {
		defer wg.Done()
		err := p2.DoClientHandshake(opts)
		if err == nil {
			t.Errorf("Client handshake should have failed due to protocol version")
		}
	}()

	wg.Wait()
}

func TestHandshake_DuplicateNodeID(t *testing.T) {
	p1, p2, _, _, _ := setupTestNodes()
	p2.config.NodeID = "node1" // duplicate

	opts := HandshakeValidationOptions{
		ExpectedNetworkID:      "nebula-testnet",
		ExpectedGenesisHash:    "gen123",
		MinimumProtocolVersion: 1,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_ = p1.DoHandshake(opts)
	}()

	go func() {
		defer wg.Done()
		err := p2.DoClientHandshake(opts)
		if err == nil {
			t.Errorf("Client handshake should have failed due to duplicate node id")
		}
	}()

	wg.Wait()
}

func TestHeartbeat_And_Timeout(t *testing.T) {
	p1, p2, _, _, _ := setupTestNodes()

	opts := HandshakeValidationOptions{
		ExpectedNetworkID:      "nebula-testnet",
		ExpectedGenesisHash:    "gen123",
		MinimumProtocolVersion: 1,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = p1.DoHandshake(opts)
	}()
	go func() {
		defer wg.Done()
		_ = p2.DoClientHandshake(opts)
	}()
	wg.Wait()

	// Wait for heartbeats
	time.Sleep(150 * time.Millisecond)

	if p1.Info().LastPong.IsZero() {
		t.Errorf("Expected LastPong to be updated after heartbeat")
	}

	// Close p2 to force timeout on p1
	p2.Disconnect("test")

	// Wait for timeout
	time.Sleep(300 * time.Millisecond)

	// Context should be cancelled due to timeout
	if p1.ctx.Err() == nil {
		t.Errorf("Expected p1 to disconnect due to heartbeat timeout")
	}
}
