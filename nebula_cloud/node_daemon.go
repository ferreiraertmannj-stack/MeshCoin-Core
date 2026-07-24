package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ==================== CONFIGURAÇÃO ====================

const (
	PORT                = ":8000"
	STORAGE_DIR         = "./node_storage"
	LOG_FILE            = "./node_daemon.log"
	RSA_KEY_SIZE        = 4096
	AES_KEY_SIZE        = 32
	NONCE_SIZE          = 12
	MAX_FRAGMENT_SIZE   = 100 * 1024 * 1024 // 100MB
	HEARTBEAT_INTERVAL  = 10 * time.Second
)

// ==================== ESTRUTURAS ====================

type NodeStatus struct {
	Status              string    `json:"status"`
	Node                string    `json:"node"`
	FragmentosArmazenados int     `json:"fragmentos_armazenados"`
	TamanhoTotalBytes   int64     `json:"tamanho_total_bytes"`
	Timestamp           time.Time `json:"timestamp"`
	Uptime              string    `json:"uptime"`
	Versao              string    `json:"versao"`
}

type FragmentMetadata struct {
	NomeArquivo  string `json:"nome_arquivo"`
	Tamanho      int64  `json:"tamanho"`
	Hash         string `json:"hash"`
	Timestamp    time.Time `json:"timestamp"`
	Algoritmo    string `json:"algoritmo"`
}

// ==================== VARIÁVEIS GLOBAIS ====================

var (
	nodeID      = "nebula-node-01"
	startTime   = time.Now()
	chavePublicaRSA *rsa.PublicKey
	chavePrivadaRSA *rsa.PrivateKey
	logger      *log.Logger
)

// ==================== INICIALIZAÇÃO ====================

func init() {
	// Criar diretório de armazenamento
	os.MkdirAll(STORAGE_DIR, 0755)

	// Configurar logging
	logFile, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	logger = log.New(logFile, "[NODE_DAEMON] ", log.LstdFlags)

	// Gerar chaves RSA (fallback para PQC)
	gerarChavesRSA()

	logger.Println("Node Daemon inicializado com sucesso")
}

func gerarChavesRSA() {
	var err error
	chavePrivadaRSA, err = rsa.GenerateKey(rand.Reader, RSA_KEY_SIZE)
	if err != nil {
		logger.Fatalf("Erro ao gerar chaves RSA: %v", err)
	}
	chavePublicaRSA = &chavePrivadaRSA.PublicKey
	logger.Println("Chaves RSA-4096 geradas com sucesso")
}

// ==================== HANDLERS HTTP ====================

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Contar fragmentos e tamanho total
	fragmentos := 0
	var tamanhoTotal int64 = 0

	files, err := ioutil.ReadDir(STORAGE_DIR)
	if err != nil {
		http.Error(w, "Erro ao ler diretório", http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".bin") {
			fragmentos++
			tamanhoTotal += file.Size()
		}
	}

	uptime := time.Since(startTime).String()

	status := NodeStatus{
		Status:              "online",
		Node:                nodeID,
		FragmentosArmazenados: fragmentos,
		TamanhoTotalBytes:   tamanhoTotal,
		Timestamp:           time.Now(),
		Uptime:              uptime,
		Versao:              "1.0.0-PQC",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
	logger.Printf("Status solicitado: %d fragmentos, %d bytes", fragmentos, tamanhoTotal)
}

func handleUploadFragmento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Validar tamanho máximo
	r.Body = http.MaxBytesReader(w, r.Body, MAX_FRAGMENT_SIZE)

	// Obter nome do fragmento
	nomeFragmento := r.FormValue("nome")
	if nomeFragmento == "" {
		http.Error(w, "Nome do fragmento não fornecido", http.StatusBadRequest)
		return
	}

	// Validar nome (prevenir path traversal)
	if strings.Contains(nomeFragmento, "..") || strings.Contains(nomeFragmento, "/") {
		http.Error(w, "Nome de arquivo inválido", http.StatusBadRequest)
		return
	}

	// Ler arquivo enviado
	file, _, err := r.FormFile("fragmento")
	if err != nil {
		http.Error(w, "Erro ao ler arquivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Salvar fragmento
	caminhoFragmento := filepath.Join(STORAGE_DIR, nomeFragmento)
	outFile, err := os.Create(caminhoFragmento)
	if err != nil {
		http.Error(w, "Erro ao salvar arquivo", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, "Erro ao copiar arquivo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"mensagem": fmt.Sprintf("Fragmento %s salvo com sucesso", nomeFragmento),
	})
	logger.Printf("Fragmento salvo: %s", nomeFragmento)
}

func handleDownloadFragmento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	nomeFragmento := r.URL.Query().Get("nome")
	if nomeFragmento == "" {
		http.Error(w, "Nome do fragmento não fornecido", http.StatusBadRequest)
		return
	}

	// Validar nome
	if strings.Contains(nomeFragmento, "..") || strings.Contains(nomeFragmento, "/") {
		http.Error(w, "Nome de arquivo inválido", http.StatusBadRequest)
		return
	}

	caminhoFragmento := filepath.Join(STORAGE_DIR, nomeFragmento)

	// Verificar se arquivo existe
	if _, err := os.Stat(caminhoFragmento); os.IsNotExist(err) {
		http.Error(w, "Fragmento não encontrado", http.StatusNotFound)
		return
	}

	// Enviar arquivo
	http.ServeFile(w, r, caminhoFragmento)
	logger.Printf("Fragmento baixado: %s", nomeFragmento)
}

