package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("🚀 MESHCOIN FULL NODE (PC) INICIADO 🚀")
	fmt.Println("==================================================")
	fmt.Println("Este nó escutará os celulares da rede P2P e salvará")
	fmt.Println("os blocos validados no Ledger Mestre (ledger.json).")
	fmt.Println("O Ledger será feito o backup na Nebula Cloud (porta 8000).")
	fmt.Println("==================================================")

	// Inicializa o Ledger (Lê do disco ou cria Genesis)
	initLedger()

	// Inicia os listeners de rede
	startNetwork()

	// Mantém o daemon rodando até receber SIGINT (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("\nEncerrando MeshCoin Full Node. Salvando ledger final...")
	saveLedger()
	log.Println("Adeus.")
}
