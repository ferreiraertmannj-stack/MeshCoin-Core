# 11 STORAGE

## Banco e Persistência
Não se utiliza um SGBD. A persistência baseia-se em serialização de structs para JSON local (`ledger.json`). 

## Snapshots e Recovery
O arquivo Go `pc_node/ledger.go` lê o JSON do disco no `initLedger()`. Se o arquivo estiver corrompido, ele emite um log e silenciosamente **recria um novo Bloco Gênesis**. Isso constitui um risco massivo de perda de dados.

## Serialização
O uso de `json.MarshalIndent` no salvamento (`saveLedger()`) é lento para bancos grandes e não escala para um ledger real com milhões de blocos, pois exige alocação do arquivo inteiro na RAM.
