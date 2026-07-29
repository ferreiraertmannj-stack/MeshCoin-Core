import os

out_dir = "PROJECT_EVIDENCE"
os.makedirs(out_dir, exist_ok=True)

files = {
    "001_LEDGER_ATOMICITY.md": """# Data Loss no Ledger Corrompido

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
""",

    "002_NETWORK_DDOS.md": """# Upload Cloud sem Autenticação

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
""",

    "003_GLOBAL_LOCK.md": """# Gargalo de Mutex Global (Global Lock)

## Resumo
Tanto o Ledger (chain completa) quanto a Mempool utilizam Mutexes globais únicos em Go, o que obriga a execução síncrona de blocos e transações em um ambiente que deveria ser massivamente concorrente.

## Severidade
Alto

## Arquivos envolvidos
- `pc_node/ledger.go`
- `pc_node/network.go`

## Funções envolvidas
- `handleNewBlock()`
- `HandleNewTransaction()`

## Fluxo de execução
Network Listener TCP
↓
Goroutine `handleConnection`
↓
`handleNewBlock()`
↓
`ledger.mu.Lock()` (Bloqueia todas as outras conexões)

## Evidência encontrada
Em `ledger.go`, há a definição `var ledger = Ledger{Chain: []Block{}, mu: sync.RWMutex{}}`. Em `handleNewBlock`, a primeira instrução é `ledger.mu.Lock()`, que só dá release via defer. O mesmo ocorre com `mempoolMutex`. 

## Como reproduzir
Submeter milhares de blocos e transações via requisições TCP paralelas para a porta 5556. A vazão (throughput) cairá drasticamente pois todas as goroutines entrarão em contenção pelos mesmos locks de maneira sequencial.

## Impacto
Impossibilidade de escalabilidade vertical no nó.

## Consequências
Com alta latência devido à contenção de lock, a rede pode droppar blocos legítimos por timeout, causando divisões (forks).

## Módulos afetados
- pc_node (Consensus)

## Protocolos afetados
- CORE_PROTOCOL.md

## Invariantes afetados
- SYSTEM_INVARIANTS.md (Escalabilidade e Concorrência)

## Causa provável
Design prematuro e ingênuo das estruturas de dados. A modelagem priorizou consistência forte na memória sem considerar bancos de dados concorrentes nativos ou channel architecture (padrão Go).

## Possíveis soluções
1. Substituir a estrutura JSON em memória por LevelDB.
2. Passar o gerenciamento de estado para uma goroutine singular consumidora de canais (Actor Model).
3. Adotar fine-grained locking, bloqueando apenas as partes da chain que estão sendo modificadas.

## Complexidade estimada
Alta

## Risco da correção
Alto

## Ordem recomendada
3
""",

    "004_FALSE_MESH.md": """# Broadcast Estático Falso-Mesh

## Resumo
Apesar da documentação citar uso de protocolos Mesh avançados (B.A.T.M.A.N.), a infraestrutura de rede atual é estática, operando exclusivamente via broadcasts UDP em LAN (Broadcast Address) e multiplexação TCP linear sem qualquer controle de topologia Mesh, grafos de roteamento ou métricas de salto (hops).

## Severidade
Médio

## Arquivos envolvidos
- `pc_node/network.go`
- Documentos de Arquitetura.

## Funções envolvidas
- `broadcastPresence()`
- `broadcastTCP()`

## Fluxo de execução
Node Inicializa
↓
UDP SendTo `255.255.255.255:5555`
↓
Ao receber dado TCP, envia for loop para todos os `activeTCPClients`.

## Evidência encontrada
A função `broadcastPresence()` em `network.go` apenas tenta discar para `255.255.255.255`. A função `broadcastTCP()` simplesmente itera no mapa de `activeTCPClients` e faz um `.Write()`. Não há código em Go que calcule rotas, gerencie TTL de pacotes, implemente roteadores ad-hoc ou lide com B.A.T.M.A.N. virtualizado.

## Como reproduzir
Executar o código em instâncias que não estejam no mesmo domínio de broadcast L2 (ex: duas subnets diferentes numa AWS ou VPS distintas). Eles não se acharão via UDP.

## Impacto
A rede não funciona na Internet nem além da mesma rede Wi-Fi/LAN física, a menos que os IPs TCP sejam hardcoded ou conhecidos.

## Consequências
A premissa principal do projeto ("Mesh") não está materializada no nó validador primário.

## Módulos afetados
- pc_node (Network)

## Protocolos afetados
- CORE_PROTOCOL.md

## Invariantes afetados
- SYSTEM_INVARIANTS.md (Descentralização, Off-Grid Capability)

## Causa provável
Documentação está apontando para o roadmap futuro (ou para implementações delegadas apenas aos scripts Python / App Flutter mobile que podem tentar Bluetooth), mas não foi implementada no Go (Core Node).

## Possíveis soluções
1. Implementar Kademlia DHT para peering dinâmico sobre TCP/UDP fora da rede local.
2. Adicionar TTL e IDs de pacote no protocolo de broadcast TCP para evitar tempestades de loop (packet storms).
3. Integrar libp2p para abstrair o roteamento mesh complexo.

## Complexidade estimada
Muito Alta

## Risco da correção
Alto

## Ordem recomendada
4
""",

    "005_JSON_STORAGE.md": """# Armazenamento Block-State em JSON

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
""",
    
    "MASTER_EVIDENCE_INDEX.md": """# MASTER EVIDENCE INDEX

## Quantidade Total de Problemas Analisados
5 evidências críticas estruturadas (representando as anomalias centrais do ecossistema mapeado).

## Quantidade por Severidade
- **Crítico:** 2 (Ledger Atomicity, Network DDoS)
- **Alto:** 2 (Global Lock, JSON Storage)
- **Médio:** 1 (False Mesh)
- **Baixo:** 0

## Arquivos Mais Críticos (Focos de Instabilidade)
1. `pc_node/ledger.go` (Responsável por 3 dos 5 problemas)
2. `nebula_cloud/node_daemon.go` (Responsável pela brecha de Cloud)
3. `pc_node/network.go` (Responsável pelo falso-mesh e gargalo de I/O)

## Módulos Mais Afetados
- **pc_node** (A fundação do consenso e validação demonstra arquitetura prematura inadequada para ambientes hostis e distribuídos).

## Ordem Sugerida de Correção
A fila de correção (Prioridade Técnica) deve seguir a fundação do sistema (do persistente para o roteamento):
1. `001_LEDGER_ATOMICITY.md` (Risco de perda total)
2. `002_NETWORK_DDOS.md` (Risco de exaustão externa)
3. `005_JSON_STORAGE.md` (Permitirá a correção arquitetural de locks e desempenho)
4. `003_GLOBAL_LOCK.md` (Sem JSON bloqueante, os locks podem ser fine-grained)
5. `004_FALSE_MESH.md` (Pode aguardar libp2p ou implementação DHT robusta)

## Dependências entre problemas
A correção do `003_GLOBAL_LOCK.md` (Performance/Gargalos) está intimamente acoplada à solução de `005_JSON_STORAGE.md`. Substituir o JSON por um banco KV (LevelDB) resolverá naturalmente os tempos de bloqueio de I/O, exigindo que a refatoração do storage preceda a refatoração de concorrência.
"""
}

for filename, content in files.items():
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)
