# Broadcast Estático Falso-Mesh

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
