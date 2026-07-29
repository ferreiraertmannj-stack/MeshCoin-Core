package script

import (
	"context"
	"encoding/hex"
	"time"
)

type ScriptEngine struct {
	policy   ScriptPolicy
	events   ScriptEvents
	stats    *ScriptStatistics
	parser   *ScriptParser
	executor *ScriptExecutor
	registry *OpcodeRegistry
	cache    *ScriptCache
	queue    *ScriptQueue
}

func NewScriptEngine(
	policy ScriptPolicy,
	events ScriptEvents,
	hashEng HashEngine,
	sigEng SignatureEngine,
	pubKeyEng PublicKeyEngine,
) *ScriptEngine {
	registry := NewOpcodeRegistry()
	parser := NewScriptParser(policy)
	executor := NewScriptExecutor(policy, registry, hashEng, sigEng, pubKeyEng)
	cache := NewScriptCache(policy.CacheTTL)
	stats := &ScriptStatistics{}

	engine := &ScriptEngine{
		policy:   policy,
		events:   events,
		stats:    stats,
		parser:   parser,
		executor: executor,
		registry: registry,
		cache:    cache,
	}

	queue := NewScriptQueue(policy.QueueSize, policy.MaxWorkers, engine)
	engine.queue = queue

	return engine
}

func (e *ScriptEngine) Start(ctx context.Context) {
	e.queue.Start(ctx)
}

func (e *ScriptEngine) Stop() {
	e.queue.Stop()
}

// RegisterOpcode allows extending the script engine dynamically
func (e *ScriptEngine) RegisterOpcode(op Opcode, handler OpcodeHandler) {
	e.registry.Register(op, handler)
}

func (e *ScriptEngine) ParseScript(raw []byte) ([]Instruction, error) {
	return e.parser.Parse(raw)
}

// Execute performs the full validation of unlocking + locking script.
// It combines scriptSig + scriptPubKey sequentially (or executes separately and transfers stack in real impls).
// For simplicity, we concatenate scriptSig and scriptPubKey as P2PKH usually does.
func (e *ScriptEngine) Execute(scriptSig, scriptPubKey, txHash []byte) error {
	// First check cache
	if e.cache.Get(scriptSig, scriptPubKey, txHash) {
		e.stats.IncCacheHits()
		return nil
	}
	e.stats.IncCacheMisses()

	job := &ScriptJob{
		ScriptSig:    scriptSig,
		ScriptPubKey: scriptPubKey,
		TxHash:       txHash,
		ResultChan:   make(chan error, 1),
	}

	err := e.queue.Enqueue(job)
	if err != nil {
		e.stats.IncQueueOverflows()
		return err
	}

	// Wait for async validation (or timeout)
	return <-job.ResultChan
}

func (e *ScriptEngine) executeInternal(scriptSig, scriptPubKey, txHash []byte) error {
	start := time.Now()

	// Simplest execution model: concatenate scripts
	combined := append([]byte{}, scriptSig...)
	combined = append(combined, scriptPubKey...)

	instructions, err := e.parser.Parse(combined)
	if err != nil {
		e.stats.IncScriptsFailed()
		return err
	}

	err = e.executor.Execute(instructions, txHash)

	e.stats.RecordExecutionTime(time.Since(start))

	txHashStr := hex.EncodeToString(txHash)

	if err != nil {
		e.stats.IncScriptsFailed()
		if e.events.OnScriptFailed != nil {
			go e.events.OnScriptFailed(txHashStr, err.Error())
		}
		return err
	}

	e.stats.IncScriptsExecuted()
	if e.events.OnScriptExecuted != nil {
		go e.events.OnScriptExecuted(txHashStr)
	}

	// Cache successful executions
	e.cache.Set(scriptSig, scriptPubKey, txHash)

	return nil
}

func (e *ScriptEngine) GetStatistics() ScriptStatistics {
	return e.stats.Snapshot()
}
