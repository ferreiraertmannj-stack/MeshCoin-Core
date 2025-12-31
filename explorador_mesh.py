import json
import os

ARQUIVO_REDE = "rede_mesh_publica.txt"

def auditar_rede():
    print("==========================================")
    print("      🔍  MESHCOIN BLOCK EXPLORER  🔍     ")
    print("==========================================")
    
    if not os.path.exists(ARQUIVO_REDE):
        print("A rede ainda não tem registros.")
        return

    transacoes_encontradas = 0
    # Dicionário para guardar o saldo de todo mundo
    saldos = {} 

    print("\n📜 HISTÓRICO DE TRANSAÇÕES:")
    print("-" * 50)

    with open(ARQUIVO_REDE, "r", encoding="utf-8") as f:
        for linha in f:
            # Procura apenas as linhas que são transações
            if "[SISTEMA] NOVA TRANSAÇÃO:" in linha:
                transacoes_encontradas += 1
                # Limpa o texto para pegar só o JSON (a parte entre chaves)
                parte_json = linha.split("NOVA TRANSAÇÃO: ")[1].strip()
                
                try:
                    dados = json.loads(parte_json)
                    remetente = dados['remetente'][:15] + "..." # Encurta o nome
                    destinatario = dados['destinatario']
                    valor = float(dados['valor'])
                    
                    print(f"💸 {remetente} enviou {valor} MESH para -> {destinatario}")
                    
                    # Atualiza os saldos (Lógica Contábil Simples)
                    # Quem recebe, ganha (+)
                    saldos[destinatario] = saldos.get(destinatario, 0) + valor
                    # Quem envia, a gente registra como saída (apenas informativo aqui)
                    # (Numa blockchain real, checaríamos se o remetente tinha saldo antes)
                    
                except:
                    print(f"⚠️ Erro ao ler bloco: {linha[:20]}...")

    print("-" * 50)
    print(f"\n📊 RESUMO DE SALDOS (Estado Atual da Rede):")
    
    if not saldos:
        print("Nenhum saldo movimentado ainda.")
    else:
        for usuario, total in saldos.items():
            print(f"💰 {usuario}: Possui {total} MESH")

    print(f"\nTotal de Transações Processadas: {transacoes_encontradas}")

if __name__ == "__main__":
    auditar_rede()