#!/usr/bin/env python3
import os, json, hashlib
from reedsolo import RSCodec
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

K, M = 10, 4
SALT = b'nebula_cloud_pqc_2026'

def _derivar_chave(senha):
    return hashlib.pbkdf2_hmac('sha256', senha.encode(), SALT, 480000)[:32]

def criptografar(dados, senha):
    chave = _derivar_chave(senha)
    nonce = os.urandom(12)
    return AESGCM(chave).encrypt(nonce, dados, b'nebula_auth'), nonce

def descriptografar(dados_enc, nonce, senha):
    return AESGCM(_derivar_chave(senha)).decrypt(nonce, dados_enc, b'nebula_auth')

def fragmentar_arquivo(caminho, senha, pasta_saida):
    os.makedirs(pasta_saida, exist_ok=True)
    dados = open(caminho, 'rb').read()
    file_hash = hashlib.sha256(dados).hexdigest()
    dados_enc, nonce = criptografar(dados, senha)
    resto = len(dados_enc) % K
    if resto: dados_enc += b'\x00' * (K - resto)
    tam = len(dados_enc) // K
    rsc = RSCodec(M)
    for i in range(K):
        frag = dados_enc[i*tam:(i+1)*tam]
        open(f"{pasta_saida}/frag_{i:02d}.bin", 'wb').write(bytes(rsc.encode(frag)))
    json.dump({"arquivo_original": os.path.basename(caminho), "file_hash": file_hash,
        "tamanho_original": len(dados), "nonce": nonce.hex(), "tam_fragmento": tam,
        "fragmentos_dados": K, "encryption": "AES-256-GCM", "kdf": "PBKDF2-SHA256-480k"},
        open(f"{pasta_saida}/meta.json", 'w'), indent=2)
    print(f"✅ {K+M} fragmentos | Hash: {file_hash[:32]}... | KDF: PBKDF2-SHA256-480k")

def reconstruir_arquivo(pasta, senha, saida):
    meta = json.load(open(f"{pasta}/meta.json"))
    rsc = RSCodec(M)
    frags = [bytearray(open(f"{pasta}/frag_{i:02d}.bin", 'rb').read()) for i in range(meta['fragmentos_dados'])]
    dados_enc = b''.join(bytes(rsc.decode(f)[0]) for f in frags)[:meta['tamanho_original']]
    dados = descriptografar(dados_enc, bytes.fromhex(meta['nonce']), senha)
    if hashlib.sha256(dados).hexdigest() != meta['file_hash']:
        raise ValueError("❌ Integridade comprometida!")
    open(saida, 'wb').write(dados)
    print(f"✅ Reconstruído: {saida} | Integridade: VERIFICADA ✓")
