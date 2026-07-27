# RFC-0001: Estabilização de Persistência Atômica do Ledger

## 1. Problema
A gravação e inicialização do estado principal da blockchain (Ledger) não possui atomicidade (Atomic Save) nem proteção contra arquivos corrompidos (Safe Init).
- **Funções:** `saveLedger()` e `initLedger()`.
- **Arquivo:** `pc_node/ledger.go`
- **Fluxo:** Ao salvar a blockchain via `saveLedger()`, o sistema invoca `ioutil.WriteFile()` para despejar o JSON recém-serializado. Esta função no sistema operacional executa um `O_TRUNC` (esvazia o arquivo primeiro) e depois inicia a escrita. Se ocorrer interrupção (ex: queda de energia) neste microsegundo, o arquivo `ledger.json` ficará vazio ou parcialmente escrito (corrompido). No próximo boot, `initLedger()` tentará ler este arquivo quebrado, interceptará o erro e, incorretamente, apagará tudo recriando um novo Bloco Gênesis, gerando perda total irreparável de dados (Amnésia).

## 2. Evidências
- **Funções:** Em `initLedger()` (linha ~60-84):
  ```go
  if jsonErr := json.Unmarshal(file, &ledger.Chain); jsonErr == nil && len(ledger.Chain) > 0 { ... }
  // Fallback:
  ledger.Chain = []Block{genesis}
  saveLedger()
  ```
  O fallback para qualquer erro de leitura é recriar o Gênesis.
- **Funções:** Em `saveLedger()` (linha ~86-96):
  ```go
  ioutil.WriteFile(ledgerFile, data, 0644)
  ```
  Não existe gravação em arquivo temporário (ex: `ledger.json.tmp`) nem renomeação.
- **Fluxo de execução:** `handleNewBlock` -> Acata Bloco Válido em RAM -> `go saveLedger()` -> `ioutil.WriteFile` destrutivo.

## 3. Causa
A implementação atual de persistência foi criada como um protótipo focado no "caminho feliz", negligenciando as primitivas de entrada e saída (I/O) atômicas do POSIX/Windows. Além disso, houve mistura das semânticas de "Arquivo não existe" com "Arquivo corrompido" dentro de um único fallback condicional (`if/else`) no boot do nó.

## 4. Impacto
- **Quando ocorre:** Quedas de energia, travamentos (Kernel Panic) ou OOM Killers que interrompam a thread Go exata e cirurgicamente durante a chamada `ioutil.WriteFile`, bem como qualquer edição manual malfeita no arquivo.
- **Dados perdidos:** A totalidade dos blocos e transações que existiam antes do crash. 100% de perda de dados validados localmente.
- **Módulos afetados:** `Ledger` (Consensus Node). Toda a sincronização P2P (Network) falhará, pois o nó voltará ao bloco 0 propagando sua própria chain como superior caso forje novos blocos, gerando forks.

## 5. Arquivos que serão alterados
Exatamente 1 arquivo:
- `pc_node/ledger.go`

## 6. Arquivos que NÃO serão alterados
- `pc_node/network.go`
- `nebula_cloud/node_daemon.go`
- O diretório Flutter/Dart (Wallet).
- O script Python Miner.

## 7. Estratégia
A correção focará puramente no método de I/O em JSON (etapa Quick Win antes de adotar LevelDB futuramente), implementando as seguintes etapas em Go:
1. **Em `saveLedger()`:** Serializar para JSON e escrever os dados em um arquivo temporário `ledger.json.tmp` utilizando `os.WriteFile()`. Em seguida, usar `os.Rename("ledger.json.tmp", "ledger.json")`. A operação Rename em sistemas de arquivos modernos é atômica no nível do VFS, prevenindo qualquer truncate no meio do caminho.
2. **Em `initLedger()`:** Tratar erros diferentemente. Se `os.IsNotExist` for verdade, prosseguir com a criação natural do Gênesis. Caso o arquivo exista, mas o `json.Unmarshal` falhe (corrupção por bit-rot ou O.S.), lançar um erro crítico (`log.Fatalf`) instruindo o operador a verificar backups manuais (ex: `.bak`), abortando o nó e impedindo a sobrescrita imediata com um novo Bloco Gênesis vazio.

## 8. Riscos
- Risco nulo à criptografia (NeonHash permanece intacto).
- Risco operacional baixíssimo (apenas mudança do driver de manipulação de string para arquivo).
- Risco de regressão zero, visto que nenhuma interface exportada para o pacote Network será modificada.

## 9. Plano de rollback
O plano de rollback será reverter o commit de refatoração diretamente via Git (`git revert <hash>`). A leitura do JSON voltará a ser interpretada pelo padrão antigo, uma vez que o formato serializado (a estrutura dos campos do `Block`) não mudará. O rollback não causará corrupção aos arquivos salvos pela nova versão.

## 10. Critérios de aceite
- Desligar o nó, injetar lixo string dentro do arquivo `ledger.json` e reiniciar o PC node: O nó deve exibir um `FATAL` e morrer, ao invés de prosseguir e criar um Gênesis em cima do arquivo antigo.
- O salvamento assíncrono (a cada bloco novo) deve exibir a criação momentânea (e desaparecimento) de um arquivo `.tmp`, comprovando a escrita atômica.
