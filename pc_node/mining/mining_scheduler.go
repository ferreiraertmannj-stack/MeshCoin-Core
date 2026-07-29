package mining

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type MiningScheduler struct {
	queue      *MiningQueue
	events     MiningEvents
	policy     MiningPolicy
	stats      *MiningStatistics
	cache      *MiningCache
	pipeline   HashPipeline
	difficulty MiningDifficultyProvider
	network    NetworkProvider

	currentJob *MiningJob
	mu         sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewMiningScheduler(
	queue *MiningQueue,
	events MiningEvents,
	policy MiningPolicy,
	stats *MiningStatistics,
	cache *MiningCache,
	pipeline HashPipeline,
	difficulty MiningDifficultyProvider,
	network NetworkProvider,
) *MiningScheduler {
	return &MiningScheduler{
		queue:      queue,
		events:     events,
		policy:     policy,
		stats:      stats,
		cache:      cache,
		pipeline:   pipeline,
		difficulty: difficulty,
		network:    network,
	}
}

func (s *MiningScheduler) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Launch workers in the queue
	validator := NewShareValidator()
	generator := NewNonceGenerator()

	s.queue.LaunchWorkers(s.pipeline, validator, generator, s.events, s.stats, s.network)

	s.wg.Add(1)
	go s.timer()
}

func (s *MiningScheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.cancelCurrentJob()
	s.wg.Wait()
}

func (s *MiningScheduler) ScheduleJob(tmpl *BlockTemplate) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Cancel existing job
	if s.currentJob != nil {
		// Ignore if the template is older or same height (unless requested explicitly)
		if tmpl.Height < s.currentJob.Template.Height {
			return
		}

		s.currentJob.Cancel()
		s.stats.IncJobsCancelled()
		if s.events.OnJobCancelled != nil {
			go s.events.OnJobCancelled(s.currentJob.ID, "New template arrived")
		}
	}

	// 2. Check Cache
	if cached, valid := s.cache.Get(); valid && cached.Template.PreviousHash == tmpl.PreviousHash {
		// Same block height and previous hash, but maybe transactions updated.
		// If we strictly want to cache by ID we could, but let's assume we reuse if valid.
		// For PoW, generating a new job is cheap, but let's respect the cache TTL.
		s.stats.IncCacheHits()
	} else {
		s.stats.IncCacheMisses()
	}

	// 3. Create new Job
	target := s.difficulty.GetCurrentTarget()
	job := NewMiningJob(tmpl, target, s.policy.JobTimeout)
	s.currentJob = job

	s.cache.Set(job)
	s.stats.IncJobsCreated()

	if s.events.OnJobCreated != nil {
		go s.events.OnJobCreated(job.ID, tmpl.Height)
	}

	// 4. Enqueue (feed to workers)
	err := s.queue.Enqueue(job)
	if err != nil {
		if s.events.OnMiningError != nil {
			go s.events.OnMiningError(fmt.Errorf("failed to enqueue job: %w", err))
		}
	}
}

func (s *MiningScheduler) cancelCurrentJob() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.currentJob != nil {
		s.currentJob.Cancel()
	}
}

func (s *MiningScheduler) timer() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.policy.RefreshPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Refresh check - if the job is too old, or diff changed
			s.mu.Lock()
			if s.currentJob != nil {
				currentTarget := s.difficulty.GetCurrentTarget()
				if currentTarget != s.currentJob.Target {
					// Difficulty changed! Must restart job.
					s.mu.Unlock()
					if s.events.OnDifficultyUpdated != nil {
						go s.events.OnDifficultyUpdated(currentTarget)
					}
					// Real implementation would request a new template. We will just cancel.
					s.cancelCurrentJob()
					continue
				}
			}
			s.mu.Unlock()
		}
	}
}
