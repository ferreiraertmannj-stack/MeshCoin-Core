# Guia de Implementação: PQC, Go e Flutter - Nébula Cloud

**Data:** Abril de 2026  
**Autor:** Manus AI  
**Cliente:** Jean Ertmann (Ertmann Tech)  
**Projeto:** Nébula Cloud (Armazenamento Descentralizado)

---

## 1. Visão Geral da Atualização

Este guia detalha a implementação de três grandes melhorias na arquitetura do Nébula Cloud:

1.  **Criptografia Pós-Quântica (PQC):** Implementação do padrão NIST FIPS 203 (ML-KEM/Kyber) e FIPS 204 (ML-DSA/Dilithium) para garantir segurança contra ataques de computadores quânticos (Harvest Now, Decrypt Later).
2.  **Node Daemon em Go:** Reescrita do servidor de armazenamento em Go (Golang) para altíssima performance, menor consumo de memória e concorrência nativa.
3.  **Cliente Mobile em Flutter:** Desenvolvimento da base do aplicativo mobile para Android/iOS, permitindo que os usuários finais enviem e recuperem arquivos diretamente pelo celular.

---

## 2. Criptografia Pós-Quântica Híbrida

A criptografia clássica (RSA/ECC) está ameaçada pelo algoritmo de Shor em computadores quânticos. Para mitigar isso, implementamos um sistema **Híbrido (PQC + Clássico)**.

### 2.1. Arquitetura do `pqc_crypto_lite.py`
O módulo foi desenhado para ser resiliente. Se a biblioteca C `liboqs` estiver instalada, ele usa PQC. Caso contrário, faz fallback automático para RSA-4096.

*   **Encapsulação de Chave (KEM):** Usa `ML-KEM-768` (Kyber). O cliente gera uma chave de sessão AES-256 aleatória e a encapsula usando a chave pública PQC do destinatário.
*   **Assinatura Digital (DSA):** Usa `ML-DSA-65` (Dilithium) para assinar os fragmentos e garantir que não foram adulterados pelos nós.
*   **Criptografia Simétrica:** Usa `AES-256-GCM`. A chave final é uma derivação híbrida (XOR) entre a chave de sessão KEM e a senha do usuário (PBKDF2HMAC).

### 2.2. Como Testar o PQC no Python
```bash
# 1. Instalar dependências (se necessário)
sudo apt-get install -y liboqs-dev
pip3 install liboqs-python cryptography

# 2. Executar o teste do módulo PQC
python3 pqc_crypto_lite.py

# 3. Testar a fragmentação PQC
python3 fragmenter_pqc.py
```

---

## 3. Node Daemon em Go (Alta Performance)

O antigo `node_daemon.py` em FastAPI foi substituído por `node_daemon.go`.

### 3.1. Vantagens do Go para o Nébula Cloud
*   **Performance:** Go é compilado para código de máquina, processando uploads/downloads de fragmentos até 10x mais rápido que Python.
*   **Concorrência:** O uso de *goroutines* permite que um único nó de smartphone (ARM) lide com milhares de conexões simultâneas sem travar.
*   **Binário Único:** Não requer instalação do Python no celular do afiliado. Basta rodar o executável compilado.

### 3.2. Como Compilar e Executar
```bash
# 1. Compilar o código Go
go build -o node_daemon_go node_daemon.go

# 2. Executar o binário
./node_daemon_go

# O servidor iniciará na porta 8000 e criará a pasta ./node_storage
```

### 3.3. Endpoints Disponíveis
*   `GET /status`: Retorna uptime, quantidade de fragmentos e tamanho total em disco.
*   `POST /fragmento/upload`: Recebe o arquivo `.bin` (multipart/form-data).
*   `GET /fragmento/download?nome=frag_01.bin`: Retorna o arquivo solicitado.
*   `DELETE /fragmento/delete?nome=frag_01.bin`: Exclui o fragmento.
*   `GET /chave-publica`: Retorna a chave pública do nó para KEM.
*   `GET /fragmentos`: Lista todos os fragmentos armazenados no nó.

---

## 4. Cliente Mobile (Flutter)

Foi criada a base do aplicativo mobile (`flutter_app_main.dart`), desenhado para ser o "Google Drive" descentralizado do cliente final.

### 4.1. Funcionalidades do App
*   **Conexão Direta:** Permite inserir o IP/Porta do Node Daemon (ou do Tracker, na arquitetura final) e verificar o status da rede.
*   **Upload Fragmentado:** Interface para selecionar arquivos do celular, aplicar hash SHA-256 e enviar via multipart request.
*   **Gerenciamento:** Lista fragmentos armazenados na rede e permite o download direto para o armazenamento interno do celular.
*   **Interface Dark Mode:** Design moderno e alinhado com a identidade visual da Ertmann Tech.

### 4.2. Como Integrar no seu Ambiente Flutter
1.  Crie um novo projeto Flutter: `flutter create nebula_cloud_app`
2.  Substitua o arquivo `lib/main.dart` pelo conteúdo do `flutter_app_main.dart` fornecido.
3.  Substitua o `pubspec.yaml` pelo arquivo fornecido.
4.  Execute `flutter pub get` para baixar as dependências.
5.  Para suporte a PQC real no Flutter, recomendamos integrar a biblioteca C `liboqs` via **FFI (Foreign Function Interface)** no futuro, já que a criptografia nativa do Dart ainda não suporta Kyber/Dilithium nativamente.

---

## 5. Próximos Passos para o "Vibe Coding"

Como você desenvolve com IA (Vibe Coding), aqui está o roadmap para seus próximos prompts:

1.  **Prompt 1 (Tracker em Go):** *"Traduza o meu tracker.py para Go, mantendo a integração com a biblioteca reedsolo (pode usar a lib 'klauspost/reedsolomon' do Go)."*
2.  **Prompt 2 (App Flutter - PQC):** *"Crie um plugin Flutter via FFI que encapsule a biblioteca liboqs em C, para que meu app mobile possa gerar chaves ML-KEM-768 nativamente."*
3.  **Prompt 3 (Rede P2P):** *"Adicione descoberta de nós P2P (DHT - Kademlia) no meu node_daemon.go, para que os nós se encontrem sem precisar de um IP fixo."*

---
**Arquivos Entregues neste Pacote:**
*   `pqc_crypto_lite.py`: Módulo criptográfico híbrido (Kyber/Dilithium + AES + RSA).
*   `fragmenter_pqc.py`: Lógica de Reed-Solomon integrada com o módulo PQC.
*   `node_daemon.go`: Servidor de armazenamento reescrito em Go.
*   `node_daemon_go`: Binário Linux compilado e pronto para uso.
*   `flutter_app_main.dart`: Código-fonte da tela principal do app mobile.
*   `pubspec.yaml`: Dependências do projeto Flutter.
