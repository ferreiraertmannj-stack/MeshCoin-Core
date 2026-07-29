package sync

import (
	"encoding/json"
	"sync"
	"testing"
)

func createValidBlockJSON(index int) []byte {
	b := blockTemplate{
		Index:        index,
		Timestamp:    1625097600,
		PreviousHash: "prev_hash",
		Hash:         "valid_hash",
		Nonce:        12345,
		Transactions: []transactionTemplate{
			{ID: "tx1", SenderAddress: "A", ReceiverAddress: "B", Amount: 10},
			{ID: "tx2", SenderAddress: "C", ReceiverAddress: "D", Amount: 20},
		},
	}
	if index == 0 {
		b.PreviousHash = "" // Genesis block
	}
	data, _ := json.Marshal(b)
	return data
}

func TestValidator_ValidBlock(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})

	validData := createValidBlockJSON(1)
	err := v.ValidateBlock(validData)
	if err != nil {
		t.Fatalf("Expected valid block, got error: %v", err)
	}

	stats := v.Statistics()
	if stats.BlocksValidated != 1 || stats.BlocksRejected != 0 {
		t.Fatalf("Invalid stats: %+v", stats)
	}
}

func TestValidator_InvalidHash(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})

	b := blockTemplate{
		Index:        1,
		Timestamp:    1625097600,
		PreviousHash: "prev_hash",
		Hash:         "", // Missing hash
		Nonce:        12345,
	}
	data, _ := json.Marshal(b)

	err := v.ValidateBlock(data)
	if err == nil {
		t.Fatalf("Expected error for missing hash")
	}
}

func TestValidator_InvalidPreviousHash(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})

	b := blockTemplate{
		Index:        1,
		Timestamp:    1625097600,
		PreviousHash: "", // Missing prev hash for non-genesis
		Hash:         "valid",
		Nonce:        12345,
	}
	data, _ := json.Marshal(b)

	err := v.ValidateBlock(data)
	if err == nil {
		t.Fatalf("Expected error for missing previous hash")
	}
}

func TestValidator_InvalidNonce(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})

	b := blockTemplate{
		Index:        1,
		Timestamp:    1625097600,
		PreviousHash: "prev",
		Hash:         "valid",
		Nonce:        -1, // Invalid nonce
	}
	data, _ := json.Marshal(b)

	err := v.ValidateBlock(data)
	if err == nil {
		t.Fatalf("Expected error for invalid nonce")
	}
}

func TestValidator_InvalidTimestamp(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})

	b := blockTemplate{
		Index:        1,
		Timestamp:    0, // Invalid timestamp
		PreviousHash: "prev",
		Hash:         "valid",
		Nonce:        123,
	}
	data, _ := json.Marshal(b)

	err := v.ValidateBlock(data)
	if err == nil {
		t.Fatalf("Expected error for invalid timestamp")
	}
}

func TestValidator_CorruptedBytes(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})
	err := v.ValidateBlock([]byte("{invalid_json: true}"))
	if err == nil {
		t.Fatalf("Expected error for corrupted bytes")
	}
}

func TestValidator_DuplicateTransactions(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})

	b := blockTemplate{
		Index:        1,
		Timestamp:    1625097600,
		PreviousHash: "prev",
		Hash:         "valid",
		Nonce:        123,
		Transactions: []transactionTemplate{
			{ID: "tx1", Amount: 10},
			{ID: "tx1", Amount: 20}, // Duplicate ID
		},
	}
	data, _ := json.Marshal(b)

	err := v.ValidateBlock(data)
	if err == nil {
		t.Fatalf("Expected error for duplicate transaction IDs")
	}
}

func TestValidator_ChunkValidation(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})

	chunk := DownloadedChunk{
		StartHeight: 1,
		EndHeight:   2,
		Blocks: [][]byte{
			createValidBlockJSON(1),
			createValidBlockJSON(2),
		},
	}

	err := v.ValidateChunk(chunk)
	if err != nil {
		t.Fatalf("Expected chunk to be valid, got: %v", err)
	}

	// Partially invalid chunk
	invalidChunk := DownloadedChunk{
		StartHeight: 3,
		EndHeight:   4,
		Blocks: [][]byte{
			createValidBlockJSON(3),
			[]byte("invalid block data"),
		},
	}

	err = v.ValidateChunk(invalidChunk)
	if err == nil {
		t.Fatalf("Expected error for partially invalid chunk")
	}
}

func TestValidator_Concurrency(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})
	validData := createValidBlockJSON(1)

	var wg sync.WaitGroup
	// 300 goroutines validando blocos simultaneamente
	for i := 0; i < 300; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := v.ValidateBlock(validData)
			if err != nil {
				t.Errorf("Unexpected concurrency error: %v", err)
			}
		}()
	}

	wg.Wait()

	stats := v.Statistics()
	if stats.BlocksValidated != 300 {
		t.Fatalf("Expected 300 validated blocks, got %d", stats.BlocksValidated)
	}
	if stats.Errors != 0 {
		t.Fatalf("Expected 0 errors, got %d", stats.Errors)
	}
}

func TestValidator_Stress(t *testing.T) {
	v := NewBlockValidator(BlockValidatorEventHandlers{})
	validData := createValidBlockJSON(1)

	// Validar 20.000 blocos
	for i := 0; i < 20000; i++ {
		err := v.ValidateBlock(validData)
		if err != nil {
			t.Fatalf("Unexpected error at block %d: %v", i, err)
		}
	}

	stats := v.Statistics()
	if stats.BlocksValidated != 20000 {
		t.Fatalf("Expected 20000 validated blocks, got %d", stats.BlocksValidated)
	}
	if stats.BlocksRejected != 0 {
		t.Fatalf("Expected 0 rejected blocks, got %d", stats.BlocksRejected)
	}
}
