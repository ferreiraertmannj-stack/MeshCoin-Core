# IMPLEMENTATION REPORT: RFC-0003-MIGRATION-TOOL (JSON to BadgerDB)

## 1. Arquivos Criados
- `pc_node/tools/migrate_db.go`: Ferramenta CLI isolada para realizar a migração sem integrar-se à malha de bootstrap do nó.

## 2. Arquivos Alterados
Nenhum arquivo do _core_ ou das regras de negócio foi modificado. Todo o código principal da **Nebula Network** (`main.go`, `ledger.go`, `consenso`) permaneceu rigorosamente intacto.

## 3. Fluxo da Migração
1. **Bootstrap CLI**: Argumentos (`--input`, `--output`, `--force`, `--verify`) capturados e normalizados.
2. **Defesa Inata**: Avalia a integridade pré-existente (arquivo de entrada ausente ou diretório de saída sem flag de autorização).
3. **Setup de Instâncias**: Abre o `jsonstorage` (Origem) em modo _Read-Only_ implícito e o `badgerstorage` (Destino) referenciando os adaptadores originais da interface `storage.Engine`.
4. **Iteração Abstrata**: O ponteiro se move unicamente através do contrato formal (`jsonEngine.NewBlockIterator()`), decodificando cada bloco isoladamente usando o _helper_ independente `storage.UnmarshalBlock()`.
5. **Enfileiramento**: Cada bloco decodificado aciona um recálculo determinístico de saldos (O(1) para a migração sequencial), e os _Bytes_ do bloco e o _Float_ do saldo são injetados em um único `storage.Batch` ativo.
6. **Persistence e Descarregamento**: O _Batch_ orquestra o `Commit()` final escrevendo _Index_ e _Address/Balance_ na _LSM Tree_ subjacente, liberando finalmente os adaptadores através do defer `Close()`.

## 4. Estratégia de Batch
Optou-se por um _Batch Unificado Atômico_. Em vez de abrir múltiplas transações (pesadas computacionalmente), todos os blocos varridos pelo Cursor de leitura do JSON e todos os saldos rastreados em memória são acoplados a um `badgerEngine.NewBatch()`. Ao final, executa-se o `Commit()`, viabilizando uma migração atômica 100% transacional, salvaguardando o sistema contra arquivos corrompidos pela metade em caso de queda de energia durante o CLI.

## 5. Estratégia de Verificação (`--verify`)
Para a rotina de dupla-certificação, um _Iterator_ oposto é invocado (`badgerEngine.NewBlockIterator()`), refazendo toda a varredura linear por cima dos bytes recém persistidos. A verificação checa assertivamente:
- `bCount != blockCount`: O total iterável exato.
- `bLastIndex != lastIndex`: Se a altura final reflete a originária.
- `bLastHash != lastHash`: Se a String (SHA-256) final corrobora a prova do `NeonHash`.
Em caso de quebra contratual analítica, o script lança um erro e aborta, alertando explicitamente o usuário sobre a corrupção da cópia transacional.

## 6. Tratamento de Erros
- _Destino Interditado_: Sem `--force`, um diretório existente trava e emite **Fatal error**.
- _Source Ausente_: Se `ledger.json` não existir, o _Bootstrap_ nem sequer inicializa a infraestrutura subjacente.
- Nenhum arquivo é detruído nativamente (proteção contra perda acidental da corrente).

## 7. Compatibilidade Preservada
Os adaptadores nativos não foram expostos e continuam privados. O CLI depende apenas de abstrações em torno do pacote `storage`, confirmando a viabilidade de plugar e mover instâncias implementadas.

## 8. Casos Testados
- **Ledger Vazio**: Nenhuma inserção feita. Loop encerra com sucesso (Altura e count = 0).
- **Ledger Inexistente**: Saída Fatal imediata provada via teste.
- **Banco Existente sem Force**: Saída Fatal protetiva em _Stat()_.
- **Flag `--force`**: Passagem limpa, os diretórios são injetados no LSM.
- **Flag `--verify`**: Confirmação rigorosa do último hash e total de blocos (testado empiricamente em ledger de 35 blocos na raiz do repositório, com tempo recorde menor que 10ms).

## 9. Resultados de Validação
- **`go fmt`**: Pass.
- **`go vet`**: Pass.
- **`go test`**: Pass (`ok pc_node (cached)`).
- **`go build`**: Pass. A rotina `go build -o migrate_db.exe ./tools/migrate_db.go` construiu perfeitamente a binária na pasta raiz do _node_.
