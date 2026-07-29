package utxo

import "fmt"

type TransactionInput struct {
	PreviousOutPoint OutPoint
	Signature        []byte
}

type TransactionOutput struct {
	Value  uint64
	Script []byte
}

type Transaction struct {
	Hash    string
	Inputs  []TransactionInput
	Outputs []TransactionOutput
}

type TransactionValidator struct {
	policy   UTXOPolicy
	detector *DoubleSpendDetector
	sigVal   SignatureValidator
	utxoSet  *UTXOSet
	cache    *UTXOCache
}

func NewTransactionValidator(
	policy UTXOPolicy,
	detector *DoubleSpendDetector,
	sigVal SignatureValidator,
	utxoSet *UTXOSet,
	cache *UTXOCache,
) *TransactionValidator {
	return &TransactionValidator{
		policy:   policy,
		detector: detector,
		sigVal:   sigVal,
		utxoSet:  utxoSet,
		cache:    cache,
	}
}

func (v *TransactionValidator) Validate(tx *Transaction) error {
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction must have at least one input")
	}
	if len(tx.Inputs) > v.policy.MaxInputs {
		return fmt.Errorf("transaction exceeds max inputs")
	}
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction must have at least one output")
	}
	if len(tx.Outputs) > v.policy.MaxOutputs {
		return fmt.Errorf("transaction exceeds max outputs")
	}

	// Double Spend & Existence check
	inputs := make([]OutPoint, len(tx.Inputs))
	for i, in := range tx.Inputs {
		inputs[i] = in.PreviousOutPoint
	}

	if err := v.detector.Detect(inputs); err != nil {
		return err
	}

	var sumInputs uint64
	var sumOutputs uint64

	// Gather inputs sum and verify signatures
	for _, in := range tx.Inputs {
		// Try Cache first
		u, ok := v.cache.Get(in.PreviousOutPoint)
		if ok {
			// update stats externally
		} else {
			u, _ = v.utxoSet.Lookup(in.PreviousOutPoint)
			v.cache.Set(in.PreviousOutPoint, u)
		}

		sumInputs += u.Value

		if !v.sigVal.Verify(tx.Hash, u.Script, in.Signature) {
			return fmt.Errorf("invalid signature for input %v", in.PreviousOutPoint)
		}
	}

	for _, out := range tx.Outputs {
		if out.Value == 0 { // Explicitly rejecting 0 value outputs if policy dictates
			return fmt.Errorf("invalid 0 value output")
		}
		sumOutputs += out.Value
	}

	if sumInputs < sumOutputs {
		return fmt.Errorf("sum of inputs (%d) is less than sum of outputs (%d)", sumInputs, sumOutputs)
	}

	fee := sumInputs - sumOutputs
	if fee < v.policy.MinFee {
		return fmt.Errorf("insufficient fee: %d < %d", fee, v.policy.MinFee)
	}

	return nil
}
