package utxo

import "time"

type UTXOPolicy struct {
	MinFee         uint64
	MaxInputs      int
	MaxOutputs     int
	CacheTTL       time.Duration
	MaxSnapshotAge time.Duration
	QueueSize      int
	MaxWorkers     int
}

func DefaultUTXOPolicy() UTXOPolicy {
	return UTXOPolicy{
		MinFee:         1,
		MaxInputs:      1000,
		MaxOutputs:     1000,
		CacheTTL:       60 * time.Second,
		MaxSnapshotAge: 24 * time.Hour,
		QueueSize:      10000,
		MaxWorkers:     4,
	}
}
