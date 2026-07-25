# 🌌 Nebula Network — A Blockchain Invisível

![Status do Projeto](https://img.shields.io/badge/Status-Alpha%20Funcional-blueviolet)
![Licença](https://img.shields.io/badge/License-MIT-green)
![Origem](https://img.shields.io/badge/Origem-Extrema%2FMG%20🇧🇷-red)
![Token](https://img.shields.io/badge/Token-NBL-cyan)
![Supply](https://img.shields.io/badge/Max%20Supply-100M%20NBL-gold)

> *"A rede está lá. Você só não vê."*

A **Nebula Network** é um ecossistema completo de blockchain descentralizada, mensageria criptografada e armazenamento distribuído que funciona **com ou sem internet**. Transformamos smartphones, PCs e até módulos ESP32 em nós de uma rede global resiliente, soberana e privada — como uma nebulosa cósmica: invisível a olho nu, mas composta por bilhões de partículas interagindo.

---

## 🚀 Diferenciais Únicos

| Recurso | Descrição |
|---|---|
| 🛰️ **Conectividade Off-Grid** | Transacione e comunique-se via Bluetooth BLE, Wi-Fi Direct, Hotspot ou qualquer conexão disponível. Funciona em zonas de sombra, apagões e regiões censuradas. |
| ⛏️ **Mineração ASIC-Resistant** | `NeonHash` (ARM/Mobile) + `RandomX` (x86/PC) — impossibilita GPUs e ASICs. Uma CPU, um voto. |
| 💬 **Nebula Chat (E2EE)** | Mensageria estilo Telegram 100% descentralizada. Textos, fotos e áudios (até 1 min), tudo fragmentado, criptografado ponta-a-ponta (AES-256-GCM) e armazenado na Nebula Cloud. |
| ☁️ **Nebula Cloud** | Armazenamento descentralizado com Reed-Solomon. Cada usuário doa no mínimo 5GB e recebe boost na mineração. O Ledger completo nunca fica no celular — apenas fragmentos criptografados. |
| 🔀 **Protocolo B.A.T.M.A.N.** | Roteamento multi-hop inteligente. Dados saltam de nó em nó até o destino, mesmo sem linha direta. |
| 🔒 **Soberania Digital** | Zero servidores centrais. Zero rastreamento. Zero censura. |

---

## 🪙 Tokenomics — NBL (Nebula)

O **NBL** é o token utilitário nativo da Nebula Network, funcionando como **gás** para todas as operações da rede.

| Parâmetro | Valor |
|---|---|
| **Nome** | Nebula |
| **Ticker** | NBL |
| **Supply Máximo** | 100.000.000 (100 Milhões) |
| **Block Time** | 2 minutos |
| **Recompensa Inicial** | 50 NBL / bloco |
| **Halving** | A cada 2.100.000 blocos (~8 anos) |
| **Consenso** | Proof-of-Work Híbrido (NeonHash + RandomX) |
| **Criptografia** | ECDSA secp256k1 · Base58Check |
| **Boost de Mineração** | +25% para nós que doam ≥5GB à Nebula Cloud |

### Cronograma de Emissão

| Era | Blocos | Recompensa | NBL Emitidos | Acumulado |
|---|---|---|---|---|
| 1ª | 0 – 2.099.999 | 50 NBL | 105.000.000* | ~100M |
| 2ª | 2.1M – 4.199.999 | 25 NBL | — | — |
| 3ª | 4.2M – 6.299.999 | 12.5 NBL | — | — |
| ... | ... | ... | ... | ≤100M |

> *O supply é limitado a 100M. A 1ª era atingirá o teto e a emissão cessará automaticamente.*

---

## 📱 Componentes do Ecossistema

### 1. Nebula App (Flutter — Android/iOS)
O aplicativo principal para smartphones. Inclui:
- **Carteira NBL** — Gere carteiras reais (ECDSA secp256k1), envie e receba NBL via QR Code.
- **Nebula Chat** — Mensageria E2EE descentralizada com suporte a texto, fotos e áudios.
- **Minerador Mobile** — Mine blocos diretamente do seu celular usando o algoritmo NeonHash.
- **Painel de Rede** — Visualize peers conectados, tabela de roteamento B.A.T.M.A.N e status da mesh.
- **Nebula Cloud** — Dashboard de armazenamento descentralizado e fragmentos do ledger.

### 2. Nebula Full Node (Go — PC x86/x64)
Nó completo para PCs que valida blocos, mantém o Ledger Mestre e sincroniza com a Nebula Cloud.
```
pc_node/
├── main.go              # Daemon principal
├── ledger.go            # Blockchain e validação de blocos
├── network.go           # UDP/TCP P2P (discovery + sync)
└── nebula_integration.go # Upload do Ledger para a Nebula Cloud
```

### 3. Nebula Cloud (Go — Armazenamento Descentralizado)
Servidor de armazenamento distribuído com fragmentação Reed-Solomon. O Ledger e os dados do Nebula Chat são criptografados, fragmentados e espalhados entre os nós participantes.
```
nebula-cloud/
├── node_daemon.go       # Servidor HTTP (porta 8000)
├── shard_manager.go     # Fragmentação e reconstrução
└── storage/             # Diretório de shards
```

### 4. Hardware: ESP32 Repeaters (Futuro)
Módulos ESP32 com antenas de longo alcance atuando como repetidores cegos da rede mesh, expandindo a cobertura da Nebula para quilômetros sem internet.

---

## 🏗️ Ecossistema de Hardware

| Dispositivo | Função | Status |
|---|---|---|
| 📱 Smartphone (ARM) | Nó móvel, minerador, chat, carteira | ✅ Funcional |
| 💻 PC (x86/x64) | Full Node, validador, Ledger Mestre | ✅ Funcional |
| 📡 ESP32 + Antena | Repetidor mesh off-grid (LoRa/Wi-Fi) | 🔜 Planejado |
| ☀️ Solar Box | Nó autônomo para áreas rurais | 🔜 Planejado |

---

## 📡 Casos de Uso

- **Comunicação em Desastres:** Alertas de emergência e chat mesmo sem torres de celular.
- **Comércio Local Off-Grid:** Pagamentos em NBL quando a internet do estabelecimento falha.
- **Privacidade Total:** Chat e transações financeiras sem intermediários ou vigilância.
- **Logística 4.0:** Transporte de dados via "Mulas de Dados" (caminhões, veículos) por todo o Brasil 🇧🇷.
- **Áreas Censuradas:** Transações e comunicação livres em regiões com bloqueios governamentais.
- **IoT Descentralizado:** Micro-pagamentos e telemetria via rede mesh para sensores e dispositivos.

---

## 🛠️ Stack Tecnológica

| Camada | Tecnologia |
|---|---|
| **App Mobile** | Flutter/Dart |
| **Full Node (PC)** | Go (Golang) |
| **Nebula Cloud** | Go (Golang) |
| **Criptografia** | ECDSA secp256k1, AES-256-GCM, SHA-256 |
| **Rede P2P** | UDP Broadcast, TCP, BLE, Wi-Fi Direct |
| **Roteamento** | B.A.T.M.A.N. (Better Approach To Mobile Adhoc Networking) |
| **Fragmentação** | Reed-Solomon Erasure Coding |

---

## 🗺️ Roadmap

- [x] Concepção do Protocolo e Algoritmo NeonHash
- [x] Protótipo do Core (Blockchain, Wallet, Miner)
- [x] App Flutter funcional com 6 abas (Home, Carteira, Chat, Mineração, Rede, Cloud)
- [x] Rede P2P entre smartphones (UDP/TCP + BLE + Wi-Fi Direct)
- [x] Full Node PC em Golang com validação de blocos
- [x] Integração com Nebula Cloud (upload automático do Ledger)
- [x] Nebula Chat com criptografia E2EE
- [x] Rebranding completo para Nebula Network
- [ ] RandomX para mineração x86/x64 no PC
- [ ] Sistema de usernames (@usuario) vinculado à blockchain
- [ ] Envio de fotos e áudios comprimidos no Nebula Chat
- [ ] Boost de mineração para doadores de storage (Proof of Storage)
- [ ] Módulo repetidor ESP32 com antena de longo alcance
- [ ] Whitepaper v2 da Nebula Network
- [ ] Lançamento oficial do Bloco Gênesis

---

## ⚡ Quick Start

### Pré-requisitos
- **Flutter SDK** (para o app mobile)
- **Go 1.21+** (para o Full Node e Nebula Cloud)
- **Android Studio / VSCode** (IDE)

- **Desenvolvedor:** Jean Ertmann
- **E-mail:** ferreiraertmannj@gmail.com
- **GitHub:** [github.com/ferreiraertmannj-stack/MeshCoin-Core](https://github.com/ferreiraertmannj-stack/MeshCoin-Core)

### 1. Iniciar a Nebula Cloud
```bash
cd nebula-cloud
go run .
# Servidor rodando na porta 8000
```

### 2. Iniciar o Full Node do PC
```bash
cd pc_node
go run .
# 🚀 NEBULA NETWORK FULL NODE (PC) INICIADO 🚀
```

### 3. Compilar e Instalar o App
```bash
cd meshcoin_flutter
flutter pub get
flutter build apk --release
# APK gerado em build/app/outputs/flutter-apk/app-release.apk
```

---

## 📜 Licença

Este projeto está licenciado sob a **MIT License** — veja o arquivo [LICENSE](LICENSE) para detalhes.

---

## 🤝 Contribuição

Contribuições são bem-vindas! Abra uma issue ou envie um pull request. Para mudanças significativas, abra uma discussão primeiro.

---

<p align="center">
  <strong>🌌 Nebula Network</strong><br>
  <em>A rede está lá. Você só não vê.</em><br>
  <sub>Feito com ❤️ em Extrema/MG 🇧🇷</sub>
</p>
