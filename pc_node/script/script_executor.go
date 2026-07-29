package script

import "fmt"

type ScriptExecutor struct {
	policy    ScriptPolicy
	registry  *OpcodeRegistry
	hashEng   HashEngine
	sigEng    SignatureEngine
	pubKeyEng PublicKeyEngine
}

func NewScriptExecutor(
	policy ScriptPolicy,
	registry *OpcodeRegistry,
	hashEng HashEngine,
	sigEng SignatureEngine,
	pubKeyEng PublicKeyEngine,
) *ScriptExecutor {
	return &ScriptExecutor{
		policy:    policy,
		registry:  registry,
		hashEng:   hashEng,
		sigEng:    sigEng,
		pubKeyEng: pubKeyEng,
	}
}

func (e *ScriptExecutor) Execute(instructions []Instruction, txHash []byte) error {
	stack := NewScriptStack(e.policy.MaxStackDepth)

	ctx := &ExecutionContext{
		Stack:      stack,
		TxHash:     txHash,
		HashEngine: e.hashEng,
		SigEngine:  e.sigEng,
		PubKeyEng:  e.pubKeyEng,
	}

	for _, instr := range instructions {
		if instr.IsData {
			err := stack.Push(instr.Data)
			if err != nil {
				return err
			}
		} else {
			handler, err := e.registry.Get(instr.Opcode)
			if err != nil {
				return err
			}
			err = handler(ctx)
			if err != nil {
				return fmt.Errorf("opcode 0x%x execution failed: %w", instr.Opcode, err)
			}
		}
	}

	// The script is valid if the top element is true (not 0 and not empty)
	top, err := stack.Pop()
	if err != nil {
		return fmt.Errorf("script finished with empty stack")
	}

	if len(top) == 0 || (len(top) == 1 && top[0] == 0) {
		return fmt.Errorf("script finished with false on top of stack")
	}

	return nil
}
