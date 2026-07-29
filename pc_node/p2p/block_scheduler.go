package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type BlockRequest struct {
	Hash     string
	Height   uint64
	Retries  int
	Assigned string // PeerID
	SentAt   time.Time
}

type BlockRequestScheduler struct {
	pendingRequests map[string]*BlockRequest // hash -> request
	queue           []*BlockRequest          // Unassigned requests
	mu              sync.RWMutex

	router      *MessageRouter
	peerManager PeerManager
	stats       *BlockchainSyncStatistics
	events      BlockchainSyncEvents

	timeout    time.Duration
	maxRetries int
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewBlockRequestScheduler(
	router *MessageRouter,
	pm PeerManager,
	stats *BlockchainSyncStatistics,
	events BlockchainSyncEvents,
	timeout time.Duration,
	maxRetries int,
) *BlockRequestScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &BlockRequestScheduler{
		pendingRequests: make(map[string]*BlockRequest),
		queue:           make([]*BlockRequest, 0),
		router:          router,
		peerManager:     pm,
		stats:           stats,
		events:          events,
		timeout:         timeout,
		maxRetries:      maxRetries,
		ctx:             ctx,
		cancel:          cancel,
	}
}

func (s *BlockRequestScheduler) Start() {
	go s.loop()
}

func (s *BlockRequestScheduler) Stop() {
	s.cancel()
}

func (s *BlockRequestScheduler) Schedule(hash string, height uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pendingRequests[hash]; exists {
		return // Already pending
	}

	req := &BlockRequest{
		Hash:    hash,
		Height:  height,
		Retries: 0,
	}
	s.pendingRequests[hash] = req
	s.queue = append(s.queue, req)
}

func (s *BlockRequestScheduler) MarkReceived(hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pendingRequests, hash)

	// Remove from queue if it was there
	for i, req := range s.queue {
		if req.Hash == hash {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			break
		}
	}
}

func (s *BlockRequestScheduler) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pendingRequests)
}

func (s *BlockRequestScheduler) loop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.processQueue()
			s.checkTimeouts()
		}
	}
}

func (s *BlockRequestScheduler) processQueue() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.queue) == 0 {
		return
	}

	peers := s.peerManager.GetConnectedPeers()
	if len(peers) == 0 {
		return
	}

	// Simple round-robin or random distribution
	peerIdx := 0

	var remaining []*BlockRequest
	for _, req := range s.queue {
		peer := peers[peerIdx%len(peers)]
		peerIdx++

		req.Assigned = peer.Info().NodeID
		req.SentAt = time.Now()

		// Construct and send message
		msg := P2PMessage{
			Type: MsgTypeGetData,
			Payload: MsgGetData{
				Items: []InventoryItem{{ObjectType: InventoryBlock, ObjectHash: req.Hash}},
			},
		}

		err := s.router.SendToPeer(peer, msg)
		if err != nil {
			// Failed to send, keep in queue
			req.Assigned = ""
			remaining = append(remaining, req)
		} else {
			s.stats.IncBlocksRequested(1)
			if s.events.OnBlocksRequested != nil {
				go s.events.OnBlocksRequested(peer.Info().NodeID, 1)
			}
		}
	}

	s.queue = remaining
}

func (s *BlockRequestScheduler) checkTimeouts() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for hash, req := range s.pendingRequests {
		if req.Assigned != "" && now.Sub(req.SentAt) > s.timeout {
			req.Retries++
			if req.Retries > s.maxRetries {
				// Drop it or trigger sync failed
				delete(s.pendingRequests, hash)
				if s.events.OnSyncFailed != nil {
					go s.events.OnSyncFailed(req.Assigned, fmt.Sprintf("timeout requesting block %s", hash))
				}
			} else {
				// Re-queue
				req.Assigned = ""
				s.queue = append(s.queue, req)
			}
		}
	}
}
