package utxo

type UTXOEvents struct {
	OnUTXOCreated          func(outpoint OutPoint, utxo UTXO)
	OnUTXOSpent            func(outpoint OutPoint)
	OnTransactionValidated func(txHash string)
	OnTransactionRejected  func(txHash string, reason string)
	OnSnapshotCreated      func(snapshotID string)
	OnRollback             func(height uint64)
}
