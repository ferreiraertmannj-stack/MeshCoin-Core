package sync

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"pc_node/storage/mockstorage"
)

func createValidTestChunk(start uint64, count int) DownloadedChunk {
	blocks := make([][]byte, count)
	for i := 0; i < count; i++ {
		b := blockTemplate{
			Index:        int(start) + i,
			Timestamp:    1625097600,
			PreviousHash: "prev",
			Hash:         "valid",
			Nonce:        123,
			Transactions: []transactionTemplate{
				{ID: "tx1", Amount: 10},
			},
		}
		if b.Index == 0 {
			b.PreviousHash = ""
		}
		blocks[i], _ = json.Marshal(b)
	}
	return DownloadedChunk{
		StartHeight:  start,
		EndHeight:    start + uint64(count) - 1,
		Blocks:       blocks,
		PeerID:       "peer_1",
		DownloadTime: 10 * time.Millisecond,
	}
}

func TestSyncController_Pipeline(t *testing.T) {
	pool := NewPeerPool()
	p := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p1", nil, 5000), sleepTime: 1 * time.Millisecond}
	pool.AddPeer(p)

	manager := NewSyncManager(pool)
	queue := NewDownloadQueue(3)
	downloader := NewDownloader(queue, pool, 2, 50*time.Millisecond)

	// Create Validator
	validator := NewBlockValidator(BlockValidatorEventHandlers{})

	// Create Importer
	engine := mockstorage.NewMockEngine()
	engine.Open("")
	defer engine.Close()
	importer := NewBlockImporter(engine, BlockImporterEventHandlers{})

	var stateMu sync.Mutex
	stateChanges := []SyncState{}

	events := SyncControllerEventHandlers{
		OnStateChanged: func(oldState, newState SyncState) {
			stateMu.Lock()
			stateChanges = append(stateChanges, newState)
			stateMu.Unlock()
		},
		OnValidationError: func(err error, chunk DownloadedChunk) {
			t.Logf("Validation Error: %v", err)
		},
		OnChunkValidated: func(chunk DownloadedChunk) {
			t.Logf("Chunk Validated! blocks: %d", len(chunk.Blocks))
		},
		OnChunkImported: func(chunk DownloadedChunk) {
			t.Logf("Chunk Imported! blocks: %d", len(chunk.Blocks))
		},
		OnPipelineError: func(err error) {
			t.Logf("Pipeline Error: %v", err)
		},
	}

	controller := NewSyncController(manager, downloader, validator, importer, events)

	// Injetar blocos diretamente para o mock downloader retornar (ou sobrescrever enqueue localmente)
	// Como a lógica do mock downloader foi simplificada em runStateMachine,
	// precisamos garantir que chunks válidos cheguem na fila de downloded.

	err := controller.Start(100)
	if err != nil {
		t.Fatalf("Failed to start controller: %v", err)
	}

	// Como o downloader original mockado nos testes dependia do worker falso,
	// vamos injetar chunks baixados diretamente na fila dele
	go func() {
		time.Sleep(20 * time.Millisecond)
		downloader.mu.Lock()
		downloader.downloadedChunks = append(downloader.downloadedChunks, createValidTestChunk(1, 100))
		downloader.mu.Unlock()

		time.Sleep(20 * time.Millisecond)

		downloader.queue.MarkCompleted(DownloadChunk{StartHeight: 1, EndHeight: 100})

		downloader.mu.Lock()
		downloader.queue.pendingChunks = []DownloadChunk{} // Esvazia pendentes
		downloader.mu.Unlock()
	}()

	time.Sleep(300 * time.Millisecond)

	status := controller.Status()

	if status.ImportedBlocks != 100 {
		t.Errorf("Expected 100 imported blocks, got %d", status.ImportedBlocks)
	}

	if status.ValidatedBlocks != 100 {
		t.Errorf("Expected 100 validated blocks, got %d", status.ValidatedBlocks)
	}

	if status.CurrentState != StateCompleted {
		t.Errorf("Expected StateCompleted, got %v", status.CurrentState)
	}
}

func TestSyncController_CancelAndPause(t *testing.T) {
	pool := NewPeerPool()
	p := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p1", nil, 5000), sleepTime: 10 * time.Millisecond}
	pool.AddPeer(p)

	manager := NewSyncManager(pool)
	queue := NewDownloadQueue(3)
	downloader := NewDownloader(queue, pool, 2, 50*time.Millisecond)
	validator := NewBlockValidator(BlockValidatorEventHandlers{})
	engine := mockstorage.NewMockEngine()
	importer := NewBlockImporter(engine, BlockImporterEventHandlers{})

	cancelled := false
	events := SyncControllerEventHandlers{
		OnCancelled: func() { cancelled = true },
	}

	controller := NewSyncController(manager, downloader, validator, importer, events)
	controller.Start(5000)
	time.Sleep(10 * time.Millisecond)

	controller.Pause()
	status := controller.Status()
	if status.CurrentState == StateCompleted {
		t.Fatalf("Should not be completed")
	}

	controller.Resume()
	controller.Cancel()
	time.Sleep(10 * time.Millisecond)

	if !cancelled {
		t.Fatalf("OnCancelled event not fired")
	}
	if controller.Status().CurrentState != StateIdle {
		t.Fatalf("Expected StateIdle after cancel, got %v", controller.Status().CurrentState)
	}
}

func TestSyncController_ConcurrencyStress(t *testing.T) {
	pool := NewPeerPool()
	p := &mockFastTCPPeer{TCPPeer: NewTCPPeer("p1", nil, 5000), sleepTime: 1 * time.Millisecond}
	pool.AddPeer(p)

	manager := NewSyncManager(pool)
	queue := NewDownloadQueue(3)
	downloader := NewDownloader(queue, pool, 10, 50*time.Millisecond)
	validator := NewBlockValidator(BlockValidatorEventHandlers{})
	engine := mockstorage.NewMockEngine()
	importer := NewBlockImporter(engine, BlockImporterEventHandlers{})

	controller := NewSyncController(manager, downloader, validator, importer, SyncControllerEventHandlers{})
	controller.Start(10000)

	var wg sync.WaitGroup
	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			action := idx % 5
			switch action {
			case 0:
				controller.Status()
			case 1:
				controller.Pause()
				time.Sleep(1 * time.Millisecond)
				controller.Resume()
			case 2:
				_ = controller.Status().ETASeconds
			case 3:
				_ = controller.Status().SpeedBlocksSec
			case 4:
				if idx == 499 { // Only one cancel at the end
					time.Sleep(50 * time.Millisecond)
					controller.Cancel()
				}
			}
		}(i)
	}

	wg.Wait()
}
