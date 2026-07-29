package p2p

// ChainProvider provides read-only access to the local blockchain
// Required by ChainLocator and BlockchainSyncManager
type ChainProvider interface {
	// GetBlockHash returns the hash of the block at the given height
	GetBlockHash(height uint64) (string, error)
	// GetBlockHeight returns the height of a block given its hash
	GetBlockHeight(hash string) (uint64, error)
	// GetTip returns the current highest block hash and its height
	GetTip() (string, uint64)
	// HasBlock checks if a block is stored in the local chain
	HasBlock(hash string) bool
}

type ChainLocator struct {
	provider ChainProvider
}

func NewChainLocator(provider ChainProvider) *ChainLocator {
	return &ChainLocator{provider: provider}
}

// BuildLocatorHashes creates a sparse list of block hashes starting from the tip,
// jumping exponentially backwards, and ending with the genesis block.
func (cl *ChainLocator) BuildLocatorHashes() []string {
	tipHash, tipHeight := cl.provider.GetTip()

	if tipHeight == 0 {
		return []string{tipHash}
	}

	var locators []string
	step := uint64(1)
	currentHeight := tipHeight

	for {
		hash, err := cl.provider.GetBlockHash(currentHeight)
		if err == nil {
			locators = append(locators, hash)
		}

		if currentHeight == 0 {
			break
		}

		// First 10 hashes are sequential
		if len(locators) > 10 {
			step *= 2
		}

		if currentHeight > step {
			currentHeight -= step
		} else {
			currentHeight = 0 // Genesis
		}
	}

	return locators
}

// FindCommonBlock finds the highest common block between our local chain
// and a list of locators provided by a peer.
func (cl *ChainLocator) FindCommonBlock(peerLocators []string) (string, uint64) {
	for _, hash := range peerLocators {
		if cl.provider.HasBlock(hash) {
			height, err := cl.provider.GetBlockHeight(hash)
			if err == nil {
				return hash, height
			}
		}
	}
	// If none found, we assume genesis is the only common block (hash "0" or predefined)
	// But in this decoupled layer, we just return empty if completely unknown.
	return "", 0
}
