package script

import (
	"bytes"
	"fmt"
)

func (r *OpcodeRegistry) registerControlOps() {
	r.Register(OP_VERIFY, func(ctx *ExecutionContext) error {
		val, err := ctx.Stack.Pop()
		if err != nil {
			return err
		}
		if bytes.Equal(val, []byte{0}) || len(val) == 0 { // Simplistic false
			return fmt.Errorf("OP_VERIFY failed")
		}
		return nil
	})

	r.Register(OP_EQUALVERIFY, func(ctx *ExecutionContext) error {
		val1, err := ctx.Stack.Pop()
		if err != nil {
			return err
		}
		val2, err := ctx.Stack.Pop()
		if err != nil {
			return err
		}

		if !bytes.Equal(val1, val2) {
			return fmt.Errorf("OP_EQUALVERIFY failed")
		}
		return nil
	})

	r.Register(OP_RETURN, func(ctx *ExecutionContext) error {
		return fmt.Errorf("script terminated by OP_RETURN")
	})
}
