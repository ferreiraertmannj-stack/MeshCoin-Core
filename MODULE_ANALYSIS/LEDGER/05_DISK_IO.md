# 05 DISK I/O

### Leitura Inicial (`initLedger()`)
- Síncrona. 
- Carrega o arquivo `ledger.json` inteiro usando `ioutil.ReadFile()`.
- OOM (Out Of Memory) Risk imediato na leitura de chains gigabytes.

### Salvamento Diário/Continuo (`saveLedger()`)
- Síncrono (dentro da Goroutine spawnada pelo `handleNewBlock`).
- Possui Lock estrito em `ledger.mu.Lock()`, serializa a cadeia *inteira* com `json.MarshalIndent`.
- Escreve *por cima* do arquivo com `ioutil.WriteFile()`. O arquivo tem bytes = tamanho total da chain atual.
- Operação extremamente ineficiente com tempo Big-O de *O(N)* a cada bloco minerado.

### Resposta de Leitura Externa (`getLedgerJSON()`)
- Tenta ler direto do disco, gerando contenção de I/O na máquina física.
- Se falhar, faz Marshal global bloqueando o RWMutex.
