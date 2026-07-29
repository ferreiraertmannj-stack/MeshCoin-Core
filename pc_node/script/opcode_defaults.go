package script

func (r *OpcodeRegistry) registerDefaults() {
	r.registerStackOps()
	r.registerCryptoOps()
	r.registerControlOps()
	r.registerNumericOps()
}
