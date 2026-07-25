package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
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
	MinerStorage int           `json:"minerStorageGB"` // Nebula Cloud Storage Pledge
	StorageType  string        `json:"storageType"`    // "SSD" ou "HDD"
	Transactions []Transaction `json:"transactions"`
}

type Ledger struct {
	Chain []Block `json:"chain"`
	mu    sync.RWMutex
}

var ledger = Ledger{
	Chain: []Block{},
}

const ledgerFile = "ledger.json"

// Load or create genesis block
func initLedger() {
	if _, err := os.Stat(ledgerFile); os.IsNotExist(err) {
		log.Println("Criando novo Ledger Mestre com Bloco Gênesis...")
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
		ledger.Chain = append(ledger.Chain, genesis)
		saveLedger()
	} else {
		log.Println("Carregando Ledger Mestre existente...")
		file, _ := ioutil.ReadFile(ledgerFile)
		json.Unmarshal(file, &ledger.Chain)
		log.Printf("Ledger carregado com %d blocos.\n", len(ledger.Chain))
	}
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
		log.Printf("Bloco %d rejeitado: PreviousHash inválido\n", block.Index)
		return false
	}

	// Verifica o Hash do bloco recebido
	computed := calculateHash(block)
	if computed != block.Hash {
		log.Printf("Bloco %d rejeitado: Hash calculado %s != %s\n", block.Index, computed, block.Hash)
		return false
	}

	ledger.Chain = append(ledger.Chain, block)
	log.Printf("✅ Bloco #%d validado e adicionado ao Ledger Mestre!\n", block.Index)

	// A cada 10 blocos, envia para a Nebula Cloud
	if block.Index%10 == 0 {
		go uploadToNebulaCloud()
	}

	// Fora do Lock, salva no disco
	go saveLedger()
	return true
}

func getLedgerJSON() []byte {
	ledger.mu.RLock()
	defer ledger.mu.RUnlock()
	data, _ := json.Marshal(ledger.Chain)
	return data
}
