import os

out_dir = "RFC/RFC-0002"

files = {
    "01_TRANSACTION_FLOW.md": """# 01 TRANSACTION FLOW

## Mapeamento de Execução (Tx)
1. **Quem chama:** `handleConnection` (network.go) ou `handleWebSocket` (main.go).
2. **Quem é chamado:** `handleNewTransactionPacket` -> `HandleNewTransaction` -> `VerifyTransaction`.
3. **Ordem de execução:** 
   - Parse JSON do payload TCP/WS para struct.
   - `HandleNewTransaction(tx)` é acionado.
   - `VerifyTransaction(tx)` verifica ECDSA e Hash (sem Lock).
   - `ledger.mu.RLock()` acionado.
   - Varredura O(N) no `ledger.Chain` para contabilidade de saldo local (cálculo iterativo).
   - `ledger.mu.RUnlock()` liberado.
   - `mempoolMutex.Lock()` acionado.
   - Abate saldo de Txs pendentes.
   - Rejeita se saldo < Tx.Amount + Tx.Fee.
   - Anexa à `PendingTransactions`.
   - `mempoolMutex.Unlock()` liberado.
   - Retorno booleano em cascata.
4. **Locks:** `ledger.mu.RLock` e `mempoolMutex.Lock`.
5. **Estruturas:** `Transaction`, `PendingTransactions`.
6. **I/O:** Nenhum (puramente em memória).
7. **Rede:** Após sucesso, `broadcastTCP` faz broadcast para vizinhos e WebSockets.
""",
    "02_BLOCK_FLOW.md": """# 02 BLOCK FLOW

## Mapeamento de Execução (Block)
1. **Quem chama:** `handleConnection` (via TCP NEW_BLOCK) ou `handleWebSocket`.
2. **Quem é chamado:** `handleNewBlockPacket` -> `handleNewBlock` -> `VerifyNeonHash` -> `VerifyTransaction`.
3. **Ordem de execução:** 
   - Decode do JSON para struct `Block`.
   - `handleNewBlock` acionado.
   - **Gargalo:** `ledger.mu.Lock()` congela TODO o ledger (Escrita exclusiva).
   - Validação estrutural: Index e PreviousHash.
   - Validação PoW: `VerifyNeonHash` roda matemática com 4KB vector array.
   - Validação Interna: `VerifyTransaction` repetido para cada Tx interna.
   - Acatado: `ledger.Chain = append(ledger.Chain, block)`.
   - Remoção de Txs confirmadas do Mempool via `mempoolMutex.Lock()`.
   - `ledger.mu.Unlock()` liberado via defer!
   - `go saveLedger()` disparado paralelamente (que tentará re-adquirir o `ledger.mu.Lock()`!).
4. **Locks:** `ledger.mu.Lock()` mantido *durante toda validação matemática do bloco*.
5. **Estruturas:** `Block`, `Ledger`, `PendingTransactions`.
6. **I/O:** `go saveLedger()` aciona persistência em disco JSON.
7. **Rede:** Respondido ACK de sucesso ou rejeição no socket.
""",
    "03_NETWORK_FLOW.md": """# 03 NETWORK FLOW

## Mapeamento de Execução (P2P / Sockets)
1. **Componentes:** `listenUDP`, `broadcastPresence`, `listenTCP`, `startSidecarAPI`.
2. **Ordem e Rotinas:**
   - `startNetwork` invoca goroutines separadas para escuta UDP e TCP.
   - UDP (Porta 5555) recebe pings e varre rede local. `broadcastPresence` joga beacon a cada 5s.
   - TCP (Porta 5556) recebe `Accept()`, spawna `handleConnection` por peer.
   - Sidecar API / WS (Porta 8080) atende Flutter app via `handleWebSocket`.
3. **Locks:** 
   - `tcpMutex` protege o map `activeTCPClients`.
4. **Estruturas Compartilhadas:** `activeTCPClients`, `clients` (websocket), `broadcast` chan.
5. **Condições de Corrida:** Socket TCP lendo/escrevendo depende do `tcpMutex`. A latência de rede retém esse lock por tempos imprevisíveis durante o `.Write()` do `broadcastTCP`.
""",
    "04_LOCK_ANALYSIS.md": """# 04 LOCK ANALYSIS

## Regiões Críticas e Locks
1. **`ledger.mu` (RWMutex)**
   - `Lock()`: Em `handleNewBlock` (duração longa: processa hashes e Txs); Em `saveLedger` (duração longa: MarshalIndent da chain toda).
   - `RLock()`: Em `getLedgerJSON` e em `HandleNewTransaction` (varre blocos).
   - **Risco de Contenção:** Muito Alto. A validação matemática (PoW) do bloco é feita *dentro* do Lock. Sincronização do disco em `saveLedger` tenta obter o mesmo Lock gerando engasgos severos na recepção.
2. **`mempoolMutex` (Mutex)**
   - `Lock()`: Em `HandleNewTransaction` (para bater salto e inserir) e `handleNewBlock` (para purgar). Duração curta.
3. **`tcpMutex` (Mutex)**
   - `Lock()`: Em `handleConnection` (add/remove) e `broadcastTCP` (loop de write).
   - **Risco de Starvation:** `broadcastTCP` faz `.Write` em socket de rede DENTRO do lock de mapa, o que é um anti-pattern letal para alta carga de rede, pois um cliente TCP lento travará o envio a todos os outros.
""",
    "05_GLOBAL_STATE.md": """# 05 GLOBAL STATE

## Variáveis Globais Acessadas
- `ledger`: (Struct `Ledger`). Lida e Escrita pelos nós de consenso, protegida por `ledger.mu`.
- `PendingTransactions`: Array em RAM de mempool. Protegida por `mempoolMutex`.
- `ledgerFile`: Path base em string. Modificado para `var` na Sprint 2.0. Lida em I/O.
- `activeTCPClients`: Map de Net.Conn (rede P2P). Protegido por `tcpMutex`.
- `clients`: Map de Net.Conn (WebSockets). Acessado em loop sem Mutex em main.go `handleWebSocket` vs `handleMessages`. **CRÍTICO: Condição de corrida severa de Data Race detectada (main.go linha 114 vs 154).**
- `broadcast`: Chan de envio para WS.
""",
    "06_IO_FLOW.md": """# 06 IO FLOW

## Mapeamento de I/O em Disco
1. **Operação Inicial:** `initLedger()` (leitura do disco para RAM).
2. **Operação Contínua:** `saveLedger()` chamado via `go saveLedger()` ao receber um bloco.
   - Obtém `ledger.mu.Lock()`.
   - Roda `json.MarshalIndent`.
   - Roda `os.CreateTemp`, `Write`, `Sync`, `os.Rename`.
3. **Operação de Cloud:** `uploadToNebulaCloud` (mock no código, envia para cloud externa).
4. **Problema I/O:** `go saveLedger` bloqueia as threads TCP porque disputa o mesmo Mutex com `handleNewBlock`. O dump atinge O(N) do tamanho da Chain.
""",
    "07_CRITICAL_SECTIONS.md": """# 07 CRITICAL SECTIONS

As três seções de maior perigo identificadas:
1. **Loop de Envio TCP (`network.go`)**: O `broadcastTCP` itera sobre `activeTCPClients` segurando o `tcpMutex`. Se um cliente estiver travado (timeout), nenhum outro cliente recebe mensagens.
2. **Loop de Envio WS (`main.go`)**: Map `clients` sendo modificado (`clients[ws] = true`) na thread de conexão sem Mutex, e iterado na thread `handleMessages()`, violando as regras do Go map (Data Race e fatal panic).
3. **Validação de Bloco (`ledger.go`)**: `VerifyNeonHash` roda pesado na CPU (algoritmo Memory-Hard com loop). Atualmente roda dentro de `ledger.mu.Lock()`! O nó inteiro para e não aceita mais transações até a validação criptográfica acabar.
""",
    "08_CALL_GRAPH.md": """# 08 CALL GRAPH

## Fluxo Textual Real (Recepção de Bloco)

```text
[Network Socket TCP] ou [WebSocket]
↓
handleConnection / handleWebSocket
↓
handleNewBlockPacket (faz json decode)
↓
handleNewBlock
|
├── [LOCK: ledger.mu.Lock()]
|
├── VerifyNeonHash (pesado na CPU)
|
├── VerifyTransaction (loop p/ cada tx)
|
├── append(ledger.Chain)
|
├── [LOCK: mempoolMutex.Lock()]
|     └── Purgar mempool
|     └── [UNLOCK]
|
├── [UNLOCK: ledger.mu.Unlock()]
↓
go saveLedger() (Assíncrono)
|
├── [LOCK: ledger.mu.Lock()] (Aguarda liberação)
|
└── I/O Atômico em Disco (Cria, Write, Sync, Rename)
      └── [UNLOCK: ledger.mu.Unlock()]
```
""",
    "09_RFC-0002-DRAFT.md": """# RFC-0002 DRAFT: Lock Optimization & Data Race Fixes

## Problema
O nó PC possui três gargalos extremos e fatais de sincronização:
1. Validação pesada de Proof of Work (NeonHash) rodando dentro do Lock global do Ledger, congelando o sistema inteiro a cada bloco recebido.
2. Anti-pattern de rede onde `.Write()` de socket é executado dentro do `tcpMutex` de mapa.
3. Map de WebSockets não possui Mutex, o que invariavelmente gerará crash (fatal Data Race Map Write) em produção ao conectar/desconectar clientes simultâneos.

## Impacto
Lentidão massiva conforme a rede cresce (travamento de thread). Queda aleatória (Crash) por corrupção de memória (Data Race em Go map).

## Arquivos que deverão ser alterados
- `pc_node/ledger.go`
- `pc_node/network.go`
- `pc_node/main.go`

## Arquivos proibidos de alteração
- Nenhum módulo de lógica central de negócio, hash, blockchain ou cloud.
- O formato do JSON.

## Estratégia
1. Em `ledger.go`: Remover validação `VerifyNeonHash` e checagem indexada `block.Index` para FORA da área do `ledger.mu.Lock()`. Travar o lock apenas no momento de fazer o `append()`.
2. Em `network.go`: Adicionar um buffer de cópia no `tcpMutex`. Clonar os peers, liberar o lock, e só então iterar chamando `.Write()` sem prender o Mutex de mapa global.
3. Em `main.go`: Introduzir `wsMutex sync.Mutex` protegendo a leitura e escrita do `map clients`.

## Riscos
- Possíveis Txs duplicadas na mempool caso não seja cuidadoso a inversão do lock na validação, porém a lógica de anexação atômica prevenirá.
- Crash zero.

## Plano de Rollback
Reverter via Git para a tag base.

## Critérios de Aceite
- Executar o programa rodando `-race` flag sem disparar logs de Data Race.
- Verificar via PProf que o nó não segura a lock de `tcpMutex` para IO de rede.
""",
    "10_EXECUTIVE_SUMMARY.md": """# 10 EXECUTIVE SUMMARY

O mapeamento exaustivo (Sprint 3) revelou que a "Amnésia do JSON" (resolvida na Sprint 1/2) não é o único calcanhar de Aquiles do Node P2P. A arquitetura sofre de severos gargalos de **Sincronização Invertida** (Locking Inversion) e **Data Races Fatais** explícitas em concorrência Go.

Apesar da modularidade abstrata existir nas classes, as estruturas Go subjacentes (Channels, Maps e Mutexes) foram inseridas sem respeitar as primitivas idiomáticas. 
* Funções densas (`VerifyNeonHash`) bloqueiam I/O global.
* Mutexes protegem Rede ao invés de Memória local em `network.go`.
* A ausência imperdoável de Mutex no map de WebSockets em `main.go` criará crashes assim que 2 conexões abrirem ao mesmo tempo no Frontend.

O pacote propõe a `RFC-0002` (Draft) para correção imediata das travas no PC Node. A Fase atual documentou 15+ funções e múltiplos mutexes globais de estado, atestando a urgência refatoratória.
"""
}

for filename, content in files.items():
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)
