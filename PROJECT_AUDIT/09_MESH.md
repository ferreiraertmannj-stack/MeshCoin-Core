# 09 MESH

## Discovery e Routing
A descoberta na rede local é feita com Broadcasts UDP na porta 5555. Nós anunciam `NEBULA_NODE:<portaTCP>` a cada 5 segundos. 

## Multi-hop e Store and Forward
Embora a documentação fale muito de protocolo B.A.T.M.A.N. e comunicações off-grid com saltos, a análise do nó Go (pc_node/network.go) revela um encaminhamento simples via TCP. Mensagens (Blocos, Transações e Chat) são recebidas, validadas e redistribuídas (broadcast TCP/WS).

## Bugs
O modelo de roteamento atual envia de volta para as conexões TCP ativas. A falta de verificação avançada de roteamento (ex: IDs de pacotes únicos) pode levar a loops de broadcast infinitos na rede local (packet storm) caso haja loops na topologia e as validações (ex: mempool já tem a tx, ou ledger já tem o bloco) não abortarem a propagação rápido o suficiente.
