import 'dart:async';
import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';
import '../services/api_service.dart';
import '../main.dart';
import 'login_screen.dart';

class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  Timer? _timer;
  Map<String, dynamic>? _stats;
  List<dynamic> _nodes = [];
  List<dynamic> _notifications = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadData();
    _timer = Timer.periodic(const Duration(seconds: 10), (_) => _loadData());
  }

  Future<void> _loadData() async {
    final stats = await ApiService.getNetworkStats();
    final nodes = await ApiService.getNodes();
    final notifs = await ApiService.getNotifications();
    if (mounted) {
      setState(() {
        _stats = stats;
        _nodes = nodes;
        _notifications = notifs.take(5).toList();
        _loading = false;
      });
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  Future<void> _logout() async {
    await ApiService.logout();
    if (mounted) {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => const LoginScreen()),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final totalNodes = _stats?['totalNodes'] ?? 0;
    final onlineNodes = _stats?['onlineNodes'] ?? 0;
    final totalStorageTB = (_stats?['totalStorageGB'] ?? 0) / 1024;
    final usedStorageTB = (_stats?['usedStorageGB'] ?? 0) / 1024;
    final totalEarnings = _stats?['totalEarnings'] ?? 0.0;

    return Scaffold(
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator(color: Color(0xFF00D4FF)))
            : RefreshIndicator(
                onRefresh: _loadData,
                color: const Color(0xFF00D4FF),
                child: SingleChildScrollView(
                  physics: const AlwaysScrollableScrollPhysics(),
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
                              const Text(
                                'NÉBULA CORE',
                                style: TextStyle(
                                  fontSize: 20,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.white,
                                  letterSpacing: 3,
                                ),
                              ),
                              Text(
                                'Ertmann Tech — Admin',
                                style: TextStyle(
                                  fontSize: 11,
                                  color: Colors.white.withOpacity(0.4),
                                ),
                              ),
                            ],
                          ),
                          Row(
                            children: [
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                                decoration: BoxDecoration(
                                  color: onlineNodes > 0
                                      ? const Color(0xFF00FF88).withOpacity(0.1)
                                      : Colors.red.withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(20),
                                  border: Border.all(
                                    color: onlineNodes > 0
                                        ? const Color(0xFF00FF88).withOpacity(0.3)
                                        : Colors.red.withOpacity(0.3),
                                  ),
                                ),
                                child: Row(
                                  children: [
                                    Icon(
                                      Icons.circle,
                                      size: 8,
                                      color: onlineNodes > 0 ? const Color(0xFF00FF88) : Colors.red,
                                    ),
                                    const SizedBox(width: 5),
                                    Text(
                                      '$onlineNodes ONLINE',
                                      style: TextStyle(
                                        fontSize: 10,
                                        fontWeight: FontWeight.bold,
                                        color: onlineNodes > 0 ? const Color(0xFF00FF88) : Colors.red,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              const SizedBox(width: 8),
                              IconButton(
                                icon: Icon(Icons.logout, color: Colors.white.withOpacity(0.4), size: 20),
                                onPressed: _logout,
                              ),
                            ],
                          ),
                        ],
                      ),
                      const SizedBox(height: 24),

                      // Stats grid
                      GridView.count(
                        crossAxisCount: 2,
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        crossAxisSpacing: 12,
                        mainAxisSpacing: 12,
                        childAspectRatio: 1.4,
                        children: [
                          _StatCard(
                            label: 'Nós Totais',
                            value: '$totalNodes',
                            sub: '$onlineNodes online',
                            icon: Icons.hub,
                            color: const Color(0xFF00D4FF),
                          ),
                          _StatCard(
                            label: 'Storage Total',
                            value: '${totalStorageTB.toStringAsFixed(1)} TB',
                            sub: '${usedStorageTB.toStringAsFixed(2)} TB usado',
                            icon: Icons.storage,
                            color: const Color(0xFF7B2FFF),
                          ),
                          _StatCard(
                            label: 'Receita Total',
                            value: 'R\$ ${totalEarnings.toStringAsFixed(2)}',
                            sub: 'mês atual',
                            icon: Icons.attach_money,
                            color: const Color(0xFF00FF88),
                          ),
                          _StatCard(
                            label: 'Afiliados',
                            value: '${_stats?['totalAffiliates'] ?? 0}',
                            sub: 'cadastrados',
                            icon: Icons.people,
                            color: const Color(0xFFFFB800),
                          ),
                        ],
                      ),
                      const SizedBox(height: 20),

                      // Storage chart
                      if (totalStorageTB > 0) ...[
                        _SectionTitle(title: 'Uso de Storage'),
                        const SizedBox(height: 12),
                        Container(
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: const Color(0xFF0D1117),
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(color: Colors.white.withOpacity(0.08)),
                          ),
                          child: Column(
                            children: [
                              Row(
                                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                children: [
                                  Text(
                                    '${(usedStorageTB / totalStorageTB * 100).toStringAsFixed(1)}% utilizado',
                                    style: const TextStyle(color: Colors.white, fontSize: 14),
                                  ),
                                  Text(
                                    '${usedStorageTB.toStringAsFixed(2)} / ${totalStorageTB.toStringAsFixed(1)} TB',
                                    style: TextStyle(
                                      color: Colors.white.withOpacity(0.4),
                                      fontSize: 12,
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 12),
                              ClipRRect(
                                borderRadius: BorderRadius.circular(8),
                                child: LinearProgressIndicator(
                                  value: totalStorageTB > 0 ? usedStorageTB / totalStorageTB : 0,
                                  minHeight: 10,
                                  backgroundColor: Colors.white.withOpacity(0.08),
                                  valueColor: const AlwaysStoppedAnimation<Color>(Color(0xFF00D4FF)),
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 20),
                      ],

                      // Nodes status
                      _SectionTitle(title: 'Status dos Nós'),
                      const SizedBox(height: 12),
                      ..._nodes.take(5).map((node) => _NodeCard(node: node)),
                      const SizedBox(height: 20),

                      // Notifications
                      if (_notifications.isNotEmpty) ...[
                        _SectionTitle(title: 'Alertas Recentes'),
                        const SizedBox(height: 12),
                        ..._notifications.map((n) => _NotifCard(notif: n)),
                      ],
                    ],
                  ),
                ),
              ),
      ),
    );
  }
}

class _SectionTitle extends StatelessWidget {
  final String title;
  const _SectionTitle({required this.title});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(width: 3, height: 16, decoration: BoxDecoration(
          color: const Color(0xFF00D4FF),
          borderRadius: BorderRadius.circular(2),
        )),
        const SizedBox(width: 8),
        Text(title, style: const TextStyle(
          fontSize: 14, fontWeight: FontWeight.w600, color: Colors.white,
        )),
      ],
    );
  }
}

