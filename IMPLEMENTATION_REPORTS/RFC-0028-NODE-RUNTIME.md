# IMPLEMENTATION REPORT: RFC-0028-NODE-RUNTIME (Nebula Network)

## 1. Visão Geral (Arquitetura)
A Fase 49 implementa o Motor Maestro da Nebula Network: o `Node Runtime & Module Orchestration Engine`. A função deste módulo é estritamente orquestral. Ele abstrai toda a infraestrutura subjacente (Storage, P2P, Mempool, UTXO, Consensus) e unifica a máquina sob um único guarda-chuva de inicialização determinística. Evitando espaguete de código, utiliza injeção de dependência via contêineres (`ServiceContainer`), notificações reativas via `EventBus` e gerenciamento topológico de *boot* via grafos diretos acíclicos (`DependencyGraph`).

## 2. Inicialização e Shutdown (Lifecycle & Dependency Graph)
- **DependencyGraph**: Para evitar travamentos em *deadlocks* cíclicos (A aguardando B, e B aguardando A), o Grafo resolve e desenha uma trilha de inicialização utilizando o Algoritmo de *Kahn* para ordenação topológica.
- **LifecycleManager**: Executa estritamente a trilha do grafo. Injeta `Contexts` portando os Limites de Timeout da Política. Se a Storage demora muito a abrir os arquivos em disco, o *startup* é imediatamente abortado de forma limpa. No Shutdown, o fluxo é estritamente reverso, derrubando a Rede primeiro e o Storage por último.

## 3. Comunicação Intermodular (Event Bus & Service Container)
- **ServiceContainer (Dependency Injection):** O acoplamento duro foi suprimido. Instâncias (Providers) são inseridas no balde, e Módulos os resgatam quando de suas instâncias de construção.
- **Event Bus:** Adota o padrão Observador genérico para comunicações indiretas. Canais Go emulando *Topics* permitem que o `P2P` avise a chegada de um bloco emitindo "BlockArrived", que por sua vez acorda a esteira de `Consensus` sem envolver *imports* cíclicos. 

## 4. Health Manager e Scheduler
- **HealthManager:** Efetua chamadas constantes nas interfaces expostas (Checkables). Mantém a telemetria do Nó (Degraded, Failed, Running).
- **RuntimeScheduler:** Cron nativo que processa lógicas periódicas internas como Varredura, Compactação da Database (Storage) ou envio de Pings (P2P).

## 5. Testes de Integração
Os testes simulam perfeitamente as dependências pesadas da rede. A suíte atesta:
- Cenário 1: Booting de `Storage -> Script -> UTXO -> Consensus -> Mempool -> Mining -> P2P`. Comprova 7 instâncias blindadas sem intercorrências.
- Concorrência de EventBus suportando massiva transição assíncrona sob `Mutex`, suportando 1000 chamadas instantâneas simuladas.
- Varredura circular com reprovação sumária (Test_CircularDependency).

---

## PENDÊNCIAS HERDADAS (OBRIGATÓRIO)

### Mocks Remanescentes:
- **`MockSignatureValidator` / `SignatureEngine`:** A verificação criptográfica do Script (Fase 47) ainda depende de um empacotamento abstrato comparativo (Mock) já que a curva final ECDSA ou Ed25519 não foi acoplada aos buffers oficiais.
- **MockStorage (Memória):** O Engine de *Storage* (Fase 48) salva, indexa e recupera perfeitamente os pacotes, todavia toda estrutura permanece atrelada à *heap* de memória volátil sob uso extensivo de `maps`. Necessita obrigatoriamente de *bindings* CGO para LevelDB/RocksDB no futuro para persistência real *on-disk*.

### Interfaces Provisórias:
- **Undo-Logs:** Rollback de Blocos na UTXO (Fase 46) suprime instâncias em memória, mas o Log reverso estrito em disco (para restaurar referências pretéritas gastas) está em espera atrelada à futura finalização da infraestrutura completa no Storage real.

### Dívida Técnica Acumulada:
- A agregação no Container (`ServiceContainer`) suporta `interface{}`, perdendo tipagem forte, podendo mascarar erros de conversão no Boot.

### Dependências de Fases Futuras:
- Falta o *Binding Real* com o Database Engine LevelDB/RocksDB.
- O Node Runtime orquestra *Mock Modules* nos testes. A fusão oficial de todos os pacotes das fases `34-48` no binário final do nó depende da Fase Integrativa final.

---

## STATUS DA DÍVIDA TÉCNICA (OBRIGATÓRIO)
**Mocks restantes:** Storage Map (Disco Volátil), Signature Verification (Sem ECDSA), Undo-log simplificado.
**Interfaces provisórias:** Interfaces de resolução `interface{}` no Service Container e EventBus.
**Integrações concluídas:** Orquestração completa, Topologia DFS/Kahn validada.
**Integrações pendentes:** O Bootstrap do nó raiz injetando as classes concretas criadas no `main.go`.
**Riscos arquiteturais:** O excesso de Go routines injetadas assincronamente pelo `EventBus` sem limitador rígido estrito além do tamanho do canal, podendo exaurir OS Threads se não mitigado.
**Compatibilidade retroativa:** Plena, as abstrações mantém o isolamento intocável.
**Impacto em desempenho:** A criação de um Event Bus genérico com Reflection indireto ou cast em tempo de execução adiciona sobrecarga fracional.
**Impacto em memória:** Instâncias do Container ficam na Heap persistentemente; Fila do Event Bus pré-alocada pode exigir ajustes finos.
**Próximos módulos candidatos à substituição de Mock:** Database Engine (LevelDB integration) ou ECC Signature Verification.
