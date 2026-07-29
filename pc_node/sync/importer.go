package sync

import (
	"fmt"
	"sync"
	"time"

	"pc_node/storage"
)

// ImportStatistics tracks metrics for the BlockImporter.
type ImportStatistics struct {
	ImportedBlocks uint64
	ImportedChunks uint64
	ImportedBytes  uint64
	ImportSpeed    float64
	StartTime      time.Time
	ElapsedTime    time.Duration
}

// BlockImporterEventHandlers holds callbacks for the import pipeline.
type BlockImporterEventHandlers struct {
	OnChunkImported   func(chunk DownloadedChunk)
	OnImportCompleted func(stats ImportStatistics)
	OnImportError     func(err error)
}

// BlockImporter handles persisting downloaded chunks to the storage engine.
type BlockImporter struct {
	mu     sync.RWMutex
	engine storage.Engine
	events BlockImporterEventHandlers

	stats ImportStatistics
}

// NewBlockImporter initializes a new importer with the given storage engine.
func NewBlockImporter(engine storage.Engine, events BlockImporterEventHandlers) *BlockImporter {
	return &BlockImporter{
		engine: engine,
		events: events,
		stats: ImportStatistics{
			StartTime: time.Now(),
		},
	}
}

// ImportChunk imports a single chunk atomically using storage.Batch.
func (im *BlockImporter) ImportChunk(chunk DownloadedChunk) error {
	batch := im.engine.NewBatch()
	defer batch.Discard()

	var bytesInChunk uint64
	currentHeight := chunk.StartHeight

	for _, blockData := range chunk.Blocks {
		if err := batch.PutBlock(currentHeight, blockData); err != nil {
			if im.events.OnImportError != nil {
				im.events.OnImportError(fmt.Errorf("failed to put block at height %d: %w", currentHeight, err))
			}
			return err
		}
		bytesInChunk += uint64(len(blockData))
		currentHeight++
	}

	if err := batch.Commit(); err != nil {
		if im.events.OnImportError != nil {
			im.events.OnImportError(fmt.Errorf("failed to commit batch: %w", err))
		}
		return err
	}

	// Update stats
	im.mu.Lock()
	im.stats.ImportedBlocks += uint64(len(chunk.Blocks))
	im.stats.ImportedChunks++
	im.stats.ImportedBytes += bytesInChunk

	im.stats.ElapsedTime = time.Since(im.stats.StartTime)
	if im.stats.ElapsedTime.Seconds() > 0 {
		im.stats.ImportSpeed = float64(im.stats.ImportedBlocks) / im.stats.ElapsedTime.Seconds()
	}
	im.mu.Unlock()

	if im.events.OnChunkImported != nil {
		im.events.OnChunkImported(chunk)
	}

	return nil
}

// ImportChunks imports a slice of chunks sequentially.
func (im *BlockImporter) ImportChunks(chunks []DownloadedChunk) error {
	for _, chunk := range chunks {
		if err := im.ImportChunk(chunk); err != nil {
			return err
		}
	}

	im.mu.RLock()
	finalStats := im.stats
	im.mu.RUnlock()

	if im.events.OnImportCompleted != nil {
		im.events.OnImportCompleted(finalStats)
	}

	return nil
}

// ImportedBlocks returns the total number of successfully imported blocks.
func (im *BlockImporter) ImportedBlocks() uint64 {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.stats.ImportedBlocks
}

// ResetStatistics resets all metrics.
func (im *BlockImporter) ResetStatistics() {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.stats = ImportStatistics{
		StartTime: time.Now(),
	}
}

// Statistics returns a copy of the current statistics
func (im *BlockImporter) Statistics() ImportStatistics {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.stats
}
