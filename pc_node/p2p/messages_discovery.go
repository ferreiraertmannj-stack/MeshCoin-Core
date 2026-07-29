package p2p

import "time"

const (
	MsgTypeGetPeers         = "GET_PEERS"
	MsgTypePeers            = "PEERS"
	MsgTypePeerAnnouncement = "PEER_ANNOUNCEMENT"
	MsgTypePeerGoodbye      = "PEER_GOODBYE"
)

type PeerAddress struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

type MsgGetPeers struct {
	Limit int `json:"limit"`
}

type MsgPeers struct {
	Peers []PeerRecord `json:"peers"`
}

type MsgPeerAnnouncement struct {
	NodeID       string       `json:"node_id"`
	Address      PeerAddress  `json:"address"`
	Capabilities []Capability `json:"capabilities"`
	Version      string       `json:"version"`
}

type MsgPeerGoodbye struct {
	Reason string `json:"reason"`
}

// PeerRecord is used for persistent storage and sharing topology
type PeerRecord struct {
	NodeID       string       `json:"node_id"`
	Address      PeerAddress  `json:"address"`
	Capabilities []Capability `json:"capabilities"`
	Version      string       `json:"version"`
	Reliability  float64      `json:"reliability"`
	Latency      int64        `json:"latency_ms"` // in milliseconds for JSON
	Uptime       int64        `json:"uptime_s"`   // in seconds
	LastSuccess  time.Time    `json:"last_success"`
	Failures     int          `json:"failures"`
	IsSeed       bool         `json:"is_seed"`
}
