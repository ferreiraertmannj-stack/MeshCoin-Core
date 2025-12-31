# 🌐 MeshCoin (MESH) - Core Prototype

> **Conectividade Financeira Off-Grid para Áreas Remotas e IoT.**

A MeshCoin é um protocolo de Camada 0 focado em transportar valor e dados através de redes mesh (Bluetooth/LoRa) sem necessidade de conexão constante com a internet.

## 🚀 Funcionalidades Atuais (Simulação v0.1)

Este repositório contém a Prova de Conceito (PoC) em Python demonstrando:
- **NeonHash Logic:** Simulação do algoritmo de mineração PoW.
- **Mesh Wallet:** Geração de chaves públicas/privadas (Secp256k1 mock).
- **P2P Gossip:** Protocolo de chat descentralizado via arquivos locais.
- **Ledger:** Sistema de contabilidade distribuída auditável.

## 🛠 Como Rodar a Simulação

1. Clone o repositório.
2. Execute `python minerador_carteira.py` para gerar moedas.
3. Execute `python chat_mesh.py` em dois terminais para simular a rede.
4. Use `python enviar_pagamento.py` para transacionar valores.

## 📍 Roadmap
- [x] Simulação Lógica (Python)
- [ ] Implementação em Hardware (Raspberry Pi Zero 2 W)
- [ ] Integração com Bluetooth Low Energy (BLE)
- [ ] Teste de Campo em Extrema/MG

---
*Desenvolvido por Jean Ertmann.*
