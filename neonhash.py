import hashlib
import time
import struct
import array

class NeonHash:
    """
    Simulação do NeonHash: Um algoritmo de Proof-of-Work focado em ARM (Smartphones).
    Ele tenta ser resistente a ASICs (Application-Specific Integrated Circuits) 
    exigindo alocação pesada de memória (Scratchpad) e acessos aleatórios.
    """
    
    def __init__(self, scratchpad_size_mb=10):
        # Para testes locais em Python, manteremos o scratchpad pequeno (10MB). 
        # Num celular real (C++), usaremos 50MB a 200MB.
        self.scratchpad_size = scratchpad_size_mb * 1024 * 1024
        self.scratchpad = None
        
    def _initialize_scratchpad(self, seed_block):
        """Preenche a memória RAM com dados pseudo-aleatórios pesados."""
        # Cria um array de bytes do tamanho do scratchpad (preenchido com zeros)
        self.scratchpad = bytearray(self.scratchpad_size)
        
        # Semente baseada no bloco anterior para garantir que cada bloco tenha uma memória diferente
        seed = hashlib.sha512(seed_block.encode('utf-8')).digest()
        
        # Expande a semente por todo o scratchpad (simulação de gargalo de memória)
        for i in range(0, self.scratchpad_size, 64):
            seed = hashlib.sha512(seed + struct.pack('<I', i)).digest()
            end = min(i + 64, self.scratchpad_size)
            self.scratchpad[i:end] = seed[:end - i]
            
    def compute_hash(self, block_data, nonce):
        """Calcula o hash realizando saltos aleatórios no scratchpad."""
        if self.scratchpad is None:
            raise ValueError("Scratchpad não inicializado. Chame _initialize_scratchpad(seed) antes.")
            
        # O estado inicial depende dos dados do bloco + nonce
        mix = hashlib.sha256(f"{block_data}{nonce}".encode('utf-8')).digest()
        
        # Realiza 1000 acessos pseudo-aleatórios à memória (Memory-Hardness)
        # Em ASICs, acessar RAM de forma aleatória quebra a pipeline de execução.
        state = int.from_bytes(mix[:8], 'little')
        
        for _ in range(1000):
            # Calcula um índice no scratchpad baseado no estado atual
            index = state % (self.scratchpad_size - 64)
            
            # Lê 64 bytes do scratchpad
            data_read = self.scratchpad[index:index+64]
            
            # Mistura os dados lidos com o estado atual usando XOR e Hash
            # (Em C++ com ARM NEON, isso seria feito com instruções SIMD vetorizadas de 128-bits)
            mix_hash = hashlib.sha256(mix + data_read).digest()
            
            # Atualiza o estado
            state = int.from_bytes(mix_hash[:8], 'little')
            mix = mix_hash
            
        return mix.hex()
        
    def mine_block(self, seed_block, block_data, target_prefix="0000"):
        print(f"🧠 Inicializando NeonHash Scratchpad (Alocando RAM)...")
        inicio = time.time()
        self._initialize_scratchpad(seed_block)
        print(f"✅ Scratchpad pronto em {round(time.time() - inicio, 2)}s.")
        
        print("🔨 Iniciando Mineração NeonHash (Resistente a ASICs)...")
        nonce = 0
        inicio_mineracao = time.time()
        
        while True:
            hash_result = self.compute_hash(block_data, nonce)
            
            if nonce % 1000 == 0:
                print(f"Tentando Nonce: {nonce} | Hash: {hash_result[:16]}...")
                
            if hash_result.startswith(target_prefix):
                fim = time.time()
                print(f"\n🎉 BLOCO MINERADO COM NEONHASH!")
                print(f"Nonce: {nonce}")
                print(f"Hash: {hash_result}")
                print(f"Tempo total: {round(fim - inicio_mineracao, 2)}s")
                return nonce, hash_result
                
            nonce += 1

if __name__ == "__main__":
    minerador = NeonHash(scratchpad_size_mb=5) # 5MB para teste rápido
    # Tenta minerar um bloco
    minerador.mine_block(seed_block="bloco_genesis", block_data="tx_joao_para_maria_50MESH", target_prefix="000")
