package p2p

import "time"

const (
	MsgTypeHello      = "HELLO"
	MsgTypeHelloAck   = "HELLO_ACK"
	MsgTypeReject     = "REJECT"
	MsgTypePing       = "PING"
	MsgTypePong       = "PONG"
	MsgTypeDisconnect = "DISCONNECT"
)

type Capability string

const (
	CapFastSync  Capability = "FastSync"
	CapMining    Capability = "Mining"
	CapRelay     Capability = "Relay"
	CapArchive   Capability = "Archive"
	CapLightNode Capability = "LightNode"
	CapAPI       Capability = "API"
	CapWallet    Capability = "Wallet"
)

// P2PMessage is the envelope for all P2P handshake and heartbeat messages.
type P2PMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type MsgHello struct {
	Version         string       `json:"version"`
	Agent           string       `json:"agent"`
	ProtocolVersion int          `json:"protocol_version"`
	NetworkID       string       `json:"network_id"`
	GenesisHash     string       `json:"genesis_hash"`
	NodeID          string       `json:"node_id"`
	Capabilities    []Capability `json:"capabilities"`
	Height          uint64       `json:"height"`
}

type MsgHelloAck struct {
	NodeID       string       `json:"node_id"`
	Capabilities []Capability `json:"capabilities"`
}

type MsgReject struct {
	Reason string `json:"reason"`
}

type MsgPing struct {
	Timestamp int64 `json:"timestamp"`
}

type MsgPong struct {
	Timestamp int64 `json:"timestamp"`
}

type MsgDisconnect struct {
	Reason string `json:"reason"`
}

type PeerInfo struct {
	NodeID          string
	Version         string
	Agent           string
	ProtocolVersion int
	NetworkID       string
	GenesisHash     string
	Capabilities    []Capability
	Latency         time.Duration
	Height          uint64
	ConnectedSince  time.Time
	LastPing        time.Time
	LastPong        time.Time
	Address         string
}

// HandshakeValidationOptions contains network rules that must match during handshake
type HandshakeValidationOptions struct {
	ExpectedNetworkID      string
	ExpectedGenesisHash    string
	MinimumProtocolVersion int
}
