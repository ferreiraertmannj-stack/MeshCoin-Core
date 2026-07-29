package consensus

type ConsensusEvents struct {
	OnBlockAccepted     func(hash string, height uint64)
	OnBlockRejected     func(hash string, reason string)
	OnRewardInvalid     func(hash string, expected uint64, actual uint64)
	OnDifficultyChanged func(oldTarget string, newTarget string)
	OnChainUpdated      func(height uint64)
	OnTipChanged        func(newTip string)
}
