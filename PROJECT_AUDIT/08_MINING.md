# 08 MINING

## Algoritmo
Utiliza `NeonHash`. O algoritmo aloca um vetor em memória de 4096 bytes semeado a partir de um SHA-256 inicial do cabeçalho. Realiza 128 iterações de saltos pseudo-aleatórios e mutações, finalizando com um SHA-256 no vetor mutado.

## Reward
Fórmula: `CalculateBlockReward()`. Recompensa base é 50 NBL, sofrendo halving a cada 2.100.000 blocos. Existe um "Proof of Storage" bônus onde armazenar dados na Nebula Cloud rende NBL extras (0.5 por GB no SSD, 0.1 por GB no HDD).

## Candidate Block e Performance
Mineração otimizada para evitar vantagens de ASICs. O uso do vetor de 4KB exige acesso a memória que encarece ASICs enquanto mantém smartphones e CPUs normais competitivos.
