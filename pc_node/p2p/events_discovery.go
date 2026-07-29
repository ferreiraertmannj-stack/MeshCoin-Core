package p2p

// DiscoveryEventHandlers contains callbacks for peer discovery events.
type DiscoveryEventHandlers struct {
	OnPeerDiscovered    func(address string)
	OnPeerValidated     func(record PeerRecord)
	OnPeerConnected     func(nodeID string, address string)
	OnPeerRemoved       func(nodeID string, reason string)
	OnPeerExpired       func(nodeID string)
	OnDiscoveryFinished func(peersFound int)
}
