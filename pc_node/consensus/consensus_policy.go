package consensus

import "time"

type ConsensusPolicy struct {
	MaxFutureTime     time.Duration
	AllowedClockDrift time.Duration
	MaxBlockWeight    uint64
	BlockVersion      uint32
	HalvingInterval   uint64
	GenesisHash       string
	InitialSubsidy    uint64
}

func DefaultConsensusPolicy() ConsensusPolicy {
	return ConsensusPolicy{
		MaxFutureTime:     2 * time.Hour,
		AllowedClockDrift: 15 * time.Second,
		MaxBlockWeight:    4000000,
		BlockVersion:      1,
		HalvingInterval:   210000,
		GenesisHash:       "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
		InitialSubsidy:    50 * 100000000, // 50 Coins
	}
}
