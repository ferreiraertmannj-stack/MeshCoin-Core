# 09 RECOVERY

### Cold Boot
- O Nó inicia, lê JSON e confia inteiramente no conteúdo local. Não há mecanismo formal atual implementado de "Verificar a integridade hash por hash do Bloco 0 ao Bloco N na subida". Se um atacante alterar o JSON localmente na máquina do minerador ajustando saldos de blocos passados, o nó carregará normalmente.

### Chain Fork Resolution
- Totalmente ausente na arquitetura interna do Ledger atual. O bloco é apenas aceito se `block.PreviousHash == lastBlock.Hash`. Se dois blocos válidos da mesma altura chegarem em instantes diferentes, o primeiro ganha. Não há armazenamento lateral (Orphan Blocks/Forks) nem função de Reorg (Retroceder X blocos e adotar uma corrente mais longa).
