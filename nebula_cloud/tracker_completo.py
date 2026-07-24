#!/usr/bin/env python3
import json, os, hashlib, requests
from datetime import datetime
from pqc_reed_solomon import PQCReedSolomonFragmenter

NODES_FILE = "nodes.json"
META_FILE = "files_metadata.json"

def _hash_senha(s): return hashlib.sha256(s.encode()).hexdigest()[:16]
def carregar_nodes(): return json.load(open(NODES_FILE)) if os.path.exists(NODES_FILE) else {}
def carregar_meta(): return json.load(open(META_FILE)) if os.path.exists(META_FILE) else {}
def salvar_meta(m): json.dump(m, open(META_FILE,'w'), indent=2)

def listar_nodes():
    nodes = carregar_nodes()
    if not nodes: print("❌ Nenhum nó registrado!"); return
    print("\n📍 Nós:")
    for nid, info in nodes.items():
        try:
            r = requests.get(f"http://{info['ip']}:{info['port']}/health", timeout=2)
            s = "✅ Online" if r.status_code == 200 else "❌ Offline"
        except: s = "❌ Offline"
        print(f"  {s} {nid}: {info['ip']}:{info['port']}")

def distribuir_arquivo():
    fp = input("Caminho do arquivo: ").strip()
    if not os.path.exists(fp): print(f"❌ Não encontrado: {fp}"); return
    pw = input("Senha (min 8 chars): ").strip()
    if len(pw) < 8: print("❌ Senha muito curta!"); return
    nodes = carregar_nodes()
    if not nodes: print("❌ Sem nós! Use Gerenciar Nós."); return
    nos_ok = {}
    for nid, info in nodes.items():
        try:
            r = requests.get(f"http://{info['ip']}:{info['port']}/health", timeout=2)
            if r.status_code == 200: nos_ok[nid] = info
        except: pass
    if not nos_ok: print("❌ Nenhum nó online!"); return
    print(f"\n🔐 Fragmentando {fp}...")
    frag = PQCReedSolomonFragmenter(pw)
    result = frag.fragmentar(fp)
    fragments = result['fragments']
    fhash = result['metadata']['file_hash']
    print(f"✅ {len(fragments)} fragmentos | Hash: {fhash[:32]}...")
    node_list = list(nos_ok.items())
    dist, ok = {}, 0
    for i, (fid, fdata) in enumerate(fragments.items()):
        nid, ninfo = node_list[i % len(node_list)]
        try:
            r = requests.post(f"http://{ninfo['ip']}:{ninfo['port']}/store",
                json={'fragment_id': fid, 'content': fdata['data']}, timeout=5)
            if r.status_code == 200:
                print(f"  ✅ {fid} → {nid}")
                dist[fid] = {'node_id': nid, 'chunk_idx': fdata['chunk_idx'],
                    'frag_idx': fdata['frag_idx'], 'total_frags': fdata['total_frags']}
                ok += 1
            else: print(f"  ❌ {fid} → erro {r.status_code}")
        except Exception as e: print(f"  ❌ {fid} → {e}")
    if ok > 0:
        meta = carregar_meta()
        fid_key = os.path.basename(fp).replace('.','_')
        meta[fid_key] = {'original_name': os.path.basename(fp), 'password_hint': _hash_senha(pw),
            'file_hash': fhash, 'fragments': dist, 'total_fragments': len(fragments),
            'success_fragments': ok, 'encryption': 'AES-256-GCM', 'kdf': 'PBKDF2-SHA256-480k',
            'reed_solomon': '10+4', 'timestamp': datetime.now().isoformat()}
        salvar_meta(meta)
        print(f"\n✅ Distribuído! ({ok}/{len(fragments)} fragmentos)")
    else: print("❌ Falha total na distribuição!")

