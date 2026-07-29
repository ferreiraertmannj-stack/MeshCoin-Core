package mempool

import (
	"fmt"
	"time"
)

// Transaction is a generic representation decoupled from Ledger
type Transaction interface {
	GetHash() string
	GetSender() string
	GetFee() uint64
	GetSize() uint64
	GetTimestamp() int64
	GetNonce() uint64
}

type MempoolValidator struct {
	policy MempoolPolicy
	events MempoolEvents
}

func NewMempoolValidator(policy MempoolPolicy, events MempoolEvents) *MempoolValidator {
	return &MempoolValidator{
		policy: policy,
		events: events,
	}
}

func (v *MempoolValidator) Validate(tx Transaction) error {
	if tx == nil {
		v.fireFailed("", "transaction is nil")
		return fmt.Errorf("transaction is nil")
	}

	hash := tx.GetHash()
	if hash == "" {
		v.fireFailed(hash, "empty hash")
		return fmt.Errorf("empty hash")
	}

	if tx.GetSender() == "" {
		v.fireFailed(hash, "empty sender")
		return fmt.Errorf("empty sender")
	}

	if tx.GetFee() < v.policy.MinFee {
		v.fireFailed(hash, "fee too low")
		return fmt.Errorf("fee too low")
	}

	if tx.GetSize() > v.policy.MaxMemoryBytes { // Unlikely, but just basic cap
		v.fireFailed(hash, "transaction too large")
		return fmt.Errorf("transaction too large")
	}

	// Basic TTL check (Timestamp shouldn't be too old)
	age := time.Since(time.Unix(tx.GetTimestamp(), 0))
	if age > v.policy.TTL {
		v.fireFailed(hash, "transaction expired before entering pool")
		return fmt.Errorf("transaction expired")
	}

	return nil
}

func (v *MempoolValidator) fireFailed(hash string, reason string) {
	if v.events.OnValidationFailed != nil {
		go v.events.OnValidationFailed(hash, reason)
	}
}
