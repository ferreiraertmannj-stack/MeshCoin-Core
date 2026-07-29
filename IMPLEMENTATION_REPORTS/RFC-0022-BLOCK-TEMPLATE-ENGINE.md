# IMPLEMENTATION REPORT: RFC-0022-BLOCK-TEMPLATE-ENGINE (Nebula Network)

## 1. Visão Geral (Arquitetura)
A Fase 43 introduz o **Block Template Engine**, um módulo independente projetado para preparar blocos candidatos de mineração. Operando sob o princípio de Injeção de Dependências, este motor está rigorosamente desacoplado das regras de Consenso, Blockchain ou Storage. Ele recebe apenas *Interfaces* limpas (`BlockchainProvider`, `MempoolProvider`, `ConsensusProvider`, `NetworkProvider`) para extrair os dados necessários e abstrair operações complexas, unificando-as na criação do `BlockTemplate`.

## 2. Seleção de Transações (`TransactionSelector`)
Para montar o candidato, uma imagem no tempo (*Snapshot*) da Mempool é capturada. 
- **Ordenação:** Ordena-se as transações da maior *Fee Density* (Taxa dividida pelo Peso/Tamanho) para a menor, priorizando retornos ao minerador.
- **Deduplicação:** Uma verificação de Hash O(1) impede que múltiplas transações iguais adentrem o mesmo bloco.
- **Encaixe em Limites:** Transações são apensadas ao Bloco Candidato até que o limite estrito da política (`MaxBlockWeight`, `MaxTransactions`) seja alcançado.

## 3. Cálculos Abstraídos (`FeeCalculator` e `WeightCalculator`)
Os cálculos operam através da extração das Interfaces transacionais.
- `WeightCalculator`: Valida de maneira robusta contra `MaxBlockWeight` e `MaxTransactions` da política instanciada.
- `FeeCalculator`: Avalia de modo transparente o total arrecadado no escopo do bloco que será vertido ao Minerador.

## 4. Montagem e Geração (`CoinbaseBuilder` e `BlockAssembler`)
O `BlockAssembler` orquestra a síntese do bloco.  
- **Coinbase**: Constrói deterministicamente uma transação Coinbase de recompensa agregando o *BlockReward* embutido na `TemplatePolicy` às *Fees* extraídas do bloco.
- **Merkle Root**: Uma array linear de hashes transacionais (encabeçada pela Coinbase) é submetida ao `ConsensusProvider` injetado para a obtenção do hash central (Merkle Root).

## 5. Cache e Scheduler Inteligentes (`TemplateCache` e `TemplateScheduler`)
Para prevenir exaustão de processamento do nó:
- **Cache**: Implementado cache *Thread-Safe* assente sob TTL configurável (e.g. 30 segundos). Retorna a cópia do bloco candidato em `O(1)` enquanto estiver fresco.
- **Scheduler**: Rotina de plano de fundo via `goroutines` e `Context` que recicla e destrói automaticamente os templates através de *Tickers* ou Interrupções Ativas (ex: recepção de *MsgTransaction* na rede). Dispara callbacks `OnTemplateUpdated`.

## 6. Eventos e Estatísticas
O motor foi desenhado no formato de observador, emitindo engatilhos (Callbacks) definidos pela interface `TemplateEvents` em momentos vitais (`OnTemplateCreated`, `OnTemplateUpdated`, `OnTemplateExpired`, `OnTemplateRejected`, `OnCoinbaseGenerated`, `OnTransactionSelected`, `OnValidationFailed`).
As contabilidades também são rigorosamente atômicas (`TemplateStatistics`), logando número de templates gerados, transações preteridas por peso, além de traçar métricas de densidade (média de weight e fees por bloco gerado).

## 7. Limitações Conhecidas
- Como o Consenso está perfeitamente desacoplado, o motor fia-se nas validações prévias da Mempool quanto a assinaturas e balanços (UTXO). 
- O cálculo de dependência unspent inter-bloco (Child-Pays-For-Parent) encontra-se em design aberto e pode ser injetado em futuras expansões da abstração da interface `Transaction`.
