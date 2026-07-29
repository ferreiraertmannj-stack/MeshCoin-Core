# IMPLEMENTATION REPORT: RFC-0012-BOOTSTRAP-INTEGRATION (Nebula Network)

## 1. Visão Geral (Overview)
A Fase 33 marca a ativação do ecossistema Fast Sync no ciclo de vida real da Nebula Network. O motor P2P assíncrono (Downloader, Validator, Importer, Controller) foi conectado no topo da função `main()`, atuando como um *Gatekeeper* obrigatório antes que as goroutines vitais da rede (Mineração, Mempool, Broadcast e Consenso de Tempo Real) possam acordar.

## 2. Arquitetura e Integração

O Bootstrap foi encapsulado em um componente lógico dedicado (`pc_node/bootstrap.go`) exposto como `RunFastSyncBootstrap()`. O fluxo é executado sequencialmente na rotina principal da `main()`, bloqueando o prosseguimento da inicialização até que um veredito seja emitido:
1. **Descoberta de Peers (`discoverPeers`)**: O Pool é populado (mock para esta fase mantendo isolamento de P2P físico, respeitando o requerimento "Nunca alterar P2P existente").
2. **Comparação de Alturas (`getHighestKnownHeight`)**:
   - Se `remoteHeight <= localHeight`: O nó loga que está atualizado e o Bootstrap libera a `main()` para iniciar o `startNetwork()`.
   - Se `remoteHeight > localHeight`: O *Pipeline* completo do Fast Sync é alocado.
3. **Isolamento Concorrente**:
   - Utilizando `context.WithCancel`, o Controller foi amarrado ao Bootstrap.
   - Qualquer sinalização de falha, cancelamento pelo usuário ou colisão crítica acionará o `Cancel()` de todas as sub-rotinas e abortará a inicialização da Nebula.

## 3. Estados e Sequência
Enquanto o Fast Sync estiver ativo, o nó estará formalmente em um estado "cúbico" fechado:
- Não abrirá portas TCP para broadcast comum.
- Não submeterá blocos de mineradores.
- Não consultará o mempool.
Apenas quando a Callback `OnCompleted` for disparada no Controller, o bloqueio do `<-ctx.Done()` será resolvido e o nó retomará seu comportamento padrão.

## 4. Testes e Resiliência
A suíte `bootstrap_test.go` foi anexada contendo cenários de:
- **Nó Atualizado**: Confirma que o Fast Sync desiste do processo instantaneamente e retorna `nil` ao topo.
- **Nó Atrasado**: O processo inicia o `SyncController` sem modificar as regras de Consenso e avança.
- **Cancelamento Abruto**: Contextos foram disparados limitando o tempo simulando o aborto de uma sincronização pesada, avaliando que as Goroutines do `Downloader` e `Manager` param imediatamente, prevenindo falhas de estado.

## 5. Limitações
Como a regra absoluta da Fase exigia *não alterar* o P2P existente (`network.go`), a etapa formal de Discovery (onde IPs cruzam handshakes) foi mantida isolada. O Bootstrap injeta stubs na `discoverPeers()` para simular as descobertas. No mundo real, na Fase final de Transporte, esse stub será trocado pelas regras P2P definitivas.

## 6. Próximos Passos
O próximo passo orgânico é integrar a camada P2P real (Mensagens em Socket, Gob/JSON na rede, Ping/Pong e RequestHeaders/Blocks), acoplando os protocolos ao `TCPPeer`. Isso finalizará o projeto permitindo que máquinas distintas transfiram megabytes paralelamente pelo Fast Sync.
