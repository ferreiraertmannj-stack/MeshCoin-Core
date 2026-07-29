package script

import "fmt"

type Instruction struct {
	Opcode Opcode
	Data   []byte
	IsData bool
}

type ScriptParser struct {
	policy ScriptPolicy
}

func NewScriptParser(policy ScriptPolicy) *ScriptParser {
	return &ScriptParser{
		policy: policy,
	}
}

// Parse converts a raw script byte slice into a slice of Instructions.
// Very basic implementation:
// If byte <= 0x4b, it's a data push (push next N bytes).
// Else, it's an Opcode.
func (p *ScriptParser) Parse(raw []byte) ([]Instruction, error) {
	if len(raw) > p.policy.MaxScriptSize {
		return nil, fmt.Errorf("script size exceeds max limit")
	}

	var instructions []Instruction
	i := 0
	count := 0

	for i < len(raw) {
		if count > p.policy.MaxOpcodeCount {
			return nil, fmt.Errorf("script opcode count exceeded")
		}

		b := raw[i]
		if b > 0 && b <= 0x4b {
			// Push data
			length := int(b)
			if i+1+length > len(raw) {
				return nil, fmt.Errorf("push data length out of bounds")
			}
			data := raw[i+1 : i+1+length]
			instructions = append(instructions, Instruction{
				Opcode: 0,
				Data:   data,
				IsData: true,
			})
			i += 1 + length
		} else {
			// Opcode
			instructions = append(instructions, Instruction{
				Opcode: Opcode(b),
				IsData: false,
			})
			i++
		}
		count++
	}

	return instructions, nil
}
