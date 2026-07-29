package p2p

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type GossipManager struct {
	cache  *GossipCache
	queue  *GossipQueue
	stats  *GossipStatistics
	events GossipEvents

	router  *MessageRouter
	peerMgr PeerManager

	defaultTTL int

	mu sync.RWMutex
}

func NewGossipManager(
	router *MessageRouter,
	peerMgr PeerManager,
	events GossipEvents,
	cacheExpiration time.Duration,
	maxQueueSize int,
	maxWorkers int,
	maxRetries int,
	queueTimeout time.Duration,
	defaultTTL int,
) *GossipManager {
	stats := &GossipStatistics{StartTime: time.Now()}
	cache := NewGossipCache(cacheExpiration)

	m := &GossipManager{
		cache:      cache,
		stats:      stats,
		events:     events,
		router:     router,
		peerMgr:    peerMgr,
		defaultTTL: defaultTTL,
	}

	m.queue = NewGossipQueue(maxQueueSize, maxWorkers, maxRetries, queueTimeout, m)

	return m
}

func (m *GossipManager) Start() {
	m.queue.Start()

	// Register base generic handler for gossip wrapper if possible,
	// but normally gossip envelopes are inside standard types.
	// We'll expose Receive to be called by router handlers.
}

func (m *GossipManager) Stop() {
	m.queue.Stop()
	m.cache.Stop()
}

func (m *GossipManager) Running() bool {
	return m.queue.isRunning
}

// Publish is used by the local node to originate a new gossip message
func (m *GossipManager) Publish(msgType string, payload interface{}, msgID string) error {
	m.stats.IncPublished()

	if m.events.OnMessagePublished != nil {
		go m.events.OnMessagePublished(msgID, msgType)
	}

	env := GossipEnvelope{
		MessageID:  msgID,
		Timestamp:  time.Now().UnixNano(),
		TTL:        m.defaultTTL,
		HopCount:   0,
		OriginNode: "local", // local node id could be injected
		Payload:    payload,
	}

	// Add to our own cache so we don't process echoes
	m.cache.Add(msgID)

	// Forward to network
	return m.Forward(env, nil)
}

// Receive is called by the Router handlers when a GossipEnvelope arrives
func (m *GossipManager) Receive(env GossipEnvelope, peer *Peer, msgType string) (interface{}, error) {
	m.stats.IncReceived()

	peerID := ""
	if peer != nil {
		peerID = peer.Info().NodeID
	}

	// 1. Validar envelope
	if env.MessageID == "" {
		m.stats.IncDropped()
		return nil, fmt.Errorf("invalid message ID")
	}

	// 2. Verificar cache
	if !m.cache.Add(env.MessageID) {
		m.stats.IncDuplicate()
		m.stats.IncCacheHits()
		if m.events.OnDuplicateMessage != nil {
			go m.events.OnDuplicateMessage(env.MessageID, peerID)
		}
		return nil, fmt.Errorf("duplicate message")
	}
	m.stats.IncCacheMisses()

	if m.events.OnMessageReceived != nil {
		go m.events.OnMessageReceived(env.MessageID, msgType, peerID)
	}

	// 3. Verificar TTL
	if env.TTL <= 0 {
		m.stats.IncTTLExpired()
		if m.events.OnTTLExpired != nil {
			go m.events.OnTTLExpired(env.MessageID)
		}
		return nil, fmt.Errorf("ttl expired")
	}

	// 6. Encaminhar aos demais peers
	// Note: We dispatch to the queue which handles it asynchronously
	_ = m.Forward(env, peer)

	// 5. Entregar localmente (Return the inner payload to the caller)
	return env.Payload, nil
}

// Forward enqueues the message to be broadcasted
func (m *GossipManager) Forward(env GossipEnvelope, peer *Peer) error {
	return m.queue.Enqueue(GossipJob{
		Envelope: env,
		Peer:     peer,
	})
}

// doForward is the actual logic executed by queue workers
func (m *GossipManager) doForward(env GossipEnvelope, peer *Peer) error {
	start := time.Now()

	env.TTL--
	env.HopCount++

	if env.TTL <= 0 {
		m.stats.IncTTLExpired()
		if m.events.OnTTLExpired != nil {
			go m.events.OnTTLExpired(env.MessageID)
		}
		return nil
	}

	peers := m.peerMgr.GetConnectedPeers()
	m.stats.UpdateActivePeers(len(peers))

	count := 0
	for _, p := range peers {
		// 7. Nunca reenviar ao peer de origem
		if peer != nil && p.Info().NodeID == peer.Info().NodeID {
			if m.events.OnPeerIgnored != nil {
				go m.events.OnPeerIgnored(p.Info().NodeID, "origin_peer")
			}
			continue
		}

		// Convert back to P2PMessage wrapping GossipEnvelope
		msg := P2PMessage{
			// Since GossipEnvelope doesn't store the exact MsgType natively,
			// typically we wrap it or infer it.
			// For this architecture, we assume the router knows, or we encode the type in the payload.
			// Let's assume the router broadcast accepts the generic envelope.
			Type:    "GOSSIP", // Fallback, real implementation might inject original type
			Payload: env,
		}

		_ = m.router.SendToPeer(p, msg)
		count++
	}

	m.stats.IncForwarded(time.Since(start))

	if m.events.OnMessageForwarded != nil {
		go m.events.OnMessageForwarded(env.MessageID, "GOSSIP", count)
	}

	if m.events.OnPropagationFinished != nil {
		go m.events.OnPropagationFinished(env.MessageID)
	}

	return nil
}

func (m *GossipManager) AddPeer(peer *Peer) {
	// Not strictly necessary since we query PeerManager directly, but keeping interface
}

func (m *GossipManager) RemovePeer(nodeID string) {
	// Same
}

func (m *GossipManager) PeerCount() int {
	return len(m.peerMgr.GetConnectedPeers())
}

func (m *GossipManager) Statistics() GossipStatistics {
	return m.stats.Snapshot()
}

// ExtractGossipPayload is a helper to unmarshal the generic interface back to struct
func ExtractGossipPayload(raw interface{}, out interface{}) error {
	b, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
