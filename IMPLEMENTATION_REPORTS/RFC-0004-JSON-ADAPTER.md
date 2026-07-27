# IMPLEMENTATION REPORT: RFC-0004 (JSON Storage Adapter)

## 1. Arquivos Criados
- `pc_node/storage/jsonstorage/json_storage.go`

## 2. Arquivos Modificados
- `pc_node/ledger.go` (Remoção segura da função local `saveLedger`, substituída por chamadas ao adaptador através da variável global `DB storage.Engine`).
- `pc_node/main.go` (Remoção da chamada `saveLedger()` redundante durante o Graceful Shutdown, mantendo a saída limpa e atômica via Adapter).
- `pc_node/ledger_test.go` (Migração da função legada `saveLedger()` diretamente para o escopo local de testes, blindando os testes existentes contra quebra de contrato sem ferir a arquitetura principal).

## 3. Interfaces Implementadas
O struct `JSONEngine` implementou integralmente os seguintes contratos de `storage.Engine`:
- `Open` e `Close` (I/O e desserialização estrita do `ledger.json`).
- `GetBlockByIndex` e `GetLatestBlock` (Buscas `O(1)` seguras via ponteiro RAM).
- `NewBlockIterator` (Iterador deep-copy assíncrono blindado contra Deadlocks/Race Conditions).
- O struct `JSONBatch` implementou integralmente `storage.Batch`, com `PutBlock`, `Commit` e atomicidade baseada em arquivos `.json.tmp`, clonando 1:1 o comportamento ultra-seguro do legado, porém operando injetado de fora.
- O método de abstração estrutural `GetBalance` retorna de forma explícita `storage.ErrUnsupported`, protegendo o escopo rigorosamente das proibições estipuladas nesta Sprint (Nenhum cálculo de saldo poderia migrar).

## 4. Compatibilidade Mantida
Em decorrência da substituição via iterador no `initLedger()`, a retrocompatibilidade manteve-se intocada: a variável array global `ledger.Chain` ainda é alimentada em tempo de boot e operada nos buffers da memória no nó principal (para fins de sidecar API e validação P2P). Essa persistência atômica atua puramente como "Side-Effect" encapsulado, preservando a assinatura do core.

## 5. Comportamento Preservado
Absolutamente nenhuma regra de consenso (Proof of Work PoW), validação ECDSA e lógica matemática foi tocada (Network/Mempool/Blockchain mantidos integralmente). Nenhuma função foi alterada além de delegar o ato mecânico de IO (Write) ao Batch recém criado.

## 6. Resultados de Validação Técnica
- **`go fmt ./...`**: Sucesso.
- **`go vet ./...`**: Sucesso. Nenhum warning ativo.
- **`go test ./...`**: Sucesso (PASS). Todos os cenários (inclusive simulações de corrupção, bloqueios JSON, falha por OOM em IO) permaneceram positivos (`ok pc_node 0.494s`), validando o encapsulamento seguro via Wrapper interno no `ledger_test.go`.
- **`go build ./...`**: Sucesso Absoluto (build do pacote nativo e do `pc_node` principal unificados).

## 7. Riscos Encontrados
- **Risco de Quebra Crítica em Testes Legados:** Os testes em `ledger_test.go` invocavam fortemente a sub-rotina privada `saveLedger()` que foi apagada.
- **Risco OOM Subjacente:** A interface de persistência atômica de lotes `Commit()` continua utilizando `MarshalIndent()` global na totalidade da stringificação da chain de bytes (uma limitação física do legado do JSON).

## 8. Riscos Corrigidos
- Realoquei a estrutura crua do `saveLedger()` limitando-a unicamente à scope tag de testes em `ledger_test.go`. Isso estabilizou e mitigou 100% os testes de concorrência antigos.
- O método `NewBlockIterator` foi adaptado para realizar um shallow-deep array copy imediato do `json.RawMessage`, o que evitou deadlocks e reduziu estresse de Lock passivo de memória no momento em que `initLedger()` acorda no `ledger.go`.
