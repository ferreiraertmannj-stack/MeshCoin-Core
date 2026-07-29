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

var blockData = []byte(`{"index":1,"timestamp":1700000000,"previousHash":"0000","merkleRoot":"root","nonce":12345,"hash":"0000abcd","minerStorage":100,"storageType":"SSD","transactions":[]}`)

func setupEngine(b *testing.B, engineType string) (storage.Engine, string, func()) {
	b.Helper()
	var engine storage.Engine
	var connStr string
	dir := b.TempDir()

	switch engineType {
	case "Mock":
		engine = mockstorage.NewMockEngine()
		connStr = ""
	case "JSON":
		engine = jsonstorage.NewJSONEngine()
		connStr = filepath.Join(dir, "bench_ledger.json")
		os.WriteFile(connStr, []byte("[]"), 0644)
	case "Badger":
		engine = badgerstorage.NewBadgerEngine()
		connStr = filepath.Join(dir, "bench_badger")
	}

	if err := engine.Open(connStr); err != nil {
		b.Fatalf("Failed to open engine %s: %v", engineType, err)
	}

	cleanup := func() {
		engine.Close()
	}

	return engine, connStr, cleanup
}

func populateEngine(b *testing.B, engine storage.Engine, numBlocks int) {
	b.Helper()
	for i := 0; i < numBlocks; i++ {
		batch := engine.NewBatch()
		batch.PutBlock(uint64(i), blockData)
		batch.PutBalance(fmt.Sprintf("ADDR_%d", i), float64(i*10))
		batch.Commit()
	}
}

// 1. Escrita de 1 bloco
func BenchmarkWrite1Block(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) {
			engine, _, cleanup := setupEngine(b, e)
			defer cleanup()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				batch := engine.NewBatch()
				batch.PutBlock(uint64(i), blockData)
				batch.Commit()
			}
		})
	}
}

// 2, 3, 4. Escrita de N blocos seq (simulando sync inicial ou carga pesada)
func benchmarkWriteNBlocks(b *testing.B, engineType string, n int) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		engine, _, cleanup := setupEngine(b, engineType)
		b.StartTimer()

		for j := 0; j < n; j++ {
			batch := engine.NewBatch()
			batch.PutBlock(uint64(j), blockData)
			batch.Commit()
		}

		b.StopTimer()
		cleanup()
		b.StartTimer()
	}
}

func BenchmarkWrite100Blocks(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) { benchmarkWriteNBlocks(b, e, 100) })
	}
}

func BenchmarkWrite1000Blocks(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		if e == "JSON" {
			continue // Pular JSON para 1000 blocos pois é O(N^2) via arquivos grandes e causará timeout no CI
		}
		b.Run(e, func(b *testing.B) { benchmarkWriteNBlocks(b, e, 1000) })
	}
}

func BenchmarkWrite10000Blocks(b *testing.B) {
	for _, e := range []string{"Mock", "Badger"} {
		b.Run(e, func(b *testing.B) { benchmarkWriteNBlocks(b, e, 10000) })
	}
}

// 5. GetLatestBlock
func BenchmarkGetLatestBlock(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) {
			engine, _, cleanup := setupEngine(b, e)
			defer cleanup()
			populateEngine(b, engine, 10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = engine.GetLatestBlock()
			}
		})
	}
}

// 6. GetBlockByIndex
func BenchmarkGetBlockByIndex(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) {
			engine, _, cleanup := setupEngine(b, e)
			defer cleanup()
			populateEngine(b, engine, 100)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = engine.GetBlockByIndex(uint64(i % 100))
			}
		})
	}
}

// 7. Iteração completa
func BenchmarkIterateChain(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) {
			engine, _, cleanup := setupEngine(b, e)
			defer cleanup()
			populateEngine(b, engine, 100)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				it := engine.NewBlockIterator()
				for it.Next() {
					_ = it.Value()
				}
				it.Close()
			}
		})
	}
}

// 8. GetBalance
func BenchmarkGetBalance(b *testing.B) {
	for _, e := range []string{"Mock", "Badger"} {
		b.Run(e, func(b *testing.B) {
			engine, _, cleanup := setupEngine(b, e)
			defer cleanup()
			populateEngine(b, engine, 10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = engine.GetBalance("ADDR_5")
			}
		})
	}
}

// 9. Batch Commit isolado (overhead)
func BenchmarkBatchCommit(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) {
			engine, _, cleanup := setupEngine(b, e)
			defer cleanup()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				batch := engine.NewBatch()
				batch.PutBalance("TEST", float64(i)) // JSON suporta gracefully com unsupported
				batch.Commit()
			}
		})
	}
}

// 10. Open
func BenchmarkOpen(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				engine, connStr, cleanup := setupEngine(b, e)
				engine.Close() // Fecha o que foi aberto no setup
				b.StartTimer()

				_ = engine.Open(connStr)

				b.StopTimer()
				cleanup()
				b.StartTimer()
			}
		})
	}
}

// 11. Close
func BenchmarkClose(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				engine, _, _ := setupEngine(b, e)
				b.StartTimer()

				_ = engine.Close()
			}
		})
	}
}

// 12. Reopen
func BenchmarkReopen(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) {
			engine, connStr, cleanup := setupEngine(b, e)
			defer cleanup()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				engine.Close()
				_ = engine.Open(connStr)
			}
		})
	}
}

// 13. Leitura concorrente
func BenchmarkConcurrentRead(b *testing.B) {
	for _, e := range []string{"Mock", "JSON", "Badger"} {
		b.Run(e, func(b *testing.B) {
			engine, _, cleanup := setupEngine(b, e)
			defer cleanup()
			populateEngine(b, engine, 100)

			b.ResetTimer()
			var wg sync.WaitGroup
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					_, _ = engine.GetBlockByIndex(uint64(idx % 100))
					_, _ = engine.GetLatestBlock()
				}(i)
			}
			wg.Wait()
		})
	}
}

// 14. Escrita concorrente
func BenchmarkConcurrentWrite(b *testing.B) {
	for _, e := range []string{"Mock", "Badger"} { // JSON serializa em array, concorrência no append causa gargalos enormes
		b.Run(e, func(b *testing.B) {
			engine, _, cleanup := setupEngine(b, e)
			defer cleanup()

			b.ResetTimer()
			var wg sync.WaitGroup
			for i := 0; i < b.N; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					batch := engine.NewBatch()
					batch.PutBlock(uint64(idx), blockData)
					batch.Commit()
				}(i)
			}
			wg.Wait()
		})
	}
}
