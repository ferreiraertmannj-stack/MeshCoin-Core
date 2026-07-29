package utxo

import "fmt"

type DoubleSpendDetector struct {
	set *UTXOSet
}

func NewDoubleSpendDetector(set *UTXOSet) *DoubleSpendDetector {
	return &DoubleSpendDetector{set: set}
}

// Detect checks if the given inputs:
// 1. Actually exist in the UTXO set.
// 2. Are not spent more than once in the same transaction.
func (d *DoubleSpendDetector) Detect(inputs []OutPoint) error {
	seen := make(map[OutPoint]bool)

	for _, in := range inputs {
		// 1. Same transaction double spend check
		if seen[in] {
			return fmt.Errorf("double spend detected within the same transaction: %v", in)
		}
		seen[in] = true

		// 2. Existence check (cannot spend what doesn't exist)
		if !d.set.Exists(in) {
			return fmt.Errorf("attempt to spend non-existent or already spent UTXO: %v", in)
		}
	}

	return nil
}
