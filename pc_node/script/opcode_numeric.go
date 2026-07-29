package script

import "bytes"

func (r *OpcodeRegistry) registerNumericOps() {
	r.Register(OP_EQUAL, func(ctx *ExecutionContext) error {
		val1, err := ctx.Stack.Pop()
		if err != nil {
			return err
		}
		val2, err := ctx.Stack.Pop()
		if err != nil {
			return err
		}

		if bytes.Equal(val1, val2) {
			return ctx.Stack.Push([]byte{1})
		}
		return ctx.Stack.Push([]byte{0})
	})
}
