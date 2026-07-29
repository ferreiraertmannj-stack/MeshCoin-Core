package p2p

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PeerManager defines the interface for retrieving active peers.
type PeerManager interface {
	GetConnectedPeers() []*Peer
	GetPeer(nodeID string) (*Peer, bool)
	DisconnectPeer(nodeID string, reason string)
}

type MessageRouter struct {
	registry   *MessageRegistry
	dispatcher *MessageDispatcher
	stats      *RouterStatistics
	events     RouterEvents

	peerMgr PeerManager
	secMgr  *SecurityManager

	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
	mu        sync.RWMutex
}

func NewMessageRouter(peerMgr PeerManager, secMgr *SecurityManager, events RouterEvents) *MessageRouter {
	stats := &RouterStatistics{StartTime: time.Now()}
	registry := NewMessageRegistry()
	dispatcher := NewMessageDispatcher(registry, stats, events)

	return &MessageRouter{
		registry:   registry,
		dispatcher: dispatcher,
		stats:      stats,
		events:     events,
		peerMgr:    peerMgr,
		secMgr:     secMgr,
	}
}

func (r *MessageRouter) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.isRunning {
		return
	}

	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.isRunning = true

	if r.events.OnRouterStarted != nil {
		go r.events.OnRouterStarted()
	}
}

func (r *MessageRouter) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.isRunning {
		return
	}

	r.cancel()
	r.isRunning = false

	if r.events.OnRouterStopped != nil {
		go r.events.OnRouterStopped()
	}
}

func (r *MessageRouter) Running() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.isRunning
}

func (r *MessageRouter) RegisterHandler(msgType string, handler MessageHandler) {
	r.registry.RegisterHandler(msgType, handler)
}

func (r *MessageRouter) RemoveHandler(msgType string) {
	r.registry.RemoveHandler(msgType)
}

func (r *MessageRouter) HandlerCount() int {
	return r.registry.HandlerCount()
}

func (r *MessageRouter) Dispatch(peer *Peer, msg P2PMessage) error {
	if !r.Running() {
		return fmt.Errorf("router is not running")
	}
	return r.dispatcher.Dispatch(peer, msg)
}

func (r *MessageRouter) Broadcast(msg P2PMessage) {
	if !r.Running() || r.peerMgr == nil {
		return
	}

	r.stats.AddBroadcast()
	peers := r.peerMgr.GetConnectedPeers()

	for _, peer := range peers {
		// Do not block broadcast if one peer blocks
		go func(p *Peer) {
			err := r.SendToPeer(p, msg)
			if err != nil {
				// Handle broadcast failure
				nodeID := p.Info().NodeID
				// e.g. apply SecurityManager penalties or remove peer
				// In a real scenario, network issues happen, but for malicious behavior we'd penalize.
				// We'll disconnect if we can't send.
				r.peerMgr.DisconnectPeer(nodeID, "broadcast send failed")
			}
		}(peer)
	}
}

func (r *MessageRouter) SendToPeer(peer *Peer, msg P2PMessage) error {
	if peer == nil {
		return fmt.Errorf("peer is nil")
	}

	if r.secMgr != nil && peer.Info().Address != "" {
		ip := peer.Info().Address
		// Not strictly required to check rate limit on our own sends, but good practice
		if r.secMgr.IsBlacklisted(ip) {
			return fmt.Errorf("peer is blacklisted")
		}
	}

	// We utilize the peer's own SendMessage which has a WriteDeadline internally
	err := peer.SendMessage(msg.Type, msg.Payload)
	if err != nil {
		return err
	}

	r.stats.AddSent()
	return nil
}

func (r *MessageRouter) GetStatistics() RouterStatistics {
	return r.stats.Snapshot()
}
