# RFC-0004: Storage Abstraction Layer (Interface Design)

## 1. Problema
Atualmente, o `ledger.go` acopla diretamente a lógica de negócios da blockchain (consenso, validação de transações) com o mecanismo de persistência (`os.CreateTemp`, `json.Marshal`, `ioutil.ReadFile`). Se precisarmos trocar para o BadgerDB (conforme RFC-0003), adicionar paginação, ou mockar o banco para testes unitários, seríamos obrigados a reescrever o núcleo do Ledger. Essa falta de separação viola o princípio de responsabilidade única (SRP) e impede evolução segura.

## 2. Objetivos
- **Desacoplamento Total:** O core da blockchain não deve saber se os blocos estão em JSON, BadgerDB, Postgres ou na memória (Testes).
- **Substituição Transparente:** Permitir ligar o novo banco sem quebrar a rede.
- **Testabilidade:** Facilitar a criação de testes 100% isolados utilizando `MockStorage`.
- **Mockabilidade:** Injeção de dependência na inicialização do nó.
- **Compatibilidade Futura:** Preparar terreno para Sidecar API e Sincronização P2P sem carregar toda a chain na RAM.

## 3. Interface Proposta
Desenhada com base no uso estrito e real do `ledger.go` atual:

```go
package storage

import "MeshCoin-Core/pc_node/models" // Abstração hipotética dos tipos locais

type Engine interface {
    // Lifecycle
    Open(connectionString string) error
    Close() error

    // Block Operations
    SaveBlock(block models.Block) error
    GetBlockByIndex(index uint64) (*models.Block, error)
    GetLatestBlock() (*models.Block, error)
    IterateBlocks(callback func(block models.Block) (stop bool)) error

    // State Operations (Balances / UTXO)
    // O(1) balance check para `HandleNewTransaction`
    GetBalance(address string) (float64, error)
    
    // API & Sync Support
    // Retorna a chain no formato antigo (JSON Array) para compatibilidade com a API Flutter atual
    DumpChainJSON() ([]byte, error) 
}
```

## 4. Separação de Responsabilidades
- **Ledger:** Atua exclusivamente como Orquestrador de Estado. Ele recebe o bloco validado e pede ao `Storage` para salvá-mo.
- **Storage:** Implementa a interface `Engine`. Responsável pelo I/O, commit atômico, compactação (Badger) e índices de balanço.
- **Consensus:** Isola as regras matemáticas matemáticas (`VerifyNeonHash`, estrutura de hash). Não sabe nada sobre disco.
- **Network:** Mantém o roteamento P2P, Sockets, WS e TCP estritos.
- **API:** Módulo Sidecar. Agora ele chama `Storage.DumpChainJSON()` ao invés de serializar uma variável global.
- **Wallet / Mining:** Permanecem desconectados do storage direto (acessam via API RPC ou P2P).
- **Cloud:** O nó `nebula_integration` acessa apenas os blocos finalizados através do `Storage.GetBlockByIndex()`.

## 5. Fluxo de Chamadas (Arquitetura Futura)

```text
[Network (P2P)] -> Recebe Bloco
       ↓
[Ledger] -> Roda Consensus e Verifica Hashes
       ↓
[Ledger] -> Solicita Storage.GetLatestBlock() (Double Check)
       ↓
[Ledger] -> Aprova Bloco e Pede Storage.SaveBlock()
       |
       |----> [Storage Adapter] -> Salva o Payload (KV)
       |----> [Storage Adapter] -> Atualiza Índice de Balanço daquele Bloco
```

## 6. Injeção de Dependência
O `ledger.go` perderá o acesso global `var ledgerFile` e a estrutura em memória `var ledger Ledger`. Em substituição:
```go
var DB storage.Engine

func InitNode(storageImpl storage.Engine) {
    DB = storageImpl
    DB.Open("./meshcoin_data")
}
```
No `main.go`, passaremos a implementação concreta na inicialização, facilitando a migração.

## 7. Estratégia de Implementação (Roadmap)
- **Etapa 1:** Criar a pasta `storage/` e o arquivo da interface (contrato puro em Go).
- **Etapa 2:** Criar o `JSONStorageAdapter` (implementa a interface varrendo o `ledger.json` e mantendo a array em RAM temporariamente para preservar o nó atual rodando enquanto refatoramos).
- **Etapa 3:** Criar o `BadgerStorageAdapter` (implementa a interface utilizando índices KV puros O(1)).
- **Etapa 4:** Troca transparente no `main.go` passando `BadgerStorageAdapter` no lugar do JSON. (Momento do Go Live).
- **Etapa 5:** Remover o código legado (`json.MarshalIndent` do `ledger.go`), deletando os vestígios da variável global em RAM.

## 8. Riscos
- Se a abstração vazar regras de negócio (ex: o Storage calcular recompensas de bloco), corromperá a arquitetura. A validação deve permanecer no Ledger; o Storage apenas escreve a verdade.
- Quebra do `getLedgerJSON()` na API. O adaptador deverá recriar o JSON fielmente para não quebrar o Flutter app.

## 9. Rollback
Se o novo Storage falhar, basta reiniciar o nó trocando a injeção em `main.go` para usar o `JSONStorageAdapter` novamente.

## 10. Critérios de Aceite
- O repositório deverá compilar perfeitamente.
- O `ledger.go` não deverá possuir nenhum import do pacote `os` relacionado à arquivos locais, `ioutil` ou menção ao disco.
- Todos os testes unitários da Sprint 2.0 deverão ser portados para injetar um `MockStorage` baseado em RAM.
