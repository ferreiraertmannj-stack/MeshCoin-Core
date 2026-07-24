#!/usr/bin/env python3
"""
Nébula Cloud - Tracker com PQC v1.1
Criptografia AES-256-GCM + PBKDF2-SHA256
"""
import json
import os
import requests
from pqc_fragmenter import PQCFragmenter
from datetime import datetime

NODES_FILE = "nodes.json"
METADATA_FILE = "files_metadata.json"

def carregar_nodes():
    if os.path.exists(NODES_FILE):
        with open(NODES_FILE, 'r') as f:
            return json.load(f)
    return {}

def carregar_metadata():
    if os.path.exists(METADATA_FILE):
        with open(METADATA_FILE, 'r') as f:
            return json.load(f)
    return {}

def salvar_metadata(metadata):
    with open(METADATA_FILE, 'w') as f:
        json.dump(metadata, f, indent=2)

def listar_nodes():
    nodes = carregar_nodes()
    if not nodes:
        print("❌ Nenhum nó registrado!")
        return
    
    print("\n📍 Nós Online:")
    for node_id, info in nodes.items():
        try:
            resp = requests.get(f"http://{info['ip']}:{info['port']}/health", timeout=2 )
            status = "✅ Online" if resp.status_code == 200 else "❌ Offline"
        except:
            status = "❌ Offline"
        print(f"  {status} {node_id}: {info['ip']}:{info['port']}")

def distribuir_arquivo():
    filepath = input("Caminho do arquivo: ").strip()
    if not os.path.exists(filepath):
        print(f"❌ Arquivo não encontrado: {filepath}")
        return
    
    password = input("Senha de criptografia: ").strip()
    
    nodes = carregar_nodes()
    if not nodes:
        print("❌ Nenhum nó disponível!")
        return
    
    print(f"\n🔐 Fragmentando {filepath} com AES-256-GCM...")
    fragmenter = PQCFragmenter(password)
    result = fragmenter.fragmentar(filepath)
    fragments = result['fragments']
    
    print(f"✅ {len(fragments)} fragmentos criados (criptografados)")
    
    node_list = list(nodes.items())
    distribution = {}
    success_count = 0
    
    for i, (frag_id, frag_data) in enumerate(fragments.items()):
        node_id, node_info = node_list[i % len(node_list)]
        url = f"http://{node_info['ip']}:{node_info['port']}/store"
        
        try:
            resp = requests.post(url, json={
                'fragment_id': frag_id,
                'content': json.dumps(frag_data )  # JSON serializado
            }, timeout=5)
            if resp.status_code == 200:
                print(f"  ✅ {frag_id} → {node_id}")
                distribution[frag_id] = node_id
                success_count += 1
            else:
                print(f"  ❌ {frag_id} → {node_id} (erro {resp.status_code})")
        except Exception as e:
            print(f"  ❌ {frag_id} → {node_id} (erro: {e})")
    
    if success_count > 0:
        metadata = carregar_metadata()
        file_id = os.path.basename(filepath).replace('.', '_')
        metadata[file_id] = {
            'original_path': filepath,
            'fragments': distribution,
            'total_fragments': len(fragments),
            'success_fragments': success_count,
            'file_hash': result['metadata']['file_hash'],
            'encryption': 'AES-256-GCM',
            'kdf': 'PBKDF2-SHA256-480k',
            'timestamp': datetime.now().isoformat()
        }
        salvar_metadata(metadata)
        print(f"\n✅ Arquivo distribuído com sucesso! ({success_count}/{len(fragments)} fragmentos)")
        print(f"   Hash: {result['metadata']['file_hash'][:16]}...")
    else:
        print(f"\n❌ Falha ao distribuir arquivo!")

def recuperar_arquivo_menu():
    metadata = carregar_metadata()
    
    if not metadata:
        print("❌ Nenhum arquivo armazenado!")
        return
    
    print("\n📂 Arquivos Armazenados:")
    files_list = list(metadata.items())
    for i, (file_id, info) in enumerate(files_list, 1):
        print(f"  {i}. {file_id} ({info['success_fragments']}/{info['total_fragments']} fragmentos) - {info['encryption']}")
    
    try:
        choice = int(input("\nEscolha um arquivo: ")) - 1
        if choice < 0 or choice >= len(files_list):
            print("❌ Opção inválida!")
            return
    except:
        print("❌ Entrada inválida!")
        return
    
    file_id, info = files_list[choice]
    password = input("Senha de descriptografia: ").strip()
    output_path = input("Caminho de saída: ").strip()
    
    nodes = carregar_nodes()
    fragments = {}
    
    print(f"\n📥 Recuperando fragmentos...")
    for frag_id, node_id in info['fragments'].items():
        if node_id not in nodes:
            print(f"  ❌ {frag_id} - Nó {node_id} não encontrado")
            continue
        
        node_info = nodes[node_id]
        url = f"http://{node_info['ip']}:{node_info['port']}/retrieve/{frag_id}"
        
        try:
            resp = requests.get(url, timeout=5 )
            if resp.status_code == 200:
                data = resp.json()
                fragments[frag_id] = json.loads(data['content'])
                print(f"  ✅ {frag_id} ← {node_id}")
            else:
                print(f"  ❌ {frag_id} - Erro {resp.status_code}")
        except Exception as e:
            print(f"  ❌ {frag_id} - Erro: {e}")
    
    if len(fragments) == info['total_fragments']:
        try:
            fragmenter = PQCFragmenter(password)
            fragmenter.recuperar(fragments, output_path)
            print(f"\n✅ Arquivo recuperado em: {output_path}")
        except Exception as e:
            print(f"\n❌ Erro ao recuperar arquivo: {e}")
    else:
        print(f"\n❌ Não foi possível recuperar todos os fragmentos ({len(fragments)}/{info['total_fragments']})")

def listar_arquivos():
    metadata = carregar_metadata()
    
    if not metadata:
        print("❌ Nenhum arquivo armazenado!")
        return
    
    print("\n📂 Arquivos Armazenados:")
    print(f"{'ID':<20} {'Fragmentos':<15} {'Criptografia':<20} {'Data':<20}")
    print("-" * 75)
    
    for file_id, info in metadata.items():
        frag_status = f"{info['success_fragments']}/{info['total_fragments']}"
        timestamp = info['timestamp'][:10]
        encryption = info.get('encryption', 'N/A')
        print(f"{file_id:<20} {frag_status:<15} {encryption:<20} {timestamp:<20}")

def menu():
    while True:
        print("\n" + "="*60)
        print("🚀 NÉBULA CLOUD TRACKER v1.1 - Com Criptografia PQC")
        print("="*60)
        print("1 - Distribuir arquivo (AES-256-GCM)")
        print("2 - Recuperar arquivo")
        print("3 - Listar arquivos armazenados")
        print("4 - Ver nós online")
        print("5 - Sair")
        print("="*60)
        
        opcao = input("\nEscolha uma opção: ").strip()
        
        if opcao == "1":
            distribuir_arquivo()
        elif opcao == "2":
            recuperar_arquivo_menu()
        elif opcao == "3":
            listar_arquivos()
        elif opcao == "4":
            listar_nodes()
        elif opcao == "5":
            print("\n👋 Até logo!")
            break
        else:
            print("❌ Opção inválida!")

if __name__ == '__main__':
    menu()
