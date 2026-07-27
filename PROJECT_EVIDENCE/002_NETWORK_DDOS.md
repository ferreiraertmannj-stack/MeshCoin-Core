# Upload Cloud sem Autenticação

## Resumo
A Nebula Cloud possui uma rota de upload HTTP que permite envio de até 100MB por requisição sem qualquer validação criptográfica, autenticação, ou rate limiting, tornando o armazenamento distribuído trivialmente explorável.

## Severidade
Crítico

## Arquivos envolvidos
- `nebula_cloud/node_daemon.go`

## Funções envolvidas
- `handleUploadFragmento()`

## Fluxo de execução
Agente Externo
↓
POST `/upload`
↓
Nebula Cloud (`handleUploadFragmento`)
↓
Disk Write

## Evidência encontrada
Em `node_daemon.go`, o `handleUploadFragmento` lê diretamente de `r.FormFile` e escreve no `STORAGE_DIR` com um limite de `MAX_FRAGMENT_SIZE` (100MB). Não há nenhum cabeçalho de autenticação ou verificação de chaves.

## Como reproduzir
Executar um loop em qualquer terminal externo: `while true; do curl -X POST -F "fragmento=@dummy.bin" -F "nome=dummy-$RANDOM.bin" http://<ip>:8000/upload; done`

## Impacto
O disco do provedor de armazenamento na Nebula Cloud será esgotado em minutos.

## Consequências
Negação de serviço (DDoS) sistêmica na camada de nuvem descentralizada, impedindo que dados reais do ledger sejam salvos.

## Módulos afetados
- nebula_cloud (Storage)

## Protocolos afetados
- CORE_PROTOCOL.md

## Invariantes afetados
- SYSTEM_INVARIANTS.md (Segurança de Rede, Prevenção de Abuso)

## Causa provável
Protótipo focado apenas no "caminho feliz" (Happy Path) para provar o conceito de fragmentação e armazenamento, negligenciando controles de acesso.

## Possíveis soluções
1. Exigir assinatura ECDSA de uma carteira válida nos headers.
2. Implementar desafio PoW (Hashcash) antes do upload.
3. Adicionar verificação de saldo da carteira na Blockchain para autorizar Storage (Proof of Stake/Payment).

## Complexidade estimada
Média

## Risco da correção
Baixo

## Ordem recomendada
2
