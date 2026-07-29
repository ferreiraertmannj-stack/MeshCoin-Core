package p2p

type BlockchainSyncEvents struct {
	OnHeadersReceived        func(peerID string, count int)
	OnHeadersValidated       func(peerID string, count int)
	OnBlocksRequested        func(peerID string, count int)
	OnBlockReceived          func(peerID string, hash string)
	OnBlockImported          func(hash string, height uint64)
	OnForkDetected           func(forkPoint string)
	OnReorganizationDetected func(oldTip string, newTip string)
	OnSyncCompleted          func(newHeight uint64)
	OnSyncFailed             func(peerID string, reason string)
}
