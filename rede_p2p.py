import socket
import threading
import json
import time

# Configurações da Rede Mesh P2P
UDP_PORT = 5555
TCP_PORT = 5556
MAGIC_WORD = "MESHCOIN_NODE"

class MeshNode:
    def __init__(self, node_name="Node"):
        self.node_name = node_name
        self.peers = set() # Lista de IPs (e portas) conhecidos
        self.callbacks = []
        
        self.running = False

    def start(self):
        self.running = True
        
        # Thread para descobrir outros nós na rede (UDP)
        threading.Thread(target=self._udp_broadcast_sender, daemon=True).start()
        threading.Thread(target=self._udp_broadcast_listener, daemon=True).start()
        
        # Thread para escutar mensagens diretas (TCP)
        threading.Thread(target=self._tcp_listener, daemon=True).start()
        
        print(f"📡 [Rede P2P] Nó '{self.node_name}' iniciado na rede local.")

    def on_message(self, callback):
        """Registra uma função para ser chamada quando chegar mensagem"""
        self.callbacks.append(callback)

    def _udp_broadcast_sender(self):
        """Envia um 'Olá' a cada 5 segundos para a rede local"""
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_BROADCAST, 1)
        
        while self.running:
            try:
                mensagem = f"{MAGIC_WORD}:{TCP_PORT}"
                # Broadcast para a rede local
                sock.sendto(mensagem.encode('utf-8'), ('<broadcast>', UDP_PORT))
            except Exception as e:
                pass
            time.sleep(5)
            
    def _udp_broadcast_listener(self):
        """Fica escutando 'Olá' de outros nós na rede"""
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        
        # Permitir reutilizar endereço
        try:
            sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        except AttributeError:
            pass
            
        sock.bind(('', UDP_PORT))
        
        while self.running:
            try:
                data, addr = sock.recvfrom(1024)
                ip_peer = addr[0]
                
                texto = data.decode('utf-8')
                if texto.startswith(MAGIC_WORD):
                    porta_tcp = int(texto.split(":")[1])
                    peer_id = (ip_peer, porta_tcp)
                    
                    if peer_id not in self.peers:
                        self.peers.add(peer_id)
            except Exception as e:
                pass

    def _tcp_listener(self):
        """Escuta conexões TCP para receber dados (transações, blocos, chat)"""
        server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        
        # Tenta conectar na porta padrão. Se estiver em uso, tenta a próxima
        porta = TCP_PORT
        while True:
            try:
                server.bind(('', porta))
                break
            except OSError:
                porta += 1
                
        server.listen(5)
        # Atualiza a porta TCP real que estamos usando
        global TCP_PORT
        TCP_PORT = porta
        
        while self.running:
            try:
                client_sock, addr = server.accept()
                threading.Thread(target=self._handle_client, args=(client_sock,), daemon=True).start()
            except Exception as e:
                pass

    def _handle_client(self, client_sock):
        try:
            data = client_sock.recv(4096)
            if data:
                pacote = json.loads(data.decode('utf-8'))
                # Chama os callbacks registrados
                for callback in self.callbacks:
                    callback(pacote)
        except Exception as e:
            pass
        finally:
            client_sock.close()

    def broadcast_data(self, data_dict):
        """Envia um dicionário (JSON) para todos os nós conhecidos"""
        pacote_json = json.dumps(data_dict).encode('utf-8')
        
        peers_removidos = set()
        
        for peer in self.peers:
            ip, porta = peer
            try:
                s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                s.settimeout(2)
                s.connect((ip, porta))
                s.sendall(pacote_json)
                s.close()
            except Exception as e:
                # Se falhar, removemos o nó inativo
                peers_removidos.add(peer)
                
        for peer in peers_removidos:
            self.peers.remove(peer)
