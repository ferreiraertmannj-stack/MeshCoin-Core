# RFC-0001 TESTS REPORT (SPRINT 2.0)

## 1. Quantidade de testes criados
Foram criados **8 testes unitários reais** independentes que testam os fluxos vitais do módulo Ledger sem uso de mocks, instanciando os ciclos reais em disco (TempDir) e controlando o fluxo da infraestrutura Go:
1. `TestInitLedgerFileDoesNotExist`
2. `TestInitLedgerFileValid`
3. `TestInitLedgerFileCorrupted` (com spawn de subprocesso para bypassar `log.Fatalf`)
4. `TestSaveLedgerNormal`
5. `TestSaveLedgerAtomic`
6. `TestIntegritySaveReload`
7. `TestSaveLedgerWriteFailure`
8. `TestSaveLedgerConcurrency`

## 2. Quais cenários foram aprovados
**Todos os 8 cenários** passaram na validação do pacote de testes (`PASS`). O comportamento da engine é determinístico sob estresse.

## 3. Quais falharam
Nenhum. Zero falhas.

## 4. Cobertura obtida
A cobertura geral do pacote resultou em **5.4% of statements**. Note que este percentual reflete a fatia em relação a todo o core do nó de rede (`pc_node/*`). As funções de I/O em disco do `ledger.go` (`saveLedger` e `initLedger`) atingiram cobertura aproximada de 100% de seus ramos.

## 5. Limitações restantes
* As assinaturas `CalculateBlockReward`, `VerifyNeonHash`, `VerifyTransaction`, `handleNewBlock` e `HandleNewTransaction` ainda não estão cobertas pelos testes implementados, dado que fogem do escopo da estabilização de persistência (RFC-0001) para qual os testes foram encomendados.
* A conversão temporária de `const` para `var` para injeção de dependência na path do disco é um anti-pattern em Go, denotando a urgência da refatoração da Arquitetura Orientada a Interface/Struct, mas que neste escopo restritivo garantiu a execução segura dos testes.
