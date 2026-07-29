package script

import "time"

type ScriptPolicy struct {
	MaxScriptSize  int
	MaxStackDepth  int
	MaxOpcodeCount int
	CacheTTL       time.Duration
	QueueSize      int
	MaxWorkers     int
}

func DefaultScriptPolicy() ScriptPolicy {
	return ScriptPolicy{
		MaxScriptSize:  10000,
		MaxStackDepth:  1000,
		MaxOpcodeCount: 201,
		CacheTTL:       60 * time.Second,
		QueueSize:      10000,
		MaxWorkers:     4,
	}
}
