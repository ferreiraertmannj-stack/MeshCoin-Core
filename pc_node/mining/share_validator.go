package mining

import (
	"math/big"
)

type ShareValidator struct {
}

func NewShareValidator() *ShareValidator {
	return &ShareValidator{}
}

func (v *ShareValidator) Validate(hash string, target string) bool {
	// Simple big.Int comparison for PoW abstraction

	hashInt := new(big.Int)
	hashInt.SetString(hash, 16)

	targetInt := new(big.Int)
	targetInt.SetString(target, 16)

	// If hash <= target, it's valid
	return hashInt.Cmp(targetInt) <= 0
}
