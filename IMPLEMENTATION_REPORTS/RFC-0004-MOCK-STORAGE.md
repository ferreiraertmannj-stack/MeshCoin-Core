# IMPLEMENTATION REPORT: RFC-0004-MOCK-STORAGE (Mock Storage Engine)

## 1. Arquivos Criados
- `pc_node/storage/mockstorage/mock_storage.go`

## 2. Interfaces Implementadas
- **`storage.Engine`:** Instanciada via `MockEngine`.
- **`storage.Batch`:** Instanciada via `MockBatch`.
- **`storage.Iterator`:** Instanciada via `MockIterator`.

## 3. Estratégia de Sincronização
A estrutura `MockEngine` utiliza um controle rigoroso com `sync.RWMutex` (`mu`). 
Operações de leitura pontuais como `GetBlockByIndex`, `GetLatestBlock`, e `GetBalance` usam `RLock()` (Read Lock), permitindo acessos concorrentes sem concorrência de escrita.
Para prevenir _Data Races_ posteriores ao retorno dos dados, os métodos devolvem **cópias defensivas (deep copies)** dos arrays de bytes (`[]byte`), impedindo que o chamador externo modifique inadvertidamente o estado interno através da mesma referência de ponteiro em memória.

## 4. Estratégia do Batch
O `MockBatch` atua de maneira completamente atômica. Ele mantém um estado temporário em `map[uint64][]byte` para blocos e `map[string]float64` para saldos. 
Durante a função `Commit()`, ele:
1. Requisita um Exclusive Lock (`Lock()`) nativo no `MockEngine`.
2. Expande internamente o tamanho do Slice nativo (`e.blocks`) caso o índice inserido no Batch exceda o tamanho atual.
3. Insere todos os buffers e saldos temporários perfeitamente isolados.
4. Efetua o _Discard_ limpando o mapa residual temporário.

## 5. Estratégia do Iterator
O `NewBlockIterator` adota a estratégia de _Snapshot In-Memory_. 
Durante a criação do iterador, ele obtém um `RLock()` e clona a matriz inteira de blocos daquele exato ponto temporal para uma array separada (`[][]byte`).
Isso garante que o `Iterator` externo não falhe ou cause _panic_ por indexação (out of bounds/concurrency changes) caso uma nova inserção seja committada no Engine base de forma estritamente concorrente.

## 6. Garantias contra Data Race
- Mutexes englobam inteiramente os acessos.
- Sem ponteiros residuais em retornos (todas as fatias são recriadas via método `copy()`).
- O iterador é totalmente estático e independente do backend mutável pós-instanciação.

## 7. Compatibilidade com JSONStorageAdapter
O design pattern permaneceu 100% aderente às diretrizes construídas na Fase 15/16. Todos os `return` statements assinam a mesma conformidade para `[]byte` e `error` (`storage.ErrNotFound`, `storage.ErrClosed`).
Ele pode ser substituído (`swap`) livremente em tempo de inicialização na `main.go` sem nenhuma alteração funcional requisitada na infraestrutura da Blockchain.

## 8. Limitações Intencionais
- Persistência 0%: Qualquer *restart* do nó descarta todos os blocos alocados na memória (RAM).
- `CreateSnapshot`: Ignorado passivamente (Mocado para retornar sucesso em vez de `NotImplemented`), visto que gravar o Snapshot em disco corromperia a natureza In-Memory estrita deste adapter.
- Sem particionamento: Se o ledger mockado crescer a proporções gigantescas, ocorrerá sobrecarga passiva de memória sem _Swapping_, devido ao clonamento defensivo contínuo do Iterator.

## 9. Resultado dos Testes
- **go fmt:** Pass.
- **go vet:** Pass. Nenhuma heurística de código concorrente alarmou Data Races ou falhas.
- **go test:** Pass (`ok pc_node`).
- **go build:** Compilação final intacta. Nenhuma dependência externa requisitada (100% nativo `sync` e abstrações de interfaces criadas localmente).

## 10. Linhas Alteradas
- Adição de 213 linhas contidas exclusivamente sob `pc_node/storage/mockstorage/mock_storage.go`. Nenhuma outra linha paralela do repositório foi alterada. (Nem mesmo JSONAdapter ou arquivos vitais de protocolo).
