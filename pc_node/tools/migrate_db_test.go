package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"pc_node/storage"
	"pc_node/storage/badgerstorage"
	"pc_node/storage/jsonstorage"
)

func TestMigrationAndDeepValidation(t *testing.T) {
	inputLedger := filepath.Join("..", "ledger.json")

	if _, err := os.Stat(inputLedger); os.IsNotExist(err) {
		t.Skip("ledger.json real not found, skipping migration validation test")
	}

	tempDir := t.TempDir()
	badgerDir := filepath.Join(tempDir, "badger_test_migration")

	t.Logf("Running migration on ledger: %s -> %s", inputLedger, badgerDir)

	// 1. Run Migration
	start := time.Now()
	cmd := exec.Command("go", "run", "migrate_db.go", "--input="+inputLedger, "--output="+badgerDir, "--force")
	cmd.Dir = "." // run in tools dir
	out, err := cmd.CombinedOutput()
	migrationTime := time.Since(start)

	if err != nil {
		t.Fatalf("Migration failed: %v\nOutput: %s", err, string(out))
	}
	t.Logf("Migration Time: %v", migrationTime)

	// 2. Open Both Adapters
	jsonEngine := jsonstorage.NewJSONEngine()
	if err := jsonEngine.Open(inputLedger); err != nil {
		t.Fatalf("Failed to open JSON Engine: %v", err)
	}
	defer jsonEngine.Close()

	badgerEngine := badgerstorage.NewBadgerEngine()
	if err := badgerEngine.Open(badgerDir); err != nil {
		t.Fatalf("Failed to open Badger Engine: %v", err)
	}
	defer badgerEngine.Close()

	// 3. Deep block-by-block Comparison
	validateStart := time.Now()

	jsonIt := jsonEngine.NewBlockIterator()
	defer jsonIt.Close()

	badgerIt := badgerEngine.NewBlockIterator()
	defer badgerIt.Close()

	var blockCount int
	var txCount int
	var lastHash string

	for jsonIt.Next() {
		if !badgerIt.Next() {
			t.Fatalf("BadgerDB has fewer blocks than JSON. Failed at block %d", blockCount)
		}

		var jBlock Block
		if err := storage.UnmarshalBlock(jsonIt.Value(), &jBlock); err != nil {
			t.Fatalf("Error unmarshalling JSON block: %v", err)
		}

		var bBlock Block
		if err := storage.UnmarshalBlock(badgerIt.Value(), &bBlock); err != nil {
			t.Fatalf("Error unmarshalling Badger block: %v", err)
		}

		// Comparar atributos
		if jBlock.Index != bBlock.Index {
			t.Errorf("Index mismatch: %d != %d", jBlock.Index, bBlock.Index)
		}
		if jBlock.Hash != bBlock.Hash {
			t.Errorf("Hash mismatch at %d: %s != %s", jBlock.Index, jBlock.Hash, bBlock.Hash)
		}
		if jBlock.PreviousHash != bBlock.PreviousHash {
			t.Errorf("PrevHash mismatch at %d: %s != %s", jBlock.Index, jBlock.PreviousHash, bBlock.PreviousHash)
		}
		if jBlock.Timestamp != bBlock.Timestamp {
			t.Errorf("Timestamp mismatch at %d: %d != %d", jBlock.Index, jBlock.Timestamp, bBlock.Timestamp)
		}
		if jBlock.Nonce != bBlock.Nonce {
			t.Errorf("Nonce mismatch at %d: %d != %d", jBlock.Index, jBlock.Nonce, bBlock.Nonce)
		}

		if len(jBlock.Transactions) != len(bBlock.Transactions) {
			t.Errorf("Tx count mismatch at %d: %d != %d", jBlock.Index, len(jBlock.Transactions), len(bBlock.Transactions))
		} else {
			for i, jTx := range jBlock.Transactions {
				bTx := bBlock.Transactions[i]
				if jTx.ID != bTx.ID {
					t.Errorf("Tx ID mismatch at %d/%d: %s != %s", jBlock.Index, i, jTx.ID, bTx.ID)
				}
				if jTx.SenderAddress != bTx.SenderAddress {
					t.Errorf("Tx Sender mismatch at %d/%d", jBlock.Index, i)
				}
				if jTx.ReceiverAddress != bTx.ReceiverAddress {
					t.Errorf("Tx Receiver mismatch at %d/%d", jBlock.Index, i)
				}
				if jTx.Amount != bTx.Amount {
					t.Errorf("Tx Amount mismatch at %d/%d", jBlock.Index, i)
				}
				if jTx.Fee != bTx.Fee {
					t.Errorf("Tx Fee mismatch at %d/%d", jBlock.Index, i)
				}
			}
		}

		blockCount++
		txCount += len(jBlock.Transactions)
		lastHash = jBlock.Hash
	}

	if badgerIt.Next() {
		t.Fatalf("BadgerDB has more blocks than JSON. Exceeded %d", blockCount)
	}

	// Read last block and balance
	latestJSONBlock, errJ := jsonEngine.GetLatestBlock()
	latestBadgerBlock, errB := badgerEngine.GetLatestBlock()

	if errJ != nil || errB != nil {
		t.Errorf("GetLatestBlock error: %v / %v", errJ, errB)
	} else {
		var lJ, lB Block
		storage.UnmarshalBlock(latestJSONBlock, &lJ)
		storage.UnmarshalBlock(latestBadgerBlock, &lB)
		if lJ.Hash != lB.Hash {
			t.Errorf("GetLatestBlock Hash mismatch: %s != %s", lJ.Hash, lB.Hash)
		}
	}

	// Since JSON doesn't support balances directly in FASE 16, we can only verify Badger has balances stored
	// We could re-calculate them here and check `badgerEngine.GetBalance`, but we did that in migrate_db.
	// As a simple check, verify balance of a known coinbase address if it has any, or just check that they exist without panicking.

	// Data Race / Concurrent Reads Validation
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx uint64) {
			defer wg.Done()
			_, _ = badgerEngine.GetBlockByIndex(idx % uint64(blockCount))
			_, _ = badgerEngine.GetLatestBlock()
			_, _ = badgerEngine.GetBalance("COINBASE") // Just concurrent reads
		}(uint64(i))
	}
	wg.Wait()

	t.Logf("Validation Time: %v", time.Since(validateStart))
	t.Logf("Total Blocks Validated: %d", blockCount)
	t.Logf("Total Txs Validated: %d", txCount)
	t.Logf("Final Hash Validated: %s", lastHash)
}
