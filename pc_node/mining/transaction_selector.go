package mining

import (
	"sort"
)

type TransactionSelector struct {
	weightCalc *WeightCalculator
	feeCalc    *FeeCalculator
}

func NewTransactionSelector(w *WeightCalculator, f *FeeCalculator) *TransactionSelector {
	return &TransactionSelector{
		weightCalc: w,
		feeCalc:    f,
	}
}

// Select filters and orders transactions from the mempool snapshot
func (s *TransactionSelector) Select(snapshot []Transaction) ([]Transaction, uint64) {
	// Order by fee density (Highest first)
	sort.SliceStable(snapshot, func(i, j int) bool {
		return s.feeCalc.Density(snapshot[i]) > s.feeCalc.Density(snapshot[j])
	})

	var selected []Transaction
	var currentWeight uint64

	dedup := make(map[string]bool)

	for _, tx := range snapshot {
		hash := tx.GetHash()

		// Avoid duplicates
		if dedup[hash] {
			continue
		}

		// Check block boundaries
		if !s.weightCalc.CheckLimit(currentWeight, len(selected), tx) {
			continue
		}

		selected = append(selected, tx)
		currentWeight += s.weightCalc.Calculate(tx)
		dedup[hash] = true
	}

	return selected, currentWeight
}
