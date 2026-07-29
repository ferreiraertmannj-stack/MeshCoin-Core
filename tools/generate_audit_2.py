import os

out_dir = "PROJECT_AUDIT"
os.makedirs(out_dir, exist_ok=True)

files = {
    "06_CONSENSUS.md": """# 06 CONSENSUS

## Fluxo Completo
O consenso se baseia majoritariamente na Prova de Trabalho (PoW - NeonHash). Ao receber um bloco via TCP (em `handleNewBlockPacket`), o nó valida se `PreviousHash` corresponde ao último bloco, e recalcula o PoW.

## Forks e Rollback
**INEXISTENTE.** Atualmente o código apenas rejeita blocos (retornando `false`) se o índice for `<=` ao atual ou se o `PreviousHash` for inválido. Não há implementação de reestruturação de cadeia (reorg/rollback) se uma cadeia paralela mais pesada for detectada.

## Mempool
A mempool (`PendingTransactions`) é um array global em memória protegido por um Mutex (`mempoolMutex`). Quando um bloco é validado, as transações nele contidas são removidas da mempool. Transações são transmitidas via P2P.

## Estado Atual
Altamente simplificado. É um consenso frágil para redes distribuídas instáveis, sendo mais semelhante a uma cadeia de logs anexáveis com verificação de hash.
""",

    "07_WALLET.md": """# 07 WALLET

## Criação e Chaves
As carteiras são criadas usando a curva elíptica `secp256k1`. A geração de endereços e gestão do par de chaves parece estar suportada tanto em código Go quanto no app Flutter (implementação Dart). No backend Go, usa-se a biblioteca `decred/dcrd/dcrec/secp256k1`.

## Assinatura
As transações assinam o pacote (Timestamp + Amount + Fee + Sender + Receiver) com a chave privada. No arquivo Go, o código (implícito em `VerifyTransaction` que não foi lido por completo) verifica a assinatura ECDSA do `SenderPubKey`. Há também o campo `PQCSignature` previsto.

## Saldo e Histórico
Calculado iterando sobre todo o `ledger.json` e somando entradas e subtraindo saídas. Não existe modelo UTXO, é um sistema baseado em Contas/Saldos calculados on-the-fly.

## Fonte da Verdade
A única fonte da verdade é a reconstrução do estado baseada na lista sequencial de blocos confirmados no `ledger.json`.
""",

    "08_MINING.md": """# 08 MINING

## Algoritmo
Utiliza `NeonHash`. O algoritmo aloca um vetor em memória de 4096 bytes semeado a partir de um SHA-256 inicial do cabeçalho. Realiza 128 iterações de saltos pseudo-aleatórios e mutações, finalizando com um SHA-256 no vetor mutado.

## Reward
Fórmula: `CalculateBlockReward()`. Recompensa base é 50 NBL, sofrendo halving a cada 2.100.000 blocos. Existe um "Proof of Storage" bônus onde armazenar dados na Nebula Cloud rende NBL extras (0.5 por GB no SSD, 0.1 por GB no HDD).

## Candidate Block e Performance
Mineração otimizada para evitar vantagens de ASICs. O uso do vetor de 4KB exige acesso a memória que encarece ASICs enquanto mantém smartphones e CPUs normais competitivos.
""",

    "09_MESH.md": """# 09 MESH

## Discovery e Routing
A descoberta na rede local é feita com Broadcasts UDP na porta 5555. Nós anunciam `NEBULA_NODE:<portaTCP>` a cada 5 segundos. 

## Multi-hop e Store and Forward
Embora a documentação fale muito de protocolo B.A.T.M.A.N. e comunicações off-grid com saltos, a análise do nó Go (pc_node/network.go) revela um encaminhamento simples via TCP. Mensagens (Blocos, Transações e Chat) são recebidas, validadas e redistribuídas (broadcast TCP/WS).

## Bugs
O modelo de roteamento atual envia de volta para as conexões TCP ativas. A falta de verificação avançada de roteamento (ex: IDs de pacotes únicos) pode levar a loops de broadcast infinitos na rede local (packet storm) caso haja loops na topologia e as validações (ex: mempool já tem a tx, ou ledger já tem o bloco) não abortarem a propagação rápido o suficiente.
""",

    "10_NETWORK.md": """# 10 NETWORK

## TCP e UDP
- **UDP (Porta 5555)**: Usado exclusivamente para descoberta local contínua via broadcast (255.255.255.255).
- **TCP (Porta 5556)**: Usado para transporte confiável de `NEW_BLOCK`, `NEW_TRANSACTION`, e pacotes de mensagens do tipo `DATA_ROUTE` ou `CHAT`.

## Bluetooth e Wi-Fi Direct
A implementação no Go não abrange o rádio local. A delegação dessas funções depende inteiramente da implementação em Dart no aplicativo móvel Flutter, o qual atuaria como uma ponte (bridge) entre essas redes e a rede TCP local do PC.

## Reconexão e Heartbeat
Há mensagens `PING` e `OGM` descritas no router TCP, tratadas de forma silenciosa para manutenção de sockets abertos. 
"""
}

for filename, content in files.items():
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)
