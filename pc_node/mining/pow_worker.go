package mining

import (
	"time"
)

type PoWWorker struct {
	id        int
	pipeline  HashPipeline
	validator *ShareValidator
	generator *NonceGenerator
	events    MiningEvents
	stats     *MiningStatistics
	network   NetworkProvider
}

func NewPoWWorker(
	id int,
	pipeline HashPipeline,
	validator *ShareValidator,
	generator *NonceGenerator,
	events MiningEvents,
	stats *MiningStatistics,
	network NetworkProvider,
) *PoWWorker {
	return &PoWWorker{
		id:        id,
		pipeline:  pipeline,
		validator: validator,
		generator: generator,
		events:    events,
		stats:     stats,
		network:   network,
	}
}

func (w *PoWWorker) Mine(job *MiningJob) {
	header := BlockHeader{
		Version:    job.Template.Version,
		PrevHash:   job.Template.PreviousHash,
		MerkleRoot: job.Template.MerkleRoot,
		Timestamp:  job.Template.Timestamp, // Starts with template timestamp
		Target:     job.Target,
	}

	hashesComputed := uint64(0)
	startTime := time.Now()

	for {
		select {
		case <-job.Ctx.Done():
			// Job cancelled
			w.stats.AddHashes(hashesComputed)
			w.updateHashrate(hashesComputed, time.Since(startTime))
			return
		default:
			// Mining loop
			nonce, extraNonce := w.generator.Generate()
			header.Nonce = nonce
			header.ExtraNonce = extraNonce

			// Periodically update timestamp
			if hashesComputed%10000 == 0 {
				header.Timestamp = w.network.GetNetworkTimestamp()
			}

			hashResult := header.Hash(w.pipeline)
			hashesComputed++

			if w.validator.Validate(hashResult, job.Target) {
				// Share/Block found!
				w.stats.AddHashes(hashesComputed)
				w.updateHashrate(hashesComputed, time.Since(startTime))
				w.stats.IncSharesFound()

				if w.events.OnShareFound != nil {
					go w.events.OnShareFound(hashResult, 0) // diff abstracted
				}

				// Reconstruct full block
				block := &Block{
					Header:       header,
					Transactions: append([]Transaction{job.Template.Coinbase}, job.Template.Transactions...),
				}

				w.stats.IncBlocksFound()
				if w.events.OnBlockFound != nil {
					go w.events.OnBlockFound(block)
				}

				// Stop the job since we found the block
				job.Cancel()
				return
			}
		}
	}
}

func (w *PoWWorker) updateHashrate(hashes uint64, elapsed time.Duration) {
	secs := elapsed.Seconds()
	if secs > 0 {
		hps := float64(hashes) / secs
		w.stats.UpdateHashRate(hps) // Note: this is a simple overwrite, in reality it should be moving average
	}
}
