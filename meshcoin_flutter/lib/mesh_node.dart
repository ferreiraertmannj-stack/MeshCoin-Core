import 'dart:io';
import 'dart:convert';
import 'dart:async';
import 'batman_router.dart';

class MeshNode {
  static const int udpPort = 5555;
  static const int tcpPort = 5556;
  static const String magicWord = "MESHCOIN_NODE";

  final String nodeName;
  final Set<String> directPeers = {}; // Vizinhos de um pulo (Diretos)
  late BatmanRouter router;
  
  bool isBridgeNode = false; // Se for True, esse celular tem acesso 4G/Wi-Fi On-Grid
  final String globalServerUrl = "http://meshcoin.global/api/sync"; 
  
  final List<Function(Map<String, dynamic>)> onMessageCallbacks = [];
  
  bool isRunning = false;
  ServerSocket? _serverSocket;
  RawDatagramSocket? _udpSocket;

  MeshNode(this.nodeName) {
    router = BatmanRouter(nodeName);
  }

  Future<void> start() async {
    isRunning = true;
    _startUdpListener();
    _startUdpBroadcaster();
    _startTcpListener();
    _startOgmBroadcaster(); // B.A.T.M.A.N OGM
  }

  void stop() {
    isRunning = false;
    _serverSocket?.close();
    _udpSocket?.close();
  }

  void onMessage(Function(Map<String, dynamic>) callback) {
    onMessageCallbacks.add(callback);
  }

  Future<void> _startUdpListener() async {
    try {
      _udpSocket = await RawDatagramSocket.bind(InternetAddress.anyIPv4, udpPort);
      _udpSocket?.broadcastEnabled = true;
      
      _udpSocket?.listen((RawSocketEvent event) {
        if (event == RawSocketEvent.read) {
          Datagram? dg = _udpSocket?.receive();
          if (dg != null) {
            String msg = utf8.decode(dg.data);
            if (msg.startsWith(magicWord)) {
              String ip = dg.address.address;
              int port = int.parse(msg.split(":")[1]);
              String peerId = "$ip:$port";
              if (!directPeers.contains(peerId)) {
                directPeers.add(peerId);
              }
            }
          }
        }
      });
    } catch (e) {
      print("UDP Listener Error: \$e");
    }
  }

  Future<void> _startUdpBroadcaster() async {
    Timer.periodic(Duration(seconds: 5), (timer) {
      if (!isRunning) {
        timer.cancel();
        return;
      }
      try {
        String msg = "\$magicWord:\$tcpPort";
        List<int> data = utf8.encode(msg);
        _udpSocket?.send(data, InternetAddress("255.255.255.255"), udpPort);
      } catch (e) {}
    });
  }

  /// Inicia o Broadcast de Mensagens OGM do protocolo B.A.T.M.A.N
  Future<void> _startOgmBroadcaster() async {
    Timer.periodic(Duration(seconds: 10), (timer) {
      if (!isRunning) return;
      router.cleanStaleRoutes();
      Map<String, dynamic> ogm = router.generateOGM();
      _sendToAllDirectPeers(ogm);
    });
  }

  Future<void> _startTcpListener() async {
    try {
      _serverSocket = await ServerSocket.bind(InternetAddress.anyIPv4, tcpPort);
      _serverSocket?.listen((Socket client) {
        String senderIp = client.remoteAddress.address;
        
        client.listen((List<int> data) {
          try {
            String jsonStr = utf8.decode(data);
            Map<String, dynamic> packet = jsonDecode(jsonStr);
            
            if (packet["tipo"] == "OGM") {
              // B.A.T.M.A.N. Protocol
              Map<String, dynamic>? updatedOgm = router.processOGM(packet, senderIp);
              if (updatedOgm != null) {
                // Rota atualizada, repassar para os vizinhos!
                _sendToAllDirectPeers(updatedOgm, excludeIp: senderIp);
              }
            } else if (packet["tipo"] == "DATA_ROUTE") {
              // Roteamento de Dados Híbrido (Pode ser Transação ou Chat)
              String finalDest = packet["final_destination"];
              if (finalDest == nodeName || finalDest == "BROADCAST") {
                // É pra mim!
                for (var callback in onMessageCallbacks) {
                  callback(packet["payload"]);
                }
              } else {
                // Se eu for bridge e a mensagem for broadcast, eu posso mandar pra internet
                if (isBridgeNode && finalDest == "BROADCAST") {
                   _syncWithGlobalNetwork(packet);
                }
                
                // Não é pra mim, devo repassar de acordo com a tabela de roteamento
                String? nextHop = router.getNextHop(finalDest);
                if (nextHop != null) {
                  _sendToIp(nextHop, packet);
                } else {
                  // Se não sei a rota, posso fazer fallback para broadcast local
                  _sendToAllDirectPeers(packet, excludeIp: senderIp);
                }
              }
            } else {
              // Mensagem legada P2P sem roteamento B.A.T.M.A.N
              for (var callback in onMessageCallbacks) {
                callback(packet);
              }
            }
          } catch (e) {
            print("Error parsing TCP packet: \$e");
          }
        });
      });
    } catch (e) {
      print("TCP Listener Error: \$e");
    }
  }

  /// Envia um dado encapsulado na camada de roteamento
  void sendRoutedData(String finalDestination, Map<String, dynamic> payload) {
    Map<String, dynamic> packet = {
      "tipo": "DATA_ROUTE",
      "final_destination": finalDestination,
      "payload": payload,
    };
    
    if (finalDestination == "BROADCAST") {
      _sendToAllDirectPeers(packet);
      return;
    }

    String? nextHop = router.getNextHop(finalDestination);
    if (nextHop != null) {
      _sendToIp(nextHop, packet);
    } else {
      // Se não conhece a rota, manda pra todo mundo tentar descobrir
      _sendToAllDirectPeers(packet);
    }
  }

  void _sendToIp(String ip, Map<String, dynamic> data) {
    String jsonStr = jsonEncode(data);
    List<int> bytes = utf8.encode(jsonStr);
    
    Socket.connect(ip, tcpPort, timeout: Duration(seconds: 2)).then((socket) {
      socket.add(bytes);
      socket.destroy();
    }).catchError((e) {
      // Peer inalcançável
    });
  }

  void _sendToAllDirectPeers(Map<String, dynamic> data, {String? excludeIp}) {
    String jsonStr = jsonEncode(data);
    List<int> bytes = utf8.encode(jsonStr);
    List<String> toRemove = [];
    
    for (String peer in directPeers) {
      String ip = peer.split(":")[0];
      int port = int.parse(peer.split(":")[1]);
      
      if (ip == excludeIp) continue;
      
      Socket.connect(ip, port, timeout: Duration(seconds: 2)).then((socket) {
        socket.add(bytes);
        socket.destroy();
      }).catchError((e) {
        toRemove.add(peer);
      });
    }
    directPeers.removeAll(toRemove);
  }
  
  /// (Mock) Envia os dados para a rede global se este celular estiver com internet
  void _syncWithGlobalNetwork(Map<String, dynamic> packet) {
      print("🌐 [BRIDGE NODE] Sincronizando pacote com a Internet (Global Network): \${packet['tipo']}");
      // Aqui entraria a chamada HTTP real (http.post(globalServerUrl, body: ...))
  }
}
