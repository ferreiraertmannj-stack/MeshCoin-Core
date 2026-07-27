# 06 REGRESSION MATRIX

| Bug Identificado | Área de Correção | Módulos em Risco de Quebra (Regressão) |
|---|---|---|
| **Ledger Atomicity** | Refatoração IO | Parsing de saldo na Wallet (App Flutter), Validação retroativa de assinaturas antigas. |
| **Network DDoS** | Filtros TCP / Upload Auth | Sincronização de peers legítimos que falhem novos handshakes, PC node perdendo sync. |
| **Global Lock** | Ledger.go (Channels) | Mempool perdendo/duplicando TX, Data Races fantasma na hora do mineiro ler a Mempool para fechar bloco. |
| **JSON Storage** | N/A (Migração BD) | Estrutura de dados `Block` serializável do Go pode perder tags ou chaves caso o codec mude, quebrando P2P. |
| **False Mesh** | Network.go (P2P routing) | Dispositivos mobile em LAN podem parar de se "enxergar" se o UDP simples for desligado cedo demais para adotar DHT pesado. |
