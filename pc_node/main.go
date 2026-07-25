package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/gorilla/websocket"
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
	
	// Exemplo: 100 GB Total (50GB Ledger / 50GB Cloud)
	log.Println("Política de Armazenamento: 100GB Alocados (50% Ledger / 50% Nebula Cloud)")
	err := AllocateNebulaCloudStorage(50)
	if err != nil {
		log.Println("Aviso na alocação Nebula Cloud:", err)
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

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan []byte)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func startSidecarAPI() {
	http.HandleFunc("/api/ledger", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(getLedgerJSON())
	})
	
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		status := fmt.Sprintf(`{"status": "online", "storage_type": "%s", "cloud_allocated_gb": 100, "blocks": %d}`, DetectStorageType(), len(ledger.Chain))
		w.Write([]byte(status))
	})

	// P2P Internet Gateway - Nebula WebSocket Chat Relay
	http.HandleFunc("/ws", handleWebSocket)

	go handleMessages()

	log.Println("📡 Sidecar API & WebSocket rodando na porta 8080 (Aguardando Frontend Flutter)")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		log.Fatal("Falha ao iniciar API Sidecar:", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Erro no WebSocket Upgrade:", err)
		return
	}
	defer ws.Close()

	clients[ws] = true
	log.Println("Novo cliente WebSocket conectado!")

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Erro de leitura WS, desconectando cliente:", err)
			delete(clients, ws)
			break
		}
		// Recebeu mensagem PQC, vamos rotear para todos online
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Erro ao enviar via WS:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
