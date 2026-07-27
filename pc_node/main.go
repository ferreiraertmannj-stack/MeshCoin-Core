package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func displayLocalIPs() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	fmt.Println("--------------------------------------------------")
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				// Ignora interfaces virtuais APIPA e Docker se possível foca nas principais
				if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") {
					fmt.Printf("✅ PC Node rodando no IP: %s\n", ip)
				}
			}
		}
	}
	fmt.Println("--------------------------------------------------")
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 NEBULA NETWORK FULL NODE (PC) INICIADO 🚀")
	fmt.Println("==================================================")
	displayLocalIPs()
	fmt.Println("Este nó escutará os celulares da rede P2P e salvará")
	fmt.Println("os blocos validados no Ledger Mestre (ledger.json).")
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

	log.Println("\nEncerrando Nebula Full Node...")
	log.Println("Adeus.")
}

var clients = make(map[*websocket.Conn]bool)
var wsMutex sync.Mutex
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

	wsMutex.Lock()
	clients[ws] = true
	wsMutex.Unlock()
	log.Println("Novo cliente WebSocket conectado!")

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Erro de leitura WS, desconectando cliente:", err)
			wsMutex.Lock()
			delete(clients, ws)
			wsMutex.Unlock()
			break
		}

		// Converte a mensagem para verificar se é um Bloco
		var packet map[string]interface{}
		if err := json.Unmarshal(msg, &packet); err == nil {
			// Verifica se está encapsulado
			tipo, ok := packet["tipo"].(string)
			if ok && tipo == "DATA_ROUTE" {
				if payload, pOk := packet["payload"].(map[string]interface{}); pOk {
					tipo = payload["tipo"].(string)
					packet = payload
				}
			}

			if ok {
				if tipo == "NEW_BLOCK" {
					handleNewBlockPacket(packet, nil)
				} else if tipo == "HANDSHAKE" {
					log.Println("🤝 Handshake recebido do aplicativo Flutter!")
				}
			}
		}

		// Roteia a mensagem para todos os outros nós conectados
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		wsMutex.Lock()
		clientsCopy := make([]*websocket.Conn, 0, len(clients))
		for client := range clients {
			clientsCopy = append(clientsCopy, client)
		}
		wsMutex.Unlock()

		for _, client := range clientsCopy {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Erro ao enviar via WS:", err)
				client.Close()
				wsMutex.Lock()
				delete(clients, client)
				wsMutex.Unlock()
			}
		}
	}
}

func getLedgerJSON() []byte {
	var buf bytes.Buffer
	buf.WriteString("[\n")
	it := DB.NewBlockIterator()
	defer it.Close()
	first := true
	for it.Next() {
		if !first {
			buf.WriteString(",\n")
		}
		buf.Write(it.Value())
		first = false
	}
	buf.WriteString("\n]")
	return buf.Bytes()
}
