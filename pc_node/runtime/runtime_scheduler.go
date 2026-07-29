package runtime

import (
	"context"
	"sync"
	"time"
)

type ScheduledTask struct {
	ID       string
	Interval time.Duration
	Action   func(ctx context.Context)
	cancel   context.CancelFunc
}

type RuntimeScheduler struct {
	tasks map[string]*ScheduledTask
	mu    sync.RWMutex
	wg    sync.WaitGroup
}

func NewRuntimeScheduler() *RuntimeScheduler {
	return &RuntimeScheduler{
		tasks: make(map[string]*ScheduledTask),
	}
}

func (s *RuntimeScheduler) Schedule(id string, interval time.Duration, action func(ctx context.Context)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[id]; exists {
		task.cancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	task := &ScheduledTask{
		ID:       id,
		Interval: interval,
		Action:   action,
		cancel:   cancel,
	}

	s.tasks[id] = task
	s.wg.Add(1)

	go func(t *ScheduledTask, c context.Context) {
		defer s.wg.Done()
		ticker := time.NewTicker(t.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-c.Done():
				return
			case <-ticker.C:
				t.Action(c)
			}
		}
	}(task, ctx)
}

func (s *RuntimeScheduler) Cancel(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task, exists := s.tasks[id]; exists {
		task.cancel()
		delete(s.tasks, id)
	}
}

func (s *RuntimeScheduler) StopAll() {
	s.mu.Lock()
	for id, task := range s.tasks {
		task.cancel()
		delete(s.tasks, id)
	}
	s.mu.Unlock()
	s.wg.Wait()
}
