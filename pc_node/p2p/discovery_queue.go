package p2p

import (
	"context"
	"sync"
)

type PeerDiscoveryQueue struct {
	queue     chan PeerRecord
	mu        sync.Mutex
	seenPeers map[string]bool

	maxQueueSize int
}

func NewPeerDiscoveryQueue(maxQueueSize int) *PeerDiscoveryQueue {
	return &PeerDiscoveryQueue{
		queue:        make(chan PeerRecord, maxQueueSize),
		seenPeers:    make(map[string]bool),
		maxQueueSize: maxQueueSize,
	}
}

// Enqueue adds a peer to the discovery queue if it hasn't been seen recently and the queue isn't full.
func (q *PeerDiscoveryQueue) Enqueue(peer PeerRecord) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Deduplication
	if q.seenPeers[peer.NodeID] {
		return false
	}

	select {
	case q.queue <- peer:
		q.seenPeers[peer.NodeID] = true
		return true
	default:
		return false // Queue full
	}
}

// ProcessAsync starts worker goroutines to process the discovery queue.
func (q *PeerDiscoveryQueue) ProcessAsync(ctx context.Context, workers int, handler func(peer PeerRecord)) {
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case peer := <-q.queue:
					handler(peer)
				}
			}
		}()
	}
}

// ResetSeen clears the deduplication map, usually called periodically
func (q *PeerDiscoveryQueue) ResetSeen() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.seenPeers = make(map[string]bool)
}

// Clear removes all pending peers
func (q *PeerDiscoveryQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Drain channel
	for {
		select {
		case <-q.queue:
		default:
			return
		}
	}
}
