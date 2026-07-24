package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

// O endereço do Nebula Cloud Daemon (Nó de Armazenamento na máquina ou rede)
const nebulaNodeURL = "http://127.0.0.1:8000"

// uploadToNebulaCloud envia o ledger atual como um arquivo (fragmento) para o Nebula Daemon
func uploadToNebulaCloud() {
	log.Println("☁️  Iniciando backup do Ledger na Nebula Cloud...")

	ledgerData := getLedgerJSON()
	if len(ledgerData) == 0 {
		log.Println("Ledger vazio, ignorando backup.")
		return
	}

	// Criando um arquivo temporário para o upload Multipart
	tempFileName := "ledger_backup.json"
	os.WriteFile(tempFileName, ledgerData, 0644)
	defer os.Remove(tempFileName)

	file, err := os.Open(tempFileName)
	if err != nil {
		log.Println("Erro ao ler ledger temporário:", err)
		return
	}
	defer file.Close()

	var requestBody bytes.Buffer
	multiPartWriter := multipart.NewWriter(&requestBody)

	// O node_daemon.go da Nebula Cloud pede "nome" e o arquivo em "fragmento"
	_ = multiPartWriter.WriteField("nome", "meshcoin_ledger_master.json")

	fileWriter, err := multiPartWriter.CreateFormFile("fragmento", "meshcoin_ledger_master.json")
	if err != nil {
		log.Println("Erro no multipart:", err)
		return
	}

	_, _ = io.Copy(fileWriter, file)
	multiPartWriter.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/upload", nebulaNodeURL), &requestBody)
	if err != nil {
		log.Println("Erro ao criar requisição Nebula:", err)
		return
	}

	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("❌ Erro ao enviar para Nebula Cloud. O node_daemon.go está rodando na porta 8000?")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("✅ Backup na Nebula Cloud concluído com sucesso!")
	} else {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("❌ Falha no backup Nebula (Status %d): %s\n", resp.StatusCode, string(bodyBytes))
	}
}
