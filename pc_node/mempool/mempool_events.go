package mempool

type MempoolEvents struct {
	OnTransactionAdded     func(hash string)
	OnTransactionRemoved   func(hash string)
	OnTransactionExpired   func(hash string)
	OnDuplicateTransaction func(hash string)
	OnPoolOverflow         func()
	OnValidationFailed     func(hash string, reason string)
}
