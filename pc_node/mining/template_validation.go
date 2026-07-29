package mining

import "fmt"

type TemplateValidator struct {
	policy    TemplatePolicy
	consensus ConsensusProvider
	events    TemplateEvents
	stats     *TemplateStatistics
}

func NewTemplateValidator(
	policy TemplatePolicy,
	consensus ConsensusProvider,
	events TemplateEvents,
	stats *TemplateStatistics,
) *TemplateValidator {
	return &TemplateValidator{
		policy:    policy,
		consensus: consensus,
		events:    events,
		stats:     stats,
	}
}

func (v *TemplateValidator) Validate(tmpl *BlockTemplate) error {
	if tmpl.TotalWeight > v.policy.MaxBlockWeight {
		return v.fail("weight exceeded")
	}

	if len(tmpl.Transactions) > v.policy.MaxTransactions {
		return v.fail("max transactions exceeded")
	}

	if tmpl.Coinbase == nil {
		return v.fail("missing coinbase")
	}

	// Structural check without full consensus loop
	if tmpl.PreviousHash == "" {
		return v.fail("missing previous hash")
	}

	if tmpl.Target == "" {
		return v.fail("missing difficulty target")
	}

	// Verify merkle root locally based on abstraction
	hashes := []string{tmpl.Coinbase.GetHash()}
	for _, tx := range tmpl.Transactions {
		hashes = append(hashes, tx.GetHash())
	}

	expectedRoot := v.consensus.CalculateMerkleRoot(hashes)
	if tmpl.MerkleRoot != expectedRoot {
		return v.fail("invalid merkle root")
	}

	return nil
}

func (v *TemplateValidator) fail(reason string) error {
	v.stats.IncValidationFailures()
	if v.events.OnValidationFailed != nil {
		go v.events.OnValidationFailed(reason)
	}
	return fmt.Errorf("template validation failed: %s", reason)
}
