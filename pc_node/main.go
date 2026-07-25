package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 NEBULA NETWORK FULL NODE (PC) INICIADO 🚀")
	fmt.Println("==================================================")
	fmt.Println("Este nó escutará os celulares da rede P2P e salvará")
	fmt.Println("os blocos validados no Ledger Mestre (ledger.json).")
	fmt.Println("O Ledger será feito o backup na Nebula Cloud (porta 8000).")
	fmt.Println("==================================================")

	// Inicializa o Ledger (Lê do disco ou cria Genesis)
	initLedger()
	
	// Aloca o armazenamento Nebula Cloud e detecta Hardware
	storageType := DetectStorageType()
	log.Printf("Hardware Detectado: %s\n", storageType)
	
	// Exemplo: 100 GB para Desktop Cloud
	err := AllocateNebulaCloudStorage(100)
	if err != nil {
		log.Println("Aviso na alocação Nebula:", err)
	}

	// Inicia os listeners de rede UDP/TCP (Mesh P2P)
	startNetwork()

	// Inicia a API REST (Sidecar) para o Frontend Flutter
	go startSidecarAPI()

	// Mantém o daemon rodando até receber SIGINT (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("\nEncerrando Nebula Full Node. Salvando ledger final...")
	saveLedger()
	log.Println("Adeus.")
}

func startSidecarAPI() {
	http.HandleFunc("/api/ledger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(getLedgerJSON())
	})
	
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := fmt.Sprintf(`{"status": "online", "storage_type": "%s", "cloud_allocated_gb": 100, "blocks": %d}`, DetectStorageType(), len(ledger.Chain))
		w.Write([]byte(status))
	})

	log.Println("📡 Sidecar API rodando na porta 8080 (Aguardando Frontend Flutter)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Falha ao iniciar API Sidecar:", err)
	}
}
