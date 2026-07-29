package storage

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestStorageEngine_PersistAndRecover(t *testing.T) {
	policy := DefaultStoragePolicy()
	engine := NewStorageEngine(policy, StorageEvents{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)
	defer engine.Close()

	// 1. Save Block
	b := &Block{Hash: "b1", Height: 1, Data: []byte("blockdata")}
	err := engine.SaveBlock(b)
	if err != nil {
		t.Fatalf("Failed to save block: %v", err)
	}

	loadedB, err := engine.LoadBlock("b1")
	if err != nil || loadedB.Height != 1 {
		t.Fatalf("Failed to load block")
	}

	// 2. Save Tx
	tx := &Transaction{Hash: "tx1", Data: []byte("txdata")}
	err = engine.SaveTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to save tx: %v", err)
	}

	loadedTx, err := engine.LoadTransaction("tx1")
	if err != nil || loadedTx.Hash != "tx1" {
		t.Fatalf("Failed to load tx")
	}

	// 3. Save UTXO
	u := &UTXOEntry{Outpoint: "tx1:0", Data: []byte("utxodata")}
	err = engine.SaveUTXO(u)
	if err != nil {
		t.Fatalf("Failed to save utxo: %v", err)
	}

	loadedU, err := engine.LoadUTXO("tx1:0")
	if err != nil || loadedU.Outpoint != "tx1:0" {
		t.Fatalf("Failed to load utxo")
	}

	// 4. Chain State
	cs := ChainState{Height: 100, BestBlock: "best"}
	engine.SaveChainState(cs)

	loadedCS, err := engine.LoadChainState()
	if err != nil || loadedCS.Height != 100 {
		t.Fatalf("Failed to load chain state")
	}
}

func TestStorageEngine_Concurrency(t *testing.T) {
	policy := DefaultStoragePolicy()
	engine := NewStorageEngine(policy, StorageEvents{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)
	defer engine.Close()

	routines := 1000
	var wg sync.WaitGroup

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			b := &Block{Hash: fmt.Sprintf("b%d", id), Height: uint64(id), Data: []byte("d")}
			_ = engine.SaveBlock(b)
		}(i)
	}

	wg.Wait()

	stats := engine.GetStatistics()
	if stats.BlocksSaved != uint64(routines) {
		t.Fatalf("Expected %d blocks saved, got %d", routines, stats.BlocksSaved)
	}
}

func TestStorageEngine_SnapshotAndCompaction(t *testing.T) {
	policy := DefaultStoragePolicy()
	engine := NewStorageEngine(policy, StorageEvents{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)
	defer engine.Close()

	_, err := engine.Snapshot()
	if err != nil {
		t.Fatalf("Failed to create snapshot: %v", err)
	}

	err = engine.Compact(ctx)
	if err != nil {
		t.Fatalf("Failed to compact: %v", err)
	}

	stats := engine.GetStatistics()
	if stats.SnapshotsCreated != 1 {
		t.Fatalf("Expected 1 snapshot, got %d", stats.SnapshotsCreated)
	}
	if stats.Compactions != 1 {
		t.Fatalf("Expected 1 compaction, got %d", stats.Compactions)
	}
}
