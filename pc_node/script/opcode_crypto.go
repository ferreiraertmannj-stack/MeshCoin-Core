package script

func (r *OpcodeRegistry) registerCryptoOps() {
	r.Register(OP_HASH256, func(ctx *ExecutionContext) error {
		val, err := ctx.Stack.Pop()
		if err != nil {
			return err
		}
		hash := ctx.HashEngine.Hash256(val)
		return ctx.Stack.Push(hash)
	})

	r.Register(OP_CHECKSIG, func(ctx *ExecutionContext) error {
		pubKeyBytes, err := ctx.Stack.Pop()
		if err != nil {
			return err
		}
		sigBytes, err := ctx.Stack.Pop()
		if err != nil {
			return err
		}

		pubKey, err := ctx.PubKeyEng.ParsePubKey(pubKeyBytes)
		if err != nil {
			return ctx.Stack.Push([]byte{0}) // push false
		}

		valid := ctx.SigEngine.Verify(ctx.TxHash, pubKey, sigBytes)
		if valid {
			return ctx.Stack.Push([]byte{1}) // true
		}
		return ctx.Stack.Push([]byte{0}) // false
	})
}
