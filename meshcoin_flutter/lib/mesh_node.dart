import 'dart:io';
import 'dart:convert';
import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:network_info_plus/network_info_plus.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'batman_router.dart';
import 'network/nearby_service.dart';

class MeshNode {
  static const int udpPort = 5555;
  static const int tcpPort = 5556;
  static const String magicWord = "NEBULA_NODE";

  final String nodeName;
  final Set<String> directPeers = {}; // Vizinhos de um pulo (Diretos)
  late BatmanRouter router;
  late NearbyService nearbyService;
  
  String? myIpAddress;
  WebSocketChannel? _wsChannel; // Conexão Global via PC Node
  
  bool isBridgeNode = false; // Se for True, esse celular tem acesso 4G/Wi-Fi On-Grid
  final String globalServerUrl = "ws://127.0.0.1:8080/ws"; // PC Node local ou IP da nuvem
  
  final List<Function(Map<String, dynamic>)> onMessageCallbacks = [];
  
  bool isRunning = false;
  ServerSocket? _serverSocket;
  RawDatagramSocket? _udpSocket;

  MeshNode(this.nodeName) {
    router = BatmanRouter(nodeName);
    nearbyService = NearbyService(nodeName);
    
    // Conecta o recebimento via Nearby (BLE) aos callbacks principais
    nearbyService.onDataReceived = (data) {
      for (var cb in onMessageCallbacks) {
        cb(data);
      }
    };
  }

  Future<void> start() async {
    isRunning = true;
    final prefs = await SharedPreferences.getInstance();
    bool enableP2p = prefs.getBool('enable_p2p_offline') ?? true;
    bool enableBridge = prefs.getBool('enable_bridge_online') ?? true;

    if (!kIsWeb) {
      await _detectMyIp();
      
      if (enableP2p) {
        // Inicia a camada Nearby (BLE + Wi-Fi Direct) apenas se não for Web
        await nearbyService.start();
        
        // Inicia a camada UDP/TCP (Hotspot/LAN)
        _startUdpListener();
        _startUdpBroadcaster();
        _startTcpListener();
        _startOgmBroadcaster(); // B.A.T.M.A.N OGM
      }
    }
    
    if (enableBridge) {
      // Inicia a conexão global P2P via WebSocket (suportada na Web também)
      _startGlobalWebSocket();
    }
  }

  Timer? _wsPingTimer;

  Future<void> _startGlobalWebSocket() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      String bridgeIp = prefs.getString('pc_node_ip') ?? '127.0.0.1';
      if (bridgeIp.contains(':')) bridgeIp = bridgeIp.split(':')[0];
      if (bridgeIp.isEmpty) bridgeIp = '127.0.0.1';
      
      final url = "ws://$bridgeIp:8080/ws";
      print("Conectando ao PC Node / Bridge em: $url");
      
      _wsChannel = WebSocketChannel.connect(Uri.parse(url));
      
      // Realizar o Handshake Real
      _wsChannel?.sink.add(jsonEncode({
        "tipo": "HANDSHAKE",
        "node": nodeName,
        "timestamp": DateTime.now().millisecondsSinceEpoch
      }));

      // Heartbeat a cada 15 segundos para evitar que o SO mate o socket
      _wsPingTimer?.cancel();
      _wsPingTimer = Timer.periodic(const Duration(seconds: 15), (timer) {
        if (_wsChannel != null) {
          try {
            _wsChannel?.sink.add(jsonEncode({"tipo": "PING"}));
          } catch (e) {
            timer.cancel();
          }
        }
      });

