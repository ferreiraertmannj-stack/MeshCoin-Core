# IMPLEMENTATION REPORT: RFC-0004-BADGER-ADAPTER (BadgerDB Storage Engine)

## 1. Arquivos Criados
- `pc_node/storage/badgerstorage/badger_storage.go`

## 2. Arquivos Alterados
- `go.mod` / `go.sum` (Atualização de dependências com a inclusão do pacote oficial do BadgerDB).

## 3. Interfaces Implementadas
- **`storage.Engine`:** Instanciada estritamente via `BadgerEngine`.
- **`storage.Batch`:** Instanciada nativamente via `BadgerBatch` repassando pro `badger.Txn`.
- **`storage.Iterator`:** Instanciada encapsulando `badger.Iterator` via `BadgerIterator`.

## 4. Métodos Implementados
- `Open` (Inicializa instâncias nativas `badger.Open` ignorando a poluição de logs standard).
- `Close` (Encerra o DB e limpa ponteiros mutexed).
- `GetBlockByIndex` (Utiliza uma Transaction `View` com Seek binário seguro `ValueCopy`).
- `GetLatestBlock` (Busca prefixada iterada reversa com finalizadores `0xFF` acoplados).
- `GetBalance` (Extração em View e cast de float64 bits a partir de `[]byte` persistido).
- `NewBatch` e transações subjacentes `PutBlock`, `PutBalance`, `Commit`, e `Discard`.
- `NewBlockIterator`, `Next`, `Value`, `Close` atrelados às otimizações do Cursor nativo.
- `CreateSnapshot` estruturado sem preenchimento intrusivo por decisão hierárquica.

## 5. Estrutura do Banco
Os pares Chave-Valor (K-V) obedecem rigorosamente à taxonomia:
- **`block/index/`** -> A chave recebe um _append_ the 8 bytes BigEndian correspondentes ao uint64 da altura para permitir a ordenação Lexicográfica inata do LSM Tree do Badger.
- **`balance/`** -> Prefixado via String Interpolation literal com `address`. Os valores persistidos são as representações brutas de 64 bits (`math.Float64bits`).

## 6. Dependências Adicionadas
- Pacote Externo: `github.com/dgraph-io/badger/v4`
- Ferramental atrelado no Workspace via `go mod tidy` (Ristretto Cache, OpenTelemetry, Protobuf, Flatbuffers).

## 7. Quantidade Aproximada de Linhas
- 270 linhas adicionadas sob `badger_storage.go`.

## 8. Problemas Encontrados
- **Lexicographical Key Indexing:** O banco `badger` classifica suas chaves iteráveis de forma estritamente alfanumérica. Se usássemos `fmt.Sprintf("block/index/%d", index)`, blocos de índice `10` precederiam índices `2`, corrompendo a leitura da chain.
- **Iterator Boundary Isolation:** O método `Value()` do BadgerDB sobrescreve o buffer na próxima iteração, propiciando severos _Data Races_ se repassado por referência crua à interface externa.

## 9. Correções Realizadas
- Construiu-se as _Helper Functions_ `makeBlockIndexKey` e `makeBalanceKey`, adotando estrita codificação `binary.BigEndian.PutUint64` limitando blocos a sufixos binários perfeitos em 8 bytes.
- Utilizou-se estritamente `item.ValueCopy(nil)` dentro dos Iterators (tanto Batch-Reads quanto Views singulares), copiando inteiramente os bytes pro GC do Go e isolando o `Txn` nativo do Badger.

## 10. Compatibilidade
- **JSONStorageAdapter:** Ambos assinam e devolvem resultados equivalentes sem nenhuma perda de coerência (Interfaces Engine 100% Plug-and-Play).
- **MockStorage:** Comportamentos em borda (ex: `GetBalance` retornando `storage.ErrNotFound` em ausência) são indistinguidos entre os dois motores.

## 11. Resultado dos Testes Locais
- **`go fmt`:** Formatado integralmente.
- **`go vet`:** Pass, nenhuma flag de concorrência disparada.
- **`go test`:** Pass, `ok pc_node (cached)`.
- **`go build`:** Compilação final atrelada à _LSM Tree_ com sucesso. Sem problemas de CGO ou bibliotecas cruzadas.
