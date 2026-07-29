package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// DoHandshake executes the server-side handshake
func (p *Peer) DoHandshake(opts HandshakeValidationOptions) error {
	p.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	var msg P2PMessage
	if err := p.decoder.Decode(&msg); err != nil {
		return fmt.Errorf("failed to read handshake: %v", err)
	}

	if msg.Type != MsgTypeHello {
		return p.rejectAndDisconnect("expected HELLO message")
	}

	var hello MsgHello
	b, _ := json.Marshal(msg.Payload)
	json.Unmarshal(b, &hello)

	// Validations
	if hello.NetworkID != opts.ExpectedNetworkID {
		return p.rejectAndDisconnect(fmt.Sprintf("invalid network id: expected %s, got %s", opts.ExpectedNetworkID, hello.NetworkID))
	}
	if hello.GenesisHash != opts.ExpectedGenesisHash {
		return p.rejectAndDisconnect("invalid genesis hash")
	}
	if hello.ProtocolVersion < opts.MinimumProtocolVersion {
		return p.rejectAndDisconnect("protocol version not supported")
	}
	if hello.NodeID == "" || hello.NodeID == p.config.NodeID {
		return p.rejectAndDisconnect("invalid or duplicate node id")
	}

	if !p.secManager.RegisterNodeID(hello.NodeID) {
		return p.rejectAndDisconnect("node id already connected")
	}

	p.mu.Lock()
	p.info = PeerInfo{
		NodeID:          hello.NodeID,
		Version:         hello.Version,
		Agent:           hello.Agent,
		ProtocolVersion: hello.ProtocolVersion,
		NetworkID:       hello.NetworkID,
		GenesisHash:     hello.GenesisHash,
		Capabilities:    hello.Capabilities,
		Height:          hello.Height,
		ConnectedSince:  time.Now(),
		Address:         p.conn.RemoteAddr().String(),
	}
	p.mu.Unlock()

	// Accept connection
	ack := MsgHelloAck{
		NodeID:       p.config.NodeID,
		Capabilities: p.config.Capabilities,
	}
	if err := p.SendMessage(MsgTypeHelloAck, ack); err != nil {
		p.Disconnect("failed to send HELLO_ACK")
		return err
	}

	if p.events.OnPeerAuthenticated != nil {
		p.events.OnPeerAuthenticated(p.info)
	}

	// Start Heartbeat and Message Loop
	go p.heartbeatLoop()
	go p.messageLoop()

	return nil
}

func (p *Peer) rejectAndDisconnect(reason string) error {
	p.SendMessage(MsgTypeReject, MsgReject{Reason: reason})
	p.conn.Close()
	p.cancelFunc()
	if p.events.OnPeerRejected != nil {
		p.events.OnPeerRejected(p.conn.RemoteAddr().String(), reason)
	}
	return errors.New(reason)
}

// DoClientHandshake executes the client-side handshake (initiator)
func (p *Peer) DoClientHandshake(opts HandshakeValidationOptions) error {
	// 1. Send Hello
	hello := MsgHello{
		Version:         p.config.Version,
		Agent:           p.config.Agent,
		ProtocolVersion: p.config.ProtocolVersion,
		NetworkID:       p.config.NetworkID,
		GenesisHash:     p.config.GenesisHash,
		NodeID:          p.config.NodeID,
		Capabilities:    p.config.Capabilities,
		Height:          p.config.Height,
	}

	if err := p.SendMessage(MsgTypeHello, hello); err != nil {
		return p.rejectAndDisconnect("failed to send hello")
	}

	// 2. Wait for HelloAck
	p.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var msg P2PMessage
	if err := p.decoder.Decode(&msg); err != nil {
		return fmt.Errorf("failed to read hello_ack: %v", err)
	}

	if msg.Type == MsgTypeReject {
		var rej MsgReject
		b, _ := json.Marshal(msg.Payload)
		json.Unmarshal(b, &rej)
		p.conn.Close()
		p.cancelFunc()
		return fmt.Errorf("handshake rejected: %s", rej.Reason)
	}

	if msg.Type != MsgTypeHelloAck {
		return p.rejectAndDisconnect("expected HELLO_ACK message")
	}

	var ack MsgHelloAck
	b, _ := json.Marshal(msg.Payload)
	json.Unmarshal(b, &ack)

	if ack.NodeID == "" || ack.NodeID == p.config.NodeID {
		return p.rejectAndDisconnect("invalid or duplicate node id in ack")
	}

	if !p.secManager.RegisterNodeID(ack.NodeID) {
		return p.rejectAndDisconnect("node id already connected")
	}

	p.mu.Lock()
	p.info = PeerInfo{
		NodeID:         ack.NodeID,
		Capabilities:   ack.Capabilities,
		ConnectedSince: time.Now(),
		Address:        p.conn.RemoteAddr().String(),
	}
	p.mu.Unlock()

	if p.events.OnPeerAuthenticated != nil {
		p.events.OnPeerAuthenticated(p.info)
	}

	go p.heartbeatLoop()
	go p.messageLoop()

	return nil
}
