package p2p

import "math/rand"

type SeedManager struct {
	seedNodes []PeerAddress
}

func NewSeedManager(seeds []PeerAddress) *SeedManager {
	return &SeedManager{
		seedNodes: seeds,
	}
}

func (sm *SeedManager) NeedsSeeds(store *PeerStore) bool {
	peers, err := store.GetAllPeers()
	if err != nil || len(peers) == 0 {
		return true
	}
	return false
}

func (sm *SeedManager) GetRandomSeed() (PeerAddress, bool) {
	if len(sm.seedNodes) == 0 {
		return PeerAddress{}, false
	}
	return sm.seedNodes[rand.Intn(len(sm.seedNodes))], true
}

func (sm *SeedManager) GetAllSeeds() []PeerAddress {
	return sm.seedNodes
}
