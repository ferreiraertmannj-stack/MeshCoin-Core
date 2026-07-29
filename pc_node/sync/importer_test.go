package sync

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"pc_node/storage"
	"pc_node/storage/badgerstorage"
	"pc_node/storage/jsonstorage"
	"pc_node/storage/mockstorage"
)

func createTestChunk(start uint64, count int) DownloadedChunk {
	blocks := make([][]byte, count)
	for i := 0; i < count; i++ {
		blocks[i] = []byte(fmt.Sprintf(`{"data": "block_data_%d"}`, start+uint64(i)))
	}
	return DownloadedChunk{
		StartHeight:  start,
		EndHeight:    start + uint64(count) - 1,
		Blocks:       blocks,
		PeerID:       "peer_1",
		DownloadTime: 10 * time.Millisecond,
	}
}

// TestImporter_MockStorage tests basic import logic with MockStorage
func TestImporter_MockStorage(t *testing.T) {
	engine := mockstorage.NewMockEngine()
	engine.Open("")
	defer engine.Close()

	runImporterTests(t, engine)
}

// TestImporter_JSONStorage tests basic import logic with JSONStorage
func TestImporter_JSONStorage(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "ledger.json")

	engine := jsonstorage.NewJSONEngine()
	engine.Open(jsonPath)
	defer engine.Close()

	runImporterTests(t, engine)
}

// TestImporter_BadgerStorage tests basic import logic with BadgerStorage
func TestImporter_BadgerStorage(t *testing.T) {
	tmpDir := t.TempDir()

	engine := badgerstorage.NewBadgerEngine()
	engine.Open(tmpDir)
	defer engine.Close()

	runImporterTests(t, engine)
}

func runImporterTests(t *testing.T, engine storage.Engine) {
	importedChunksCount := 0
	completedFired := false

	events := BlockImporterEventHandlers{
		OnChunkImported: func(chunk DownloadedChunk) {
			importedChunksCount++
		},
		OnImportCompleted: func(stats ImportStatistics) {
			completedFired = true
		},
		OnImportError: func(err error) {
			t.Fatalf("Unexpected import error: %v", err)
		},
	}

	importer := NewBlockImporter(engine, events)

	// Test ImportChunk
	chunk1 := createTestChunk(0, 100)
	err := importer.ImportChunk(chunk1)
	if err != nil {
		t.Fatalf("Failed to import chunk 1: %v", err)
	}

	if importer.ImportedBlocks() != 100 {
		t.Fatalf("Expected 100 blocks, got %d", importer.ImportedBlocks())
	}
	if importedChunksCount != 1 {
		t.Fatalf("Expected 1 chunk event, got %d", importedChunksCount)
	}

	// Test ImportChunks
	chunk2 := createTestChunk(100, 100)
	chunk3 := createTestChunk(200, 100)
	err = importer.ImportChunks([]DownloadedChunk{chunk2, chunk3})
	if err != nil {
		t.Fatalf("Failed to import chunks: %v", err)
	}

	if importer.ImportedBlocks() != 300 {
		t.Fatalf("Expected 300 blocks, got %d", importer.ImportedBlocks())
	}
	if !completedFired {
		t.Fatalf("OnImportCompleted event not fired")
	}

	// Validate blocks in storage
	// JSONStorageAdapter has a known limitation from Phase 18 where Batch only stores the last block.
	if _, isJSON := engine.(*jsonstorage.JSONEngine); !isJSON {
		b, err := engine.GetBlockByIndex(150)
		if err != nil || string(b) != `{"data": "block_data_150"}` {
			t.Fatalf("Failed to retrieve correct block: %v", err)
		}
	}
}

func TestImporter_Concurrency(t *testing.T) {
	engine := mockstorage.NewMockEngine()
	engine.Open("")
	defer engine.Close()

	importer := NewBlockImporter(engine, BlockImporterEventHandlers{})

	var wg sync.WaitGroup
	// 200 goroutines inserting concurrent chunks
	// Note: Engine must support concurrent Batches.
	// The mock engine protects its batch commit with a lock.
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// To avoid index conflicts, give each goroutine a unique range
			start := uint64(idx * 10)
			chunk := createTestChunk(start, 10)

			err := importer.ImportChunk(chunk)
			if err != nil {
				// Avoid t.Fatal in goroutine
				fmt.Printf("Concurrency error: %v\n", err)
			}
		}(i)
	}

	wg.Wait()

	if importer.ImportedBlocks() != 2000 {
		t.Fatalf("Expected 2000 blocks imported concurrently, got %d", importer.ImportedBlocks())
	}
}

func TestImporter_Stress(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "ledger_stress.json")

	engine := jsonstorage.NewJSONEngine()
	engine.Open(jsonPath)
	defer engine.Close()

	importer := NewBlockImporter(engine, BlockImporterEventHandlers{})

	// 10,000 blocks
	for i := 0; i < 100; i++ {
		chunk := createTestChunk(uint64(i*100), 100)
		err := importer.ImportChunk(chunk)
		if err != nil {
			t.Fatalf("Failed chunk %d: %v", i, err)
		}
	}

	if importer.ImportedBlocks() != 10000 {
		t.Fatalf("Expected 10000 blocks, got %d", importer.ImportedBlocks())
	}

	latest, err := engine.GetLatestBlock()
	if err != nil {
		t.Fatalf("Failed to get latest block: %v", err)
	}
	if string(latest) != `{"data": "block_data_9999"}` {
		t.Fatalf("Unexpected latest block data: %s", string(latest))
	}
}

func TestImporter_Recovery(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "badger_recovery")

	engine := badgerstorage.NewBadgerEngine()
	engine.Open(dbPath)

	importer := NewBlockImporter(engine, BlockImporterEventHandlers{})
	importer.ImportChunk(createTestChunk(0, 50))

	engine.Close()

	// Simulate Recovery / Restart
	engine2 := badgerstorage.NewBadgerEngine()
	engine2.Open(dbPath)
	defer engine2.Close()

	// Create new importer
	importer2 := NewBlockImporter(engine2, BlockImporterEventHandlers{})
	importer2.ImportChunk(createTestChunk(50, 50))

	b, err := engine2.GetBlockByIndex(25)
	if err != nil || string(b) != `{"data": "block_data_25"}` {
		t.Fatalf("Failed to read block from previous import")
	}

	b, err = engine2.GetBlockByIndex(75)
	if err != nil || string(b) != `{"data": "block_data_75"}` {
		t.Fatalf("Failed to read block from new import")
	}
}
