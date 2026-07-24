import 'dart:convert';
import 'dart:typed_data';
import 'package:nearby_connections/nearby_connections.dart';

class NearbyService {
  final String userName;
  final Strategy strategy = Strategy.P2P_CLUSTER; // Suporta M x N conexões (Rede Mesh)
  final Set<String> connectedEndpoints = {};
  
  Function(Map<String, dynamic>)? onDataReceived;

  NearbyService(this.userName);

  Future<void> start() async {
    try {
      bool a = await Nearby().startAdvertising(
        userName,
        strategy,
        onConnectionInitiated: _onConnectionInit,
        onConnectionResult: (id, status) {
          if (status == Status.CONNECTED) {
            connectedEndpoints.add(id);
          } else {
            connectedEndpoints.remove(id);
          }
        },
        onDisconnected: (id) {
          connectedEndpoints.remove(id);
        },
      );

      bool d = await Nearby().startDiscovery(
        userName,
        strategy,
        onEndpointFound: (id, name, serviceId) {
          // Automaticamente pede conexão ao achar outro nó MeshCoin
          Nearby().requestConnection(
            userName,
            id,
            onConnectionInitiated: _onConnectionInit,
            onConnectionResult: (id, status) {
              if (status == Status.CONNECTED) {
                connectedEndpoints.add(id);
              } else {
                connectedEndpoints.remove(id);
              }
            },
            onDisconnected: (id) {
              connectedEndpoints.remove(id);
            },
          );
        },
        onEndpointLost: (id) {
          print("Nó perdido (Nearby): $id");
        },
      );
      
      print("Nearby Connections (BLE/Wi-Fi Direct) Iniciado: Adv=$a, Disc=$d");
    } catch (e) {
      print("Erro ao iniciar Nearby Connections: $e");
    }
  }

  void _onConnectionInit(String id, ConnectionInfo info) {
    // Aceita conexão automaticamente para formar a Mesh
    Nearby().acceptConnection(
      id,
      onPayLoadRecieved: (endpointId, payload) {
        if (payload.type == PayloadType.BYTES && payload.bytes != null) {
          try {
            String msg = utf8.decode(payload.bytes!);
            Map<String, dynamic> data = jsonDecode(msg);
            if (onDataReceived != null) {
              onDataReceived!(data);
            }
          } catch(e) {}
        }
      },
      onPayloadTransferUpdate: (endpointId, payloadTransferUpdate) {},
    );
  }

  void broadcastData(Map<String, dynamic> data) {
    if (connectedEndpoints.isEmpty) return;
    String msg = jsonEncode(data);
    Uint8List bytes = Uint8List.fromList(utf8.encode(msg));
    
    for (String endpoint in connectedEndpoints) {
      Nearby().sendBytesPayload(endpoint, bytes);
    }
  }

  Future<void> stop() async {
    await Nearby().stopAdvertising();
    await Nearby().stopDiscovery();
    await Nearby().stopAllEndpoints();
    connectedEndpoints.clear();
  }
}
