# 14 NEBULA

## Tudo Que Pertence ao Nebula Cloud
A nuvem (`nebula_cloud/node_daemon.go`) expõe endpoints HTTP (`:8000`) para upload/download de fragmentos (`.bin`). 

## Integração e Estado
A integração ocorre no `pc_node/ledger.go` que a cada 10 blocos aciona uma goroutine para enviar backups. O estado é totalmente funcional, mas altamente ingênuo: aceita uploads de qualquer um via POST (sem autenticação) desde que com no máximo 100MB de tamanho.

## Roadmap Inferido
Transição de simples armazenamento para contratos de fragmentação distribuída (Reed-Solomon) onde os doadores de storage receberão recompensas diretamente atreladas às "Proofs of Storage" em blockchain.
