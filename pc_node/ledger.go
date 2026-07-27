package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// BlockHeader and Transaction structures
type Transaction struct {
	ID              string  `json:"id"`
	SenderPubKey    string  `json:"senderPubKey"`
	SenderAddress   string  `json:"senderAddress"`
	ReceiverAddress string  `json:"receiverAddress"`
	Amount          float64 `json:"amount"`
	Fee             float64 `json:"fee"`
	Timestamp       int64   `json:"timestamp"`
	Signature       string  `json:"signature"`      // Assinatura ECDSA
	PQCSignature    string  `json:"pqcSignature"`   // [Futuro] Assinatura Pós-Quântica (Dilithium)
}

type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	PreviousHash string        `json:"previousHash"`
	MerkleRoot   string        `json:"merkleRoot"`
	Nonce        int           `json:"nonce"`
	Hash         string        `json:"hash"`
	MinerStorage int           `json:"minerStorage"` // Nebula Cloud Storage Pledge
	StorageType  string        `json:"storageType"`    // "SSD" ou "HDD"
	Transactions []Transaction `json:"transactions"`
}

type Ledger struct {
	Chain []Block `json:"chain"`
	mu    sync.RWMutex
}

var (
	ledger = Ledger{Chain: []Block{}}
	PendingTransactions []Transaction
	mempoolMutex sync.Mutex
)

const ledgerFile = "ledger.json"

// Load or create genesis block
func initLedger() {
	var file []byte
	var err error

	if file, err = ioutil.ReadFile(ledgerFile); err == nil {
		if jsonErr := json.Unmarshal(file, &ledger.Chain); jsonErr == nil && len(ledger.Chain) > 0 {
			log.Printf("Ledger Mestre existente carregado com %d blocos.\n", len(ledger.Chain))
			return
		} else {
			log.Println("⚠️ ledger.json corrompido ou vazio. Recriando Bloco Gênesis...")
		}
	} else {
		log.Println("Criando novo Ledger Mestre com Bloco Gênesis...")
	}

	genesis := Block{
		Index:        0,
		Timestamp:    1672531200000,
		PreviousHash: "0000000000000000000000000000000000000000000000000000000000000000",
		MerkleRoot:   "",
		Nonce:        0,
		Hash:         "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f",
		MinerStorage: 0,
		StorageType:  "SSD",
		Transactions: []Transaction{},
	}
	ledger.Chain = []Block{genesis}
	saveLedger()
}

func saveLedger() {
	ledger.mu.Lock()
	defer ledger.mu.Unlock()

	data, err := json.MarshalIndent(ledger.Chain, "", "  ")
	if err != nil {
		log.Println("Erro ao serializar ledger:", err)
		return
	}
	ioutil.WriteFile(ledgerFile, data, 0644)
}

func calculateHash(b Block) string {
	var record string
	if b.MinerStorage > 0 || (b.StorageType != "" && b.StorageType != "Mobile") {
		record = fmt.Sprintf("%d%d%s%s%d%d%s", b.Index, b.Timestamp, b.PreviousHash, b.MerkleRoot, b.Nonce, b.MinerStorage, b.StorageType)
	} else {
		record = fmt.Sprintf("%d%d%s%s%d", b.Index, b.Timestamp, b.PreviousHash, b.MerkleRoot, b.Nonce)
	}
	
	// NeonHash v1.0 (Vector Math / Memory Hard Simulation)
	h := sha256.New()
	h.Write([]byte(record))
	seedHash := h.Sum(nil)

	// Aloca um vetor de memória de 4KB (4096 bytes)
	memoryVector := make([]byte, 4096)
	for i := 0; i < 4096; i++ {
		memoryVector[i] = (seedHash[i%32] ^ byte(i&255)) & 255
	}

	// Mistura o vetor (operações pseudo-vetoriais matemáticas simples)
	state := uint32(seedHash[0])
	for i := 0; i < 128; i++ {
		idx := state % 4096
		val := uint32(memoryVector[idx])
		state = (state*31 + val) & 0xFFFFFFFF
		memoryVector[(idx+1)%4096] ^= byte(state & 255)
	}

	// Hash final de tudo
	hf := sha256.New()
	hf.Write(memoryVector)
	finalDigest := hf.Sum(nil)
	return hex.EncodeToString(finalDigest)
}

// CalculateBlockReward aplica a fórmula do Halving e do Bônus de Armazenamento
func CalculateBlockReward(blockIndex int, minerStorageGB int, storageType string) float64 {
	halvings := blockIndex / 2100000
	baseReward := 50.0

	// Halving
	for i := 0; i < halvings; i++ {
		baseReward /= 2
	}

	// Nebula Cloud Storage Bonus (Proof of Storage)
	bonus := 0.0
	if storageType == "SSD" {
		bonus = float64(minerStorageGB) * 0.5 // 100GB = 50 NBL
	} else {
		bonus = float64(minerStorageGB) * 0.1 // 100GB = 10 NBL
	}

	return baseReward + bonus
}

func VerifyNeonHash(block Block) bool {
	computed := calculateHash(block)
	return computed == block.Hash
}

