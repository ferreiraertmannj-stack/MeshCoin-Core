import os

out_dir = "PROJECT_AUDIT"
os.makedirs(out_dir, exist_ok=True)

files = {
    "01_EXECUTIVE_SUMMARY.md": """# 01 EXECUTIVE SUMMARY

## Resumo Executivo
O projeto Nebula Network consiste em um ecossistema offline-first focado em comunicação em malha (Mesh Networking) e um armazenamento distribuído (Nebula Cloud). O componente blockchain (MeshCoin Core) existe como uma camada de registro e recompensa para nós da rede.

## Estado Geral
O projeto encontra-se em um estado funcional de protótipo (Alpha), possuindo implementações híbridas em Go (Full Node, Cloud) e Python (Protótipos), juntamente com um aplicativo Flutter para dispositivos móveis. Há forte ênfase em criptografia (transição planejada para PQC) e resiliência offline.

## Arquitetura Encontrada
- **Mobile/Client:** App Flutter que lida com UI, Carteira, e comunicação Bluetooth/Wi-Fi Direct.
- **Node PC:** Implementação em Go (`pc_node/main.go`) que lida com validação de blocos TCP/UDP, memória de transações (mempool), e interface com o armazenamento Cloud.
- **Cloud Node:** Servidor Go (`nebula_cloud/node_daemon.go`) para armazenamento descentralizado (fragmentação via Reed-Solomon).
- **Scripts em Python:** Protótipos de roteamento, tracker, mineração e integração criptográfica.

## Riscos
- Múltiplas linguagens para lógicas chave (Python para prototipação e Go para Node) podem causar inconsistências no protocolo.
- O consenso baseado em arquivo local JSON (`ledger.json`) apresenta alto risco de dessincronização e conflitos em um ambiente distribuído se a camada de rede não garantir consistência forte.
- Dificuldade estática na mineração sem um mecanismo robusto de ajuste contínuo (aparentemente dependente de halving hardcoded).

## Pontos Fortes
- Foco explícito em resistência à censura e operação off-grid.
- Inovação no algoritmo de mineração focado em CPU/Mobile (NeonHash).
- Divisão clara de responsabilidades entre cliente móvel, full node de validação e infraestrutura de cloud.

## Pontos Fracos
- Arquitetura de armazenamento do ledger em arquivo único JSON.
- Forte acoplamento de lógica de negócio em scripts Python dispersos.
- Gerenciamento de rede híbrido (UDP/TCP) propenso a particionamento de rede.
""",

    "02_PROJECT_STRUCTURE.md": """# 02 PROJECT STRUCTURE

## Estrutura de Diretórios
- `/docs/architecture/`: Documentação oficial imutável.
- `/meshcoin_flutter/`: Aplicativo mobile principal, UI, cliente P2P mobile. Depende da rede local para propagar transações.
- `/pc_node/`: Código Go do Full Node. Responsável pelo consenso, manutenção do `ledger.json` e mineração via TCP/UDP. Depende da Nebula Cloud para long-term storage.
- `/nebula_cloud/`: Servidor de armazenamento descentralizado (Go) que recebe fragmentos de dados (Reed-Solomon).
- `/tools/`: Reservado para ferramentas internas de engenharia.
- `/*.py` (raiz): Scripts Python de prototipação (p.ex. `chat_mesh.py`, `rede_p2p.py`, `pqc_crypto.py`).

## Dependências entre Componentes
- **Flutter App** -> **PC Node**: O app mobile transmite transações P2P via UDP/TCP, que chegam ao PC Node para validação.
- **PC Node** -> **Nebula Cloud**: O PC Node faz o upload do ledger e snapshots a cada X blocos (ex: 10 blocos) para a nuvem.
- **Python Scripts** -> **N/A**: Operam de maneira independente, principalmente como simuladores e validadores de conceito para a arquitetura P2P e PQC.
""",

    "03_DEPENDENCY_GRAPH.md": """# 03 DEPENDENCY GRAPH

## Dependências de Linguagem
- **Go Modules (`go.mod`)**: Dependências em `github.com/decred/dcrd/dcrec/secp256k1/v4` para curva elíptica e ECDSA.
- **Python (`requirements.txt`)**: Dependências para protótipos de criptografia e rede.
- **Flutter (`pubspec.yaml`)**: Plugins de UI, armazenamento local, criptografia.

## Acoplamentos e Riscos
- **Acoplamento forte** entre `pc_node` e o sistema de arquivos local (`ledger.json`). Não utiliza um banco de dados KV puro (como LevelDB ou RocksDB).
- **Dependência Cíclica / Forte** na propagação de blocos e transações entre `network.go` e `ledger.go` usando mutex globais (`ledger.mu`, `mempoolMutex`), o que pode causar gargalos.
""",

    "04_MODULE_ANALYSIS.md": """# 04 MODULE ANALYSIS

## PC Node (Go)
- **Responsabilidade**: Manter estado global da blockchain, minerar blocos, propagar estado pela rede P2P (TCP/UDP).
- **Maturidade**: Alpha.
- **Riscos**: Serialização do ledger inteiro em JSON, sem uso de Merkle Trees eficientes para sync parcial.

## Nebula Cloud (Go)
- **Responsabilidade**: Armazenamento fragmentado e redundante.
- **Maturidade**: Proof of Concept.
- **Riscos**: Falta de incentivo direto automatizado e verificação de disponibilidade (Proofs of Spacetime/Replication).

## Mobile Client (Flutter)
- **Responsabilidade**: Interação do usuário, carteira (chaves secp256k1), chat.
- **Maturidade**: Beta/Funcional.
- **Riscos**: Vazamento de chaves privadas no storage do celular se não for utilizado Keystore/Secure Enclave nativo.

## Python Prototypes
- **Responsabilidade**: Simulação PQC, Reed-Solomon, Tracking.
- **Maturidade**: Experimental.
""",

    "05_BLOCKCHAIN.md": """# 05 BLOCKCHAIN

## Como Funciona
A blockchain é mantida em memória e salva em disco como `ledger.json` no nó PC. Os blocos são empilhados sequencialmente. O bloco Gênesis é hardcoded.

## Quem Cria e Valida
- **Criação**: Nós PC e Mobile criam blocos através de mineração.
- **Validação**: Feita no momento da recepção em `handleNewBlock` (no arquivo `ledger.go`), validando índice, hash prévio, e PoW.

## Armazenamento
Arquivos JSON serializados locais (`ledger.json`) no PC. Nuvem Nebula Cloud armazena backups a cada 10 blocos.

## Dificuldade e Hash
Utiliza o algoritmo `NeonHash` (vetor em memória de 4KB, operações matemáticas pseudo-vetoriais iteradas 128 vezes, finalizando com SHA-256).

## Resolução de Conflitos
A implementação atual de `ledger.go` **rejeita sumariamente** qualquer bloco cujo índice seja menor ou igual ao último conhecido, ou cujo PreviousHash não bata. Falta um mecanismo robusto de reconciliação de forks (Longest Chain real com reorg).

## Bugs e Riscos
- Risco de split brain irreversível na rede sem um processo automático de rollback de blocos órfãos.
- Leitura do JSON inteiro na memória na inicialização.
"""
}

for filename, content in files.items():
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)
