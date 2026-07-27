# RFC-0002 DRAFT: Lock Optimization & Data Race Fixes

## Problema
O nó PC possui três gargalos extremos e fatais de sincronização:
1. Validação pesada de Proof of Work (NeonHash) rodando dentro do Lock global do Ledger, congelando o sistema inteiro a cada bloco recebido.
2. Anti-pattern de rede onde `.Write()` de socket é executado dentro do `tcpMutex` de mapa.
3. Map de WebSockets não possui Mutex, o que invariavelmente gerará crash (fatal Data Race Map Write) em produção ao conectar/desconectar clientes simultâneos.

## Impacto
Lentidão massiva conforme a rede cresce (travamento de thread). Queda aleatória (Crash) por corrupção de memória (Data Race em Go map).

## Arquivos que deverão ser alterados
- `pc_node/ledger.go`
- `pc_node/network.go`
- `pc_node/main.go`

## Arquivos proibidos de alteração
- Nenhum módulo de lógica central de negócio, hash, blockchain ou cloud.
- O formato do JSON.

## Estratégia
1. Em `ledger.go`: Remover validação `VerifyNeonHash` e checagem indexada `block.Index` para FORA da área do `ledger.mu.Lock()`. Travar o lock apenas no momento de fazer o `append()`.
2. Em `network.go`: Adicionar um buffer de cópia no `tcpMutex`. Clonar os peers, liberar o lock, e só então iterar chamando `.Write()` sem prender o Mutex de mapa global.
3. Em `main.go`: Introduzir `wsMutex sync.Mutex` protegendo a leitura e escrita do `map clients`.

## Riscos
- Possíveis Txs duplicadas na mempool caso não seja cuidadoso a inversão do lock na validação, porém a lógica de anexação atômica prevenirá.
- Crash zero.

## Plano de Rollback
Reverter via Git para a tag base.

## Critérios de Aceite
- Executar o programa rodando `-race` flag sem disparar logs de Data Race.
- Verificar via PProf que o nó não segura a lock de `tcpMutex` para IO de rede.
