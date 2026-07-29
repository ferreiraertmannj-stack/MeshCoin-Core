package p2p

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"time"
)

// Config holds the local node settings for handshake validation
type Config struct {
	NodeID          string
	Version         string
	Agent           string
	ProtocolVersion int
	NetworkID       string
	GenesisHash     string
	Capabilities    []Capability
	Height          uint64

	HeartbeatInterval time.Duration
	HeartbeatTimeout  time.Duration
}

// Peer represents a connected and authenticated P2P peer.
type Peer struct {
	conn       net.Conn
	info       PeerInfo
	secManager *SecurityManager
	events     PeerEventHandlers
	config     Config

	mu         sync.RWMutex
	cancelFunc context.CancelFunc
	ctx        context.Context

	encoder *json.Encoder
	decoder *json.Decoder

	writeMu sync.Mutex
}

func NewPeer(conn net.Conn, config Config, secManager *SecurityManager, events PeerEventHandlers) *Peer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Peer{
		conn:       conn,
		secManager: secManager,
		events:     events,
		config:     config,
		ctx:        ctx,
		cancelFunc: cancel,
		encoder:    json.NewEncoder(conn),
		decoder:    json.NewDecoder(conn),
	}
}

func (p *Peer) SendMessage(msgType string, payload interface{}) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	_ = p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	msg := P2PMessage{
		Type:    msgType,
		Payload: payload,
	}
	return p.encoder.Encode(msg)
}

func (p *Peer) Disconnect(reason string) {
	p.cancelFunc()
	p.mu.Lock()
	if p.info.NodeID != "" {
		p.secManager.UnregisterNodeID(p.info.NodeID)
	}
	p.mu.Unlock()

	_ = p.SendMessage(MsgTypeDisconnect, MsgDisconnect{Reason: reason})
	p.conn.Close()

	p.mu.RLock()
	nodeID := p.info.NodeID
	p.mu.RUnlock()

	if nodeID != "" && p.events.OnPeerDisconnected != nil {
		p.events.OnPeerDisconnected(nodeID)
	}
}

func (p *Peer) Info() PeerInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.info
}
