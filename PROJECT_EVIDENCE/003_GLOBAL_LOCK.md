# Gargalo de Mutex Global (Global Lock)

## Resumo
Tanto o Ledger (chain completa) quanto a Mempool utilizam Mutexes globais únicos em Go, o que obriga a execução síncrona de blocos e transações em um ambiente que deveria ser massivamente concorrente.

## Severidade
Alto

## Arquivos envolvidos
- `pc_node/ledger.go`
- `pc_node/network.go`

## Funções envolvidas
- `handleNewBlock()`
- `HandleNewTransaction()`

## Fluxo de execução
Network Listener TCP
↓
Goroutine `handleConnection`
↓
`handleNewBlock()`
↓
`ledger.mu.Lock()` (Bloqueia todas as outras conexões)

## Evidência encontrada
Em `ledger.go`, há a definição `var ledger = Ledger{Chain: []Block{}, mu: sync.RWMutex{}}`. Em `handleNewBlock`, a primeira instrução é `ledger.mu.Lock()`, que só dá release via defer. O mesmo ocorre com `mempoolMutex`. 

## Como reproduzir
Submeter milhares de blocos e transações via requisições TCP paralelas para a porta 5556. A vazão (throughput) cairá drasticamente pois todas as goroutines entrarão em contenção pelos mesmos locks de maneira sequencial.

## Impacto
Impossibilidade de escalabilidade vertical no nó.

## Consequências
Com alta latência devido à contenção de lock, a rede pode droppar blocos legítimos por timeout, causando divisões (forks).

## Módulos afetados
- pc_node (Consensus)

## Protocolos afetados
- CORE_PROTOCOL.md

## Invariantes afetados
- SYSTEM_INVARIANTS.md (Escalabilidade e Concorrência)

## Causa provável
Design prematuro e ingênuo das estruturas de dados. A modelagem priorizou consistência forte na memória sem considerar bancos de dados concorrentes nativos ou channel architecture (padrão Go).

## Possíveis soluções
1. Substituir a estrutura JSON em memória por LevelDB.
2. Passar o gerenciamento de estado para uma goroutine singular consumidora de canais (Actor Model).
3. Adotar fine-grained locking, bloqueando apenas as partes da chain que estão sendo modificadas.

## Complexidade estimada
Alta

## Risco da correção
Alto

## Ordem recomendada
3