class _StatCard extends StatelessWidget {
  final String label;
  final String value;
  final String sub;
  final IconData icon;
  final Color color;

  const _StatCard({
    required this.label,
    required this.value,
    required this.sub,
    required this.icon,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFF0D1117),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: color.withOpacity(0.2)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Container(
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: color.withOpacity(0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(icon, color: color, size: 18),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(value, style: TextStyle(
                fontSize: 18, fontWeight: FontWeight.bold, color: color,
              )),
              Text(label, style: const TextStyle(
                fontSize: 11, color: Colors.white, fontWeight: FontWeight.w500,
              )),
              Text(sub, style: TextStyle(
                fontSize: 10, color: Colors.white.withOpacity(0.4),
              )),
            ],
          ),
        ],
      ),
    );
  }
}

class _NodeCard extends StatelessWidget {
  final dynamic node;
  const _NodeCard({required this.node});

  @override
  Widget build(BuildContext context) {
    final isOnline = node['status'] == 'online';
    final color = isOnline ? const Color(0xFF00FF88) : Colors.red;

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFF0D1117),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.15)),
      ),
      child: Row(
        children: [
          Container(
            width: 8, height: 8,
            decoration: BoxDecoration(shape: BoxShape.circle, color: color),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(node['name'] ?? 'Nó', style: const TextStyle(
                  color: Colors.white, fontSize: 13, fontWeight: FontWeight.w500,
                )),
                Text(
                  '${node['storageUsedGB'] ?? 0} GB / ${node['storageCapacityGB'] ?? 0} GB',
                  style: TextStyle(color: Colors.white.withOpacity(0.4), fontSize: 11),
                ),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                isOnline ? 'ONLINE' : 'OFFLINE',
                style: TextStyle(
                  fontSize: 10, fontWeight: FontWeight.bold,
                  color: color, letterSpacing: 1,
                ),
              ),
              if (node['latencyMs'] != null)
                Text(
                  '${node['latencyMs']}ms',
                  style: TextStyle(color: Colors.white.withOpacity(0.4), fontSize: 10),
                ),
            ],
          ),
        ],
      ),
    );
  }
}

class _NotifCard extends StatelessWidget {
  final dynamic notif;
  const _NotifCard({required this.notif});

  @override
  Widget build(BuildContext context) {
    final type = notif['type'] as String? ?? 'info';
    final color = type == 'error' ? Colors.red
        : type == 'warning' ? const Color(0xFFFFB800)
        : const Color(0xFF00D4FF);

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withOpacity(0.05),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.2)),
      ),
      child: Row(
        children: [
          Icon(
            type == 'error' ? Icons.error_outline
                : type == 'warning' ? Icons.warning_amber_outlined
                : Icons.info_outline,
            color: color, size: 16,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              notif['message'] ?? '',
              style: const TextStyle(color: Colors.white, fontSize: 12),
            ),
          ),
        ],
      ),
    );
  }
}