func handleNewBlock(block Block) bool {
	ledger.mu.Lock()
	defer ledger.mu.Unlock()

	if len(ledger.Chain) == 0 {
		return false
	}

	lastBlock := ledger.Chain[len(ledger.Chain)-1]

	// Verifica se já temos o bloco
	if block.Index <= lastBlock.Index {
		return false
	}

	// Valida se o bloco bate com nossa chain
	if block.PreviousHash != lastBlock.Hash {
		log.Printf("❌ Bloco %d rejeitado: PreviousHash inválido\n", block.Index)
		return false
	}

	// Verifica o Hash do bloco recebido usando a validação NeonHash
	if !VerifyNeonHash(block) {
		log.Printf("❌ Bloco %d rejeitado: PoW NeonHash Inválido\n", block.Index)
		return false
	}

	// Valida cada transação dentro do bloco
	for _, tx := range block.Transactions {
		if tx.SenderAddress != "COINBASE" {
			if !VerifyTransaction(tx) {
				log.Printf("❌ Bloco %d rejeitado: Transação %s inválida\n", block.Index, tx.ID)
				return false
			}
		}
	}

	ledger.Chain = append(ledger.Chain, block)
	log.Printf("✅ Bloco #%d validado e adicionado ao Ledger Mestre!\n", block.Index)

	// Remove as transações mineradas da Mempool Global (PendingTransactions)
	mempoolMutex.Lock()
	minedTxIDs := make(map[string]bool)
	for _, tx := range block.Transactions {
		minedTxIDs[tx.ID] = true
	}
	newPending := make([]Transaction, 0)
	for _, pTx := range PendingTransactions {
		if !minedTxIDs[pTx.ID] {
			newPending = append(newPending, pTx)
		}
	}
	PendingTransactions = newPending
	mempoolMutex.Unlock()

	// A cada 10 blocos, envia para a Nebula Cloud
	if block.Index%10 == 0 {
		go uploadToNebulaCloud()
	}

	// Fora do Lock, salva no disco
	go saveLedger()
	return true
}

func getLedgerJSON() []byte {
	// Força a leitura direta do arquivo em disco
	data, err := ioutil.ReadFile(ledgerFile)
	if err != nil {
		// Fallback para memória se o arquivo não estiver disponível por algum motivo
		ledger.mu.RLock()
		defer ledger.mu.RUnlock()
		mData, _ := json.Marshal(ledger.Chain)
		return mData
	}
	return data
}

func formatDartDouble(v float64) string {
	s := fmt.Sprintf("%v", v)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

func VerifyTransaction(tx Transaction) bool {
	if tx.SenderAddress == "COINBASE" {
		return true
	}

	// 1. Rebuild txData exactly as Dart does
	txData := fmt.Sprintf("%s:%s:%s:%s:%s:%d",
		tx.SenderPubKey,
		tx.SenderAddress,
		tx.ReceiverAddress,
		formatDartDouble(tx.Amount),
		formatDartDouble(tx.Fee),
		tx.Timestamp,
	)

	// 2. Verify hash
	hashBytes := sha256.Sum256([]byte(txData))
	computedHash := hex.EncodeToString(hashBytes[:])
	if computedHash != tx.ID {
		log.Printf("❌ Tx %s rejeitada: ID/Hash inválido (Calculado: %s)\n", tx.ID, computedHash)
		return false
	}

	// 3. Verify Signature
	if len(tx.Signature) != 128 {
		return false
	}
	pubKeyBytes, err := hex.DecodeString(tx.SenderPubKey)
	if err != nil {
		return false
	}
	pubKey, err := secp256k1.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false
	}

	rBytes, _ := hex.DecodeString(tx.Signature[:64])
	sBytes, _ := hex.DecodeString(tx.Signature[64:])
	var r, s secp256k1.ModNScalar
	r.SetByteSlice(rBytes)
	s.SetByteSlice(sBytes)

	sig := ecdsa.NewSignature(&r, &s)
	if !sig.Verify(hashBytes[:], pubKey) {
		log.Printf("❌ Tx %s rejeitada: Assinatura inválida\n", tx.ID)
		return false
	}

	return true
}

func HandleNewTransaction(tx Transaction) bool {
	if !VerifyTransaction(tx) {
		return false
	}

	// Verify balance (Simple O(N) sweep of the ledger)
	ledger.mu.RLock()
	balance := 0.0
	for _, block := range ledger.Chain {
		for _, bTx := range block.Transactions {
			if bTx.ReceiverAddress == tx.SenderAddress {
				balance += bTx.Amount
			}
			if bTx.SenderAddress == tx.SenderAddress {
				balance -= (bTx.Amount + bTx.Fee)
			}
		}
	}
	ledger.mu.RUnlock()

	// Consider mempool pending balance
	mempoolMutex.Lock()
	defer mempoolMutex.Unlock()
	for _, pTx := range PendingTransactions {
		if pTx.SenderAddress == tx.SenderAddress {
			balance -= (pTx.Amount + pTx.Fee)
		}
	}

	if balance < (tx.Amount + tx.Fee) {
		log.Printf("❌ Tx %s rejeitada: Saldo insuficiente (%.2f)\n", tx.ID, balance)
		return false
	}

	PendingTransactions = append(PendingTransactions, tx)
	log.Printf("✅ Transação %s validada e adicionada à Mempool Global!\n", tx.ID)
	return true
}
