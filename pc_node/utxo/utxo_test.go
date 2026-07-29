package utxo

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestUTXOEngine_InsertAndLookup(t *testing.T) {
	policy := DefaultUTXOPolicy()
	events := UTXOEvents{}
	engine := NewUTXOEngine(policy, events, &MockSignatureValidator{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	// Create a coinbase block to mint some UTXOs
	cb := Transaction{
		Hash: "cb_tx",
		Inputs: []TransactionInput{
			{PreviousOutPoint: OutPoint{TxHash: ""}}, // empty implies coinbase
		},
		Outputs: []TransactionOutput{
			{Value: 50, Script: []byte("script1")},
			{Value: 25, Script: []byte("script2")},
		},
	}

	block := &Block{Height: 1, Hash: "block1", Transactions: []Transaction{cb}}

	err := engine.AddBlock(block)
	if err != nil {
		t.Fatalf("Failed to add block: %v", err)
	}

	// Check existence
	op1 := OutPoint{TxHash: "cb_tx", Index: 0}
	if !engine.HasUTXO(op1) {
		t.Fatalf("Expected UTXO to exist")
	}

	u, ok := engine.GetUTXO(op1)
	if !ok || u.Value != 50 {
		t.Fatalf("Expected value 50, got %d", u.Value)
	}
}

func TestUTXOEngine_DoubleSpend(t *testing.T) {
	policy := DefaultUTXOPolicy()
	engine := NewUTXOEngine(policy, UTXOEvents{}, &MockSignatureValidator{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	// Setup initial UTXO
	cb := Transaction{
		Hash:    "cb_tx",
		Inputs:  []TransactionInput{{PreviousOutPoint: OutPoint{TxHash: ""}}},
		Outputs: []TransactionOutput{{Value: 50, Script: []byte("script1")}},
	}
	engine.AddBlock(&Block{Height: 1, Transactions: []Transaction{cb}})

	// Create Tx spending same UTXO twice
	op := OutPoint{TxHash: "cb_tx", Index: 0}
	tx := Transaction{
		Hash: "tx1",
		Inputs: []TransactionInput{
			{PreviousOutPoint: op},
			{PreviousOutPoint: op}, // Double spend!
		},
		Outputs: []TransactionOutput{
			{Value: 40, Script: []byte("script2")},
		},
	}

	err := engine.ValidateTransaction(&tx)
	if err == nil {
		t.Fatalf("Expected double spend error")
	}
}

func TestUTXOEngine_NonExistentUTXO(t *testing.T) {
	policy := DefaultUTXOPolicy()
	engine := NewUTXOEngine(policy, UTXOEvents{}, &MockSignatureValidator{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	tx := Transaction{
		Hash: "tx1",
		Inputs: []TransactionInput{
			{PreviousOutPoint: OutPoint{TxHash: "ghost", Index: 0}},
		},
		Outputs: []TransactionOutput{{Value: 10, Script: []byte("script")}},
	}

	err := engine.ValidateTransaction(&tx)
	if err == nil {
		t.Fatalf("Expected non-existent UTXO error")
	}
}

func TestUTXOEngine_Snapshot(t *testing.T) {
	policy := DefaultUTXOPolicy()
	engine := NewUTXOEngine(policy, UTXOEvents{}, &MockSignatureValidator{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	cb := Transaction{
		Hash:    "cb_tx",
		Inputs:  []TransactionInput{{PreviousOutPoint: OutPoint{TxHash: ""}}},
		Outputs: []TransactionOutput{{Value: 50, Script: []byte("script1")}},
	}
	engine.AddBlock(&Block{Height: 1, Transactions: []Transaction{cb}})

	snap := engine.Snapshot()

	u, ok := snap.Get(OutPoint{TxHash: "cb_tx", Index: 0})
	if !ok || u.Value != 50 {
		t.Fatalf("Expected snapshot to contain UTXO")
	}
}

func TestUTXOEngine_ConcurrencyAndQueue(t *testing.T) {
	policy := DefaultUTXOPolicy()
	policy.MaxWorkers = 10
	engine := NewUTXOEngine(policy, UTXOEvents{}, &MockSignatureValidator{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	routines := 1000
	var wg sync.WaitGroup

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cb := Transaction{
				Hash:    fmt.Sprintf("cb_tx_%d", id),
				Inputs:  []TransactionInput{{PreviousOutPoint: OutPoint{TxHash: ""}}},
				Outputs: []TransactionOutput{{Value: 50, Script: []byte("script1")}},
			}
			block := &Block{Height: uint64(id), Transactions: []Transaction{cb}}
			_ = engine.AddBlock(block)
		}(i)
	}

	wg.Wait()

	stats := engine.GetStatistics()
	if stats.UTXOsCreated != uint64(routines) {
		t.Fatalf("Expected %d UTXOs created, got %d", routines, stats.UTXOsCreated)
	}
}

func TestUTXOEngine_Rollback(t *testing.T) {
	policy := DefaultUTXOPolicy()
	engine := NewUTXOEngine(policy, UTXOEvents{}, &MockSignatureValidator{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	cb := Transaction{
		Hash:    "cb_tx",
		Inputs:  []TransactionInput{{PreviousOutPoint: OutPoint{TxHash: ""}}},
		Outputs: []TransactionOutput{{Value: 50, Script: []byte("script1")}},
	}
	block := &Block{Height: 1, Transactions: []Transaction{cb}}
	engine.AddBlock(block)

	op := OutPoint{TxHash: "cb_tx", Index: 0}
	if !engine.HasUTXO(op) {
		t.Fatalf("UTXO should exist")
	}

	// Rollback
	engine.RollbackBlock(block)

	if engine.HasUTXO(op) {
		t.Fatalf("UTXO should be removed after rollback")
	}
}
