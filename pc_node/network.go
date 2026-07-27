package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	udpPort = 5555
	tcpPort = 5556
	magic   = "NEBULA_NODE"
)

var (
	activeTCPClients = make(map[net.Conn]bool)
	tcpMutex         sync.Mutex
)

func startNetwork() {
	go listenUDP()
	go broadcastPresence()
	go listenTCP()
}

func listenUDP() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", udpPort))
	if err != nil {
		log.Fatalf("Erro ao resolver UDP addr: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Erro ao escutar UDP: %v", err)
	}
	defer conn.Close()

	log.Printf("📡 Ouvindo Broadcasts UDP na porta %d...\n", udpPort)

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		msg := string(buffer[:n])
		if strings.HasPrefix(msg, magic) {
			// Um nó celular se anunciou!
			// Podemos mapear seu IP se precisarmos enviar algo pra ele
			_ = remoteAddr
		}
	}
}

func broadcastPresence() {
	// Cria uma conexão UDP para enviar broadcasts
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", udpPort))
	if err != nil {
		log.Println("Erro ao resolver endereço de broadcast UDP:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Println("Erro ao iniciar broadcast UDP:", err)
		return
	}
	defer conn.Close()

	msg := []byte(fmt.Sprintf("%s:%d", magic, tcpPort))

	// Envia o ping a cada 5 segundos
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		_, err := conn.Write(msg)
		if err != nil {
			// Em algumas redes, o 255.255.255.255 pode falhar dependendo da interface de rede principal
		}
	}
}

func listenTCP() {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", tcpPort))
	if err != nil {
		log.Fatalf("Erro ao iniciar TCP: %v", err)
	}
	defer listener.Close()

	log.Printf("🔌 Servidor TCP rodando na porta %d (Aguardando Blocos e Transações)...\n", tcpPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("❌ Erro ao aceitar conexão TCP:", err)
			continue
		}
		log.Printf("📥 Nova conexão TCP aceita de: %s\n", conn.RemoteAddr().String())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	// Track active TCP connection for Relay
	tcpMutex.Lock()
	activeTCPClients[conn] = true
	tcpMutex.Unlock()

	defer func() {
		tcpMutex.Lock()
		delete(activeTCPClients, conn)
		tcpMutex.Unlock()
		conn.Close()
	}()

	decoder := json.NewDecoder(conn)
	for {
		var packet map[string]interface{}
		err := decoder.Decode(&packet)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("❌ Erro leitura TCP %s: %v\n", conn.RemoteAddr().String(), err)
			}
			return // Break loop cleanly on EOF or disconnect
		}

		tipo, ok := packet["tipo"].(string)
		if !ok {
			continue
		}

		switch tipo {
		case "NEW_BLOCK":
			log.Println("⛏️ Recebido NEW_BLOCK! Validando PoW...")
			handleNewBlockPacket(packet, conn)
		case "NEW_TRANSACTION":
			log.Println("💸 Recebido NEW_TRANSACTION! Processando para Mempool...")
			handleNewTransactionPacket(packet, conn)
		case "DATA_ROUTE", "CHAT":
			log.Println("🔄 Repassando pacote de CHAT/DATA P2P...")
			broadcastTCP(packet, conn)
		case "OGM", "PING":
			// Keep-alive silencioso
		default:
			log.Printf("⚠️ TCP: Tipo desconhecido %s\n", tipo)
		}
	}
}

func broadcastTCP(packet map[string]interface{}, sender net.Conn) {
	data, err := json.Marshal(packet)
	if err != nil {
		return
	}

	// Envia para os clientes WebSocket também!
	broadcast <- data

	tcpMutex.Lock()
	clientsCopy := make([]net.Conn, 0, len(activeTCPClients))
	for client := range activeTCPClients {
		if client != sender {
			clientsCopy = append(clientsCopy, client)
		}
	}
	tcpMutex.Unlock()

	for _, client := range clientsCopy {
		_, err := client.Write(data)
		if err != nil {
			client.Close()
			tcpMutex.Lock()
			delete(activeTCPClients, client)
			tcpMutex.Unlock()
		}
	}
}

func handleNewBlockPacket(packet map[string]interface{}, conn net.Conn) {
	blockData, ok := packet["block"]
	if !ok {
		log.Println("❌ Falha na validação: pacote NEW_BLOCK não contém a chave 'block'")
		if conn != nil {
			conn.Write([]byte(`{"status": "error", "message": "Missing block data"}`))
		}
		return
	}

	bytesData, _ := json.Marshal(blockData)
	var block Block
	if err := json.Unmarshal(bytesData, &block); err != nil {
		log.Println("❌ Erro ao converter bloco do JSON:", err)
		if conn != nil {
			conn.Write([]byte(`{"status": "error", "message": "Invalid JSON format"}`))
		}
		return
	}

	success := handleNewBlock(block)
	if success {
		log.Println("✅ NEW_BLOCK processado com sucesso via TCP/WS.")
		if conn != nil {
			conn.Write([]byte(`{"status": "success", "message": "Block appended to ledger"}`))
		}
	} else {
		log.Println("❌ Bloco rejeitado pelo Ledger Mestre.")
		if conn != nil {
			conn.Write([]byte(`{"status": "error", "message": "Block rejected by validation"}`))
		}
	}
}

func handleNewTransactionPacket(packet map[string]interface{}, conn net.Conn) {
	txData, ok := packet["transaction"]
	if !ok {
		log.Println("❌ Pacote NEW_TRANSACTION inválido: sem campo 'transaction'")
		return
	}

	bytesData, err := json.Marshal(txData)
	if err != nil {
		log.Println("❌ Erro ao converter tx para JSON interno:", err)
		return
	}

	var tx Transaction
	if err := json.Unmarshal(bytesData, &tx); err == nil {
		added := HandleNewTransaction(tx)
		if added {
			broadcastTCP(packet, conn)
			if conn != nil {
				json.NewEncoder(conn).Encode(map[string]interface{}{"status": "success", "msg": "Transaction appended to Mempool"})
			}
		} else {
			if conn != nil {
				json.NewEncoder(conn).Encode(map[string]interface{}{"status": "error", "msg": "Transaction rejected"})
			}
		}
	} else {
		log.Println("❌ Falha ao dar unmarshal na transação recebida:", err)
	}
}
