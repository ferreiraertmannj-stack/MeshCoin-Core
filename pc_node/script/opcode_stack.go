package script

func (r *OpcodeRegistry) registerStackOps() {
	r.Register(OP_DUP, func(ctx *ExecutionContext) error {
		val, err := ctx.Stack.Peek()
		if err != nil {
			return err
		}
		// Deep copy to avoid reference issues
		dup := make([]byte, len(val))
		copy(dup, val)
		return ctx.Stack.Push(dup)
	})

	r.Register(OP_DROP, func(ctx *ExecutionContext) error {
		_, err := ctx.Stack.Pop()
		return err
	})
}
