package mining

type TemplateEvents struct {
	OnTemplateCreated     func(height uint64)
	OnTemplateUpdated     func(height uint64)
	OnTemplateExpired     func()
	OnTemplateRejected    func(reason string)
	OnCoinbaseGenerated   func(reward uint64, fees uint64)
	OnTransactionSelected func(hash string)
	OnValidationFailed    func(reason string)
}
