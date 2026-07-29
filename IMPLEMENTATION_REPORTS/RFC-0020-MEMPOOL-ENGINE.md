# IMPLEMENTATION REPORT: RFC-0020-MEMPOOL-ENGINE (Nebula Network)

## 1. Arquitetura Geral
A Fase 41 introduz o `Mempool Engine`, o repositório em memória responsável por acolher, triar, validar primariamente e gerenciar o ciclo de vida das transações pendentes propagadas na rede antes que o *Minerador* ou o *Ledger* os incluam num bloco válido. 

Cumprindo as diretrizes de desacoplamento rigoroso estipuladas, a Mempool ignora solenemente a verificação contextual (Consenso e UTXO). Ela valida puramente: Assinaturas topológicas (se possui Hash, se Sender existe) e integridade básica (Taxas Mínimas e Tamanho), rejeitando sujeira (Spam) antes de atingir a fila de verificação criptográfica pesada no Ledger.

### Módulos Internos:
- **Mempool**: O orquestrador (`TransactionPool`) responsável pela injeção unificada e por encapsular as outras ferramentas, bem como gerenciar varreduras periódicas de Limpeza (*TTL Sweeper*).
- **Mempool Validator**: Camada síncrona leve. Se a transação infringir qualquer `MempoolPolicy` de entrada (fee zerado, tamanho abusivo, limite temporal excedido), um `OnValidationFailed` será despachado assincronamente.
- **Mempool Cache**: Estrutura `O(1)` suportada por Multi-Index Dictionaries. Hospeda: `txs map[string]entry` para mapeamento de Hashes (Get, Remove), e `bySender map[string]map[string]bool` para agregar subconjuntos por carteira rapidamente. 
- **Mempool Queue**: Sistema assíncrono produtor-consumidor (Fila Padrão Go Channels `chan`). Intercepta submissões públicas e distribui via *N-Workers* (goroutines paralelas) a digestão da transação (validação + indexação), servindo primariamente de colchão contra ataques volumétricos (*DDoS/Backpressure*).
- **Mempool Statistics**: Painel operacional livre de Lock Contention destrutivo, computando médias temporais e transações vivas/caídas sob `sync.RWMutex`.

## 2. Eventos e Notificações (Decoupled Callbacks)
O Motor avisa o ambiente (como a camada de P2P ou Logs) da sua saúde através dos `MempoolEvents`:
- `OnTransactionAdded(hash)`
- `OnTransactionRemoved(hash)`
- `OnTransactionExpired(hash)`
- `OnDuplicateTransaction(hash)`
- `OnPoolOverflow()`
- `OnValidationFailed(hash, reason)`

## 3. Políticas de Sobrevivência (MempoolPolicy)
1. **TTL (Time to Live)**: Se a transação mofar no cache por mais de 24 horas, o limpador paralelo invoca a guilhotina e expele o lixo silenciosamente, mitigando desperdício.
2. **Capacidade & Memory Limits**: A Fila (`Queue`) tem tamanho predefinido. O Repositório dita 50,000 transações de limite. 
3. **RBF (Replace-By-Fee) Simples**: Ao atingir o overflow, caso uma nova transação possua uma taxa de pagamento (*Fee*) superior a da transação mais ineficiente (Pior Pagadora), o motor aplica um *Eviction* local descarregando o estorvo em prol do usuário que está pagando mais, estabilizando as finanças orgânicas da Mempool.

## 4. Estabilidade & Stress Test
O pacote de testes aciona um cenário cáotico de múltiplas vias (1000 sub-rotinas). Todas cravam paralelamente requisições `AddTransaction()` contendo duplicatas, inserções originais e chamadas imediatas simultâneas em `RemoveTransaction()` cruzando Hashes vivos.
Resultado:
- **Zero Panic / Zero Deadlock**. A Queue absolve os picos sem travar o Caller principal.
- **Zero Data Race**. Confirmado analiticamente que as inserções não estripulam o mapa, e os Callbacks ocorrem de forma apartada `go func()` com proteção de concorrência.
- As duplicatas foram perfeitamente detectadas via Indexador sem sobreposições indevidas.

## 5. Limitações Conhecidas
- A Evicção de RBF baseia-se num *Linear Scan* de complexidade *O(N)*, o que significa que sob lotação total (Ex: 50,000 tx), pode custar alguns milissegundos para encontrar a pior transação se submetida à ataques sucessivos de inserção. Solução viável futura: *Min-Heap O(log n)* dedicada ao Fee Index.
- Falta a conexão dessa nova arquitetura base com o motor Gossip / P2P, o que pertencerá a próximas implementações.

## 6. Próximas Expansões
- Ligar a Fila da `Mempool` aos Handlers de `MsgTransaction` originários do `Message Router`.
- Vincular a API RPC à `MempoolStatistics`.
