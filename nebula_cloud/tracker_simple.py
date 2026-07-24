#!/usr/bin/env python3
"""
Nébula Cloud - Tracker Simplificado
Gerencia distribuição de fragmentos
"""
import json
import os
import requests
from fragmenter_simple import fragmentar_arquivo

NODES_FILE = "nodes.json"

def carregar_nodes():
    """Carrega lista de nós"""
    if os.path.exists(NODES_FILE):
        with open(NODES_FILE, 'r') as f:
            return json.load(f)
    return {}

def listar_nodes():
    """Lista nós online"""
    nodes = carregar_nodes()
    if not nodes:
        print("❌ Nenhum nó registrado!")
        return
    
    print("\n📍 Nós Online:")
    for node_id, info in nodes.items():
        print(f"  ✅ {node_id}: {info['ip']}:{info['port']}")

def distribuir_arquivo():
    """Distribui arquivo entre nós"""
    filepath = input("Caminho do arquivo: ").strip()
    if not os.path.exists(filepath):
        print(f"❌ Arquivo não encontrado: {filepath}")
        return
    
    password = input("Senha de criptografia: ").strip()
    
    nodes = carregar_nodes()
    if not nodes:
        print("❌ Nenhum nó disponível!")
        return
    
    # Fragmentar arquivo
    print(f"\n🔐 Fragmentando {filepath}...")
    fragments = fragmentar_arquivo(filepath, password)
    print(f"✅ {len(fragments)} fragmentos criados")
    
    # Distribuir entre nós
    node_list = list(nodes.items())
    for i, (frag_id, frag_data) in enumerate(fragments.items()):
        node_id, node_info = node_list[i % len(node_list)]
        url = f"http://{node_info['ip']}:{node_info['port']}/store"
        
        try:
            resp = requests.post(url, json={
                'fragment_id': frag_id,
                'content': frag_data
            }, timeout=5 )
            if resp.status_code == 200:
                print(f"  ✅ {frag_id} → {node_id}")
            else:
                print(f"  ❌ {frag_id} → {node_id} (erro {resp.status_code})")
        except Exception as e:
            print(f"  ❌ {frag_id} → {node_id} (erro: {e})")

def menu():
    """Menu principal"""
    while True:
        print("\n🚀 NÉBULA CLOUD TRACKER v0.2")
        print("1 - Distribuir arquivo")
        print("2 - Listar nós")
        print("3 - Sair")
        
        opcao = input("\nEscolha uma opção: ").strip()
        
        if opcao == "1":
            distribuir_arquivo()
        elif opcao == "2":
            listar_nodes()
        elif opcao == "3":
            break
        else:
            print("❌ Opção inválida!")

if __name__ == '__main__':
    menu()
