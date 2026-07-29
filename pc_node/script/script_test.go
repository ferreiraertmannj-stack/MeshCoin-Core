package script

import (
	"context"
	"sync"
	"testing"
)

func buildP2PKH(sig, pubkey []byte) ([]byte, []byte) {
	// Simple scriptSig: <sig> <pubkey>
	scriptSig := append([]byte{byte(len(sig))}, sig...)
	scriptSig = append(scriptSig, byte(len(pubkey)))
	scriptSig = append(scriptSig, pubkey...)

	// Simple scriptPubKey: OP_DUP OP_HASH256 <pubKeyHash> OP_EQUALVERIFY OP_CHECKSIG
	hashEng := NewDefaultHashEngine()
	pubKeyHash := hashEng.Hash256(pubkey)

	scriptPubKey := []byte{byte(OP_DUP), byte(OP_HASH256), byte(len(pubKeyHash))}
	scriptPubKey = append(scriptPubKey, pubKeyHash...)
	scriptPubKey = append(scriptPubKey, byte(OP_EQUALVERIFY), byte(OP_CHECKSIG))

	return scriptSig, scriptPubKey
}

func TestScriptEngine_P2PKH_Valid(t *testing.T) {
	policy := DefaultScriptPolicy()
	engine := NewScriptEngine(
		policy,
		ScriptEvents{},
		NewDefaultHashEngine(),
		NewDefaultSignatureEngine(),
		NewDefaultPublicKeyEngine(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	pubkey := []byte("my_public_key")
	sig := []byte("my_public_key") // Mock signature matches pubkey in DefaultSignatureEngine

	sigScript, pubKeyScript := buildP2PKH(sig, pubkey)
	txHash := []byte("tx_hash")

	err := engine.Execute(sigScript, pubKeyScript, txHash)
	if err != nil {
		t.Fatalf("Expected valid script execution, got: %v", err)
	}
}

func TestScriptEngine_P2PKH_InvalidSignature(t *testing.T) {
	policy := DefaultScriptPolicy()
	engine := NewScriptEngine(
		policy,
		ScriptEvents{},
		NewDefaultHashEngine(),
		NewDefaultSignatureEngine(),
		NewDefaultPublicKeyEngine(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	pubkey := []byte("my_public_key")
	sig := []byte("invalid_sig")

	sigScript, pubKeyScript := buildP2PKH(sig, pubkey)
	txHash := []byte("tx_hash")

	err := engine.Execute(sigScript, pubKeyScript, txHash)
	if err == nil {
		t.Fatalf("Expected script execution to fail due to invalid signature")
	}
}

func TestScriptEngine_Cache(t *testing.T) {
	policy := DefaultScriptPolicy()
	engine := NewScriptEngine(
		policy,
		ScriptEvents{},
		NewDefaultHashEngine(),
		NewDefaultSignatureEngine(),
		NewDefaultPublicKeyEngine(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	pubkey := []byte("my_public_key")
	sig := []byte("my_public_key")

	sigScript, pubKeyScript := buildP2PKH(sig, pubkey)
	txHash := []byte("tx_hash")

	engine.Execute(sigScript, pubKeyScript, txHash)

	stats := engine.GetStatistics()
	if stats.CacheMisses != 1 {
		t.Fatalf("Expected 1 cache miss, got %d", stats.CacheMisses)
	}

	// Execute again, should hit cache
	engine.Execute(sigScript, pubKeyScript, txHash)

	stats = engine.GetStatistics()
	if stats.CacheHits != 1 {
		t.Fatalf("Expected 1 cache hit, got %d", stats.CacheHits)
	}
}

func TestScriptEngine_MalformedScript(t *testing.T) {
	policy := DefaultScriptPolicy()
	engine := NewScriptEngine(
		policy,
		ScriptEvents{},
		NewDefaultHashEngine(),
		NewDefaultSignatureEngine(),
		NewDefaultPublicKeyEngine(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	// Push 100 bytes, but array is empty
	script := []byte{100}

	err := engine.Execute(script, []byte{}, []byte("tx_hash"))
	if err == nil {
		t.Fatalf("Expected malformed script error")
	}
}

func TestScriptEngine_Concurrency(t *testing.T) {
	policy := DefaultScriptPolicy()
	policy.MaxWorkers = 10
	engine := NewScriptEngine(
		policy,
		ScriptEvents{},
		NewDefaultHashEngine(),
		NewDefaultSignatureEngine(),
		NewDefaultPublicKeyEngine(),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	engine.Start(ctx)
	defer engine.Stop()

	routines := 1000
	var wg sync.WaitGroup

	pubkey := []byte("my_public_key")
	sig := []byte("my_public_key")
	sigScript, pubKeyScript := buildP2PKH(sig, pubkey)

	for i := 0; i < routines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Unique tx hash to avoid cache hit masking execution
			txHash := []byte{byte(id), byte(id >> 8)}
			_ = engine.Execute(sigScript, pubKeyScript, txHash)
		}(i)
	}

	wg.Wait()

	stats := engine.GetStatistics()
	if stats.ScriptsExecuted != uint64(routines) {
		t.Fatalf("Expected %d scripts executed, got %d", routines, stats.ScriptsExecuted)
	}
}
