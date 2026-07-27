# IMPLEMENTATION REPORT: RFC-0004-CLEANUP (Desacoplamento Completo do Ledger)

## 1. Responsabilidades Removidas do Ledger
- **Leitura/Escrita Direta em Disco:** Todas as funções locais relacionadas à abstração mecânica de sistema de arquivos (OS e ioutil) foram removidas.
- **Serialização Direta de Persistência:** A chamada nativa `json.Marshal` para serialização de bytes durante gravação de blocos foi totalmente abolida do ledger principal, bem como o processo espelho `json.Unmarshal`.
- **Roteamento de Payload:** A função auxiliar e legada `getLedgerJSON` que atuava servindo a Sidecar API lendo do arquivo JSON em disco não existe mais na scope do `ledger.go`.

## 2. Responsabilidades Delegadas ao Adapter
- **I/O File System:** `os.CreateTemp`, `os.Rename`, `os.Remove`, `os.Open` e afins.
- **Domínio de Transformação Raw Byte:** Foram criados Helpers de serialização/deserialização `jsonstorage.MarshalBlock` e `jsonstorage.UnmarshalBlock` dentro do Adapter. A conversão da estrutura `Block` -> `[]byte` (bytes brutos) exigidos pela Interface Engine agora pertence contextualmente ao próprio ambiente Storage, blindando o Ledger.
- **Iterador (Block Iterator):** A responsabilidade da API via Node/HTTP foi transferida para consumir o `DB.NewBlockIterator()` nativamente no `main.go`.

## 3. Arquivos Alterados
- `pc_node/ledger.go` (Despida completamente de bibliotecas json, os, ioutil).
- `pc_node/main.go` (Recebeu a função refatorada `/api/ledger` usando iterador do DB, e a função encapsulada helper `getLedgerJSON` isolada da ledger chain).
- `pc_node/storage/jsonstorage/json_storage.go` (Aprimorada para exportar Helpers de conversão).
- `pc_node/ledger_test.go` (Diretório temporário de testes consertado perante a abstração do OS).

## 4. Quantidade de Linhas Alteradas
- **Linhas Removidas:** ~40 (remoção de pacotes `encoding/json` e `io/ioutil`, extinção da sub-rotina `getLedgerJSON` original, remoção do helper `formatDartDouble`).
- **Linhas Adicionadas:** ~30 (Adição da refatoração de iterador na `main.go`, adição dos Unmarshal Helpers na package jsonstorage).

## 5. Compatibilidade Preservada
A estrutura, lógica, criptografia e formatação nativa do MeshCoin Core permanceram idênticos ao original de produção. O nó principal continua escutando o Websocket (`main.go`) e disparando `uploadToNebulaCloud()`. Todas as restrições da Blockchain (ECDSA/Secp256k1 e NeonHash) operam sem qualquer interferência. 

## 6. Testes Executados
Os testes unificados mantiveram-se em pleno funcionamento:
- `go fmt ./...`: Sucesso.
- `go vet ./...`: Sem warnings, pacotes redundantes varridos.
- `go test ./...`: Pass! `ok pc_node 0.329s`. (Teste atômico de cleanup validado)
- `go build ./...`: Compilação binária nativa intacta.

## 7. Riscos Encontrados
- **Risco de Serialização de Domínio vs Persistência:** A Interface `Engine` do Storage demanda estritamente arrays de byte `[]byte`, enquanto o `ledger.go` opera structs `Block`. Abolir `json.Marshal` no arquivo gerou um GAP de cast direto.
- **Risco de Quebra no TestSaveLedgerAtomic:** Ao alterar as variáveis relativas ao package global `os`, a verificação de sujeira (limpeza de diretórios) do Node-Test corrompeu momentaneamente devido à discrepância com o Workspace do GitHub Actions/SO `filepath.Dir`.

## 8. Riscos Corrigidos
- Adicionados Wrappers Helpers em `json_storage.go` para servir de pipeline de transporte entre o Domínio (`Block`) e a interface Engine (`[]byte`).
- A diretiva estrita `filepath.Glob` foi consertada dentro de `ledger_test.go` alinhando a criação do lixo (Tmp) ao seu devida local folder injetado via Test Pipeline, garantindo 100% de estabilidade e Cleanup Automático.

## 9. Auditoria Final
O arquivo `ledger.go` foi rigorosamente inspecionado.
Ele não possui mais: `os`, `ioutil`, `json.Marshal`, `json.Unmarshal`.
Sua atribuição foi integralmente devolvida para:
1. Conflito PoW/Consenso (NeonHash V1.0)
2. Validação Matemática Mempool ECDSA (Assinaturas Secp256k1)
3. Regras e Indexação da Blockchain
4. Coordenação da delegação final ao objeto `DB`.
