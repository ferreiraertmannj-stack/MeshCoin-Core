# IMPLEMENTATION PLAN: RFC-0012-BOOTSTRAP-INTEGRATION (Nebula Network)

## 1. Goal Description
Integrar a arquitetura Fast Sync ao processo real de inicialização (Bootstrap) da Nebula Network, garantindo que a sincronização ocorra *antes* do nó iniciar suas atividades padrão (mineração, recebimento de blocos e transações no mempool). O sistema Fast Sync deverá ser interrompido corretamente se houver erro ou cancelamento pelo usuário, sem deixar goroutines ativas.

## 2. Proposed Changes

### `pc_node/bootstrap.go` (NEW)
Criar este arquivo para centralizar o processo de Bootstrap de forma limpa.
Ele conterá a função `RunBootstrap(localHeight uint64) error`, que será responsável por:
1. Instanciar `PeerPool` e adicionar conexões (como as regras de P2P físico ainda não foram detalhadas, podemos popular o pool de forma simulada/básica por enquanto ou usar a interface, pois os componentes devem ser "utilizados exatamente como estão").
2. Descobrir a altura máxima (`remoteHeight`).
3. Comparar `remoteHeight > localHeight`. Se falso, retorna e continua o Node normal.
4. Se verdadeiro, inicializa `SyncManager`, `Downloader`, `BlockValidator`, `BlockImporter` e `SyncController`.
5. Acopla os callbacks do Controller (ex: `OnCompleted`, `OnFailed`, `OnCancelled`) usando `context.WithCancel` para sincronizar a rotina principal de inicialização do nó.
6. Invocar `controller.Start(remoteHeight)` e travar a execução até a conclusão usando um `select` no context.

### `pc_node/main.go` (MODIFY)
O arquivo principal será refatorado minimamente para incluir o Bootstrap antes do `startNetwork()`.
- [MODIFY] `main.go`:
  ```go
  initLedger() // Carrega o ledger e define localHeight
  
  // Inicia Fast Sync
  err := RunBootstrap(uint64(len(ledger.Chain)))
  if err != nil {
      log.Fatalf("Falha crítica no Fast Sync: %v", err)
  }

  // Só então inicia a rede normal (mineração, mempool, p2p, sidecar)
  startNetwork()
  ```

### `pc_node/bootstrap_test.go` (NEW)
Testes cobrindo:
- Nó atualizado (Fast sync ignorado).
- Nó atrasado (Fast sync executado até o fim).
- Cancelamento (Testando que o contexto é liberado corretamente).
- Erro durante sincronização.

## 3. User Review Required
> [!IMPORTANT]
> A implementação real das mensagens P2P TCP (serialização JSON em socket, busca de headers pela rede, etc) no pacote `sync` ainda está pendente ou mockada nas Fases anteriores (ex: `SendMsg` é um stub em `TCPPeer`).
> Para esta integração de Bootstrap, devemos utilizar um mecanismo de injeção falso para a "Descoberta de Peers" e o Download, ou deixamos a orquestração amarrada usando `TCPPeer` real mas cientes de que ele fará *no-ops* pela rede? O plano criará a fundação arquitetural que descobre peers (mockados nos testes, no-ops na inicialização real) para satisfazer as restrições da Fase 33. Confirma se essa abordagem está de acordo com o isolamento esperado?

## 4. Verification Plan
- Executar `go test ./...` com o novo `bootstrap_test.go`.
- Validar `go build ./...` no diretório raiz e `/pc_node/`.
- Confirmar zero concorrência espúria (`goroutines`) vazando via testes de estresse.
