# RFC-0001 REVIEW: Ledger Atomic Persistence

## 1. A RFC descreve corretamente o problema?
Sim. A RFC descreve o problema com extrema precisão técnica. A evidência primária citada é a chamada `ioutil.WriteFile` sem tratamento atômico (O_TRUNC nativo do SO) e o erro lógico da função `initLedger()`, que assume que `json.Unmarshal` falhando significa que o nó deve iniciar "limpo". O arquivo `pc_node/ledger.go` (linha 60-84) demonstra exatamente esse condicional agrupado problemático.

## 2. Existem cenários não considerados?
A RFC endereça o cenário básico, mas omite cenários de concorrência e cross-platform essenciais para infraestruturas críticas:
- **Cross-platform (Windows/Linux):** A função `os.Rename` no Windows pode falhar ao tentar sobrescrever um arquivo existente (`ledger.json`) se ele não possuir permissões corretas ou estiver em lock por leituras de outras threads.
- **Falhas de fsync (Flush to Disk):** A chamada padrão de escrita delega os bytes ao buffer de I/O do Kernel. Se ocorrer queda de energia antes de um `fsync` real do sistema operacional, o `ledger.json.tmp` recém-criado pode ficar vazio ou conter bytes nulos no disco. Renomeá-lo atômica e imediatamente resultaria em um `ledger.json` vazio (perda atômica).
- **Múltiplas goroutines:** A função `handleNewBlock` realiza `go saveLedger()`. Múltiplos blocos validados rapidamente dispararão múltiplas instâncias dessa gravação. Escrever hardcoded em `ledger.json.tmp` causará *Data Races* brutais a nível de sistema de arquivos e corromperá a serialização concorrente.

## 3. A estratégia proposta é suficiente?
**Não é suficiente na sua forma pura descrita (Write + Rename simples).**
Falta incorporar obrigatoriamente:
- A execução de `file.Sync()` (Fsync) no descritor de arquivo `.tmp` logo após a escrita e antes de fechar o arquivo, garantindo persistência do kernel no disco rígido.
- A geração de arquivos temporários com nomes exclusivos (ex: usando `os.CreateTemp` em `/tmp` local ou gerando um `.tmp.<rand>`) para que múltiplas goroutines em lock não choquem I/O local antes do rename final atômico.

## 4. Existem riscos de regressão?
- O maior risco (Cross-platform) é o nó PC rodando em Windows recusar o `os.Rename` sobre um arquivo travado, causando um erro silencioso onde a RAM possui os blocos novos, mas o disco paralisou na versão antiga (regressão de I/O).

## 5. A alteração pode permanecer restrita ao ledger.go?
**Sim.** A alteração é estritamente localizada. O parsing JSON, as regras de leitura e a injeção em memória operam apenas sobre os arrays de estrutura dentro de `ledger.go`. Nenhuma assinatura de rede (`network.go`) ou comunicação externa (RPC/Cloud) necessita tomar conhecimento se o arquivo por baixo está usando rename atômico.

## 6. Existe uma solução melhor mantendo o mesmo escopo?
Mantendo o estrito limite de "usar JSON serializado" (descartando o salto para LevelDB temporariamente):
- A melhor solução (Alternativa A) exige utilizar `os.CreateTemp` para que o I/O concorrente não colida nomes, usar `file.Write()`, forçar `file.Sync()`, fechar o descritor e só então invocar um wrapper seguro de rename (ou usar locks de leitura/escrita no O.S. combinados com o `ledger.mu`). O escopo conceitual é o mesmo da RFC, porém robustecido.

## 7. Plano de testes
- **Persistência normal:** Inserir 5 transações e verificar se `ledger.json` existe com os dados intactos, validando que o flush ocorre adequadamente.
- **Corrupção proposital:** Interromper o processo, injetar sintaxe corrompida (`"chain": [ { `) no `ledger.json` manualmente, e iniciar o node. O node DEVE retornar um erro fatal (`log.Fatalf`) e recusar a formatação, impedindo perda histórica.
- **Interrupção durante gravação:** Em ambiente simulado de teste, forçar um timer ou `panic` entre a criação do arquivo temporário e seu Rename. Após re-iniciar, o original `ledger.json` não pode ter sido afetado.
- **Concorrência repetida:** Executar um benchmark submetendo simultaneamente 50 requests HTTP/TCP de Blocos e avaliar se os saves não quebram o acesso local (concurrency file safety).

## 8. Recomendação final
**APROVADA COM AJUSTES**

**Justificativa Técnica:**
A RFC diagnostica com exatidão a falha de arquitetura responsável pela possível amnésia (perda de dados) nos nodes da Nebula Network e a solução primária de Atomic Rename no O.S. é o padrão correto da indústria.
Contudo, para prosseguir com a implementação (Fase 9), o Engenheiro **obrigatoriamente** deve incluir no código:
1. `file.Sync()` para proteção real contra panics de kernel/energia.
2. Prevenção de concorrência local de nomes `.tmp` sob goroutines paralelas.
3. Compatibilidade do `os.Rename` no Windows (ou bypass seguro).
A aprovação é dada contanto que a implementação cumpra os aditivos desta revisão.