      _wsChannel?.stream.listen((message) {
        // Mensagem recebida da internet (Japão, etc) ou PC Node
        try {
          final data = json.decode(message);
          for (var cb in onMessageCallbacks) {
            cb(data);
          }
        } catch (e) {}
      }, onError: (err) {
        print("Erro WS Global: $err");
        _wsPingTimer?.cancel();
      }, onDone: () {
        print("Conexão WS fechada. Reconectando...");
        _wsPingTimer?.cancel();
        // Reconectar após um tempo se cair
        if (isRunning) {
          Future.delayed(const Duration(seconds: 5), _startGlobalWebSocket);
        }
      });
    } catch (e) {
      print("Não foi possível conectar ao WebSocket Global: $e");
    }
  }

  Future<void> _detectMyIp() async {
    try {
      final info = NetworkInfo();
      myIpAddress = await info.getWifiIP();
      print("Meu IP local: $myIpAddress");
    } catch (e) {
      print("Erro ao obter IP: $e");
    }
  }

  void stop() {
    isRunning = false;
    _serverSocket?.close();
    _udpSocket?.close();
  }

  void onMessage(Function(Map<String, dynamic>) callback) {
    onMessageCallbacks.add(callback);
  }

  void removeMessage(Function(Map<String, dynamic>) callback) {
    onMessageCallbacks.remove(callback);
  }

  Future<void> _startUdpListener() async {
    try {
      _udpSocket = await RawDatagramSocket.bind(InternetAddress.anyIPv4, udpPort);
      _udpSocket?.broadcastEnabled = true;
      try {
        _udpSocket?.joinMulticast(InternetAddress("224.0.0.1"));
      } catch(e) {}
      
      _udpSocket?.listen((RawSocketEvent event) {
        if (event == RawSocketEvent.read) {
          Datagram? dg = _udpSocket?.receive();
          if (dg != null) {
            String msg = utf8.decode(dg.data);
            if (msg.startsWith(magicWord)) {
              String ip = dg.address.address;
              // IGNORAR SE FOR EU MESMO
              if (ip == myIpAddress || ip == "127.0.0.1") return;

              int port = int.parse(msg.split(":")[1]);
              String peerId = "$ip:$port";
              if (!directPeers.contains(peerId)) {
                directPeers.add(peerId);
              }

              // Auto-descoberta do PC Node
              SharedPreferences.getInstance().then((prefs) {
                String? currentIp = prefs.getString('pc_node_ip');
                if (currentIp == null || currentIp.isEmpty || currentIp == '127.0.0.1') {
                  prefs.setString('pc_node_ip', ip);
                  print("📡 PC Node descoberto via UDP: $ip. Conectando WebSocket e baixando Ledger...");
                  _startGlobalWebSocket();
                }
              });
            }
          }
        }
      });
    } catch (e) {
      print("UDP Listener Error: $e");
    }
  }

  Future<void> _startUdpBroadcaster() async {
    final info = NetworkInfo();
    
    Timer.periodic(Duration(seconds: 5), (timer) async {
      if (!isRunning) {
        timer.cancel();
        return;
      }
      try {
        String msg = "$magicWord:$tcpPort";
        List<int> data = utf8.encode(msg);
        
        // Em Hotspots Android, 255.255.255.255 geralmente é bloqueado.
        // O ideal é usar o subnet broadcast, ex: 192.168.43.255
        String? subnetMask = await info.getWifiSubmask();
        String broadcastIp = "255.255.255.255";
        
        if (myIpAddress != null && subnetMask != null) {
          try {
             broadcastIp = _calculateBroadcastAddress(myIpAddress!, subnetMask);
          } catch(e) {}
        }
        
        // Envia para o Subnet Broadcast (muito mais confiável no Android)
        _udpSocket?.send(data, InternetAddress(broadcastIp), udpPort);
        
        // Como fallback, tenta enviar para o broadcast global e para Multicast
        _udpSocket?.send(data, InternetAddress("255.255.255.255"), udpPort);
        _udpSocket?.send(data, InternetAddress("224.0.0.1"), udpPort);
      } catch (e) {
        print("Erro UDP Broadcaster: $e");
      }
    });
  }

  String _calculateBroadcastAddress(String ip, String subnet) {
    List<int> ipParts = ip.split('.').map(int.parse).toList();
    List<int> subnetParts = subnet.split('.').map(int.parse).toList();
    
    if (ipParts.length != 4 || subnetParts.length != 4) return "255.255.255.255";
    
    List<int> broadcastParts = [];
    for (int i = 0; i < 4; i++) {
      // Inverte a máscara (bitwise NOT) e faz OR com o IP
      int invertedSubnet = ~subnetParts[i] & 0xFF;
      broadcastParts.add(ipParts[i] | invertedSubnet);
    }
    
    return broadcastParts.join('.');
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
    
    // Envia pela Internet via PC Node (Supernode Global) para chats entre continentes
    if (_wsChannel != null) {
      try {
        _wsChannel!.sink.add(jsonEncode(packet));
      } catch(e) {}
    }
    
    // Sempre envia também para a rede BLE/Wi-Fi Direct offline
    try {
      nearbyService.broadcastData(packet);
    } catch(e) {}
    
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
      socket.close();
    }).catchError((e) {
      // Peer inalcançável
    });
  }

  void broadcastMessage(Map<String, dynamic> data) {
    String jsonStr = json.encode(data);
    
    // Envia pela Internet se disponível
    if (_wsChannel != null) {
      try {
        _wsChannel!.sink.add(jsonStr);
      } catch(e) {
        print("Erro ao enviar WebSocket: \$e");
      }
    }
    
    if (kIsWeb) return; // Web só usa Websocket

    // Envia via Nearby (BLE)
    nearbyService.broadcastData(data);

    // Envia via TCP para Peers Conhecidos
    for (String peer in directPeers) {
      String ip = peer.split(":")[0];
      int port = int.parse(peer.split(":")[1]);
      _sendToIp(ip, data);
    }
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
        socket.close();
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
