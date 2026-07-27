# RFC-0004 REVIEW: Storage Abstraction Architecture

## Revisão Técnica Rigorosa

### 1. A interface Engine possui responsabilidades demais?
Sim. A interface misturou conceitos genéricos de persistência (Key-Value) com regras de apresentação (JSON). 

### 2. Existe alguma responsabilidade que deveria sair do Storage?
O método `DumpChainJSON()`. O mecanismo de Storage jamais deve ter ciência de protocolos de apresentação, como serialização JSON, XML ou gRPC.

### 3. Existe alguma responsabilidade importante que está faltando?
Sim. O suporte a operações atômicas massivas (**Batching**) e **Transações** de banco (ACID). Sem um `WriteBatch`, a inserção de um bloco com milhares de transações e atualizações de balanço engasgará com múltiplos acessos avulsos ao disco. Falta também suporte a cursores (**Iterators**) reais e controle de snapshots.

### 4. GetBalance() pertence realmente ao Storage?
Depende do paradigma. Se o `Engine` for um *Key-Value Store* puro, ele deveria expor apenas `Get(key)`. No entanto, como a interface proposta é um *Domain Storage* (específico do Ledger), `GetBalance()` é aceitável, pois atua como a interface de um **StateDB** encapsulado, escondendo a complexidade dos índices do orquestrador do Ledger.

### 5. DumpChainJSON() viola SRP?
Sim, violência direta ao Princípio da Responsabilidade Única (Single Responsibility Principle - SOLID). A camada de persistência não tem a responsabilidade de formatar dados para a API REST. A API (Sidecar) deve instanciar um `Iterator` do banco e fazer a serialização ela mesma.

### 6. Métodos Adicionais Obrigatórios
Para suportar DBs modernos como Badger/Pebble:
- **Batch:** `NewBatch() StorageBatch` e `CommitBatch(StorageBatch)`
- **Iterator:** Interface `StorageIterator` (Next, Valid, Key, Value)
- **Prefix Scan:** `SeekPrefix(prefix []byte) StorageIterator`
- **Snapshot:** `CreateSnapshot(path string)`
- **Recovery:** `RestoreSnapshot(path string)`

### 7. A interface suporta migração futura sem quebrar compatibilidade?
Da maneira como foi desenhada, suporta a troca do banco em si, mas a ausência de uma injeção de opções no `Open()` (ex: configurações de mmap para Android) limitará a migração. O `connectionString` genérico ajuda, mas precisará de parse semântico.

### 8. Há risco de lock inversion entre Ledger e Storage?
Sim. O método `IterateBlocks(callback)` é perigoso. Se o Storage segurar uma *Read Lock* interna durante a iteração, e a `callback` do Ledger tentar adquirir um `ledger.mu.Lock()` ou invocar uma mutação, pode gerar Lock Inversion e Deadlock imediato.

### 9. Há risco de deadlock?
Exatamente o cenário acima. Iteradores baseados em Callbacks fechadas (`func`) tendem a causar deadlocks acidentais. Iteradores explícitos (`it.Next()`) operados por quem chama são mais seguros arquiteturalmente.

### 10. A interface consegue suportar todos os DBs sem alterações futuras?
Não. Como BadgerDB e RocksDB dependem intensivamente de *Batches* para ter performance, a ausência estrutural de `WriteBatch` faria com que cada atualização de UTXO disparasse I/O síncrono. Isso destruiria o throughput.

### 11. A arquitetura está preparada para Android?
Em teoria sim, mas sem suporte nativo a controle de tamanho de Batch, o Android poderá sofrer *Out of Memory* ao lidar com volumes grandes de transações.

### 12. A arquitetura suporta milhões de blocos?
A interface suportaria, caso o método `DumpChainJSON` seja removido. Se ele for mantido, a API consumirá 100% da RAM para stringificar a chain inteira quando solicitada.

### 13. Violação de Princípios
- **SOLID (SRP):** `DumpChainJSON` viola SRP.
- **Clean Architecture:** A persistência formatando respostas Web contamina os anéis externos.
- **Hexagonal Architecture:** O *Port* (Interface) de Storage está assumindo o papel de um *Adapter* de API.

---

## 14. Versão Final da Interface (Ajustada - Apenas Documentação)

```go
package storage

import "MeshCoin-Core/pc_node/models"

// Engine é o contrato puro do Domínio de Persistência
type Engine interface {
    // Lifecycle e Configuração
    Open(connectionString string) error
    Close() error

    // Operações em Batch (ACID para blocos completos)
    NewBatch() Batch
    
    // Leitura Estrutural O(1)
    GetBlockByIndex(index uint64) (*models.Block, error)
    GetLatestBlock() (*models.Block, error)
    
    // Leitura de Estado O(1)
    GetBalance(address string) (float64, error)
    
    // Iteração Segura (Evita callback deadlocks e elimina DumpJSON)
    NewBlockIterator() Iterator
    
    // Snapshots para Sync e Backup
    CreateSnapshot(path string) error
}

// Batch agrupa as mutações do State e da Chain atômicamente
type Batch interface {
    PutBlock(block models.Block) error
    PutBalance(address string, balance float64) error
    Commit() error
    Discard()
}

// Iterator transfere o controle do loop para a API / Ledger
type Iterator interface {
    Next() bool
    Value() models.Block
    Error() error
    Close()
}
```

---

## 15. Recomendação Final

**STATUS: APROVADA COM AJUSTES**

A necessidade de desacoplamento é premente e inquestionável. No entanto, a versão original (Draft) continha falhas crônicas de design (violação do SRP com JSON e ausência de WriteBatch) que destruiriam a utilidade dos motores LSM de alta performance. 

Com a adoção da **Versão Final da Interface (Seção 14)**, que introduz os conceitos de `Batch` (ACID), `Iterator` (proteção contra deadlocks) e expurga os resquícios de JSON para a API Rest, a abstração está arquiteturalmente pronta para ser codificada na Fase seguinte.
