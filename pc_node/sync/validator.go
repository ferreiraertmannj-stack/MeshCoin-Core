package sync

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"pc_node/storage"
)

// ValidationStatistics tracks metrics for the BlockValidator.
type ValidationStatistics struct {
	BlocksValidated uint64
	BlocksRejected  uint64
	BytesValidated  uint64
	ValidationSpeed float64
	StartTime       time.Time
	ElapsedTime     time.Duration
	Errors          uint64
}

// BlockValidatorEventHandlers holds callbacks for the validation pipeline.
type BlockValidatorEventHandlers struct {
	OnBlockValidated      func(index int)
	OnChunkValidated      func(chunk DownloadedChunk)
	OnValidationError     func(err error)
	OnValidationCompleted func(stats ValidationStatistics)
}

// StandardBlockValidator validates downloaded blocks before import.
type StandardBlockValidator struct {
	mu     sync.RWMutex
	events BlockValidatorEventHandlers

	stats ValidationStatistics
}

// BlockTemplate defines the structure required for validation
// without importing the main ledger package to maintain decoupling.
type blockTemplate struct {
	Index        int                   `json:"index"`
	Timestamp    int64                 `json:"timestamp"`
	PreviousHash string                `json:"previousHash"`
	Hash         string                `json:"hash"`
	Nonce        int                   `json:"nonce"`
	Transactions []transactionTemplate `json:"transactions"`
}

type transactionTemplate struct {
	ID              string  `json:"id"`
	SenderAddress   string  `json:"senderAddress"`
	ReceiverAddress string  `json:"receiverAddress"`
	Amount          float64 `json:"amount"`
}

// NewBlockValidator initializes a new validator.
func NewBlockValidator(events BlockValidatorEventHandlers) *StandardBlockValidator {
	return &StandardBlockValidator{
		events: events,
		stats: ValidationStatistics{
			StartTime: time.Now(),
		},
	}
}

// ValidateBlock parses the byte data into a local block structure and validates its fields.
func (v *StandardBlockValidator) ValidateBlock(data []byte) error {
	if len(data) == 0 {
		return v.registerError(errors.New("block data is empty"))
	}

	var block blockTemplate
	if err := storage.UnmarshalBlock(data, &block); err != nil {
		return v.registerError(fmt.Errorf("failed to unmarshal block: %w", err))
	}

	if block.Index < 0 {
		return v.registerError(fmt.Errorf("invalid block index: %d", block.Index))
	}

	if block.Index > 0 && block.PreviousHash == "" {
		return v.registerError(errors.New("missing previousHash in non-genesis block"))
	}

	if block.Hash == "" {
		return v.registerError(errors.New("missing block hash"))
	}

	if block.Timestamp <= 0 {
		return v.registerError(fmt.Errorf("invalid timestamp: %d", block.Timestamp))
	}

	if block.Nonce < 0 {
		return v.registerError(fmt.Errorf("invalid nonce: %d", block.Nonce))
	}

	// Validate transactions
	txIDs := make(map[string]bool)
	for i, tx := range block.Transactions {
		if tx.ID == "" {
			return v.registerError(fmt.Errorf("missing transaction ID at index %d", i))
		}
		if txIDs[tx.ID] {
			return v.registerError(fmt.Errorf("duplicate transaction ID found: %s", tx.ID))
		}
		txIDs[tx.ID] = true
	}

	v.registerSuccess(uint64(len(data)))

	if v.events.OnBlockValidated != nil {
		v.events.OnBlockValidated(block.Index)
	}

	return nil
}

// ValidateBlocks iterates over an array of block byte slices and validates each one.
func (v *StandardBlockValidator) ValidateBlocks(blocks [][]byte) error {
	for _, blockData := range blocks {
		if err := v.ValidateBlock(blockData); err != nil {
			return err
		}
	}
	return nil
}

// ValidateChunk validates all blocks in a DownloadedChunk.
func (v *StandardBlockValidator) ValidateChunk(chunk DownloadedChunk) error {
	if err := v.ValidateBlocks(chunk.Blocks); err != nil {
		return err
	}

	if v.events.OnChunkValidated != nil {
		v.events.OnChunkValidated(chunk)
	}

	return nil
}

// registerSuccess updates statistics for a successful validation.
func (v *StandardBlockValidator) registerSuccess(bytesCount uint64) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.stats.BlocksValidated++
	v.stats.BytesValidated += bytesCount
	v.stats.ElapsedTime = time.Since(v.stats.StartTime)

	if v.stats.ElapsedTime.Seconds() > 0 {
		v.stats.ValidationSpeed = float64(v.stats.BlocksValidated) / v.stats.ElapsedTime.Seconds()
	}
}

// registerError updates statistics for a failed validation and emits an event.
func (v *StandardBlockValidator) registerError(err error) error {
	v.mu.Lock()
	v.stats.BlocksRejected++
	v.stats.Errors++
	v.mu.Unlock()

	if v.events.OnValidationError != nil {
		v.events.OnValidationError(err)
	}
	return err
}

// Statistics returns a thread-safe copy of the current validation statistics.
func (v *StandardBlockValidator) Statistics() ValidationStatistics {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.stats
}

// ResetStatistics resets all metrics.
func (v *StandardBlockValidator) ResetStatistics() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.stats = ValidationStatistics{
		StartTime: time.Now(),
	}
}

// FinishValidation emits the OnValidationCompleted event with final stats.
func (v *StandardBlockValidator) FinishValidation() {
	v.mu.RLock()
	finalStats := v.stats
	v.mu.RUnlock()

	if v.events.OnValidationCompleted != nil {
		v.events.OnValidationCompleted(finalStats)
	}
}
