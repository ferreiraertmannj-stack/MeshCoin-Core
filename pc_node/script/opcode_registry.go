package script

import "fmt"

type Opcode byte

const (
	OP_DUP         Opcode = 0x76
	OP_DROP        Opcode = 0x75
	OP_EQUAL       Opcode = 0x87
	OP_EQUALVERIFY Opcode = 0x88
	OP_HASH256     Opcode = 0xaa
	OP_CHECKSIG    Opcode = 0xac
	OP_VERIFY      Opcode = 0x69
	OP_RETURN      Opcode = 0x6a
)

type ExecutionContext struct {
	Stack      *ScriptStack
	TxHash     []byte
	HashEngine HashEngine
	SigEngine  SignatureEngine
	PubKeyEng  PublicKeyEngine
}

type OpcodeHandler func(ctx *ExecutionContext) error

type OpcodeRegistry struct {
	handlers map[Opcode]OpcodeHandler
}

func NewOpcodeRegistry() *OpcodeRegistry {
	r := &OpcodeRegistry{
		handlers: make(map[Opcode]OpcodeHandler),
	}
	r.registerDefaults()
	return r
}

func (r *OpcodeRegistry) Register(op Opcode, handler OpcodeHandler) {
	r.handlers[op] = handler
}

func (r *OpcodeRegistry) Get(op Opcode) (OpcodeHandler, error) {
	h, ok := r.handlers[op]
	if !ok {
		return nil, fmt.Errorf("unknown opcode 0x%x", op)
	}
	return h, nil
}
