package consensus

import "fmt"

type RewardValidator struct {
	policy ConsensusPolicy
}

func NewRewardValidator(policy ConsensusPolicy) *RewardValidator {
	return &RewardValidator{
		policy: policy,
	}
}

// CalculateSubsidy calculates the block subsidy based on halving intervals
func (r *RewardValidator) CalculateSubsidy(height uint64) uint64 {
	halvings := height / r.policy.HalvingInterval

	// If halving occurs 64 times, reward drops to 0 (shifts limit)
	if halvings >= 64 {
		return 0
	}

	return r.policy.InitialSubsidy >> halvings
}

// Validate checks if the declared reward matches expected Subsidy
func (r *RewardValidator) Validate(height uint64, declaredReward uint64) error {
	expected := r.CalculateSubsidy(height)
	if declaredReward != expected {
		return fmt.Errorf("invalid reward: expected %d, got %d", expected, declaredReward)
	}
	return nil
}
