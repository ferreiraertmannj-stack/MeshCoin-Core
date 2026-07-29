package consensus

import "fmt"

type CoinbaseValidator struct {
	rewardValidator *RewardValidator
}

func NewCoinbaseValidator(rewardValidator *RewardValidator) *CoinbaseValidator {
	return &CoinbaseValidator{
		rewardValidator: rewardValidator,
	}
}

func (v *CoinbaseValidator) Validate(height uint64, totalFees uint64, coinbase Transaction) error {
	if coinbase == nil {
		return fmt.Errorf("missing coinbase transaction")
	}

	// Calculate subsidy
	subsidy := v.rewardValidator.CalculateSubsidy(height)

	expectedTotal := subsidy + totalFees
	actualTotal := coinbase.GetReward() + coinbase.GetFees()

	if actualTotal != expectedTotal {
		return fmt.Errorf("coinbase value mismatch: expected %d, got %d", expectedTotal, actualTotal)
	}

	// We also delegate the pure reward validation
	return v.rewardValidator.Validate(height, coinbase.GetReward())
}
