package utxo

import (
	"fmt"
	"time"
)

type UTXOSnapshot struct {
	ID        string
	Timestamp time.Time
	utxos     map[OutPoint]UTXO
}

func (s *UTXOSnapshot) Get(op OutPoint) (UTXO, bool) {
	u, ok := s.utxos[op]
	return u, ok
}

// CreateSnapshot captures an O(N) copy of the current UTXO set.
// This is isolated and returns an immutable read-only view.
func (set *UTXOSet) CreateSnapshot() *UTXOSnapshot {
	set.mu.RLock()
	defer set.mu.RUnlock()

	snap := make(map[OutPoint]UTXO, len(set.utxos))
	for k, v := range set.utxos {
		snap[k] = v
	}

	return &UTXOSnapshot{
		ID:        fmt.Sprintf("snap-%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		utxos:     snap,
	}
}