def recuperar_arquivo_menu():
    meta = carregar_meta()
    if not meta: print("❌ Nenhum arquivo armazenado!"); return
    print("\n📂 Arquivos:")
    fl = list(meta.items())
    for i, (k, v) in enumerate(fl, 1):
        print(f"  {i}. {k} ({v['success_fragments']}/{v['total_fragments']} frags) - {v.get('encryption','AES-256-GCM')}")
    try: choice = int(input("\nEscolha: ")) - 1
    except: print("❌ Inválido!"); return
    if choice < 0 or choice >= len(fl): print("❌ Fora do range!"); return
    fid_key, info = fl[choice]
    pw = input("Senha: ").strip()
    if _hash_senha(pw) != info.get('password_hint',''):
        print("❌ Senha incorreta!"); return
    out = input("Caminho de saída: ").strip()
    nodes = carregar_nodes()
    frags = {}
    print("\n📥 Recuperando...")
    for fid, finfo in info['fragments'].items():
        nid = finfo['node_id']
        if nid not in nodes: print(f"  ❌ {fid} - nó {nid} não encontrado"); continue
        ninfo = nodes[nid]
        try:
            r = requests.get(f"http://{ninfo['ip']}:{ninfo['port']}/retrieve/{fid}", timeout=5)
            if r.status_code == 200:
                d = r.json()
                frags[fid] = {'data': d['content'], 'chunk_idx': finfo['chunk_idx'],
                    'frag_idx': finfo['frag_idx'], 'total_frags': finfo['total_frags']}
                print(f"  ✅ {fid} ← {nid}")
            else: print(f"  ❌ {fid} - erro {r.status_code}")
        except Exception as e: print(f"  ❌ {fid} - {e}")
    if len(frags) >= info['total_fragments']:
        ok = PQCReedSolomonFragmenter(pw).recuperar(frags, out)
        print(f"\n✅ Recuperado: {out}" if ok else "\n❌ Erro na reconstrução!")
    else: print(f"\n❌ Fragmentos insuficientes ({len(frags)}/{info['total_fragments']})")

def gerenciar_nodes():
    while True:
        print("\n⚙️  NÓS\n1-Adicionar 2-Remover 3-Listar 4-Voltar")
        op = input("Opção: ").strip()
        if op == "1":
            nid = input("ID do nó: ").strip()
            ip = input("IP: ").strip()
            port = input("Porta (5000): ").strip() or "5000"
            nodes = carregar_nodes()
            nodes[nid] = {'ip': ip, 'port': int(port), 'status': 'online'}
            json.dump(nodes, open(NODES_FILE,'w'), indent=2)
            print(f"✅ {nid} registrado em {ip}:{port}")
        elif op == "2":
            nid = input("ID a remover: ").strip()
            nodes = carregar_nodes()
            if nid in nodes:
                del nodes[nid]
                json.dump(nodes, open(NODES_FILE,'w'), indent=2)
                print(f"✅ {nid} removido!")
            else: print("❌ Não encontrado!")
        elif op == "3": listar_nodes()
        elif op == "4": break

def menu():
    while True:
        print("\n" + "="*50)
        print("🚀 NÉBULA CLOUD TRACKER v1.1")
        print("="*50)
        print("1-Distribuir 2-Recuperar 3-Listar 4-Nós 5-Gerenciar Nós 6-Sair")
        op = input("\nOpção: ").strip()
        if op=="1": distribuir_arquivo()
        elif op=="2": recuperar_arquivo_menu()
        elif op=="3":
            meta = carregar_meta()
            if not meta: print("❌ Nenhum arquivo!"); continue
            print(f"\n{'Nome':<25} {'Frags':<10} {'Data':<12}")
            print("-"*50)
            for k,v in meta.items():
                print(f"{k:<25} {v['success_fragments']}/{v['total_fragments']:<8} {v['timestamp'][:10]:<12}")
        elif op=="4": listar_nodes()
        elif op=="5": gerenciar_nodes()
        elif op=="6": print("\n👋 Até logo!"); break
        else: print("❌ Opção inválida!")

if __name__ == '__main__':
    menu()
