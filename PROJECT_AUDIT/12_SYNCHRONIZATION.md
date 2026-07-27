# 12 SYNCHRONIZATION

## Sincronização e Download de Blocos
Um nó baixa blocos solicitando a outros ou esperando o recebimento de `NEW_BLOCK` na porta TCP. O mecanismo no código audita se o `PreviousHash` corresponde, mas não existe um protocolo complexo de handshake (`GetBlocks`, `Inv`, `Headers`) implementado no protótipo que auditei.

## Detecção de Conflitos e Recuperação
Se um nó ficar offline, quando voltar ele terá uma cadeia de blocos menor. Ao receber um broadcast, ele não possui mecanismo nativo de "fetch missing blocks" visível na lógica superficial. Essa ausência de fast-sync é um débito técnico enorme.
