package p2p

import (
	"os"
	"testing"
	"time"
)

func setupDiscoveryEnvironment(filename string) (*DiscoveryManager, *PeerStore, *SeedManager, *SecurityManager) {
	os.Remove(filename)
	store := NewPeerStore(filename)

	seeds := []PeerAddress{
		{IP: "1.1.1.1", Port: 5556},
		{IP: "2.2.2.2", Port: 5556},
	}
	seedMgr := NewSeedManager(seeds)
	secManager := NewSecurityManager()

	events := DiscoveryEventHandlers{}
	dm := NewDiscoveryManager(store, seedMgr, secManager, events, 1000, 10)
	return dm, store, seedMgr, secManager
}

func TestDiscovery_InitialAndSeedNodes(t *testing.T) {
	dm, store, _, _ := setupDiscoveryEnvironment("test_peers_1.json")

	// Rede Vazia
	peers, _ := store.GetAllPeers()
	if len(peers) != 0 {
		t.Fatalf("Expected empty network initially")
	}

	// Trigger Start (automatically bootstraps from seeds if empty)
	dm.Start()

	time.Sleep(100 * time.Millisecond) // Let async queue process

	peersAfter, _ := store.GetAllPeers()
	if len(peersAfter) != 2 {
		t.Fatalf("Expected 2 seed nodes to be discovered, got %d", len(peersAfter))
	}

	dm.Stop()
}

func TestDiscovery_MsgPeers_Deduplication_And_Limits(t *testing.T) {
	dm, store, seedMgr, _ := setupDiscoveryEnvironment("test_peers_2.json")
	seedMgr.seedNodes = []PeerAddress{}

	// Simulate discovering 3 peers (1 valid, 1 duplicate, 1 invalid)
	msg := MsgPeers{
		Peers: []PeerRecord{
			{NodeID: "p1", Address: PeerAddress{IP: "10.0.0.1", Port: 5556}},
			{NodeID: "p1", Address: PeerAddress{IP: "10.0.0.1", Port: 5556}}, // Dup
			{NodeID: "p2", Address: PeerAddress{IP: "", Port: 0}},            // Invalid
		},
	}

	dm.Start()
	dm.OnMsgPeers(msg)

	time.Sleep(100 * time.Millisecond)

	peers, _ := store.GetAllPeers()
	if len(peers) != 1 {
		t.Fatalf("Expected exactly 1 valid unique peer, got %d", len(peers))
	}
	if peers[0].NodeID != "p1" {
		t.Fatalf("Expected p1 to be saved")
	}

	dm.Stop()
}

func TestDiscovery_Blacklist(t *testing.T) {
	dm, store, seedMgr, secManager := setupDiscoveryEnvironment("test_peers_3.json")
	seedMgr.seedNodes = []PeerAddress{}

	// Blacklist an IP
	secManager.BlacklistPeer("10.0.0.99", 5*time.Minute)

	msg := MsgPeers{
		Peers: []PeerRecord{
			{NodeID: "banned", Address: PeerAddress{IP: "10.0.0.99", Port: 5556}},
		},
	}

	dm.Start()
	dm.OnMsgPeers(msg)
	time.Sleep(100 * time.Millisecond)

	peers, _ := store.GetAllPeers()
	if len(peers) != 0 {
		t.Fatalf("Expected blacklisted peer to be dropped, got %d", len(peers))
	}

	dm.Stop()
}

func TestDiscovery_PeerFailure_And_Cleanup(t *testing.T) {
	dm, store, _, _ := setupDiscoveryEnvironment("test_peers_4.json")

	peer := PeerRecord{
		NodeID:  "p_fail",
		Address: PeerAddress{IP: "10.0.0.2", Port: 5556},
	}
	store.SavePeer(peer)

	for i := 0; i < 6; i++ {
		dm.HandlePeerFailure("p_fail")
	}

	// Should be deleted after 6 failures
	_, err := store.GetPeer("p_fail")
	if err == nil {
		t.Fatalf("Expected peer to be deleted after too many failures")
	}
}

func TestDiscovery_AutomaticRecovery(t *testing.T) {
	dm, store, _, _ := setupDiscoveryEnvironment("test_peers_5.json")

	peer := PeerRecord{
		NodeID:      "p_old",
		Address:     PeerAddress{IP: "10.0.0.2", Port: 5556},
		LastSuccess: time.Now().Add(-48 * time.Hour), // 48h ago
	}
	store.SavePeer(peer)

	dm.Start()

	// Manually trigger cleanup
	dm.cleanupLoopOnceForTest()

	_, err := store.GetPeer("p_old")
	if err == nil {
		t.Fatalf("Expected old peer to be cleaned up")
	}

	// Because table is now empty, if we trigger randomWalk (which bootstraps if empty), it recovers
	dm.randomWalkLoopOnceForTest()

	time.Sleep(100 * time.Millisecond) // Let queue process

	peers, _ := store.GetAllPeers()
	if len(peers) != 2 { // 2 seeds
		t.Fatalf("Expected automatic recovery to populate 2 seeds, got %d", len(peers))
	}

	dm.Stop()
}

// Helpers for testing unexported loops synchronously
func (dm *DiscoveryManager) cleanupLoopOnceForTest() {
	dm.queue.ResetSeen()
	peers, err := dm.store.GetAllPeers()
	if err != nil {
		return
	}
	now := time.Now()
	for _, peer := range peers {
		if now.Sub(peer.LastSuccess) > 24*time.Hour && !peer.IsSeed {
			dm.store.DeletePeer(peer.NodeID)
		}
	}
}

func (dm *DiscoveryManager) randomWalkLoopOnceForTest() {
	peers, err := dm.store.GetAllPeers()
	if err != nil || len(peers) == 0 {
		if dm.seedMgr.NeedsSeeds(dm.store) {
			dm.bootstrapFromSeeds()
		}
		return
	}
}
