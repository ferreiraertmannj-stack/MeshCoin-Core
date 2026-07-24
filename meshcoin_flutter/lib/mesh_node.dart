import 'dart:io';
import 'dart:convert';
import 'dart:async';

class MeshNode {
  static const int udpPort = 5555;
  static const int tcpPort = 5556;
  static const String magicWord = "MESHCOIN_NODE";

  final String nodeName;
  final Set<String> peers = {};
  final List<Function(Map<String, dynamic>)> onMessageCallbacks = [];
  
  bool isRunning = false;
  ServerSocket? _serverSocket;
  RawDatagramSocket? _udpSocket;

  MeshNode(this.nodeName);

  Future<void> start() async {
    isRunning = true;
    _startUdpListener();
    _startUdpBroadcaster();
    _startTcpListener();
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
              if (!peers.contains(peerId)) {
                peers.add(peerId);
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

  Future<void> _startTcpListener() async {
    try {
      _serverSocket = await ServerSocket.bind(InternetAddress.anyIPv4, tcpPort);
      _serverSocket?.listen((Socket client) {
        client.listen((List<int> data) {
          try {
            String jsonStr = utf8.decode(data);
            Map<String, dynamic> packet = jsonDecode(jsonStr);
            for (var callback in onMessageCallbacks) {
              callback(packet);
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

  void broadcastData(Map<String, dynamic> data) {
    String jsonStr = jsonEncode(data);
    List<int> bytes = utf8.encode(jsonStr);
    
    List<String> toRemove = [];
    
    for (String peer in peers) {
      List<String> parts = peer.split(":");
      String ip = parts[0];
      int port = int.parse(parts[1]);
      
      Socket.connect(ip, port, timeout: Duration(seconds: 2)).then((socket) {
        socket.add(bytes);
        socket.destroy();
      }).catchError((e) {
        toRemove.add(peer);
      });
    }
    
    peers.removeAll(toRemove);
  }
}
