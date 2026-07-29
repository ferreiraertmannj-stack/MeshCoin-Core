package mempool

type TransactionNetworkEvents struct {
	OnTransactionReceived   func(hash string)
	OnTransactionAccepted   func(hash string)
	OnTransactionRejected   func(hash string, reason string)
	OnTransactionExpired    func(hash string)
	OnTransactionPropagated func(hash string)
	OnDuplicateTransaction  func(hash string)
	OnValidationFailed      func(hash string, reason string)
	OnAnnouncementReceived  func(hash string)
}
