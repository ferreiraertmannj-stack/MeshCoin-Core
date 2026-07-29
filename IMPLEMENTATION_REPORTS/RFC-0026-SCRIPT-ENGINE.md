# IMPLEMENTATION REPORT: RFC-0026-SCRIPT-ENGINE (Nebula Network)

## 1. Visão Geral (Arquitetura)
A Fase 47 implementa o poderoso motor de scripts criptográficos (`Script Engine`), desvinculando sumariamente qualquer verificação de assinatura das regras diretas da Mempool, UTXO e Blockchain. Assumindo a figura de uma Máquina Virtual Isolada (Stack-based Virtual Machine), ela decodifica instruções em tempo real garantindo o controle preciso sobre regras de bloqueio/desbloqueio econômico (ScriptSig e ScriptPubKey).

## 2. Máquina Virtual Baseada em Pilha (Stack & Executor)
- **ScriptStack:** Fiel à lógica original P2PKH, gerencia os dados através de *Push* e *Pop*. Limitada matematicamente a uma profundidade segura (1000 itens) definida na política `ScriptPolicy`, mitigando vetores de ataque *Out-Of-Memory* (OOM).
- **Executor:** A execução inicia-se pelo `ScriptParser`, desmembrando hexadecimais em `Instruction` (Data vs Opcodes). O Executor então corre de maneira sequencial. Caso o estado final do Stack seja nulo ou falso (0), a transação é severamente abortada.

## 3. Dinâmica de Opcodes e Registros Puros
- O `OpcodeRegistry` trabalha no modelo *Plugin-like*. Inicialmente foram embarcados os comandos canônicos `OP_DUP`, `OP_HASH256`, `OP_EQUALVERIFY` e `OP_CHECKSIG`. 
- Isso garante que a Nebula Network possua a capacidade singular de expandir Contratos Inteligentes ou SegWit injetando Opcodes posteriormente na inicialização do nó, sem precisar alterar a classe base.

## 4. Hash, PublicKey e Signature Engines (Agnosticismo Criptográfico)
Desprezou-se o encapsulamento hard-coded criptográfico. Injetou-se o trio:
1. `HashEngine`: Centraliza Double SHA-256 e RIPEMD-160.
2. `PublicKeyEngine`: Delega as regras de formatação (e compressão) a um ecossistema estanque.
3. **`SignatureEngine`**: Ponto alto do isolamento. Permite transições indolores (ECDSA -> Ed25519) bastando instanciar novos Engines, suprimindo com glória o antigo MockSignatureValidator herdado da Fase 46.

## 5. ScriptCache e Fila Assíncrona (Concorrência Absoluta)
O Gargalo criptográfico foi pulverizado:
- **ScriptCache**: Assinaturas repetidas submetidas pelo Gossip (P2P) transitando pela Mempool recebem bypass imediato *O(1)* se o Hash da Tx, ScriptSig e ScriptPubKey uníssonos não estouraram o TTL temporal.
- **ScriptQueue**: Canais em Go providenciam que uma tempestade (*Flooding*) de Transações inválidas não derrube os núcleos processadores, afunilando e enfileirando execuções independentes sob o guarda-chuva de `context.Context` imunes à interrupções abruptas.

## 6. Limitações Atuais e Marcos Futuros
- **Isolamento em Sandbox Limitada:** A implementação atende puramente fluxos numéricos básicos, controles de igualdade e *CheckSig*. Funcionalidades condicionais avançadas (`OP_IF`, `OP_ELSE`) bem como tranca de relógio estrita (`OP_CHECKLOCKTIMEVERIFY`) dependem de atualizações no pacote `opcode_control.go`.
- **Validação Segregada:** O parser atual concatenará o *ScriptSig* com *ScriptPubKey*, atendendo à regra legada. No futuro, execuções particionadas garantem a mitigação completa da malleabilidade transacional.
