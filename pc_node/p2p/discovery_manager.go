package p2p

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type DiscoveryManager struct {
	store      *PeerStore
	queue      *PeerDiscoveryQueue
	seedMgr    *SeedManager
	secManager *SecurityManager
	events     DiscoveryEventHandlers

	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex

	// Configs
	maxKnownPeers  int
	maxConnections int
}

func NewDiscoveryManager(store *PeerStore, seedMgr *SeedManager, secManager *SecurityManager, events DiscoveryEventHandlers, maxKnown int, maxConn int) *DiscoveryManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DiscoveryManager{
		store:          store,
		queue:          NewPeerDiscoveryQueue(1000),
		seedMgr:        seedMgr,
		secManager:     secManager,
		events:         events,
		ctx:            ctx,
		cancel:         cancel,
		maxKnownPeers:  maxKnown,
		maxConnections: maxConn,
	}
}

func (dm *DiscoveryManager) Start() {
	// Start async queue processor (e.g. 3 workers for attempting connections)
	dm.queue.ProcessAsync(dm.ctx, 3, dm.processDiscoveredPeer)

	go dm.randomWalkLoop()
	go dm.cleanupLoop()

	if dm.seedMgr.NeedsSeeds(dm.store) {
		dm.bootstrapFromSeeds()
	}
}

func (dm *DiscoveryManager) Stop() {
	dm.cancel()
}

func (dm *DiscoveryManager) bootstrapFromSeeds() {
	for i, seed := range dm.seedMgr.GetAllSeeds() {
		dm.queue.Enqueue(PeerRecord{
			NodeID:  fmt.Sprintf("seed_%d", i),
			Address: seed,
			IsSeed:  true,
		})
	}
}

func (dm *DiscoveryManager) processDiscoveredPeer(peer PeerRecord) {
	// Validation
	if peer.Address.IP == "" || peer.Address.Port <= 0 {
		return
	}

	// Blacklist
	ipStr := peer.Address.IP
	if dm.secManager.IsBlacklisted(ipStr) {
		return
	}

	// Rate limiting connection attempts
	if !dm.secManager.AllowConnection(ipStr) {
		return
	}

	if dm.events.OnPeerValidated != nil {
		dm.events.OnPeerValidated(peer)
	}

	// In a real network, we would dial the peer here using TCP.
	// For architecture decoupling, we simulate the Dial logic or emit an event
	// that triggers the node's Network package to Dial.
	// The prompt states: "Toda implementação deverá ocorrer apenas na camada de descoberta"
	// So we don't dial real sockets here directly if it couples to network.go, but we can emit OnPeerConnected.

	// Simulate connection attempt success / failure
	success := dm.attemptConnection(peer)

	if success {
		peer.LastSuccess = time.Now()
		peer.Reliability += 1.0
		peer.Failures = 0
		dm.store.SavePeer(peer)
		if dm.events.OnPeerConnected != nil {
			dm.events.OnPeerConnected(peer.NodeID, fmt.Sprintf("%s:%d", peer.Address.IP, peer.Address.Port))
		}
	} else {
		dm.HandlePeerFailure(peer.NodeID)
	}
}

func (dm *DiscoveryManager) attemptConnection(peer PeerRecord) bool {
	// Simulated. Real connection dialing is outside the scope of Discovery Toplogy,
	// or it can be mocked in tests.
	return true
}

func (dm *DiscoveryManager) HandlePeerFailure(nodeID string) {
	if nodeID == "" {
		return
	}
	peer, err := dm.store.GetPeer(nodeID)
	if err != nil {
		return
	}

	peer.Failures++
	peer.Reliability -= 2.0

	if peer.Failures > 5 {
		dm.store.DeletePeer(nodeID)
		if dm.events.OnPeerRemoved != nil {
			dm.events.OnPeerRemoved(nodeID, "too many failures")
		}
	} else {
		dm.store.SavePeer(peer)
	}
}

// randomWalkLoop periodically asks known peers for new peers
func (dm *DiscoveryManager) randomWalkLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			peers, err := dm.store.GetAllPeers()
			if err != nil || len(peers) == 0 {
				if dm.seedMgr.NeedsSeeds(dm.store) {
					dm.bootstrapFromSeeds()
				}
				continue
			}

			// Randomly select a peer to ask
			target := peers[rand.Intn(len(peers))]

			// Emulate sending MsgGetPeers to the target
			// The actual network send would be handled by the Peer abstraction
			if dm.events.OnPeerDiscovered != nil {
				// We trigger the event to show activity
				dm.events.OnPeerDiscovered(fmt.Sprintf("%s:%d", target.Address.IP, target.Address.Port))
			}
		}
	}
}

// cleanupLoop removes dead, banned, or old peers automatically
func (dm *DiscoveryManager) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-dm.ctx.Done():
			return
		case <-ticker.C:
			dm.queue.ResetSeen() // Clear deduplication map periodically

			peers, err := dm.store.GetAllPeers()
			if err != nil {
				continue
			}

			now := time.Now()
			for _, peer := range peers {
				// If peer hasn't been successful in 24 hours
				if now.Sub(peer.LastSuccess) > 24*time.Hour && !peer.IsSeed {
					dm.store.DeletePeer(peer.NodeID)
					if dm.events.OnPeerExpired != nil {
						dm.events.OnPeerExpired(peer.NodeID)
					}
					continue
				}

				// If peer is blacklisted
				if dm.secManager.IsBlacklisted(peer.Address.IP) {
					dm.store.DeletePeer(peer.NodeID)
					if dm.events.OnPeerRemoved != nil {
						dm.events.OnPeerRemoved(peer.NodeID, "blacklisted")
					}
					continue
				}
			}
		}
	}
}

// OnMsgPeers handles receiving a MsgPeers from a remote node
func (dm *DiscoveryManager) OnMsgPeers(msg MsgPeers) {
	for _, peer := range msg.Peers {
		// Verify limits
		peersInStore, _ := dm.store.GetAllPeers()
		if len(peersInStore) >= dm.maxKnownPeers {
			break
		}

		dm.queue.Enqueue(peer)
		if dm.events.OnPeerDiscovered != nil {
			dm.events.OnPeerDiscovered(fmt.Sprintf("%s:%d", peer.Address.IP, peer.Address.Port))
		}
	}
	if dm.events.OnDiscoveryFinished != nil {
		dm.events.OnDiscoveryFinished(len(msg.Peers))
	}
}
