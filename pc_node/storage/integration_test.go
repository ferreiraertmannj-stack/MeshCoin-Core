package storage_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"pc_node/storage"
	"pc_node/storage/badgerstorage"
	"pc_node/storage/jsonstorage"
	"pc_node/storage/mockstorage"
)

func TestStorageEnginesIntegration(t *testing.T) {
	engines := []struct {
		name    string
		factory func(t *testing.T) (storage.Engine, string, func())
	}{
		{
			name: "MockStorage",
			factory: func(t *testing.T) (storage.Engine, string, func()) {
				e := mockstorage.NewMockEngine()
				return e, "", func() {}
			},
		},
		{
			name: "JSONStorage",
			factory: func(t *testing.T) (storage.Engine, string, func()) {
				e := jsonstorage.NewJSONEngine()
				dir := t.TempDir()
				path := filepath.Join(dir, "test_ledger.json")
				os.WriteFile(path, []byte("[]"), 0644)
				return e, path, func() {}
			},
		},
		{
			name: "BadgerStorage",
			factory: func(t *testing.T) (storage.Engine, string, func()) {
				e := badgerstorage.NewBadgerEngine()
				path := filepath.Join(t.TempDir(), "badger_test_db")
				return e, path, func() {}
			},
		},
	}

	for _, tc := range engines {
		t.Run(tc.name, func(t *testing.T) {
			engine, connStr, cleanup := tc.factory(t)
			defer cleanup()

			// 8. Engine Close/Open
			if err := engine.Open(connStr); err != nil {
				t.Fatalf("Failed to open engine: %v", err)
			}
			defer engine.Close()

			// 1. Inserção sequencial de blocos & 6. Batch Commit
			// JSON Engine só suporta 1 bloco por batch
			block0 := []byte(`{"index":0, "hash":"hash0"}`)
			block1 := []byte(`{"index":1, "hash":"hash1"}`)

			batch0 := engine.NewBatch()
			batch0.PutBlock(0, block0)
			if err := batch0.Commit(); err != nil {
				t.Fatalf("Batch0 Commit failed: %v", err)
			}

			batch1 := engine.NewBatch()
			batch1.PutBlock(1, block1)

			// 5. Persistência de saldos
			errBal1 := batch1.PutBalance("ADDR_A", 100.5)
			if errBal1 != nil && errBal1 != storage.ErrUnsupported {
				t.Errorf("Unexpected error in PutBalance: %v", errBal1)
			}
			if err := batch1.Commit(); err != nil {
				t.Fatalf("Batch1 Commit failed: %v", err)
			}

			// 7. Batch Discard
			block2 := []byte(`{"index":2, "hash":"hash2"}`)
			batch2 := engine.NewBatch()
			batch2.PutBlock(2, block2)
			batch2.Discard()
			// Validate block 2 is not there
			_, err := engine.GetBlockByIndex(2)
			if err != storage.ErrNotFound {
				t.Errorf("Expected ErrNotFound after Discard, got %v", err)
			}

			// 2. Recuperação do último bloco
			latest, err := engine.GetLatestBlock()
			if err != nil {
				t.Fatalf("Failed GetLatestBlock: %v", err)
			}
			var lBlock struct{ Index int }
			storage.UnmarshalBlock(latest, &lBlock)
			if lBlock.Index != 1 {
				t.Errorf("GetLatestBlock mismatch. Got index %d", lBlock.Index)
			}

			// 3. Busca por índice
			b0, err := engine.GetBlockByIndex(0)
			if err != nil {
				t.Errorf("Failed GetBlockByIndex(0): %v", err)
			}
			var b0Struct struct{ Index int }
			storage.UnmarshalBlock(b0, &b0Struct)
			if b0Struct.Index != 0 {
				t.Errorf("Block 0 mismatch")
			}

			// Leitura de saldos (5)
			bal, err := engine.GetBalance("ADDR_A")
			if err != storage.ErrUnsupported {
				if err != nil {
					t.Errorf("GetBalance error: %v", err)
				} else if bal != 100.5 {
					t.Errorf("Expected balance 100.5, got %v", bal)
				}
			}

			// 4. Iteração completa da cadeia
			it := engine.NewBlockIterator()
			count := 0
			for it.Next() {
				count++
			}
			if it.Error() != nil {
				t.Errorf("Iterator error: %v", it.Error())
			}
			it.Close()
			if count != 2 {
				t.Errorf("Iterator count expected 2, got %d", count)
			}

			// 9. Leituras concorrentes
			var wg sync.WaitGroup
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					_, _ = engine.GetLatestBlock()
					_, _ = engine.GetBlockByIndex(uint64(idx % 2))
					_, _ = engine.GetBalance("ADDR_A")
				}(i)
			}
			wg.Wait()

			// 10. Escritas concorrentes
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					cBatch := engine.NewBatch()
					// apenas escrevendo saldos de teste independentes (JSON retorna unsupported mas não deve crashar)
					_ = cBatch.PutBalance(fmt.Sprintf("ADDR_%d", idx), float64(idx))
					// Teste falharia se desse race condition.
					// Se JSON não suporta concurrent batch commit escrevendo blocos simultâneos sem corromper,
					// Ledger usa um lock global, mas a engine deve se proteger internamente se prometer ser safe.
					// O json_storage.go tem `e.mu.Lock()` no Commit. Badger usa transações isoladas.
					_ = cBatch.Commit()
				}(i)
			}
			wg.Wait()

			// 11. Snapshot (quando suportado)
			snapPath := filepath.Join(t.TempDir(), "snapshot.dat")
			errSnap := engine.CreateSnapshot(snapPath)
			if errSnap != nil && errSnap != storage.ErrUnsupported {
				t.Errorf("Unexpected error in CreateSnapshot: %v", errSnap)
			}

			// 12. Tratamento de erros padronizado
			_, errNotFound := engine.GetBlockByIndex(999)
			if errNotFound != storage.ErrNotFound {
				t.Errorf("Expected ErrNotFound, got %v", errNotFound)
			}

			// Re-open Test
			if err := engine.Close(); err != nil {
				t.Errorf("Close failed: %v", err)
			}
			if tc.name != "MockStorage" {
				if err := engine.Open(connStr); err != nil {
					t.Errorf("Re-open failed: %v", err)
				}
				defer engine.Close()
				b1, err := engine.GetBlockByIndex(1)
				if err != nil {
					t.Errorf("Data not persisted across re-open: %v", err)
				} else {
					var b1Struct struct{ Index int }
					storage.UnmarshalBlock(b1, &b1Struct)
					if b1Struct.Index != 1 {
						t.Errorf("Data not persisted across re-open: index mismatch")
					}
				}
			}
		})
	}
}
