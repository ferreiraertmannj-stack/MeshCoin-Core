# IMPLEMENTATION REPORT: RFC-0027-STORAGE-ENGINE (Nebula Network)

## 1. Visão Geral (Arquitetura)
A Fase 48 instaura a Persistência Oficial da Blockchain (`Storage Engine`). O motor opera em isolamento extremo: desvenda, armazena e recupera dados criptográficos (Blocos, Transações, UTXOs, Metadados e Estado da Cadeia) em repositórios segmentados. Abstém-se de possuir lógicas de Consenso, Mineração ou Rede, comunicando seus ritos via Padrão Observador (`Observer Pattern`) e extraindo suas dependências por *Dependency Injection*.

## 2. Repositórios e Lojas (Stores)
A arquitetura desdobra as competências de salvamento:
- **BlockStore:** Indexa os Blocos brutos valendo-se simultaneamente das chaves (Hash) e numéricas (Height), garantindo consultas bidirecionais de latência O(1) e pré-aquecendo o cache sob inserções quentes.
- **TransactionStore:** Guarda e restaura o volume das Transações valendo-se de suas assinaturas (Hash), provisionando as requisições ágeis demandadas pelo *Fast Sync* e fôrmas externas.
- **UTXOStore:** O cofre que lastreia os gastos (*Outpoints*). Carrega um mapeamento direto de chaves para que a UTXO Engine (Fase 46) resgate os registros sob demarcações exclusivas Mutex.
- **ChainState & MetadataStore:** Guardiões pontuais do *Best Block*, Dificuldade, *Total Work* e configurações fixas providas na Gênese (Magic Numbers, Versões de Banco e Identificadores).

## 3. Gestão Térmica: Cache LRU e Enfileiramento (Fila)
As barreiras mecânicas foram sobrepostas via algoritmos vitais:
- **StorageCache (LRU O(1)):** Emula o padrão "Least Recently Used". Elementos requisitados escalam para o topo do eviction list. Sob estresse, descarta itens ociosos que excedam o `CacheSize` ou sofram decaimento (TTL Expirado). As consultas (*Hits*) poupam milhares de chamadas aos discos virtuais.
- **StorageQueue (Fila de Backpressure):** Modela os *Writes* através de agrupamentos (`channels`). Submissões simultâneas (Goroutines) encontram os *Workers* estabilizados do Storage, atestando e assegurando a consistência atômica das gravações sem provocar gargalos ou atrito em *busy waiting*.

## 4. Snapshots, Recovery e Compactação
- **StorageSnapshot:** Consolida a captura imutável pontual do momento atual, engatilhando o registro histórico necessário caso se planeje provisionamentos via *Bootstrap*.
- **StorageCompaction:** Emula a compactação subjacente (LevelDB / RocksDB SSTables). Uma interface leve provida para garantir a varredura sem travar a leitura (*Read Lock-Free*), fundindo dados sobrepostos.
- **StorageRecovery:** Ao religar o Node, executa-se o atestado de segurança. Atualmente assegurando a leitura do *ChainState*, garantindo um trilho limpo de retorno frente à desligamentos súbitos. 

## 5. Próximas Expansões
Esta implementação (Mock robusto) prepara o terreno para acoplar motores em disco (*On-Disk*):
- Substituição do armazenamento interno `map` e sub-listas por instâncias ativas de bancos `Key-Value` (ex: LevelDB, RocksDB, Pebble) mantendo inalteradas as Interfaces e as Facades.
- Ampliar o histórico estrito do **Undo-Log**, registrando estritamente as ações regressas, vital para viabilizar reorganizações (`Rollbacks`) plenas.
