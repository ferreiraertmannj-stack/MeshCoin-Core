import os

files = {
    "PROJECT_BASELINE/01_CURRENT_SYSTEM_STATE.md": """# 01 CURRENT SYSTEM STATE

## Estado Atual por Módulo

- **Blockchain (Ledger):** Parcialmente Funcional
  - *Justificativa:* Consegue minerar blocos e salvar histórico JSON. Mas a resiliência contra arquivos corrompidos está quebrada e não há mecanismo nativo forte de reorg de cadeia (Forks).
- **Consensus:** Experimental
  - *Justificativa:* O algoritmo NeonHash é executado e validado corretamente, mas o nó carece de verificação em larga escala e resolve empates de blocos pela primeira vez que são vistos, sem `Longest Chain` claro implementado adequadamente no TCP stream atual.
- **Wallet:** Funcional
  - *Justificativa:* A geração de chaves `secp256k1` e verificação de assinatura de pacotes de dados operam de forma condizente e determinística.
- **Mesh:** Não Implementado (Falso) / Parcialmente Funcional (LAN apenas)
  - *Justificativa:* O protocolo fala em "Mesh" e roteamento multi-hop offline, mas o que existe no `network.go` é estritamente um TCP broadcast e um UDP de LAN. Sem roteadores Ad-Hoc ou DHT funcionais.
- **Mining:** Funcional
  - *Justificativa:* Mineração NeonHash opera perfeitamente com controle de dificuldade estático e verificação P2P coerente.
- **Networking:** Parcialmente Funcional
  - *Justificativa:* A base TCP roda estavelmente para 1:1, mas está propensa a loop infinito de rotas devido à ausência de pacotes roteáveis indexados e é frágil a DoS (lock global).
- **Nebula Cloud:** Experimental
  - *Justificativa:* Endpoint exposto, consegue enviar blocos, mas não existe autenticação, criptografia nativa no pacote base sem uso dos scripts paralelos ou prova de storage amarrada à wallet via smart-contract.
- **Storage (Local):** Quebrado (Em Escala)
  - *Justificativa:* Lidar com dump de `ledger.json` inteiro não permite escalabilidade além de algumas dezenas de megabytes antes de esgotar CPU/RAM de um nó pequeno.
- **Synchronization:** Parcialmente Funcional
  - *Justificativa:* Nodes pegam pacotes de broadcasts ao estarem online (listening). Recuperar um node atrasado desde o Gênesis é inviável no código atual.
- **Crypto:** Funcional
  - *Justificativa:* ECDSA funciona, e o pacote PQC (Dilithium) tem base sólida, restando apenas substituição final dos stubs.
""",

    "PROJECT_BASELINE/02_TEST_MATRIX.md": """# 02 TEST MATRIX

## Blockchain
- `test_append_block`: Valida se o bloco adicionado à chain incrementa índice e atualiza hash.
- `test_corrupt_ledger_init`: Garante panic/recuperação ao invés de sobrescrever genesis.

## Consensus
- `test_neonhash_valid`: Verifica corretude da prova (nonce, hash, difficulty).
- `test_reject_invalid_pow`: Garante que pacotes de rede com hashes falsos sejam barrados na porta sem lock.
- `test_fork_resolution`: Força submissão de cadeia paralela maior.

## Wallet
- `test_key_generation`: Valida curva secp256k1 e geração de endereço Base58Check.
- `test_transaction_sign`: Valida assinatura ECDSA da transação.
- `test_invalid_signature`: Transação forjada deve ser descartada.

## Mesh
- `test_broadcast_loop`: Simula topologia anelar; garante que o pacote TCP morre por TTL/ID cache.
- `test_dht_discovery`: Validar encontro de peers WAN.

## Mining
- `test_halving_reward`: Teste matemático da fórmula com storage bonuses variados.

## Networking
- `test_tcp_flood`: Conexões massivas para testar OOM.
- `test_rate_limit`: Bloqueia envio de +10 blocos/seg do mesmo peer.

## Nebula Cloud
- `test_cloud_upload_auth`: Falha intencional de POST sem signature header.
- `test_file_chunking`: Teste do algoritmo Reed-Solomon.

## Storage
- `test_leveldb_write`: Mock de migração para DB nativo, valida ACID (Atomic).

## Synchronization
- `test_fast_sync`: Nó zera e demanda cadeia completa aos peers em batch.

## Crypto
- `test_pqc_stub`: Testar futura assinatura Dilithium simulada.
""",

    "PROJECT_BASELINE/03_INTEGRATION_MATRIX.md": """# 03 INTEGRATION MATRIX

## Mapeamento de Dependências

- **Mobile App (Flutter) -> PC Node (Go):** Integração Crítica. O mobile atua apenas como assinador/interface e depende cegamente da saúde do PC Node para espalhar tx e baixar blocos da LAN.
- **PC Node -> Nebula Cloud:** Integração Moderada. O PC Node fará upload a cada 10 blocos. Se a Cloud cair, o Node deve arquivar temporariamente e não crashar.
- **Python Scripts -> PC Node / Cloud:** Experimental. Ferramentas externas de auditoria criptográfica. Sem dependência direta.

## Integrações Críticas Identificadas
1. `Mempool <-> Consensus`: A exclusão de transações da Mempool (Mempool Mutex) mediante a validação de um novo bloco (Ledger Mutex) cria um risco cruzado de deadlock.
2. `TCP Listener <-> Ledger Disk IO`: A recepção de pacotes congela a rede se o IO local do ledger for sobrecarregado, expondo a integração entre Networking e Storage como falha de design.
""",

    "04_CRITICAL_PATHS.md": """# 04 CRITICAL PATHS

## 1. Persistência do Ledger
- `handleNewBlock` -> Mutex Lock -> Validation -> Array Append -> `json.Marshal` -> Disk Write -> Mutex Unlock.
*(Caminho extremamente custoso e perigoso).*

## 2. Mineração e Broadcast
- Calculo NeonHash -> Sucesso -> Propaga `NEW_BLOCK` -> Outros nós validam -> Rejeitam se conflito / Aceitam e gravam.

## 3. Criação de Carteira
- (App Mobile) Gera Entropia -> Secp256k1 Key Pair -> ECDSA Sign -> Propaga Address -> Escuta Blockchain por balance.

## 4. Transferência
- Sign Tx -> TCP `NEW_TRANSACTION` -> Validador Mempool -> Mineração -> Ledger.

## 5. Entrada de Novo Nó
- Inicializa vazio -> UDP 5555 Broadcast -> Aceita novas tx/blocos TCP -> *(Não há Fetch Histórico = Nó inútil)*.

## 6. Sincronização
- Atualmente manual/passiva.
""",

    "PROJECT_BASELINE/04_CRITICAL_PATHS.md": """# 04 CRITICAL PATHS

## 1. Persistência do Ledger
- `handleNewBlock` -> Mutex Lock -> Validation -> Array Append -> `json.Marshal` -> Disk Write -> Mutex Unlock.
*(Caminho extremamente custoso e perigoso).*

## 2. Mineração e Broadcast
- Calculo NeonHash -> Sucesso -> Propaga `NEW_BLOCK` -> Outros nós validam -> Rejeitam se conflito / Aceitam e gravam.

## 3. Criação de Carteira
- (App Mobile) Gera Entropia -> Secp256k1 Key Pair -> ECDSA Sign -> Propaga Address -> Escuta Blockchain por balance.

## 4. Transferência
- Sign Tx -> TCP `NEW_TRANSACTION` -> Validador Mempool -> Mineração -> Ledger.

## 5. Entrada de Novo Nó
- Inicializa vazio -> UDP 5555 Broadcast -> Aceita novas tx/blocos TCP -> *(Não há Fetch Histórico = Nó inútil)*.

## 6. Sincronização
- Atualmente manual/passiva.
""",

    "PROJECT_BASELINE/05_STATE_MACHINES.md": """# 05 STATE MACHINES

## State Machine: TCP Peer Lifecycle
- **Estados:** Disconnected, Handshake (Ping/OGM), Established, Syncing, Banned.
- **Eventos:** Connect, Auth Timeout, Packet Received, Malformed Packet.
- **Falhas:** Timeout sem `OGM` desliga conexão; Ban permanente se PoW falso > 5 vezes.
- **Recuperação:** Auto-reconnect via thread background.

## State Machine: Block Validation
- **Estados:** Receiving, Hash Verification, Tx Verification, Ledger Append, Broadcast.
- **Eventos:** `NEW_BLOCK` payload, ECDSA OK/Fail, PreviousHash Match/Mismatch.
- **Falhas:** Rejeição. Se ledger local maior, envia cadeia superior para Peer.
- **Recuperação:** Drop block. Se fork detectado (mesma altura), armazena pendente.
""",

    "PROJECT_BASELINE/06_REGRESSION_MATRIX.md": """# 06 REGRESSION MATRIX

| Bug Identificado | Área de Correção | Módulos em Risco de Quebra (Regressão) |
|---|---|---|
| **Ledger Atomicity** | Refatoração IO | Parsing de saldo na Wallet (App Flutter), Validação retroativa de assinaturas antigas. |
| **Network DDoS** | Filtros TCP / Upload Auth | Sincronização de peers legítimos que falhem novos handshakes, PC node perdendo sync. |
| **Global Lock** | Ledger.go (Channels) | Mempool perdendo/duplicando TX, Data Races fantasma na hora do mineiro ler a Mempool para fechar bloco. |
| **JSON Storage** | N/A (Migração BD) | Estrutura de dados `Block` serializável do Go pode perder tags ou chaves caso o codec mude, quebrando P2P. |
| **False Mesh** | Network.go (P2P routing) | Dispositivos mobile em LAN podem parar de se "enxergar" se o UDP simples for desligado cedo demais para adotar DHT pesado. |
""",

    "PROJECT_BASELINE/07_TEST_DATASETS.md": """# 07 TEST DATASETS

Conjuntos necessários de simulação estruturada:

1. **GENESIS_DATASET:** O bloco Gênesis padrão em múltiplos formatos e cadeias derivadas de 1 a 10 blocos "limpos" para testes de reorg/long-chain.
2. **BAD_ACTORS:** Arquivos JSON com blocos forjados (assinatura falsa, hash manipulado, índice errado, transação double-spend, saldo fantasma).
3. **NETWORK_FLOOD_PAYLOADS:** Pacotes brutos de TCP malformados e estourando 100MB de limite de upload de cloud.
4. **MEMPOOL_DUMP:** 5.000 transações válidas pré-computadas para teste de engasgo (throughput testing) e lock contention no Validator Node.
5. **CORRUPTED_LEDGER:** Simulações de arquivo `ledger.json` cortados pela metade (Null byte injetado) para teste do "Safe Init" e "Atomic Save".
""",

    "PROJECT_BASELINE/08_ACCEPTANCE_CHECKLIST.md": """# 08 ACCEPTANCE CHECKLIST

Qualquer PR ou Fix proposto no Nebula Network deverá cumprir:

- [ ] 1. Não utiliza Mocks no pacote produtivo. Fakes apenas restritos aos arquivos `_test.go` ou diretório `/test`.
- [ ] 2. Não apaga a documentação, apenas adiciona ou altera status.
- [ ] 3. Se altera estrutura TCP/UDP, o mobile Flutter é compatível.
- [ ] 4. Testes unitários para nova lógica cobrem cenários negativos explícitos.
- [ ] 5. Concorrência: Código submetido não bloqueia o IO principal com `sync.Mutex` ao longo de chamadas de disco (`ioutil.Write`).
- [ ] 6. Nenhuma refatoração altera o Bloco Gênesis ou endereços legados.
- [ ] 7. Invariantes Mantidos (Sistema Descentralizado preservado, Off-grid possível).
""",

    "PROJECT_BASELINE/09_RISK_ACCEPTANCE.md": """# 09 RISK ACCEPTANCE

## Riscos Aceitáveis
1. **Lentidão em Dispositivos Low-End (Mobile):** A mineração NeonHash consumirá muita bateria; este risco é tolerado temporariamente até o balanceamento definitivo do ASIC resistance.
2. **Desconexão de Peers em Redes Celulares (4G/5G Nat via CGNAT):** Até o DHT robusto estar funcional, dependemos de IPs públicos do PC e aceita-se temporariamente instabilidade mesh.
3. **Inconsistência da Mempool Visual:** O App Flutter pode apresentar transações como "Aguardando" por longos períodos se a malha isolar o celular. Tolerável.

## Riscos Inaceitáveis
1. **Perda de Ledger Local:** Corrupção de arquivos de estado mestre nunca pode destruir o trabalho validado.
2. **Execução Remota (RCE) via Pacotes P2P:** Ausência de limites TCP/BufferOverflow.
3. **Locks Globais Permanentes:** Desligar a rede inteira por IO engasgado descaracteriza a "alta disponibilidade" da Nebula.
""",

    "PROJECT_BASELINE/10_BASELINE_SUMMARY.md": """# 10 BASELINE SUMMARY

O ecossistema Nebula Network está na transição de "Prova de Conceito" para "Arquitetura Alpha". 

A **Baseline Técnica** revela um código Go funcional (Node, Cloud) suportado por Python scripts paralelos e um App Flutter integrado via rede. Porém, a infraestrutura fundamental de persistência e malha distribuída está engessada em paradigmas síncronos (JSON e Mutexes) com rotas estáticas em LAN.

Os pontos fortes (criptografia independente, foco em descentralização raiz, mineração inovadora em mobile) são tangíveis. O risco crítico é puramente de software: resiliência de dados em disco local e blindagem defensiva (autenticação do Cloud Layer e anti-DDoS). 

A estabilização deve focar unilateralmente em proteger os dados do disco, refatorar a concorrência assíncrona, para só depois escalar a topologia Mesh Global. O código não deve evoluir funcionalidades sem que esta baseline mude para 100% "Pass" nos testes estruturais propostos.
""",

    "TEST_STRATEGY/README.md": """# Estratégia de Testes - Nebula Network

Esta pasta (atualmente vazia de código) hospedará a futura matriz contínua de Integração e Testes (CI/CD) e scripts autônomos.

## Evolução Futura
1. Implementação de Testes Puros em Go (`go test`) para o núcleo criptográfico, Consensus e Ledger.
2. Automação P2P (Testnets em Containêres Docker) orquestrados aqui, criando 5 a 20 nós isolados em rede bridge para observar partições de rede.
3. Integração com o projeto Flutter (`flutter test`) via mocks locais apenas da porta TCP e testes End-to-End nativos automatizados.
4. Testes de Stress de Carga implementados provavelmente em Python, simulando envio de milhares de pacotes `NEW_TRANSACTION`.

A infraestrutura de teste *jamais modificará* o core sem garantia prévia contra as Máquinas de Estados mapeadas em `PROJECT_BASELINE/05_STATE_MACHINES.md`.
"""
}

for filepath, content in files.items():
    with open(filepath, "w", encoding="utf-8") as f:
        f.write(content)
