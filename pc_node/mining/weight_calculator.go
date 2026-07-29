package mining

type WeightCalculator struct {
	policy TemplatePolicy
}

func NewWeightCalculator(policy TemplatePolicy) *WeightCalculator {
	return &WeightCalculator{
		policy: policy,
	}
}

// Calculate returns the weight of a transaction. For this abstraction, size = weight (simplification).
func (w *WeightCalculator) Calculate(tx Transaction) uint64 {
	return tx.GetSize()
}

// CheckLimit ensures the new transaction doesn't overflow block boundaries
func (w *WeightCalculator) CheckLimit(currentWeight uint64, txCount int, tx Transaction) bool {
	if txCount >= w.policy.MaxTransactions {
		return false
	}
	return currentWeight+w.Calculate(tx) <= w.policy.MaxBlockWeight
}
