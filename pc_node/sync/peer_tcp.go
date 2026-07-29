package sync

import (
	"errors"
	"net"
	"time"
)

// TCPPeer implements the Peer interface for Fast Sync,
// encapsulating the physical network connection.
type TCPPeer struct {
	id             string
	conn           net.Conn // A conexão real (TCP ou WebSocket do nó)
	height         uint64
	latency        time.Duration
	failures       int
	connectedSince time.Time
}

// NewTCPPeer creates a new instance of TCPPeer.
func NewTCPPeer(id string, conn net.Conn, initialHeight uint64) *TCPPeer {
	return &TCPPeer{
		id:             id,
		conn:           conn,
		height:         initialHeight,
		latency:        time.Millisecond * 50, // default placeholder
		failures:       0,
		connectedSince: time.Now(),
	}
}

func (p *TCPPeer) ID() string {
	return p.id
}

func (p *TCPPeer) Height() uint64 {
	return p.height
}

func (p *TCPPeer) Latency() time.Duration {
	return p.latency
}

func (p *TCPPeer) Failures() int {
	return p.failures
}

func (p *TCPPeer) ConnectedSince() time.Time {
	return p.connectedSince
}

// SendMsg writes any P2P message directly into the physical connection.
// Na Sprint atual, age como stub/abstração, pois não devemos integrar protocolos.
func (p *TCPPeer) SendMsg(msg interface{}) error {
	if p.conn == nil {
		return errors.New("connection is nil")
	}
	// TODO: Na Fase de integração real, serializar JSON/Gob e escrever no socket.
	return nil
}

// Receive is a stub for reading from the conn.
func (p *TCPPeer) Receive() (interface{}, error) {
	if p.conn == nil {
		return nil, errors.New("connection is nil")
	}
	return nil, nil
}

// RequestHeaders uses SendMsg to ask for a block skeleton.
func (p *TCPPeer) RequestHeaders(startHash string, limit int) error {
	msg := MsgGetHeaders{StartHash: startHash, Limit: limit}
	return p.SendMsg(msg)
}

// RequestBlocks requests binary payloads.
func (p *TCPPeer) RequestBlocks(startIndex, endIndex uint64) error {
	msg := MsgGetBlocks{StartIndex: startIndex, EndIndex: endIndex}
	return p.SendMsg(msg)
}

func (p *TCPPeer) Ping() error {
	// Atualizaria a latência e verificaria se o socket está vivo.
	return nil
}

func (p *TCPPeer) Disconnect() {
	if p.conn != nil {
		p.conn.Close()
	}
}

// --- Métodos exclusivos para injeção de testes ou updates do status ---

func (p *TCPPeer) UpdateHeight(h uint64) {
	p.height = h
}

func (p *TCPPeer) UpdateLatency(d time.Duration) {
	p.latency = d
}

func (p *TCPPeer) AddFailure() {
	p.failures++
}
