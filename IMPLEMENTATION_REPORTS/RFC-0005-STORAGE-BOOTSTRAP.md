# IMPLEMENTATION REPORT: RFC-0005-STORAGE-BOOTSTRAP (Storage Engine Factory)

## 1. Arquivos Criados
- `pc_node/storage/serialization.go` (Desacoplamento estrito das funções auxiliares de (Un)Marshal para isolar `ledger.go` do pacote `jsonstorage`).

## 2. Arquivos Alterados
- `pc_node/main.go` (Introdução da `InitStorageEngine` Factory e injeção de dependência na inicialização).
- `pc_node/ledger.go` (Remoção total de pacotes concretos. Dependência puramente baseada na interface e na factory repassada pelo bootstrap).
- `pc_node/ledger_test.go` (Injeção via Mock do `jsonstorage` restrito ao escopo de Testes de Unidade, preservando os mocks temporários de arquivo).

## 3. Fluxo de Inicialização e Storage Factory
Durante a execução de `main()`, a função `InitStorageEngine()` é a primeira regra de negócio crítica executada.
Ela analisa determinística e sequencialmente as chaves providas pelo ambiente operacional:
1. Variável de Ambiente O.S. `NEBULA_STORAGE`.
2. Flag de linha de comando nativa `-storage`.
3. Padrão nativo (Default: `json`).

Após a coleta da String, ela é reduzida para LowerCase. Um bloco `switch` valida e instancia as três _engines_ suportadas. 
Se o nome repassado divergir da lista formal, a fábrica aborta a pipeline lançando um Erro Fatal rastreável (`log.Fatalf`), que estanca a thread _main_ e impede totalmente que ouvintes vulneráveis (TCP, WebSockets e API) entrem em estado _Listen_ com _Ledger_ ausente.

## 4. Engines Suportadas e Compatibilidade Preservada
- **`json`**: O ecossistema retrô-compatível para legados que lê e escreve arquivos crus `ledger.json` via sistema O.S nativo.
- **`badger`**: A nova infraestrutura LSM Tree validada na RFC-0004.
- **`mock`**: Banco efêmero estritamente residente em RAM para simulações e cenários de testes unitários.

Nenhuma regra de negócio da Nebula Network (PoW, Mempool, ECDSA, Consenso, Parsing P2P, Broadcast) foi corrompida, alterada ou adaptada. `ledger.go` continua utilizando primitivas abstratas cegas (Interface `storage.Engine` e iteradores polimórficos) para assinar as alturas e hashear blocos, mantendo-se estritamente fiel aos fundamentos das Fases 1 a 19.

## 5. Resultados de Validação e Pipeline de CI Local
Foram feitas chamadas sistêmicas isoladas utilizando _environment variables_ (json, badger, mock). Todos os processos inicializaram nativamente sem _panics_ ou falhas de dereferência de ponteiro (`nil pointer dereference`). 

- **`go fmt`**: Pass (`serialization.go`, `main.go`, `ledger.go`, `ledger_test.go`).
- **`go vet`**: Pass.
- **`go test`**: Pass (`ok pc_node (cached)`).
- **`go build`**: Pass (Geração do binário `pc_node.exe` estática na arquitetura destino sem resquícios sintáticos).
