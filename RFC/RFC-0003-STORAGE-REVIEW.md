# RFC-0003 STORAGE REVIEW: Decision Architecture

Este documento apresenta uma revisão técnica detalhada sobre as opções de Storage Engine (KV Database) para substituir a persistência JSON nativa do MeshCoin Core.

## 1. Análise Comparativa das Opções (18 Pontos de Avaliação)

As quatro engines puramente Go (ou suportadas via Go sem necessidade pesada de CGO) analisadas são: **LevelDB (syndtr/goleveldb)**, **BadgerDB (dgraph-io/badger)**, **PebbleDB (cockroachdb/pebble)** e **BoltDB (etcd-io/bbolt)**.

| Requisito Técnico | LevelDB (goleveldb) | BadgerDB | PebbleDB | BoltDB (bbolt) |
| :--- | :--- | :--- | :--- | :--- |
| **1. Perf. Escrita** | Média/Alta (LSM) | Altíssima (WiscKey) | Altíssima (RocksDB-like) | Baixa (B+Tree, page syncs) |
| **2. Perf. Leitura** | Alta | Altíssima | Altíssima | Altíssima |
| **3. Uso de RAM** | Baixo | Alto (exige tuning) | Médio/Alto | Muito Baixo (mmap) |
| **4. Segurança (Corrupção)**| Média (alguns bugs edge) | Altíssima (Value logs) | Altíssima | Absoluta (ACID estrito) |
| **5. Recuperação Pós-Queda**| Baseado no WAL | Resiliente (Value log replay) | Resiliente (WAL estrito) | Imediata (Copy-on-write) |
| **6. Backup** | Backup offline fácil | Online backup nativo | Online via snapshots | Simples cópia de arquivo único |
| **7. Comp. Windows** | 100% nativa | 100% nativa | 100% nativa | 100% nativa |
| **8. Comp. Linux** | 100% nativa | 100% nativa | 100% nativa | 100% nativa |
| **9. Comp. Android** | Suportada (usada em light nodes) | Suportada (exige atenção à mmap RAM) | Suportada | Suportada |
| **10. Dependências Externas** | Nenhuma (pure Go) | Nenhuma (pure Go) | Nenhuma (pure Go) | Nenhuma (pure Go) |
| **11. Compl. Integração** | Muito Baixa (simples) | Média (exige Garbage Collection manual) | Média/Alta (tuning complexo) | Baixa (API limpa, transacional) |
| **12. Compl. Migração** | Baixa | Média | Média | Baixa |
| **13. Throughput Blockchain**| Adequado para early-stage | Excelente (separa Keys de Values) | Excelente | Lento (gargalo no commit do bloco) |
| **14. Escalabilidade (Mi)** | Engasga na compactação | Projetada para isso (TB+) | Projetada para isso (TB+) | Arquivo único cresce muito rápido |
| **15. Facilidade de Índices** | Prefixos simples | Suporta namespaces fáceis | Prefixos simples | Buckets aninhados (nativo) |
| **16. Suporte da Comunidade**| Antigo, mas estável | Muito ativo (Dgraph) | Muito ativo (Cockroach) | Estável/Congelado (etcd) |
| **17. Manutenção Futura** | Baixa | Exige gerenciamento de logs | Exige tuning | Zero manutenção |
| **18. Comp. Funcionalidades**| Sidecar/Explorer Ok | Ideal p/ Smart Contracts e RPC rápida | Excelente p/ Explorer | Sidecar Ok |

---

## 2. Matriz de Decisão e Notas (0 - 10)

| Solução | Nota | Ponto Forte | Ponto Fraco Fatal |
| :--- | :---: | :--- | :--- |
| **LevelDB (Go)** | **7.5** | Simplicidade e baixo consumo de recursos. | Compactações longas travam leitura temporariamente; projeto meio estagnado. |
| **PebbleDB** | **8.5** | Sucessor moderno do Level/RocksDB em Go. | Curva de aprendizado de tuning é desproporcional para o estágio atual do MeshCoin. |
| **BoltDB** | **6.0** | Segurança transacional ACID impecável (zero corrupção). | Performance de escrita (append) de blocos despenca drasticamente com o aumento da rede. |
| **BadgerDB** | **9.5** | Separação de Keys (Index) e Values (Blocos) - o WiscKey Paper. | Requer um daemon de Garbage Collection interno rodando no código Go (fácil resolução). |

---

## 3. Recomendação Final

**A solução escolhida para o MeshCoin Core é o BadgerDB (`github.com/dgraph-io/badger/v4`).**

### 3.1 Justificativa Técnica
Em uma blockchain, blocos contêm payloads muito grandes (lista de transações), mas a busca de saldos (UTXOs/Balances) se baseia puramente nos hashes e metadados. O **BadgerDB** foi arquitetado exatamente sobre o *WiscKey Paper*, onde as chaves (Keys) e valores pequenos ficam na LSM Tree (na RAM), enquanto valores grandes (Blocos serializados em bytes) são guardados em *Value Logs* não-compactados no disco.
Isso significa que a `VerifyTransaction` poderá varrer a LSM Tree procurando chaves (índices de saldo) em microssegundos, sem nunca carregar os blocos gigantes para a RAM. Isso resolve instantaneamente o gargalo mapeado pela RFC-0003. Além disso, sendo Pure Go, preserva totalmente a portabilidade Windows/Linux/Android do MeshCoin.

### 3.2 Motivos do Descarte das Demais
- **BoltDB:** Utiliza uma B+Tree. Ao inserir novos blocos sequenciais (append-only), há um rebalanceamento constante da árvore, destruindo a velocidade de gravação (TPS). Transações bloqueiam escritores concorrentes globalmente.
- **LevelDB:** Embora Geth (Ethereum) tenha usado por muito tempo, ele armazena Keys e Values juntos na LSM Tree. Blocos grandes provocam *Write Amplification* excessivo e travam a engine durante a compactação (L0 -> L1).
- **PebbleDB:** Apesar de ser uma obra de arte da engenharia, é super-otimizado para o CockroachDB. Sua API e tuning de cache seriam overkill complexo demais para uma equipe que está saindo de um arquivo `ledger.json`.

---

## 4. Riscos Mapeados (BadgerDB)

1. **Memória Mapeada (mmap) vs Android:** BadgerDB por padrão consome bastante RAM por tentar fazer cache de índices. Em dispositivos Android futuros com restrição severa de recursos, poderá causar crashes se não configurarmos as `badger.Options` com `ValueLogLoadingMode: FileIO` ao invés de `MemoryMap`.
2. **Value Log Garbage Collection:** Como blocos não sofrem "delete", a blockchain é append-only. No entanto, índices auxiliares de mempool ou saldos temporários que são atualizados deixarão rastros (*stale data*). Será necessário implementar uma goroutine periódica (`db.RunValueLogGC()`) nativa para expurgar espaço morto, senão o disco encherá desnecessariamente.

---

## 5. Plano Futuro (Implementação)

1. **Aprovação da RFC-0003 e Review:** Confirmação pela equipe Core do BadgerDB como Storage Engine Oficial.
2. **Camada de Abstração:** Criação de interface Go (`LedgerStorage`) isolando chamadas `.Get`, `.Set`, `.Iterate`.
3. **Instalação da Dependência:** `go get github.com/dgraph-io/badger/v4`.
4. **Tooling de Migração:** Implementação do `migrate_db.go` para ler o `ledger.json` e encadear os `.Set()` no Badger.
5. **Integração:** Substituição nativa no nó PC (`pc_node/ledger.go`).
