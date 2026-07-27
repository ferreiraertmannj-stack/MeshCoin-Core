# Data Loss no Ledger Corrompido

## Resumo
A função que inicializa o Ledger mestre trata erros de leitura/deserialização apagando silenciosamente o estado atual e recriando um novo Bloco Gênesis, resultando em perda total de dados locais.

## Severidade
Crítico

## Arquivos envolvidos
- `pc_node/ledger.go`

## Funções envolvidas
- `initLedger()`

## Fluxo de execução
Inicialização do Node
↓
Carregamento do Ledger (`initLedger()`)
↓
Erro de Unmarshal JSON
↓
Criação do Bloco Gênesis
↓
Sobrescrita de `ledger.json`

## Evidência encontrada
Em `pc_node/ledger.go`, a função `initLedger()` tenta ler e dar unmarshal no `ledger.json`. Se houver erro, ela cai no bloco principal que inicializa a variável genesis (linha 71) e chama `saveLedger()`, sobrescrevendo o arquivo existente corrompido sem gerar backup.

## Como reproduzir
1. Inicie o `pc_node`.
2. Pare o processo e edite o arquivo `ledger.json` para conter um JSON inválido (ex: remova uma chave).
3. Reinicie o `pc_node`.
4. O nó recriará a chain apenas com o genesis, apagando os blocos anteriores.

## Impacto
Um desligamento abrupto ou erro de disco que deixe o JSON parcial fará o nó deletar todo o histórico validado.

## Consequências
Nós validadores sofrerão "amnésia" e poderão forçar rescync total (se suportado no futuro) ou criarão forks irreconciliáveis caso voltem a minerar.

## Módulos afetados
- pc_node (Consensus / Storage)

## Protocolos afetados
- CORE_PROTOCOL.md

## Invariantes afetados
- SYSTEM_INVARIANTS.md (Imutabilidade do Ledger, Persistência Garantida)

## Causa provável
Falta de tratamento de erro robusto. O desenvolvedor agrupou o cenário "Arquivo não existe" com "Arquivo corrompido" no mesmo fallback de criação do Gênesis.

## Possíveis soluções
1. Fazer backup automático de arquivos inválidos (ex: `ledger.json.corrupt`) e abortar (panic) a inicialização.
2. Manter uma cópia .bak atualizada e reverter em caso de corrupção do principal.
3. Trocar o formato JSON por um banco de dados ACID (LevelDB, RocksDB) resistente a falhas de energia.

## Complexidade estimada
Baixa

## Risco da correção
Baixo

## Ordem recomendada
1
