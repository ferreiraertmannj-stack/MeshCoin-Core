package sync

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPeerPool_BasicOperations(t *testing.T) {
	pool := NewPeerPool()

	if pool.PeerCount() != 0 {
		t.Fatalf("Expected 0 peers, got %d", pool.PeerCount())
	}
	if pool.BestPeer() != nil {
		t.Fatalf("Expected nil BestPeer on empty pool")
	}
	if pool.FastestPeer() != nil {
		t.Fatalf("Expected nil FastestPeer on empty pool")
	}
	if pool.RandomPeer() != nil {
		t.Fatalf("Expected nil RandomPeer on empty pool")
	}

	p1 := NewTCPPeer("peer1", nil, 100)
	p2 := NewTCPPeer("peer2", nil, 200)

	// Simulate properties
	p1.UpdateLatency(20 * time.Millisecond)  // Faster
	p2.UpdateLatency(100 * time.Millisecond) // Higher

	pool.AddPeer(p1)
	pool.AddPeer(p2)

	if pool.PeerCount() != 2 {
		t.Fatalf("Expected 2 peers, got %d", pool.PeerCount())
	}

	fastest := pool.FastestPeer()
	if fastest == nil || fastest.ID() != "peer1" {
		t.Fatalf("Expected peer1 to be fastest")
	}

	highest := pool.HighestPeer()
	if highest == nil || highest.ID() != "peer2" {
		t.Fatalf("Expected peer2 to be highest")
	}

	best := pool.BestPeer()
	// p2 has height 200 (score +800) latency 100
	// p1 has height 100 (score +400) latency 20
	// p2 is expected to be best
	if best == nil || best.ID() != "peer2" {
		t.Fatalf("Expected peer2 to be best")
	}

	// Make p2 fail a lot to reduce its score
	p2.AddFailure()
	p2.AddFailure()
	p2.AddFailure()
	p2.AddFailure()
	p2.AddFailure() // 5 failures = -500 score

	bestNow := pool.BestPeer()
	if bestNow == nil || bestNow.ID() != "peer1" {
		t.Fatalf("Expected peer1 to become best after p2 failures")
	}

	list := pool.ListPeers()
	if len(list) != 2 {
		t.Fatalf("ListPeers should return 2 peers")
	}

	pool.RemovePeer("peer1")
	if pool.PeerCount() != 1 {
		t.Fatalf("Expected 1 peer after removal")
	}

	// Peer inexistente
	pool.RemovePeer("peer_inexistente")
	if pool.PeerCount() != 1 {
		t.Fatalf("Expected count to remain 1 after fake removal")
	}
}

func TestPeerPool_Concurrency(t *testing.T) {
	pool := NewPeerPool()
	var wg sync.WaitGroup

	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			peerID := fmt.Sprintf("peer_%d", idx)
			peer := NewTCPPeer(peerID, nil, uint64(idx*10))

			pool.AddPeer(peer)
			_ = pool.BestPeer()
			_ = pool.RandomPeer()
			_ = pool.ListPeers()

			if idx%2 == 0 {
				pool.RemovePeer(peerID)
			}
		}(i)
	}

	wg.Wait()

	if pool.PeerCount() != numGoroutines/2 {
		t.Fatalf("Expected %d peers remaining, got %d", numGoroutines/2, pool.PeerCount())
	}
}
