package script

type ScriptEvents struct {
	OnScriptExecuted    func(txHash string)
	OnScriptFailed      func(txHash string, reason string)
	OnSignatureVerified func()
	OnSignatureRejected func()
	OnOpcodeExecuted    func(opcode byte)
	OnCacheHit          func()
	OnCacheMiss         func()
}
