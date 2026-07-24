class RouteEntry {
  final String nextHopIp;
  final int hops;
  final int sequenceNumber;
  final int lastUpdated;

  RouteEntry({
    required this.nextHopIp,
    required this.hops,
    required this.sequenceNumber,
    required this.lastUpdated,
  });
}

class BatmanRouter {
  // Tabela de roteamento: Destino (NodeName ou IP) -> Rota
  final Map<String, RouteEntry> routingTable = {};
  
  final String myNodeName;
  int mySequenceNumber = 0;

  BatmanRouter(this.myNodeName);

  /// Cria uma mensagem OGM (Originator Message) para anunciar este nó na rede.
  Map<String, dynamic> generateOGM() {
    mySequenceNumber++;
    return {
      "tipo": "OGM",
      "originator": myNodeName,
      "sequence_number": mySequenceNumber,
      "hops": 0
    };
  }

  /// Processa um OGM recebido. Retorna o OGM atualizado se precisar ser retransmitido, 
  /// ou null se a rota já for conhecida/inferior.
  Map<String, dynamic>? processOGM(Map<String, dynamic> ogm, String senderIp) {
    String originator = ogm["originator"];
    int seqNum = ogm["sequence_number"];
    int hops = ogm["hops"];

    // Se o OGM for meu, ignoro (loop de rede)
    if (originator == myNodeName) return null;

    int newHops = hops + 1;
    int now = DateTime.now().millisecondsSinceEpoch;

    RouteEntry? currentRoute = routingTable[originator];
    bool shouldUpdate = false;

    if (currentRoute == null) {
      // Rota nova
      shouldUpdate = true;
    } else {
      // Avalia se o OGM é mais novo ou é uma rota melhor (menos pulos)
      if (seqNum > currentRoute.sequenceNumber) {
        shouldUpdate = true;
      } else if (seqNum == currentRoute.sequenceNumber && newHops < currentRoute.hops) {
        shouldUpdate = true;
      }
    }

    if (shouldUpdate) {
      routingTable[originator] = RouteEntry(
        nextHopIp: senderIp,
        hops: newHops,
        sequenceNumber: seqNum,
        lastUpdated: now,
      );

      // Prepara o OGM para ser retransmitido para os vizinhos (Gossip)
      return {
        "tipo": "OGM",
        "originator": originator,
        "sequence_number": seqNum,
        "hops": newHops
      };
    }

    return null; // Não repassa (rota antiga ou pior)
  }

  /// Retorna o IP do próximo nó (Next Hop) para alcançar um destino.
  String? getNextHop(String destination) {
    return routingTable[destination]?.nextHopIp;
  }
  
  /// Limpa rotas que não receberam OGM há mais de 30 segundos
  void cleanStaleRoutes() {
    int now = DateTime.now().millisecondsSinceEpoch;
    routingTable.removeWhere((key, route) => (now - route.lastUpdated) > 30000);
  }
}
