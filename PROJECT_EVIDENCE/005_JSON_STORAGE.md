# Armazenamento Block-State em JSON

## Resumo
O uso de `json.MarshalIndent` no salvamento (`saveLedger()`) para toda a estrutura de blockchain na memória falhará fatalmente em escala por consumo abusivo de RAM e lentidão de serialização do array dinâmico massivo.

## Severidade
Alto

## Arquivos envolvidos
- `pc_node/ledger.go`

## Funções envolvidas
- `saveLedger()`
- `getLedgerJSON()`

## Fluxo de execução
Novo Bloco Validado
↓
`ledger.Chain = append(...)`
↓
`json.MarshalIndent(ledger.Chain)`
↓
`ioutil.WriteFile` (Escreve toda a chain no disco)

## Evidência encontrada
Em `saveLedger()`, o objeto array Go `ledger.Chain` é integralmente serializado para JSON e escrito por cima de `ledgerFile` a cada única mudança. Em uma rede com 500.000 blocos, cada salvamento recodificará gigabytes inteiros do zero.

## Como reproduzir
Criar um script de teste unitário inserindo 100.000 blocos forjados no array `ledger.Chain` e medindo o tempo de execução e consumo de RAM de `saveLedger()`.

## Impacto
Gargalo extremo de IOPS. Bloqueio prolongado da rede durante o unmarshal/marshal, travando o nó por falta de memória (OOM Killer).

## Consequências
O nó perderá capacidade de sync rápido. A blockchain é inviável em produção (Mainnet) desta forma.

## Módulos afetados
- pc_node (Storage/Consensus)

## Protocolos afetados
- CORE_PROTOCOL.md

## Invariantes afetados
- SYSTEM_INVARIANTS.md (Escalabilidade e Desempenho)

## Causa provável
Uso de solução rápida de protótipo (JSON dump) ao invés de estruturação via Merkle Patricia Tries ou Key-Value stores otimizadas (ex: LevelDB).

## Possíveis soluções
1. Migrar a persistência do ledger para um banco KV (ex: goleveldb ou badger). Armazenar blocos individuais mapeados por `{prefix}_index` e `{prefix}_hash`.
2. Otimizar a leitura parcial via índices, sem carregar transações antigas na RAM global.
3. Gravar blocos binários anexáveis (Append-only) em disco bruto em vez de arrays JSON (estilo Bitcoin Core `blk0000.dat`).

## Complexidade estimada
Média

## Risco da correção
Médio

## Ordem recomendada
5
