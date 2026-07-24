package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	udpPort = 5555
	tcpPort = 5556
	magic   = "MESHCOIN_NODE"
)

func startNetwork() {
	go listenUDP()
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

func listenTCP() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", tcpPort))
	if err != nil {
		log.Fatalf("Erro ao iniciar TCP: %v", err)
	}
	defer listener.Close()

	log.Printf("🔌 Servidor TCP rodando na porta %d (Aguardando Blocos e Transações)...\n", tcpPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Erro ao aceitar TCP:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	var packet map[string]interface{}

	if err := decoder.Decode(&packet); err != nil {
		return
	}

	// Tratamento do pacote vindo da Rede Mesh
	tipo, ok := packet["tipo"].(string)
	if !ok {
		// Tenta verificar se é um pacote B.A.T.M.A.N aninhado
		if rTipo, ok := packet["tipo"].(string); ok && rTipo == "DATA_ROUTE" {
			payload, ok := packet["payload"].(map[string]interface{})
			if ok {
				tipo = payload["tipo"].(string)
				packet = payload // extrai o payload interno
			}
		}
	}

	switch tipo {
	case "NEW_BLOCK":
		handleNewBlockPacket(packet)
	case "NEW_TRANSACTION":
		log.Println("Nova transação recebida na Mempool Global!")
		// O Nó Completo guarda na sua própria mempool ou só loga
	case "CHAT":
		// Ignora mensagens de chat
	default:
		// Pacote desconhecido
	}
}

func handleNewBlockPacket(packet map[string]interface{}) {
	blockData, ok := packet["block"]
	if !ok {
		return
	}

	bytes, _ := json.Marshal(blockData)
	var block Block
	if err := json.Unmarshal(bytes, &block); err != nil {
		log.Println("Erro ao converter bloco:", err)
		return
	}

	handleNewBlock(block)
}
