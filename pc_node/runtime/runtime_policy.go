package runtime

import "time"

type RuntimePolicy struct {
	StartupTimeout       time.Duration
	ShutdownTimeout      time.Duration
	HealthInterval       time.Duration
	MaxRestartAttempts   int
	MaxConcurrentModules int
	EventQueueSize       int
}

func DefaultRuntimePolicy() RuntimePolicy {
	return RuntimePolicy{
		StartupTimeout:       30 * time.Second,
		ShutdownTimeout:      15 * time.Second,
		HealthInterval:       5 * time.Second,
		MaxRestartAttempts:   3,
		MaxConcurrentModules: 10,
		EventQueueSize:       10000,
	}
}
