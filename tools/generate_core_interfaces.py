import os

out_dir = "MODULE_ANALYSIS/CORE_INTERFACES"
os.makedirs(out_dir, exist_ok=True)

template = """# {title}

## Responsabilidade
{resp}

## Entradas
- Eventos recebidos via TCP/UDP
- Chamadas de função diretas

## Saídas
- Retorno booleano de validação
- Emissão de pacotes para a rede

## Dependências
- Módulos requeridos para execução

## Interfaces Públicas
- Métodos expostos (Ex: `Handle...`, `Verify...`)

## Interfaces Privadas
- Métodos internos não exportados

## Eventos Recebidos
- Descrição de triggers

## Eventos Enviados
- Descrição de ações disparadas

## Estruturas Compartilhadas
- Structs trafegadas

## Fluxos Críticos
- Sequências de execução vitais

## Possíveis Gargalos
- Pontos de contenção (CPU, RAM, Disco, Lock)

## Riscos Arquiteturais
- Ameaças mapeadas
"""

modules = [
    ("01_LEDGER_INTERFACES.md", "01 LEDGER INTERFACES", "Armazenar estado da blockchain e mempool."),
    ("02_CONSENSUS_INTERFACES.md", "02 CONSENSUS INTERFACES", "Garantir concordância e PoW (NeonHash)."),
    ("03_WALLET_INTERFACES.md", "03 WALLET INTERFACES", "Assinatura, geração de chaves e visualização de balanços."),
    ("04_MINING_INTERFACES.md", "04 MINING INTERFACES", "Execução de hashing para montagem de blocos e recompensas."),
    ("05_NETWORK_INTERFACES.md", "05 NETWORK INTERFACES", "Abertura de sockets TCP/UDP e routing básico."),
    ("06_MESH_INTERFACES.md", "06 MESH INTERFACES", "Descoberta P2P WAN/LAN."),
    ("07_SYNC_INTERFACES.md", "07 SYNC INTERFACES", "Reconciliação de histórico entre pares."),
    ("08_STORAGE_INTERFACES.md", "08 STORAGE INTERFACES", "Interface de disco para state (DB/KV)."),
    ("09_CRYPTO_INTERFACES.md", "09 CRYPTO INTERFACES", "Geração ECDSA, SHA256 e PQC."),
    ("10_NODE_INTERFACES.md", "10 NODE INTERFACES", "Ponto de entrada (Main) orquestrando Módulos."),
    ("11_API_INTERFACES.md", "11 API INTERFACES", "Endpoints HTTP (Nebula/Node).")
]

for filename, title, resp in modules:
    content = template.format(title=title, resp=resp)
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)

custom_files = {
    "12_DEPENDENCY_GRAPH.md": """# 12 DEPENDENCY GRAPH

## Grafo Lógico de Dependências
```text
[API/Cloud] -----> [Node Daemon]
                      |
[Wallet (Flutter)] -> [Node Network (TCP)]
                      |
[Mesh Discovery] ---> [Network] ---> [Consensus (Verify)]
                                        |
                                        V
[Storage (Disk)] <------------------ [Ledger]
                                        |
[Mining (App)] ----> [Consensus] -------+
```
""",
    "13_CALL_GRAPH.md": """# 13 CALL GRAPH

## Principais Chamadas entre Módulos
- **Network -> Ledger:** Quando um peer envia um TCP payload contendo transação, a função `HandleNewTransaction` é acionada (cruzamento do pacote de socket para a Mempool).
- **Consensus -> Crypto:** Durante a mineração ou validação (`VerifyNeonHash`), o consenso delega a geração do `sha256` e a checagem `ecdsa.Verify` à infraestrutura criptográfica em Go.
- **Ledger -> Storage:** Ao acatar um bloco em RAM, o Ledger invoca `saveLedger`, cruzando para a interface de File System.
""",
    "14_EVENT_FLOW.md": """# 14 EVENT FLOW

## Eventos Importantes Documentados
1. **Nova Transação:** Dispara validação local -> Adiciona em RAM (Mempool) -> Faz Broadcast UDP/TCP para vizinhos.
2. **Novo Bloco:** Dispara hash match -> Verifica assinaturas internas -> Acata bloco na Chain -> Invalida Txs duplicadas na Mempool -> Salva no disco.
3. **Novo Peer:** Dispara `OGM` (Originator Message) no estilo B.A.T.M.A.N -> Adiciona no map `activeTCPClients`.
4. **Sincronização:** (Flow falho atual) Nó aguarda novo bloco e insere à frente, ignorando o passado vazio.
""",
    "15_SHARED_STRUCTURES.md": """# 15 SHARED STRUCTURES

## Structs Centrais
1. **Block**: Transita entre `Network` (Socket), `Consensus` (Miner), `Ledger` (Array) e `Storage` (JSON). É o artefato de mais alto acoplamento do sistema.
2. **Transaction**: Transita entre `Wallet` (Assinatura UI), `Network` (Payload UDP/TCP), `Ledger` (Mempool), `Consensus` (Validação individual).
""",
    "16_ARCHITECTURAL_BOTTLENECKS.md": """# 16 ARCHITECTURAL BOTTLENECKS

## Gargalos de Comunicação
1. **Network -> Ledger:** O socket aguarda o Lock Global (`sync.RWMutex`) do Ledger, transformando redes assíncronas em processos síncronos e lentos.
2. **Ledger -> Storage:** Dump integral do array JSON paralisa todo o IO.
""",
    "17_INTERFACE_RISKS.md": """# 17 INTERFACE RISKS

1. **Ausência de Contratos Fortes (Interfaces Go):** Módulos conversam via struct literal e globais. Não há `type LedgerStore interface` o que impede injeção de dependência e testes unitários.
2. **API Cloud sem Middleware Auth:** O limite da interface Cloud não tem bloqueio criptográfico; aceita form-data cego.
""",
    "18_REFACTORING_BOUNDARIES.md": """# 18 REFACTORING BOUNDARIES

## Limites Seguros
- **Isolados:** O `Crypto` e `Consensus (Hash Alg)` podem ser alterados isoladamente, pois recebem um Block e devolvem um booleano.
- **Acoplados:** `Storage` e `Ledger` precisam ser alterados JUNTOS. Mudar de JSON para LevelDB obriga alterar todo o State Machine do Ledger.
- **Risco Maior:** Trocar TCP iterativo por Canais (Actor Model) no `Network`. Mudará todo o ciclo de vida dos nós.
""",
    "19_MODULE_COUPLING.md": """# 19 MODULE COUPLING

O acoplamento é extremamente ALTO entre:
- `Network` e `Ledger` (Variável global `ledger` e `PendingTransactions` é acessada por pacotes da rede).
O acoplamento é BAIXO entre:
- `Crypto` (puro) e os demais componentes.
""",
    "20_EXECUTIVE_SUMMARY.md": """# 20 EXECUTIVE SUMMARY

O mapeamento das interfaces revela um design procedural e acoplado, centrado em variáveis globais. Existem 11 módulos lógicos documentados e dezenas de transições síncronas.
A ausência de "Interfaces" (`type X interface`) reais na linguagem Go impede testes isolados.
Para as próximas fases, a quebra desse acoplamento entre `Network` e `Ledger` será o maior desafio arquitetural.
"""
}

for filename, content in custom_files.items():
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)

print("Core interface mapping files generated.")
