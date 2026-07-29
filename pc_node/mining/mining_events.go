package mining

type MiningEvents struct {
	OnMiningStarted     func()
	OnMiningStopped     func()
	OnJobCreated        func(jobID string, height uint64)
	OnJobCancelled      func(jobID string, reason string)
	OnNonceTested       func(hashes uint64) // For hashrate reporting
	OnShareFound        func(hash string, difficulty uint64)
	OnBlockFound        func(block *Block)
	OnDifficultyUpdated func(newTarget string)
	OnMiningError       func(err error)
}
