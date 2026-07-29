package p2p

const (
	MsgTypeGossipBlock        = "GOSSIP_BLOCK"
	MsgTypeGossipTransaction  = "GOSSIP_TRANSACTION"
	MsgTypeGossipInventory    = "GOSSIP_INVENTORY"
	MsgTypeGossipAnnouncement = "GOSSIP_ANNOUNCEMENT"
	MsgTypeGossipRelay        = "GOSSIP_RELAY"
	MsgTypeGossipReject       = "GOSSIP_REJECT"
	MsgTypeGossipKeepAlive    = "GOSSIP_KEEPALIVE"
)

// GossipEnvelope wraps a gossip payload with deduplication and routing metadata.
// It is embedded inside P2PMessage.Payload when sent over the wire.
type GossipEnvelope struct {
	MessageID  string      `json:"message_id"`
	Timestamp  int64       `json:"timestamp"`
	TTL        int         `json:"ttl"`
	HopCount   int         `json:"hop_count"`
	OriginNode string      `json:"origin_node"`
	Payload    interface{} `json:"payload"`
}

type MsgBlock struct {
	Data []byte `json:"data"` // Raw block data
}

type MsgTransaction struct {
	Data []byte `json:"data"` // Raw tx data
}

type MsgInventory struct {
	Items []string `json:"items"` // List of hashes
}

type MsgAnnouncement struct {
	Subject string `json:"subject"`
	Content string `json:"content"`
}

type MsgRelay struct {
	TargetNode string `json:"target_node"`
	Data       []byte `json:"data"`
}

type MsgGossipReject struct { // Overrides or aliases the base MsgReject for gossip context
	Reason string `json:"reason"`
}

type MsgKeepAlive struct {
	Timestamp int64 `json:"timestamp"`
}
