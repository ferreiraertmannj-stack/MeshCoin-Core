package storage

type StorageEvents struct {
	OnBlockSaved        func(hash string)
	OnTransactionSaved  func(hash string)
	OnUTXOUpdated       func(outpoint string)
	OnSnapshotCreated   func(id string)
	OnDatabaseCompacted func()
	OnStorageError      func(err error)
}
