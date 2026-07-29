package p2p

import "sync"

type ForkDetector struct {
	provider ChainProvider
	events   BlockchainSyncEvents

	// Track candidate branches: hash -> height
	candidateTips map[string]uint64
	mu            sync.RWMutex
}

func NewForkDetector(provider ChainProvider, events BlockchainSyncEvents) *ForkDetector {
	return &ForkDetector{
		provider:      provider,
		events:        events,
		candidateTips: make(map[string]uint64),
	}
}

// CheckForFork evaluates if an incoming header or block indicates a fork
func (fd *ForkDetector) CheckForFork(hash string, parentHash string, height uint64) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	_, tipHeight := fd.provider.GetTip()

	// If this block's parent is not our tip, and it is building on something else
	if !fd.provider.HasBlock(parentHash) {
		// Possibly a fork or just an orphan. We won't know until we walk back.
		// For decoupled detection, if it's building a longer chain:
		if height > tipHeight {
			fd.candidateTips[hash] = height
		}
	} else {
		// If the parent is in our chain, but parent is NOT the tip, it's a fork point
		parentHeight, err := fd.provider.GetBlockHeight(parentHash)
		if err == nil && parentHeight < tipHeight {
			// Fork detected
			if fd.events.OnForkDetected != nil {
				go fd.events.OnForkDetected(parentHash)
			}

			// If this new branch is longer than our tip
			if height > tipHeight {
				fd.candidateTips[hash] = height
			}
		}
	}
}

func (fd *ForkDetector) TriggerReorganization(oldTip string, newTip string) {
	if fd.events.OnReorganizationDetected != nil {
		go fd.events.OnReorganizationDetected(oldTip, newTip)
	}
}

func (fd *ForkDetector) ClearCandidate(hash string) {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	delete(fd.candidateTips, hash)
}
