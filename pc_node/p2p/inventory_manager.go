package p2p

import (
	"fmt"
	"sync"
	"time"
)

type InventoryManager struct {
	cache  *InventoryCache
	queue  *InventoryQueue
	stats  *InventoryStatistics
	events InventoryEvents

	router    *MessageRouter
	gossipMgr *GossipManager
	peerMgr   PeerManager

	maxPerMsg  int
	reqTimeout time.Duration

	pendingReqs map[string]chan MsgData
	mu          sync.RWMutex
}

func NewInventoryManager(
	router *MessageRouter,
	gossipMgr *GossipManager,
	peerMgr PeerManager,
	events InventoryEvents,
	cacheExpiration time.Duration,
	maxQueueSize int,
	maxWorkers int,
	maxRetries int,
	queueTimeout time.Duration,
	maxPerMsg int,
) *InventoryManager {
	stats := &InventoryStatistics{StartTime: time.Now()}
	cache := NewInventoryCache(cacheExpiration)

	m := &InventoryManager{
		cache:       cache,
		stats:       stats,
		events:      events,
		router:      router,
		gossipMgr:   gossipMgr,
		peerMgr:     peerMgr,
		maxPerMsg:   maxPerMsg,
		reqTimeout:  queueTimeout, // Using queueTimeout as request timeout for MsgData
		pendingReqs: make(map[string]chan MsgData),
	}

	m.queue = NewInventoryQueue(maxQueueSize, maxWorkers, maxRetries, queueTimeout, m)

	return m
}

func (m *InventoryManager) Start() {
	m.queue.Start()
}

func (m *InventoryManager) Stop() {
	m.queue.Stop()
	m.cache.Stop()
}

func (m *InventoryManager) Running() bool {
	return m.queue.isRunning
}

func (m *InventoryManager) Announce(items []InventoryItem) {
	if len(items) == 0 {
		return
	}

	// Limit per message
	if len(items) > m.maxPerMsg {
		items = items[:m.maxPerMsg]
	}

	for _, item := range items {
		m.AddKnownObject(item.ObjectHash)
	}

	msg := P2PMessage{
		Type:    MsgTypeInventory,
		Payload: MsgInventory{Items: items},
	}

	if m.router != nil {
		m.router.Broadcast(msg)
	}
	m.stats.IncInventoriesSent()
}

func (m *InventoryManager) ReceiveInventory(peer *Peer, msg MsgInventory) {
	m.stats.IncInventoriesReceived()

	if peer == nil {
		return
	}
	peerID := peer.Info().NodeID

	var unknown []InventoryItem
	for _, item := range msg.Items {
		if m.cache.Contains(item.ObjectHash) {
			m.stats.IncCacheHits()
			m.stats.IncObjectsIgnored()
			if m.events.OnInventoryIgnored != nil {
				go m.events.OnInventoryIgnored(peerID, 1, "already_known")
			}
		} else {
			m.stats.IncCacheMisses()
			unknown = append(unknown, item)
		}
	}

	if m.events.OnInventoryReceived != nil {
		go m.events.OnInventoryReceived(peerID, len(msg.Items))
	}

	if len(unknown) > 0 {
		m.RequestObjects(peer, unknown)
	}
}

func (m *InventoryManager) RequestObjects(peer *Peer, items []InventoryItem) {
	for _, item := range items {
		_ = m.queue.Enqueue(InventoryJob{
			Item:    item,
			Peer:    peer,
			Retries: 0,
		})
	}
}

func (m *InventoryManager) doRequestObject(job InventoryJob) error {
	m.stats.IncObjectsRequested()
	peerID := job.Peer.Info().NodeID

	if m.events.OnObjectRequested != nil {
		go m.events.OnObjectRequested(peerID, job.Item.ObjectHash)
	}

	reqMsg := P2PMessage{
		Type: MsgTypeGetData,
		Payload: MsgGetData{
			Items: []InventoryItem{job.Item},
		},
	}

	start := time.Now()

	// Register channel for response
	respCh := make(chan MsgData, 1)
	m.mu.Lock()
	m.pendingReqs[job.Item.ObjectHash] = respCh
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.pendingReqs, job.Item.ObjectHash)
		m.mu.Unlock()
	}()

	err := m.router.SendToPeer(job.Peer, reqMsg)
	if err != nil {
		return err // queue worker will retry if under limit
	}

	select {
	case <-respCh:
		m.stats.IncObjectsReceived(time.Since(start))

		if m.events.OnObjectReceived != nil {
			go m.events.OnObjectReceived(peerID, job.Item.ObjectHash)
		}

		// Delivery phase (mocked as forwarding to Gossip/Router)
		m.AddKnownObject(job.Item.ObjectHash)
		m.stats.IncObjectsDelivered()

		if m.events.OnObjectDelivered != nil {
			go m.events.OnObjectDelivered(job.Item.ObjectHash)
		}

		if m.events.OnSynchronizationFinished != nil {
			go m.events.OnSynchronizationFinished(peerID, 1)
		}
		return nil

	case <-time.After(m.reqTimeout):
		return fmt.Errorf("request timeout")
	}
}

func (m *InventoryManager) ReceiveObjects(peer *Peer, msg MsgData) {
	m.mu.RLock()
	ch, ok := m.pendingReqs[msg.Item.ObjectHash]
	m.mu.RUnlock()

	if ok {
		select {
		case ch <- msg:
		default:
		}
	}
}

// ReceiveGetData is for handling incoming GET_DATA requests from peers
func (m *InventoryManager) ReceiveGetData(peer *Peer, msg MsgGetData) {
	// In a real implementation, we would query the Ledger or Database.
	// Since we can't alter those, we just simulate answering NotFound
	// or we answer if it's in our mock/cache (which it usually isn't physically there).

	var notFound []InventoryItem

	for _, item := range msg.Items {
		if !m.cache.Contains(item.ObjectHash) {
			notFound = append(notFound, item)
			if m.events.OnObjectNotFound != nil {
				go m.events.OnObjectNotFound(peer.Info().NodeID, item.ObjectHash)
			}
		} else {
			// Object is known, but we don't have the bytes in cache.
			// Ideally we fetch from storage, which we cannot touch.
			// So we'll just ignore for now in this decoupled layer.
		}
	}

	if len(notFound) > 0 {
		resp := P2PMessage{
			Type:    MsgTypeNotFound,
			Payload: MsgNotFound{Items: notFound},
		}
		_ = m.router.SendToPeer(peer, resp)
	}
}

func (m *InventoryManager) PeerHasObject(peerID string, objectHash string) bool {
	// We don't store per-peer inventory in this simple design unless requested.
	return false
}

func (m *InventoryManager) AddKnownObject(objectHash string) {
	if m.cache.Add(objectHash) {
		m.stats.IncObjectsCached()
	}
}

func (m *InventoryManager) RemoveKnownObject(objectHash string) {
	m.cache.Remove(objectHash)
}

func (m *InventoryManager) Statistics() InventoryStatistics {
	return m.stats.Snapshot()
}
