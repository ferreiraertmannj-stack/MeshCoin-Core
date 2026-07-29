import os

out_dir = "MODULE_ANALYSIS/LEDGER"

files = {
    "01_OVERVIEW.md": """# 01 OVERVIEW

## Ledger Module (`pc_node/ledger.go`)

### Tipos e Estruturas
1. **Transaction**: Representa uma transferência na rede. Possui chaves, assinaturas (ECDSA `secp256k1` e stub PQC `Dilithium`), timestamps, taxas e valores.
2. **Block**: Container indexado de transações, carimbado temporalmente, possuindo referência ao bloco anterior (`PreviousHash`), raiz de Merkle (não implementado o preenchimento, apenas campo existente), Nonce, dados da Cloud (`MinerStorage`, `StorageType`) e seu próprio Hash final (NeonHash).
3. **Ledger**: Agrupador principal contendo a cadeia `Chain []Block` e um bloqueio global de Leitura/Escrita `sync.RWMutex`.

### Variáveis Globais
- `ledger`: Instância singleton persistindo o estado primário.
- `PendingTransactions`: Mempool array em RAM.
- `mempoolMutex`: Lock isolado para acesso à Mempool.
- `ledgerFile`: Path constante `ledger.json`.

### Dependências (Imports)
- Nativos Go: `crypto/sha256`, `encoding/hex`, `encoding/json`, `fmt`, `io/ioutil`, `log`, `strings`, `sync`.
- Externos (Consensus/Wallet): `github.com/decred/dcrd/dcrec/secp256k1/v4` (ECDSA parser e verifier).
""",

    "02_DATA_FLOW.md": """# 02 DATA FLOW

## Recebimento de Transação
1. **Origem:** Cliente (via `TCP NEW_TRANSACTION`).
2. **Entrada:** `HandleNewTransaction(tx)`.
3. **Validação Cryptográfica:** `VerifyTransaction()` recalcula hash da string Dart e verifica ECDSA.
4. **Validação Contábil:** Varre O(N) do `ledger` em Lock Read, e O(M) da Mempool em Lock Mutex para somar/subtrair balance.
5. **Atualização de Estado (RAM):** `PendingTransactions = append(...)`.

## Recebimento de Bloco
1. **Origem:** Rede / Mineiro Local (`TCP NEW_BLOCK`).
2. **Entrada:** `handleNewBlock(block)`.
3. **Bloqueio:** `ledger.mu.Lock()` congela o sistema.
4. **Validação Estrutural:** Index > LastIndex; PreviousHash == LastHash.
5. **Validação PoW:** `VerifyNeonHash()` recalcula a prova O(1).
6. **Validação de Txs Internas:** Re-verifica signatures.
7. **Atualização de Estado (RAM):** `ledger.Chain = append(...)`.
8. **Limpeza:** Remove Txs mineradas do `PendingTransactions`.
9. **Hooks:** Dispara Goroutine para Cloud (a cada 10 blocos).
10. **Persistência:** Dispara Goroutine para `saveLedger()`.
""",

    "03_STATE_MACHINE.md": """# 03 STATE MACHINE

## Máquina de Estado do Bloco Mestre (Ledger)

**Estado Inicial: DISCONNECTED/UNINITIALIZED**
- Dispara `initLedger()`.
- Lê arquivo. Se Válido -> **Estado: SYNCED (RAM)**.
- Se Inválido/Inexistente -> Subscreve Gênesis -> `saveLedger()` -> **Estado: SYNCED (RAM)**.

**Estado: SYNCED (RAM)**
- *Evento:* Recebe `NEW_BLOCK`
- *Transição:* Vai para **VALIDATING_BLOCK**.

**Estado: VALIDATING_BLOCK**
- *Locks Ativos:* `ledger.mu.Lock()`
- Se Sucesso: Adiciona ao Array, limpa Mempool, chama IO Assíncrono -> Volta para **SYNCED**.
- Se Falha (Fork, Bad PoW, Bad Tx): Desbloqueia, descarta bloco -> Volta para **SYNCED**.
""",

    "04_CALL_GRAPH.md": """# 04 CALL GRAPH

## Grafos Textuais de Chamada

**Fluxo de Injeção de Bloco:**
```text
(Rede/TCP) -> handleNewBlock()
                 │
                 ├─> VerifyNeonHash()
                 │       └─> calculateHash()
                 │
                 ├─> VerifyTransaction() [Para cada Tx]
                 │       ├─> formatDartDouble()
                 │       └─> secp256k1.ParsePubKey()
                 │       └─> ecdsa.Verify()
                 │
                 ├─> (Mempool Cleanup)
                 │
                 ├─> go uploadToNebulaCloud() [A cada 10]
                 │
                 └─> go saveLedger() [I/O Assíncrono]
```

**Dependências Entrantes:**
- `pc_node/network.go` -> Chama `handleNewBlock`, `HandleNewTransaction` e `getLedgerJSON`.
- `pc_node/node_daemon.go` (Cloud) -> Lê blocos para upload.
""",

    "05_DISK_IO.md": """# 05 DISK I/O

### Leitura Inicial (`initLedger()`)
- Síncrona. 
- Carrega o arquivo `ledger.json` inteiro usando `ioutil.ReadFile()`.
- OOM (Out Of Memory) Risk imediato na leitura de chains gigabytes.

### Salvamento Diário/Continuo (`saveLedger()`)
- Síncrono (dentro da Goroutine spawnada pelo `handleNewBlock`).
- Possui Lock estrito em `ledger.mu.Lock()`, serializa a cadeia *inteira* com `json.MarshalIndent`.
- Escreve *por cima* do arquivo com `ioutil.WriteFile()`. O arquivo tem bytes = tamanho total da chain atual.
- Operação extremamente ineficiente com tempo Big-O de *O(N)* a cada bloco minerado.

### Resposta de Leitura Externa (`getLedgerJSON()`)
- Tenta ler direto do disco, gerando contenção de I/O na máquina física.
- Se falhar, faz Marshal global bloqueando o RWMutex.
""",

    "06_LOCKING.md": """# 06 LOCKING

### Mutexes Ativos
1. **`ledger.mu` (sync.RWMutex)**:
   - Bloqueia *toda* a escrita do Ledger.
   - Chamado em: `saveLedger()`, `handleNewBlock()`.
   - Read-lock (`RLock`) chamado em `HandleNewTransaction()` e `getLedgerJSON()`.
   
2. **`mempoolMutex` (sync.Mutex)**:
   - Protege apenas o array global `PendingTransactions`.
   - Chamado em: `handleNewBlock()` (durante a limpeza), e `HandleNewTransaction()`.

### Riscos de Deadlock e Concorrência
- `saveLedger` (chamado via goroutine) dá Lock. Se ocorrerem múltiplos `NEW_BLOCK` num burst da rede, a fila do Mutex do Go forçará a serialização dessas chamadas, paralisando a rede (TCP handler esperando `handleNewBlock`).
""",

    "07_ERROR_HANDLING.md": """# 07 ERROR HANDLING

### Falhas de Parsing / Corrupção
- O `initLedger()` intercepta qualquer erro de Unmarshal (seja corrupção, seja arquivo em branco) e aplica o bypass: sobrescreve com um array contendo apenas o Bloco 0 (Gênesis). É uma estratégia destrutiva de fail-open onde falhar custa toda a base de dados em vez de um `panic`.

### Falhas de Assinatura
- `VerifyTransaction()` faz um bail-out seguro em chaves ruins ou strings menores que 128 chars. Apenas descarta a Tx. Não penaliza (ban) a conexão TCP chamadora porque essa função é pura, isolada da rede.

### Falhas de Hash
- `VerifyNeonHash()` também pura. Falha silenciosamente logando o erro sem quebrar o laço de execução principal.
""",

    "08_ATOMICITY.md": """# 08 ATOMICITY

### Escrita em Disco
- **Não existe atomicidade real**.
- O Go `ioutil.WriteFile()` abre, trunca (O_TRUNC) e escreve. Se a energia cair entre o `Open()` e o fim da gravação, o `ledger.json` restará corrompido (tamanho 0 ou parcial).

### Adição na Memória
- **Garantida pelo Mutex.** O array em Go nunca sofrerá *data race* graças ao `ledger.mu`. Portanto, transições *em RAM* são atômicas. Mas, se a RAM corromper ou for perdida, o disco não sustenta as ACID properties devido ao I/O ser inseguro.
""",

    "09_RECOVERY.md": """# 09 RECOVERY

### Cold Boot
- O Nó inicia, lê JSON e confia inteiramente no conteúdo local. Não há mecanismo formal atual implementado de "Verificar a integridade hash por hash do Bloco 0 ao Bloco N na subida". Se um atacante alterar o JSON localmente na máquina do minerador ajustando saldos de blocos passados, o nó carregará normalmente.

### Chain Fork Resolution
- Totalmente ausente na arquitetura interna do Ledger atual. O bloco é apenas aceito se `block.PreviousHash == lastBlock.Hash`. Se dois blocos válidos da mesma altura chegarem em instantes diferentes, o primeiro ganha. Não há armazenamento lateral (Orphan Blocks/Forks) nem função de Reorg (Retroceder X blocos e adotar uma corrente mais longa).
""",

    "10_IMPROVEMENT_OPPORTUNITIES.md": """# 10 IMPROVEMENT OPPORTUNITIES

### 1. Database ACID
- **Descrição:** Substituir Array em Memória e ioutil.WriteFile por key-value store (ex: LevelDB).
- **Motivação:** Garantir persistência sem data-loss; remover serialização JSON repetitiva inteira de O(N); viabilizar índices rápidos sem travar RAM.
- **Riscos:** Necessidade de conversão de dados pré-existentes.
- **Arquivos:** `ledger.go`
- **Complexidade:** Alta
- **Impacto esperado:** Escalabilidade Infinita de I/O, Redução de RAM para MB.
- **Prioridade:** 1 (Blocker)

### 2. Lock-free State Manager (Actor Model)
- **Descrição:** Remover `sync.Mutex` e gerenciar a Chain através de canais (Channels/Select).
- **Motivação:** Mutex retém requests TCP globais, causando timeout e partição da rede mesh sob estresse.
- **Riscos:** Eventuais data races em consultas síncronas requerem cuidados ao desenhar RPC interno.
- **Arquivos:** `ledger.go`, `network.go`
- **Complexidade:** Alta
- **Impacto esperado:** High Throughput, Não bloqueia a porta 5556 sob ataque ou carga.
- **Prioridade:** 2 (High)

### 3. Orphan Block & Fork Resolution
- **Descrição:** Modificar `handleNewBlock` para não rejeitar sumariamente blocos com `Index` ou `PreviousHash` alternativos. Guardar na memória e aplicar regra de Longest Chain.
- **Motivação:** A ausência de Fork Resolution dividirá fatalmente a rede na primeira discordância entre nós fisicamente distantes, inutilizando a moeda na WAN.
- **Riscos:** Vetor para spam se limite de ram-orphan não for fixado.
- **Arquivos:** `ledger.go`
- **Complexidade:** Muito Alta
- **Impacto esperado:** Resiliência de consenso P2P de nível profissional.
- **Prioridade:** 3 (High)

### 4. Cache UTXO / Account Balances
- **Descrição:** Em vez de fazer O(N) da Gênesis até Block[Last] para calcular saldo da Tx, manter um mapa em disco de `address -> balance`.
- **Motivação:** Envio de transação em redes avançadas levará horas apenas para verificar saldos se O(N) se mantiver.
- **Riscos:** Possível descompasso cache-storage se não atualizado atomicamente junto com bloco.
- **Arquivos:** `ledger.go`
- **Complexidade:** Média
- **Impacto esperado:** Verificação instantânea de transação O(1).
- **Prioridade:** 4 (Medium)

### 5. Safe Init (Truncate Protection)
- **Descrição:** Mudar `saveLedger` para escrever em `ledger.json.tmp` e renomear atomicamente (`os.Rename`), preservando a leitura.
- **Motivação:** Solução de low-hanging fruit que impede que um shutdown destrua 100% dos dados.
- **Riscos:** Nenhum.
- **Arquivos:** `ledger.go`
- **Complexidade:** Baixa
- **Impacto esperado:** Sobrevivência a quedas de luz.
- **Prioridade:** 1 (Quick Win)
"""
}

for path, content in files.items():
    with open(os.path.join(out_dir, path), "w", encoding="utf-8") as f:
        f.write(content)
