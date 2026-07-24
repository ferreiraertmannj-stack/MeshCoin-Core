#!/usr/bin/env python3
"""Registra um nó no Tracker"""
import json
import os

NODES_FILE = "nodes.json"

# Carregar nós existentes
if os.path.exists(NODES_FILE):
    with open(NODES_FILE, 'r') as f:
        nodes = json.load(f)
else:
    nodes = {}

# Adicionar novo nó
node_id = "infinix_hot_50"
node_ip = "192.168.0.102"
node_port = 5000

nodes[node_id] = {
    "ip": node_ip,
    "port": node_port,
    "status": "online",
    "storage_capacity": 10000  # 10GB
}

# Salvar
with open(NODES_FILE, 'w') as f:
    json.dump(nodes, f, indent=2)

print(f"✅ Nó {node_id} registrado em {node_ip}:{node_port}")
print(json.dumps(nodes, indent=2))
