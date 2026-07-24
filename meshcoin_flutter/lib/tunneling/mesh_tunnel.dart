/// ═══════════════════════════════════════════════════════════════
/// Mesh Tunneling — Transport para BTC, ETH, Solana e Pix Off-Grid
/// ═══════════════════════════════════════════════════════════════

enum TunnelChain { btc, eth, sol, pix, mesh }
enum TunnelStatus { queued, broadcasting, confirmed, failed }

class TunnelTransaction {
  final String id;
  final TunnelChain chain;
  final String fromAddress;
  final String toAddress;
  final String amount;
  final String payload; // Dados crus da transação da outra chain
  final TunnelStatus status;
  final DateTime createdAt;
  final DateTime? broadcastedAt;

  TunnelTransaction({
    required this.id,
    required this.chain,
    required this.fromAddress,
    required this.toAddress,
    required this.amount,
    required this.payload,
    this.status = TunnelStatus.queued,
    DateTime? createdAt,
    this.broadcastedAt,
  }) : createdAt = createdAt ?? DateTime.now();

  String get chainName {
    switch (chain) {
      case TunnelChain.btc: return 'Bitcoin';
      case TunnelChain.eth: return 'Ethereum';
      case TunnelChain.sol: return 'Solana';
      case TunnelChain.pix: return 'Pix';
      case TunnelChain.mesh: return 'MeshCoin';
    }
  }

  String get chainSymbol {
    switch (chain) {
      case TunnelChain.btc: return 'BTC';
      case TunnelChain.eth: return 'ETH';
      case TunnelChain.sol: return 'SOL';
      case TunnelChain.pix: return 'BRL';
      case TunnelChain.mesh: return 'MESH';
    }
  }

  String get chainEmoji {
    switch (chain) {
      case TunnelChain.btc: return '₿';
      case TunnelChain.eth: return 'Ξ';
      case TunnelChain.sol: return '◎';
      case TunnelChain.pix: return '🇧🇷';
      case TunnelChain.mesh: return '🌐';
    }
  }

  String get statusLabel {
    switch (status) {
      case TunnelStatus.queued: return 'Aguardando Bridge';
      case TunnelStatus.broadcasting: return 'Transmitindo...';
      case TunnelStatus.confirmed: return 'Confirmada';
      case TunnelStatus.failed: return 'Falhou';
    }
  }
}

class MeshTunnelQueue {
  final List<TunnelTransaction> _queue = [];

  List<TunnelTransaction> get pending =>
      _queue.where((tx) => tx.status == TunnelStatus.queued).toList();

  List<TunnelTransaction> get all => List.unmodifiable(_queue);

  int get pendingCount => pending.length;

  void enqueue(TunnelTransaction tx) {
    _queue.add(tx);
  }

  /// Flush: despacha transações pendentes quando um Bridge Node está disponível
  List<TunnelTransaction> flush() {
    List<TunnelTransaction> toSend = pending;
    // Em produção, aqui faria POST para APIs de cada chain
    for (int i = 0; i < _queue.length; i++) {
      if (_queue[i].status == TunnelStatus.queued) {
        _queue[i] = TunnelTransaction(
          id: _queue[i].id,
          chain: _queue[i].chain,
          fromAddress: _queue[i].fromAddress,
          toAddress: _queue[i].toAddress,
          amount: _queue[i].amount,
          payload: _queue[i].payload,
          status: TunnelStatus.broadcasting,
          createdAt: _queue[i].createdAt,
          broadcastedAt: DateTime.now(),
        );
      }
    }
    return toSend;
  }
}
