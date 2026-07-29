package p2p

// PeerEventHandlers contains the decoupled callbacks for all peer lifecycle events.
type PeerEventHandlers struct {
	OnPeerConnected     func(address string)
	OnPeerAuthenticated func(info PeerInfo)
	OnPeerRejected      func(address string, reason string)
	OnPeerDisconnected  func(nodeID string)
	OnHeartbeatTimeout  func(nodeID string)
	OnPeerBlacklisted   func(address string, reason string)
}
