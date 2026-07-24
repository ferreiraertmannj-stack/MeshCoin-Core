#!/usr/bin/env python3
"""
Fragmentador simples para Nébula Cloud
Divide arquivo em fragmentos com criptografia
"""
import hashlib
import os
from cryptography.fernet import Fernet
import base64

def gerar_chave(password):
    """Gera chave a partir da senha"""
    salt = b'nebula_cloud_2026'
    key = hashlib.pbkdf2_hmac('sha256', password.encode(), salt, 100000)
    return base64.urlsafe_b64encode(key[:32])

def fragmentar_arquivo(filepath, password, chunk_size=1024*100):
    """Fragmenta arquivo em pedaços criptografados"""
    if not os.path.exists(filepath):
        raise FileNotFoundError(f"Arquivo não encontrado: {filepath}")
    
    chave = gerar_chave(password)
    cipher = Fernet(chave)
    fragments = {}
    
    with open(filepath, 'rb') as f:
        chunk_num = 0
        while True:
            chunk = f.read(chunk_size)
            if not chunk:
                break
            
            encrypted = cipher.encrypt(chunk)
            frag_id = f"frag_{os.path.basename(filepath)}_{chunk_num:04d}"
            fragments[frag_id] = encrypted.decode('utf-8')
            chunk_num += 1
    
    return fragments

def recuperar_arquivo(fragments, password, output_path):
    """Recupera arquivo a partir de fragmentos"""
    chave = gerar_chave(password)
    cipher = Fernet(chave)
    
    # Ordenar fragmentos
    sorted_frags = sorted(fragments.items(), key=lambda x: int(x[0].split('_')[-1]))
    
    with open(output_path, 'wb') as f:
        for frag_id, encrypted_data in sorted_frags:
            decrypted = cipher.decrypt(encrypted_data.encode('utf-8'))
            f.write(decrypted)
    
    return output_path

if __name__ == '__main__':
    # Teste
    test_file = "test.txt"
    with open(test_file, 'w') as f:
        f.write("Teste de fragmentação Nébula Cloud")
    
    frags = fragmentar_arquivo(test_file, "senha123")
    print(f"✅ {len(frags)} fragmentos criados")
    for frag_id in list(frags.keys())[:3]:
        print(f"  - {frag_id}")
