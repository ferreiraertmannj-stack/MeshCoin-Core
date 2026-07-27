# RFC-0003: Storage Engine Migration (JSON to KV Database)

## 1. Problema
Atualmente, o nó PC do Nebula Network armazena todo o histórico da blockchain em um único arquivo de texto massivo (`ledger.json`). Toda vez que um novo bloco é adicionado, a totalidade da cadeia é serializada via `json.MarshalIndent` e reescrita no disco. Além disso, consultas de saldo (`HandleNewTransaction`) exigem varredura completa O(N) de todos os blocos e transações na RAM. Conforme a rede cresce, esse modelo garantirá um Out Of Memory (OOM) e travamento de I/O de disco intolerável.

## 2. Evidências
O mapeamento aponta que o formato atual impacta as seguintes funções:
- **`saveLedger()` (ledger.go):** Roda O(N) para transformar toda a `ledger.Chain` em um slice de bytes.
- **`initLedger()` (ledger.go):** Carrega o arquivo inteiro para a RAM durante o boot.
- **`HandleNewTransaction(tx Transaction)` (ledger.go):** Itera todos os blocos `for _, b := range ledger.Chain` e todas as transações `for _, t := range b.Transactions` para auditar saldos.
- **`getLedgerJSON()` (ledger.go):** Acionado por `main.go` no endpoint `/api/ledger`. Retorna a chain inteira na RAM para a rede.

## 3. Impacto
- **I/O Saturado:** Escrever gigabytes de dados a cada ~10 minutos para anexar poucos kilobytes de um bloco novo.
- **RAM Exaurida:** O sistema quebra se o `ledger.json` for maior do que a RAM física da máquina.
- **Consumo de CPU:** Serialização JSON de grandes estruturas trava o Garbage Collector do Go.

## 4. Escopo
- Alterar o Storage Engine nativo de serialização JSON para um banco de dados ACID Key-Value (ex: LevelDB ou BadgerDB).
- Substituir o array em RAM `ledger.Chain` por paginação e busca via DB iterators.
- Módulos restritos a essa mudança: `ledger.go`, inicialização em `main.go` e API REST.

## 5. Arquitetura Proposta
- Adoção de um DB embutido (KV Store).
- A variável global `ledger.Chain []Block` deixará de existir como array contínuo em RAM e passará a ser uma abstração de acesso ao DB.
- Criação de prefixos de banco:
  - `b_idx_<uint64>` -> Bytes do Bloco
  - `b_hash_<string>` -> Index do Bloco
  - `tx_<string>` -> Metadata da Transação
  - `bal_<address>` -> Saldo consolidado do endereço
  - `meta_latest_block` -> Index do bloco mais alto

## 6. Fluxo Atual vs Fluxo Futuro

### Fluxo Atual
1. Nó inicia.
2. `ioutil.ReadFile("ledger.json")` carrega gigabytes.
3. RAM é consumida em 100% da chain.
4. Ao receber um bloco, `.Lock()` -> `append` -> `.Unlock()` -> re-escreve `ledger.json` inteiro.
5. Verificação de Txs percorre todo o array na memória.

### Fluxo Futuro
1. Nó inicia.
2. Abre conexão KV (`db.Open()`). Carrega apenas `meta_latest_block` (O(1)).
3. RAM consumida apenas para os blocos da ponta (cache).
4. Ao receber bloco, escreve no DB de forma atômica `b_idx_<index>` (Tempo O(1), I/O mínimo).
5. Verificação de Txs usa consulta direta no índice `bal_<address>` (O(1)).

## 7. Análise de Otimização e Índices (Perguntas da Sprint)

* **Todas as chamadas que dependem do JSON atual:** `initLedger`, `saveLedger`, `getLedgerJSON`, Sidecar API `/api/ledger`.
* **Quem escreve:** `saveLedger()` e, historicamente, a API de Mock em scripts Python.
* **Quem lê:** `initLedger()`, `HandleNewTransaction()` (lê a RAM carregada pelo JSON).
* **Quem depende do formato atual:** O frontend Flutter e a ferramenta `explorador_mesh.py` dependem da saída de `getLedgerJSON()`.
* **Quais estruturas persistidas em KV:** `Block`, `Transaction`.
* **Quais índices necessários:** Tabela de blocos por Index (ordem cronológica), blocos por Hash (resolução P2P), saldos por Endereço (validação tx O(1)).
* **Quais operações passarão de O(N) para O(log N) / O(1):** `HandleNewTransaction` deixará de varrer todos os blocos; a gravação de blocos deixará de regravar a história inteira.

## 8. Estratégia de Migração (Ferramenta Automática)
Para manter compatibilidade:
1. Criar um script CLI separado em `tools/migrate_db.go`.
2. O script lerá o `ledger.json` existente e importará bloco por bloco para a nova estrutura KV.
3. Criará um flag de inicialização em `main.go`. Se existir `ledger.json` e o DB não existir, o próprio daemon invocará a migração transparente antes de abrir as portas TCP. Após sucesso, renomeia `ledger.json` para `ledger.json.backup`.

## 9. Riscos
- Quebra de compatibilidade com scripts Python que lêem o `ledger.json` de fora usando I/O de disco direto ao invés de usar a API REST (ex: `explorador_mesh.py`).
- Corrupção durante a migração inicial de nós muito grandes.

## 10. Rollback
Manter o arquivo `ledger.json.backup` intocado após migração. Caso falhe, apagar a pasta do LevelDB/Badger e remover o sufixo `.backup` para restaurar o motor em RAM.

## 11. Critérios de Aceite
- API `/api/ledger` continuar retornando a mesma estrutura JSON para o Flutter (usando iteradores do DB).
- Nós migrados inicializarem em milissegundos em vez de minutos.
- Gravação de novos blocos não excederem pico de RAM local.

## 12. Roadmap de Execução (Próximas Fases)
1. **RFC-0003-1:** Inserir e testar o Wrapper DB/KV (emulação em memória).
2. **RFC-0003-2:** Modificar `HandleNewTransaction` para usar índices de Balanço.
3. **RFC-0003-3:** Criar e testar `migrate_db.go`.
4. **RFC-0003-4:** Ligar o armazenamento persistente (LevelDB) no nó.
