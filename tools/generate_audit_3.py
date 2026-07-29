import os

out_dir = "PROJECT_AUDIT"
os.makedirs(out_dir, exist_ok=True)

files = {
    "11_STORAGE.md": """# 11 STORAGE

## Banco e Persistência
Não se utiliza um SGBD. A persistência baseia-se em serialização de structs para JSON local (`ledger.json`). 

## Snapshots e Recovery
O arquivo Go `pc_node/ledger.go` lê o JSON do disco no `initLedger()`. Se o arquivo estiver corrompido, ele emite um log e silenciosamente **recria um novo Bloco Gênesis**. Isso constitui um risco massivo de perda de dados.

## Serialização
O uso de `json.MarshalIndent` no salvamento (`saveLedger()`) é lento para bancos grandes e não escala para um ledger real com milhões de blocos, pois exige alocação do arquivo inteiro na RAM.
""",

    "12_SYNCHRONIZATION.md": """# 12 SYNCHRONIZATION

## Sincronização e Download de Blocos
Um nó baixa blocos solicitando a outros ou esperando o recebimento de `NEW_BLOCK` na porta TCP. O mecanismo no código audita se o `PreviousHash` corresponde, mas não existe um protocolo complexo de handshake (`GetBlocks`, `Inv`, `Headers`) implementado no protótipo que auditei.

## Detecção de Conflitos e Recuperação
Se um nó ficar offline, quando voltar ele terá uma cadeia de blocos menor. Ao receber um broadcast, ele não possui mecanismo nativo de "fetch missing blocks" visível na lógica superficial. Essa ausência de fast-sync é um débito técnico enorme.
""",

    "13_CRYPTOGRAPHY.md": """# 13 CRYPTOGRAPHY

## Implementações Reais
Usa de fato `secp256k1` (Go) para assinaturas clássicas e `AES-256-GCM` na fragmentação (verificado scripts Python / Go na nuvem). 

## Implementações Simuladas
O código contém placeholders estruturais para assinaturas Pós-Quânticas (campo `PQCSignature` vazio, mas scripts PQC Python com referências ao Kyber/Dilithium apontam para integração real em andamento). O NeonHash foi validado como real e customizado.

## Pontos Críticos
O vetor em memória inicializado no `NeonHash` depende unicamente do hash SHA-256 do record do bloco. Sem um pool dinâmico extra de estado global (como em Ethash), esse algoritmo ainda é vulnerável a precomputação agressiva, embora mais resistente a ASICs simples.
""",

    "14_NEBULA.md": """# 14 NEBULA

## Tudo Que Pertence ao Nebula Cloud
A nuvem (`nebula_cloud/node_daemon.go`) expõe endpoints HTTP (`:8000`) para upload/download de fragmentos (`.bin`). 

## Integração e Estado
A integração ocorre no `pc_node/ledger.go` que a cada 10 blocos aciona uma goroutine para enviar backups. O estado é totalmente funcional, mas altamente ingênuo: aceita uploads de qualquer um via POST (sem autenticação) desde que com no máximo 100MB de tamanho.

## Roadmap Inferido
Transição de simples armazenamento para contratos de fragmentação distribuída (Reed-Solomon) onde os doadores de storage receberão recompensas diretamente atreladas às "Proofs of Storage" em blockchain.
""",

    "15_SECURITY.md": """# 15 SECURITY

## Ataques Possíveis
- **Double Spend e Replay:** Alta probabilidade de sucesso devido à falta aparente de rastreamento de nonces por endereço de envio ou verificação aprofundada de modelo de conta vs. UTXO de maneira forte.
- **Spoofing e Flood:** Faltam filtros de rate-limit severos no `pc_node/network.go`. Um atacante conectando no TCP pode inundar o nó de blocos inválidos.
- **DDoS no Cloud Node:** Qualquer usuário pode usar o endpoint de `/upload` do `node_daemon.go` sem autenticação.
- **Integridade:** `saveLedger()` não salva para um arquivo temporário antes do rename atômico. Falha elétrica corromperá o `ledger.json` irremediavelmente.
"""
}

for filename, content in files.items():
    with open(os.path.join(out_dir, filename), "w", encoding="utf-8") as f:
        f.write(content)
