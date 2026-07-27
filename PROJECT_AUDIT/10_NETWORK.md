# 10 NETWORK

## TCP e UDP
- **UDP (Porta 5555)**: Usado exclusivamente para descoberta local contínua via broadcast (255.255.255.255).
- **TCP (Porta 5556)**: Usado para transporte confiável de `NEW_BLOCK`, `NEW_TRANSACTION`, e pacotes de mensagens do tipo `DATA_ROUTE` ou `CHAT`.

## Bluetooth e Wi-Fi Direct
A implementação no Go não abrange o rádio local. A delegação dessas funções depende inteiramente da implementação em Dart no aplicativo móvel Flutter, o qual atuaria como uma ponte (bridge) entre essas redes e a rede TCP local do PC.

## Reconexão e Heartbeat
Há mensagens `PING` e `OGM` descritas no router TCP, tratadas de forma silenciosa para manutenção de sockets abertos. 
