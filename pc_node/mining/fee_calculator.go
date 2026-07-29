package mining

type FeeCalculator struct {
}

func NewFeeCalculator() *FeeCalculator {
	return &FeeCalculator{}
}

// Density returns fee per byte
func (f *FeeCalculator) Density(tx Transaction) float64 {
	size := tx.GetSize()
	if size == 0 {
		return 0
	}
	return float64(tx.GetFee()) / float64(size)
}

func (f *FeeCalculator) CalculateTotal(txs []Transaction) uint64 {
	var total uint64
	for _, tx := range txs {
		total += tx.GetFee()
	}
	return total
}
