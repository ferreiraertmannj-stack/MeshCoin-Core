package main

import (
	"flag"
	"log"
	"os"

	"pc_node/storage"
	"pc_node/storage/badgerstorage"
	"pc_node/storage/jsonstorage"
)

type Transaction struct {
	ID              string  `json:"id"`
	SenderPubKey    string  `json:"senderPubKey"`
	SenderAddress   string  `json:"senderAddress"`
	ReceiverAddress string  `json:"receiverAddress"`
	Amount          float64 `json:"amount"`
	Fee             float64 `json:"fee"`
	Timestamp       int64   `json:"timestamp"`
	Signature       string  `json:"signature"`
	PQCSignature    string  `json:"pqcSignature"`
}

type Block struct {
	Index        uint64        `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	PreviousHash string        `json:"previousHash"`
	MerkleRoot   string        `json:"merkleRoot"`
	Nonce        int           `json:"nonce"`
	Hash         string        `json:"hash"`
	MinerStorage int           `json:"minerStorage"`
	StorageType  string        `json:"storageType"`
	Transactions []Transaction `json:"transactions"`
}

func main() {
	inputPtr := flag.String("input", "ledger.json", "Caminho do ledger.json legado")
	outputPtr := flag.String("output", "badger_data", "Diretório do banco Badger de destino")
	forcePtr := flag.Bool("force", false, "Permite sobrescrever banco existente")
	verifyPtr := flag.Bool("verify", false, "Verifica a consistência após a migração")

	flag.Parse()

	log.Println("Iniciando migração de JSON para BadgerDB...")

	if _, err := os.Stat(*inputPtr); os.IsNotExist(err) {
		log.Fatalf("❌ Erro: ledger legado não encontrado em %s", *inputPtr)
	}

	if _, err := os.Stat(*outputPtr); err == nil {
		if !*forcePtr {
			log.Fatalf("❌ Erro: Diretório de destino (%s) já existe. Use --force para continuar.", *outputPtr)
		}
		log.Printf("⚠️ Diretório existente %s detectado. Abrindo com flag --force.", *outputPtr)
	}

	jsonEngine := jsonstorage.NewJSONEngine()
	if err := jsonEngine.Open(*inputPtr); err != nil {
		log.Fatalf("❌ Erro ao abrir JSONStorageAdapter: %v", err)
	}
	defer jsonEngine.Close()

	badgerEngine := badgerstorage.NewBadgerEngine()
	if err := badgerEngine.Open(*outputPtr); err != nil {
		log.Fatalf("❌ Erro ao abrir BadgerStorageAdapter: %v", err)
	}
	defer badgerEngine.Close()

	it := jsonEngine.NewBlockIterator()
	defer it.Close()

	var blockCount int
	var lastHash string
	var lastIndex uint64
	balances := make(map[string]float64)

	batch := badgerEngine.NewBatch()

	for it.Next() {
		val := it.Value()
		var b Block
		if err := storage.UnmarshalBlock(val, &b); err != nil {
			log.Fatalf("❌ Erro ao decodificar bloco: %v", err)
		}

		if err := batch.PutBlock(b.Index, val); err != nil {
			log.Fatalf("❌ Erro ao enfileirar bloco no batch: %v", err)
		}

		blockCount++
		lastHash = b.Hash
		lastIndex = b.Index

		for _, tx := range b.Transactions {
			if tx.ReceiverAddress != "" {
				balances[tx.ReceiverAddress] += tx.Amount
			}
			if tx.SenderAddress != "COINBASE" && tx.SenderAddress != "" {
				balances[tx.SenderAddress] -= (tx.Amount + tx.Fee)
			}
		}
	}

	if err := it.Error(); err != nil {
		log.Fatalf("❌ Erro durante a iteração do ledger JSON: %v", err)
	}

	for addr, bal := range balances {
		if err := batch.PutBalance(addr, bal); err != nil {
			log.Fatalf("❌ Erro ao enfileirar saldo no batch: %v", err)
		}
	}

	if err := batch.Commit(); err != nil {
		log.Fatalf("❌ Erro ao commitar os dados no BadgerDB: %v", err)
	}

	log.Printf("✅ Migração concluída: %d blocos processados (Altura: %d, Último Hash: %s)", blockCount, lastIndex, lastHash)

	if *verifyPtr {
		log.Println("🔍 Iniciando verificação de consistência...")

		bIt := badgerEngine.NewBlockIterator()
		defer bIt.Close()

		var bCount int
		var bLastHash string
		var bLastIndex uint64

		for bIt.Next() {
			val := bIt.Value()
			var b Block
			if err := storage.UnmarshalBlock(val, &b); err != nil {
				log.Fatalf("❌ Verificação falhou: erro ao decodificar bloco do BadgerDB: %v", err)
			}
			bCount++
			bLastHash = b.Hash
			bLastIndex = b.Index
		}

		if err := bIt.Error(); err != nil {
			log.Fatalf("❌ Verificação falhou: erro na iteração do BadgerDB: %v", err)
		}

		if bCount != blockCount {
			log.Fatalf("❌ DIVERGÊNCIA: Blocos migrados (%d) != Blocos no BadgerDB (%d)", blockCount, bCount)
		}
		if bLastIndex != lastIndex {
			log.Fatalf("❌ DIVERGÊNCIA: Altura migrada (%d) != Altura no BadgerDB (%d)", lastIndex, bLastIndex)
		}
		if bLastHash != lastHash {
			log.Fatalf("❌ DIVERGÊNCIA: Último hash migrado (%s) != Último hash no BadgerDB (%s)", lastHash, bLastHash)
		}

		log.Println("✅ Verificação bem sucedida: A base de dados migrada é 100% idêntica.")
	}
}
