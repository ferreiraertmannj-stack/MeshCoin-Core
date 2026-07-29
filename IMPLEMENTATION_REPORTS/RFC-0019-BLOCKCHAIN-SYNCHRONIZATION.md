# IMPLEMENTATION REPORT: RFC-0019-BLOCKCHAIN-SYNCHRONIZATION (Nebula Network)

## 1. Arquitetura
A Fase 40 introduz o `BlockchainSyncManager`, consolidando a sincronização distribuída sem ferir o encapsulamento ou manipular dados do *Consenso* (Ledger). A arquitetura é orquestrada em *pc_node/p2p* através de composição pura e Injeção de Dependências, conversando com abstrações externas (`ChainProvider`) para determinar lacunas, e provendo callbacks desacoplados para injeção nativa via `BlockchainSyncEvents`.

O motor atua primariamente na estratégia **Header-First Synchronization**, desdobrando-se em:
- **Chain Locator**: Responsável por ancorar um *Checkpoint* iterativo que calcula a interseção exata (fork point comum) entre a cadeia local e a de um peer remoto, aplicando distanciamento exponencial.
- **Fork Detector**: Valida silenciosamente (sem reverter dados) bifurcações locais mapeando transições marginais.
- **Block Request Scheduler**: Máquina de estado (Fila) responsável por designar downloads (*Pulls*) a peers específicos, efetuando *Backoff*, controle ativo de timeouts e redistribuição inteligente.
- **Pending Blocks Manager**: Depósito transiente (*In-Memory*) segurando blocos prematuros/órfãos isoladamente do Ledger enquanto eles aguardam progenitores (pais) na cadeia.

## 2. Fluxos e Estratégia Header-First

O procedimento evita o desperdício tráfego da rede, recusando blocos desnecessários através do seguinte pipeline:

```text
1. [Nó A] envia MsgGetHeaders contendo LocatorHashes exponenciais.
2. [Nó B] mapeia o ponto comum através de FindCommonBlock(LocatorHashes).
3. [Nó B] envia MsgHeaders com até X headers subsequentes.
4. [Nó A] invoca OnHeadersReceived, dispara ForkDetector, e identifica Headers não-conhecidos (desconsiderando pendentes).
5. [Nó A] enfileira hashes faltantes no BlockRequestScheduler.
6. [Scheduler] despacha requests (MsgGetData) paralelos para diversos peers conectados de forma assíncrona.
7. [Nó A] invoca HandleBlockReceived que analisa a arvore. Se houver parentesco no Ledger (ChainProvider): Importa (OnBlockImported) imediatamente, resolvendo cascata de Pending Blocks. Do contrário, guarda no PendingBlocksManager e dá Request no Pai faltante.
```

## 3. Componentes Específicos

### 3.1. Fork Detection
O `ForkDetector` cruza as alturas fornecidas e alerta o `OnForkDetected` caso um pai declarado exista fisicamente no nó, porém desancorado da sub-árvore ativa (altura menor que a do *Tip* atual). 

### 3.2. Pending Blocks Manager
Um simples array de chaves trancaria a CPU. Utilizamos um *Multi-Index HashMap* em Memória:
- `orphans map[string]PendingBlock`: para acesso *Fast Lookup* via Hash.
- `waitingForParent map[string][]PendingBlock`: para resolver dependências transversais na ordem O(1).
Quando um bloco X finalmente é incorporado, verificamos instantaneamente em `waitingForParent[X]` todos os filhos orfãos, reimportando-os em cascata assíncrona (Depth-First).

### 3.3. Scheduler e Timeout Engine
- A matriz de requisição (`pendingRequests`) mapeia o par (Blockhash -> Assinatura do Peer + Timestamp de Envio).
- Um _worker_ roda um `loop()` validando a matriz. Caso `now - SentAt > Timeout`, a requisição recai pro estado Não Assinalado, incrementando *Retries*. Se exceder, levanta erro global.

## 4. Estatísticas Thread-Safe
Em compasso com os relatórios das Fases P2P, a struct `BlockchainSyncStatistics` mede (através de `sync.RWMutex`) as chaves: 
`HeadersReceived`, `HeadersValidated`, `BlocksRequested`, `BlocksReceived`, `BlocksImported`, `PendingBlocks`, `OrphanBlocks`, `ForksDetected`, e `Reorganizations`. 
Permite um painel administrativo exposto e isolado, evitando que threads do *RPC Server* conflitem com ponteiros vivos do Sincronizador de rede.

## 5. Prevenções e Concorrência Segura
O código passou no crivo irrestrito do compilador do Go (`go test -race`):
- O Handler principal de importação (`HandleBlockReceived`) foi selado através do Mutex Primário do `SyncManager`.
- Callbacks disparados para o mundo externo (`sm.events.OnBlockReceived`) foram mantidos paralelos via `go`, mas chamadas imperativas sequenciais à inserção da Blockchain (como `OnBlockImported`) agora operam de forma controladamente síncrona dentro da tranca do Ledger, proibindo o aparecimento de novos orfãos "fantasmas" devidos à dessincronização de threads.
- Teste desenhado com sobrecarga (*Stress Test* com 1000 goroutines) bombardeou blocos desordenados para testar bloqueio serial do *PendingBlocksManager*. (Taxa de acerto 100% – todos importados).

## 6. Limitações Conhecidas
- Se um bloco malicioso for injetado sem que seu Header tenha sido requisitado, e o atacante inventar milhares de pais fantasmas, o nó pode encher a RAM (DDoS em PendingBlocks). Atualmente o contador é estritamente medido em `.Count()`. Mas um `Cleanup` cronológico precisará ser introduzido no `PendingBlocksManager`.
- O Locator exponencial não escala em complexidade se uma bifurcação (Reorg) imensa com mais de 500 hashes acontecer, pois pularia um intervalo grande.

## 7. Próximas Expansões
- Adicionar o expurgo por TTL em `PendingBlocks`.
- Interligar o `OnBlockImported` e o `OnHeadersValidated` à camada real do Consenso/Storage Engine construída nas fases iniciais.
