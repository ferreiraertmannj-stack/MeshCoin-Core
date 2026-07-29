package sync

import (
	"errors"
	"net"
	"sync"
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

	enc *TransportEncoder
	dec *TransportDecoder
	mu  sync.Mutex // Protege writes concorrentes no socket
}

// NewTCPPeer creates a new instance of TCPPeer.
func NewTCPPeer(id string, conn net.Conn, initialHeight uint64) *TCPPeer {
	var enc *TransportEncoder
	var dec *TransportDecoder

	if conn != nil {
		enc = NewTransportEncoder(conn)
		dec = NewTransportDecoder(conn)
	}

	return &TCPPeer{
		id:             id,
		conn:           conn,
		height:         initialHeight,
		latency:        time.Millisecond * 50, // default placeholder
		failures:       0,
		connectedSince: time.Now(),
		enc:            enc,
		dec:            dec,
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
func (p *TCPPeer) SendMsg(msgType MsgType, msg interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.conn == nil || p.enc == nil {
		return errors.New("connection is nil")
	}

	// Read/Write timeout: 5s para controle de lentidão
	p.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

	err := p.enc.Encode(msgType, msg)
	if err != nil {
		p.failures++
	}
	return err
}

// Receive blocks and reads the next message from the connection.
func (p *TCPPeer) Receive() (*TransportMessage, error) {
	if p.conn == nil || p.dec == nil {
		return nil, errors.New("connection is nil")
	}

	// Limite de tempo de leitura
	p.conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	msg, err := p.dec.Decode()
	if err != nil {
		p.failures++
	}
	return msg, err
}

// RequestHeaders uses SendMsg to ask for a block skeleton.
func (p *TCPPeer) RequestHeaders(startHeight uint64, limit int) error {
	msg := GetHeadersMsg{StartHeight: startHeight, Limit: limit}
	return p.SendMsg(MsgTypeRequestHeaders, msg)
}

// RequestBlocks requests binary payloads.
func (p *TCPPeer) RequestBlocks(startIndex, endIndex uint64) error {
	msg := GetBlocksMsg{StartHeight: startIndex, EndHeight: endIndex}
	return p.SendMsg(MsgTypeRequestBlocks, msg)
}

func (p *TCPPeer) Ping() error {
	start := time.Now()
	msg := PingMsg{Timestamp: start}
	err := p.SendMsg(MsgTypePing, msg)
	if err != nil {
		return err
	}
	// Em um cenário assíncrono, a latência seria calculada no Pong.
	return nil
}

func (p *TCPPeer) GetStatus() error {
	return p.SendMsg(MsgTypeGetStatus, nil)
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
