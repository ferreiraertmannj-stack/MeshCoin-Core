package storage_test

import (
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"pc_node/storage"
	"pc_node/storage/badgerstorage"
	"pc_node/storage/jsonstorage"
)

func TestFaultInjection_JSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ledger.json")
	os.WriteFile(path, []byte("[]"), 0644)

	// 1, 2, 3, 4, 5. Encerramento inesperado e Reabertura
	engine := jsonstorage.NewJSONEngine()
	if err := engine.Open(path); err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	batch := engine.NewBatch()
	batch.PutBlock(0, []byte(`{"index":0,"hash":"hash0"}`))
	batch.PutBlock(1, []byte(`{"index":1,"hash":"hash1"}`)) // Overwrites pending in JSON, but let's simulate
	_ = batch.PutBalance("ADDR_A", 50.0)

	// Encerramento INESPERADO durante batch (sem commit)
	engine.Close()
	batch.Discard() // Simulate drop

	// Reabre e verifica integridade
	engine2 := jsonstorage.NewJSONEngine()
	if err := engine2.Open(path); err != nil {
		t.Fatalf("Reopen failed: %v", err)
	}
	_, errNotFound := engine2.GetBlockByIndex(0)
	if errNotFound != storage.ErrNotFound {
		t.Errorf("Expected ErrNotFound after uncommitted crash, got %v", errNotFound)
	}

	// Commit normal e encerramento abrupto após commit
	batch2 := engine2.NewBatch()
	batch2.PutBlock(0, []byte(`{"index":0,"hash":"hash0"}`))
	batch2.Commit()
	engine2.Close() // Fecha imediatamente

	engine3 := jsonstorage.NewJSONEngine()
	engine3.Open(path)
	b0, err := engine3.GetBlockByIndex(0)
	if err != nil || len(b0) == 0 {
		t.Errorf("Block lost after commit: %v", err)
	}
	engine3.Close()

	// 6, 8, 9. Corrupção Proposital e Leituras após corrupção
	os.WriteFile(path, []byte("{CORRUPTED_JSON_123*&^"), 0644)

	engine4 := jsonstorage.NewJSONEngine()
	errCorrupted := engine4.Open(path)
	if errCorrupted != storage.ErrCorruptedData {
		t.Errorf("Expected ErrCorruptedData, got %v", errCorrupted)
	}

	// 10. Abertura Simultânea
	os.WriteFile(path, []byte("[]"), 0644)
	eA := jsonstorage.NewJSONEngine()
	eB := jsonstorage.NewJSONEngine()
	_ = eA.Open(path)
	_ = eB.Open(path) // JSON allows multiple readers, so it doesn't fail natively
	eA.Close()
	eB.Close()
}

func TestFaultInjection_Badger(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "badger_db")

	engine := badgerstorage.NewBadgerEngine()
	if err := engine.Open(dir); err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// 1, 2, 3, 4, 5. Crash Recovery
	batch := engine.NewBatch()
	batch.PutBlock(0, []byte(`{"index":0,"hash":"hash0"}`))
	_ = batch.PutBalance("ADDR_A", 50.0)

	// Encerramento inesperado (discard sem commit)
	batch.Discard()
	engine.Close()

	// Reabertura
	engine2 := badgerstorage.NewBadgerEngine()
	engine2.Open(dir)
	_, err := engine2.GetBlockByIndex(0)
	if err != storage.ErrNotFound {
		t.Errorf("Block 0 should not exist, got %v", err)
	}

	// Commit e encerramento
	batch2 := engine2.NewBatch()
	batch2.PutBlock(0, []byte(`{"index":0,"hash":"hash0"}`))
	batch2.PutBalance("ADDR_A", 100.0)
	if err := batch2.Commit(); err != nil {
		t.Errorf("Commit failed: %v", err)
	}
	engine2.Close() // Fecha abruptamente

	// Recuperação
	engine3 := badgerstorage.NewBadgerEngine()
	engine3.Open(dir)

	b0, err := engine3.GetBlockByIndex(0)
	if err != nil {
		t.Errorf("Block 0 lost: %v", err)
	}
	if string(b0) != `{"index":0,"hash":"hash0"}` {
		t.Errorf("Block 0 corrupted")
	}

	bal, err := engine3.GetBalance("ADDR_A")
	if err != nil || bal != 100.0 {
		t.Errorf("Balance lost or wrong: bal=%f err=%v", bal, err)
	}

	// 11. Recuperação após Close/Open repetidos
	for i := 0; i < 5; i++ {
		engine3.Close()
		engine3 = badgerstorage.NewBadgerEngine()
		if err := engine3.Open(dir); err != nil {
			t.Fatalf("Failed open %d: %v", i, err)
		}
	}

	// 12. Execução paralela de centenas de leituras durante recuperação
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = engine3.GetBlockByIndex(0)
			_, _ = engine3.GetBalance("ADDR_A")
			_, _ = engine3.GetLatestBlock()
		}()
	}
	wg.Wait()

	// 10. Abertura simultânea (Badger should error due to lock)
	engineSimul := badgerstorage.NewBadgerEngine()
	errSimul := engineSimul.Open(dir)
	if errSimul == nil {
		t.Errorf("Expected lock error on simultaneous open, got nil")
	} else {
		engineSimul.Close() // just in case
	}

	engine3.Close()

	// 7. Corrupção proposital do BadgerDB
	// Corrompendo a MANIFEST ou arquivo .sst para forçar o Badger a falhar na abertura
	manifestPath := filepath.Join(dir, "MANIFEST")
	if _, err := os.Stat(manifestPath); err == nil {
		// Substituir o conteúdo por lixo
		garbage := make([]byte, 1024)
		rand.Read(garbage)
		os.WriteFile(manifestPath, garbage, 0644)

		engine4 := badgerstorage.NewBadgerEngine()
		errCorrupted := engine4.Open(dir)
		if errCorrupted == nil {
			t.Errorf("Expected error when opening corrupted BadgerDB")
			engine4.Close()
		}
	}
}
