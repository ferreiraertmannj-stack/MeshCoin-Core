import 'dart:async';
import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';
import '../services/node_service.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> with TickerProviderStateMixin {
  Timer? _refreshTimer;
  bool _isRunning = false;
  Map<String, String> _config = {};
  NodeStatus? _status;
  late AnimationController _pulseController;
  final List<FlSpot> _latencyHistory = [];
  int _latencyIndex = 0;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 2),
    )..repeat(reverse: true);
    _loadData();
    _refreshTimer = Timer.periodic(const Duration(seconds: 5), (_) => _loadData());
  }

  Future<void> _loadData() async {
    final config = await NodeService.getConfig();
    final status = NodeService.lastStatus;
    if (mounted) {
      setState(() {
        _config = config;
        _isRunning = NodeService.isRunning;
        _status = status;
        if (status != null) {
          _latencyHistory.add(FlSpot(_latencyIndex.toDouble(), status.latencyMs));
          _latencyIndex++;
          if (_latencyHistory.length > 20) _latencyHistory.removeAt(0);
        }
      });
    }
  }

  void _toggleDaemon() {
    if (_isRunning) {
      NodeService.stopDaemon();
    } else {
      NodeService.startDaemon();
    }
    setState(() => _isRunning = !_isRunning);
  }

  @override
  void dispose() {
    _refreshTimer?.cancel();
    _pulseController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final nodeName = _config['nodeName'] ?? 'Meu Nó';
    final storageUsed = _status?.storageUsedGB ?? 0.0;
    final storageCapacity = double.tryParse(_config['storageCapacityGb'] ?? '50') ?? 50.0;
    final storagePercent = storageCapacity > 0 ? (storageUsed / storageCapacity) : 0.0;
    final fragments = _status?.fragmentCount ?? 0;

    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        nodeName,
                        style: const TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                      ),
                      Text(
                        _config['nodeId'] ?? 'Não configurado',
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.white.withOpacity(0.4),
                          fontFamily: 'monospace',
                        ),
                      ),
                    ],
                  ),
                  // Status indicator
                  AnimatedBuilder(
                    animation: _pulseController,
                    builder: (_, __) => Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                      decoration: BoxDecoration(
                        color: _isRunning
                            ? Color.lerp(
                                const Color(0xFF00D4FF).withOpacity(0.1),
                                const Color(0xFF00D4FF).withOpacity(0.25),
                                _pulseController.value,
                              )
                            : Colors.red.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(20),
                        border: Border.all(
                          color: _isRunning
                              ? const Color(0xFF00D4FF).withOpacity(0.5)
                              : Colors.red.withOpacity(0.5),
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Container(
                            width: 8,
                            height: 8,
                            decoration: BoxDecoration(
                              shape: BoxShape.circle,
                              color: _isRunning ? const Color(0xFF00D4FF) : Colors.red,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            _isRunning ? 'ONLINE' : 'OFFLINE',
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.bold,
                              color: _isRunning ? const Color(0xFF00D4FF) : Colors.red,
                              letterSpacing: 1,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 28),

              // Toggle Button
              GestureDetector(
                onTap: _toggleDaemon,
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.symmetric(vertical: 18),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: _isRunning
                          ? [const Color(0xFF1A1A2E), const Color(0xFF16213E)]
                          : [const Color(0xFF00D4FF), const Color(0xFF7B2FFF)],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: _isRunning
                          ? Colors.red.withOpacity(0.4)
                          : Colors.transparent,
                    ),
                    boxShadow: _isRunning
                        ? []
                        : [
                            BoxShadow(
                              color: const Color(0xFF00D4FF).withOpacity(0.3),
                              blurRadius: 20,
                              spreadRadius: 2,
                            ),
                          ],
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        _isRunning ? Icons.stop_circle_outlined : Icons.play_circle_outline,
                        color: _isRunning ? Colors.red : Colors.white,
                        size: 24,
                      ),
                      const SizedBox(width: 10),
                      Text(
                        _isRunning ? 'PARAR NÓ' : 'INICIAR NÓ',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: _isRunning ? Colors.red : Colors.white,
                          letterSpacing: 2,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 24),

              // Stats Grid
              GridView.count(
                crossAxisCount: 2,
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                childAspectRatio: 1.4,
                children: [
                  _StatCard(
                    icon: Icons.storage,
                    label: 'Fragmentos',
                    value: '$fragments',
                    subtitle: 'armazenados',
                    color: const Color(0xFF7B2FFF),
                  ),
                  _StatCard(
                    icon: Icons.sd_storage,
                    label: 'Storage Usado',
                    value: '${storageUsed.toStringAsFixed(2)} GB',
                    subtitle: 'de ${storageCapacity.toStringAsFixed(0)} GB',
                    color: const Color(0xFF00D4FF),
                  ),
                  _StatCard(
                    icon: Icons.wifi,
                    label: 'Latência',
                    value: _status != null ? '${_status!.latencyMs.toStringAsFixed(0)}ms' : '--',
                    subtitle: 'última medição',
                    color: const Color(0xFF00FF88),
                  ),
                  _StatCard(
                    icon: Icons.access_time,
                    label: 'Última Sync',
                    value: _status != null
                        ? _formatTime(_status!.lastSeen)
                        : '--',
                    subtitle: 'heartbeat',
                    color: const Color(0xFFFFB800),
                  ),
                ],
              ),
              const SizedBox(height: 20),

              // Storage Progress
              _buildCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Uso de Armazenamento',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Colors.white,
                          ),
                        ),
                        Text(
                          '${(storagePercent * 100).toStringAsFixed(1)}%',
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF00D4FF),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: LinearProgressIndicator(
                        value: storagePercent.clamp(0.0, 1.0),
                        minHeight: 10,
                        backgroundColor: Colors.white.withOpacity(0.08),
                        valueColor: AlwaysStoppedAnimation(
                          storagePercent > 0.8
                              ? Colors.orange
                              : const Color(0xFF00D4FF),
                        ),
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '${storageUsed.toStringAsFixed(3)} GB usados de ${storageCapacity.toStringAsFixed(0)} GB disponíveis',
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.white.withOpacity(0.5),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),

              // Nebula Core Connection
              _buildCard(
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: const Color(0xFF7B2FFF).withOpacity(0.15),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: const Icon(Icons.cloud_outlined,
                          color: Color(0xFF7B2FFF), size: 22),
                    ),
                    const SizedBox(width: 14),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Nébula Core',
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: Colors.white,
                            ),
                          ),
                          Text(
                            _config['coreUrl']?.isNotEmpty == true
                                ? _config['coreUrl']!
                                : 'Não configurado',
                            style: TextStyle(
                              fontSize: 11,
                              color: Colors.white.withOpacity(0.4),
                              fontFamily: 'monospace',
                            ),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ],
                      ),
                    ),
                    Container(
                      width: 10,
                      height: 10,
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        color: _isRunning
                            ? const Color(0xFF00FF88)
                            : Colors.grey,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCard({required Widget child}) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF0D1117),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white.withOpacity(0.08)),
      ),
      child: child,
    );
  }

  String _formatTime(DateTime dt) {
    final now = DateTime.now();
    final diff = now.difference(dt);
    if (diff.inSeconds < 60) return '${diff.inSeconds}s atrás';
    if (diff.inMinutes < 60) return '${diff.inMinutes}m atrás';
    return '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
  }
}

class _StatCard extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;
  final String subtitle;
  final Color color;

  const _StatCard({
    required this.icon,
    required this.label,
    required this.value,
    required this.subtitle,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFF0D1117),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white.withOpacity(0.08)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: color.withOpacity(0.15),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(icon, color: color, size: 16),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  label,
                  style: TextStyle(
                    fontSize: 11,
                    color: Colors.white.withOpacity(0.5),
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                value,
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
              Text(
                subtitle,
                style: TextStyle(
                  fontSize: 10,
                  color: Colors.white.withOpacity(0.35),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
