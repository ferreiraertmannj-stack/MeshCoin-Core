import 'dart:async';
import 'package:flutter/material.dart';
import '../services/api_service.dart';

class NodesScreen extends StatefulWidget {
  const NodesScreen({super.key});

  @override
  State<NodesScreen> createState() => _NodesScreenState();
}

class _NodesScreenState extends State<NodesScreen> {
  Timer? _timer;
  List<dynamic> _nodes = [];
  bool _loading = true;
  String _filter = 'all';

  @override
  void initState() {
    super.initState();
    _loadNodes();
    _timer = Timer.periodic(const Duration(seconds: 10), (_) => _loadNodes());
  }

  Future<void> _loadNodes() async {
    final nodes = await ApiService.getNodes();
    if (mounted) {
      setState(() { _nodes = nodes; _loading = false; });
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  List<dynamic> get _filteredNodes {
    if (_filter == 'online') return _nodes.where((n) => n['status'] == 'online').toList();
    if (_filter == 'offline') return _nodes.where((n) => n['status'] != 'online').toList();
    return _nodes;
  }

  @override
  Widget build(BuildContext context) {
    final onlineCount = _nodes.where((n) => n['status'] == 'online').length;

    return Scaffold(
      body: SafeArea(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text('Nós da Rede', style: TextStyle(
                        fontSize: 22, fontWeight: FontWeight.bold, color: Colors.white,
                      )),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                        decoration: BoxDecoration(
                          color: const Color(0xFF00FF88).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(color: const Color(0xFF00FF88).withOpacity(0.3)),
                        ),
                        child: Text(
                          '$onlineCount/${_nodes.length} online',
                          style: const TextStyle(
                            fontSize: 12, color: Color(0xFF00FF88), fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  // Filter chips
                  Row(
                    children: [
                      _FilterChip(label: 'Todos', value: 'all', current: _filter,
                          onTap: (v) => setState(() => _filter = v)),
                      const SizedBox(width: 8),
                      _FilterChip(label: 'Online', value: 'online', current: _filter,
                          onTap: (v) => setState(() => _filter = v)),
                      const SizedBox(width: 8),
                      _FilterChip(label: 'Offline', value: 'offline', current: _filter,
                          onTap: (v) => setState(() => _filter = v)),
                    ],
                  ),
                ],
              ),
            ),
            Expanded(
              child: _loading
                  ? const Center(child: CircularProgressIndicator(color: Color(0xFF00D4FF)))
                  : _filteredNodes.isEmpty
                      ? Center(
                          child: Text(
                            'Nenhum nó encontrado',
                            style: TextStyle(color: Colors.white.withOpacity(0.3)),
                          ),
                        )
                      : RefreshIndicator(
                          onRefresh: _loadNodes,
                          color: const Color(0xFF00D4FF),
                          child: ListView.builder(
                            padding: const EdgeInsets.symmetric(horizontal: 20),
                            itemCount: _filteredNodes.length,
                            itemBuilder: (ctx, i) => _NodeDetailCard(node: _filteredNodes[i]),
                          ),
                        ),
            ),
          ],
        ),
      ),
    );
  }
}

class _FilterChip extends StatelessWidget {
  final String label;
  final String value;
  final String current;
  final Function(String) onTap;

  const _FilterChip({
    required this.label, required this.value,
    required this.current, required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final selected = value == current;
    return GestureDetector(
      onTap: () => onTap(value),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 7),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFF00D4FF).withOpacity(0.15) : const Color(0xFF0D1117),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(
            color: selected ? const Color(0xFF00D4FF) : Colors.white.withOpacity(0.1),
          ),
        ),
        child: Text(label, style: TextStyle(
          fontSize: 12,
          color: selected ? const Color(0xFF00D4FF) : Colors.white.withOpacity(0.5),
          fontWeight: selected ? FontWeight.bold : FontWeight.normal,
        )),
      ),
    );
  }
}

class _NodeDetailCard extends StatelessWidget {
  final dynamic node;
  const _NodeDetailCard({required this.node});

  @override
  Widget build(BuildContext context) {
    final isOnline = node['status'] == 'online';
    final statusColor = isOnline ? const Color(0xFF00FF88) : Colors.red;
    final usedGB = (node['storageUsedGB'] ?? 0).toDouble();
    final capGB = (node['storageCapacityGB'] ?? 1).toDouble();
    final pct = capGB > 0 ? usedGB / capGB : 0.0;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF0D1117),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: statusColor.withOpacity(0.15)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(node['name'] ?? 'Nó', style: const TextStyle(
                      color: Colors.white, fontSize: 15, fontWeight: FontWeight.bold,
                    )),
                    Text(node['ipAddress'] ?? '', style: TextStyle(
                      color: Colors.white.withOpacity(0.4), fontSize: 11, fontFamily: 'monospace',
                    )),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: statusColor.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: statusColor.withOpacity(0.3)),
                ),
                child: Text(
                  isOnline ? 'ONLINE' : 'OFFLINE',
                  style: TextStyle(
                    fontSize: 10, fontWeight: FontWeight.bold,
                    color: statusColor, letterSpacing: 1,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          Row(
            children: [
              _MetricChip(icon: Icons.storage, label: '${usedGB.toStringAsFixed(1)}/${capGB.toStringAsFixed(0)} GB'),
              const SizedBox(width: 8),
              if (node['latencyMs'] != null)
                _MetricChip(icon: Icons.speed, label: '${node['latencyMs']}ms'),
              const SizedBox(width: 8),
              _MetricChip(icon: Icons.folder_copy_outlined, label: '${node['fragmentCount'] ?? 0} frags'),
            ],
          ),
          const SizedBox(height: 10),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: pct.clamp(0.0, 1.0),
              minHeight: 5,
              backgroundColor: Colors.white.withOpacity(0.06),
              valueColor: AlwaysStoppedAnimation<Color>(
                pct > 0.85 ? const Color(0xFFFFB800) : const Color(0xFF00D4FF),
              ),
            ),
          ),
          const SizedBox(height: 4),
          Text(
            '${(pct * 100).toStringAsFixed(1)}% utilizado',
            style: TextStyle(color: Colors.white.withOpacity(0.3), fontSize: 10),
          ),
        ],
      ),
    );
  }
}

class _MetricChip extends StatelessWidget {
  final IconData icon;
  final String label;
  const _MetricChip({required this.icon, required this.label});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.04),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.white.withOpacity(0.08)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 11, color: Colors.white.withOpacity(0.4)),
          const SizedBox(width: 4),
          Text(label, style: TextStyle(
            fontSize: 11, color: Colors.white.withOpacity(0.6),
          )),
        ],
      ),
    );
  }
}