func handleDeleteFragmento(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	nomeFragmento := r.URL.Query().Get("nome")
	if nomeFragmento == "" {
		http.Error(w, "Nome do fragmento não fornecido", http.StatusBadRequest)
		return
	}

	// Validar nome
	if strings.Contains(nomeFragmento, "..") || strings.Contains(nomeFragmento, "/") {
		http.Error(w, "Nome de arquivo inválido", http.StatusBadRequest)
		return
	}

	caminhoFragmento := filepath.Join(STORAGE_DIR, nomeFragmento)

	err := os.Remove(caminhoFragmento)
	if err != nil {
		http.Error(w, "Erro ao deletar arquivo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"mensagem": fmt.Sprintf("Fragmento %s deletado com sucesso", nomeFragmento),
	})
	logger.Printf("Fragmento deletado: %s", nomeFragmento)
}

func handleChavePublica(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Serializar chave pública
	chavePublicaDER, err := x509.MarshalPKIXPublicKey(chavePublicaRSA)
	if err != nil {
		http.Error(w, "Erro ao serializar chave", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"chave_publica": hex.EncodeToString(chavePublicaDER),
		"algoritmo": "RSA-4096",
		"node": nodeID,
	})
	logger.Println("Chave pública solicitada")
}

func handleListaFragmentos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	files, err := ioutil.ReadDir(STORAGE_DIR)
	if err != nil {
		http.Error(w, "Erro ao ler diretório", http.StatusInternalServerError)
		return
	}

	var fragmentos []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".bin") {
			fragmentos = append(fragmentos, file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total": len(fragmentos),
		"fragmentos": fragmentos,
	})
	logger.Printf("Lista de fragmentos solicitada: %d fragmentos", len(fragmentos))
}

// ==================== MIDDLEWARE ====================

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next(w, r)
		logger.Printf("Tempo de resposta: %v", time.Since(start))
	}
}

// ==================== MAIN ====================

func main() {
	// Registrar rotas
	http.HandleFunc("/status", loggingMiddleware(handleStatus))
	http.HandleFunc("/fragmento/upload", loggingMiddleware(handleUploadFragmento))
	http.HandleFunc("/fragmento/download", loggingMiddleware(handleDownloadFragmento))
	http.HandleFunc("/fragmento/delete", loggingMiddleware(handleDeleteFragmento))
	http.HandleFunc("/chave-publica", loggingMiddleware(handleChavePublica))
	http.HandleFunc("/fragmentos", loggingMiddleware(handleListaFragmentos))

	// Iniciar servidor
	logger.Printf("Iniciando Node Daemon na porta %s", PORT)
	fmt.Printf("🚀 Node Daemon rodando em http://localhost%s\n", PORT)
	fmt.Printf("📊 Armazenamento: %s\n", STORAGE_DIR)
	fmt.Printf("🔐 Criptografia: RSA-4096 (fallback para PQC)\n")

	err := http.ListenAndServe(PORT, nil)
	if err != nil {
		logger.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

// ==================== FUNÇÕES AUXILIARES ====================

func criptografarAES(dados []byte, chave []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(chave)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, NONCE_SIZE)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, dados, nil)
	return ciphertext, nonce, nil
}

func descriptografarAES(ciphertext []byte, nonce []byte, chave []byte) ([]byte, error) {
	block, err := aes.NewCipher(chave)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func calcularHash(dados []byte) string {
	hash := sha256.Sum256(dados)
	return hex.EncodeToString(hash[:])
}
