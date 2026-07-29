package p2p

type GossipEvents struct {
	OnMessagePublished    func(msgID string, msgType string)
	OnMessageReceived     func(msgID string, msgType string, fromPeer string)
	OnMessageForwarded    func(msgID string, msgType string, peers int)
	OnMessageDropped      func(msgID string, msgType string, reason string)
	OnDuplicateMessage    func(msgID string, fromPeer string)
	OnTTLExpired          func(msgID string)
	OnQueueOverflow       func(msgID string)
	OnPeerIgnored         func(peerID string, reason string)
	OnPropagationFinished func(msgID string)
}
