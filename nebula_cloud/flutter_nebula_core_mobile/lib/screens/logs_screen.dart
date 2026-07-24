import 'dart:async';
import 'package:flutter/material.dart';
import '../services/api_service.dart';

class LogsScreen extends StatefulWidget {
  const LogsScreen({super.key});

  @override
  State<LogsScreen> createState() => _LogsScreenState();
}

class _LogsScreenState extends State<LogsScreen> {
  Timer? _timer;
  List<dynamic> _logs = [];
  bool _loading = true;
  String _severity = 'all';

  @override
  void initState() {
    super.initState();
    _loadLogs();
    _timer = Timer.periodic(const Duration(seconds: 5), (_) => _loadLogs());
  }

  Future<void> _loadLogs() async {
    final logs = await ApiService.getLogs(
      limit: 100,
      severity: _severity == 'all' ? null : _severity,
    );
    if (mounted) {
      setState(() { _logs = logs; _loading = false; });
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  Color _severityColor(String? s) {
    switch (s) {
      case 'error': return Colors.red;
      case 'warning': return const Color(0xFFFFB800);
      case 'success': return const Color(0xFF00FF88);
      default: return const Color(0xFF00D4FF);
    }
  }

  IconData _severityIcon(String? s) {
    switch (s) {
      case 'error': return Icons.error_outline;
      case 'warning': return Icons.warning_amber_outlined;
      case 'success': return Icons.check_circle_outline;
      default: return Icons.info_outline;
    }
  }

  @override
  Widget build(BuildContext context) {
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
                      const Text('Logs da Rede', style: TextStyle(
                        fontSize: 22, fontWeight: FontWeight.bold, color: Colors.white,
                      )),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: const Color(0xFF00FF88).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(color: const Color(0xFF00FF88).withOpacity(0.3)),
                        ),
                        child: const Row(
                          children: [
                            Icon(Icons.circle, color: Color(0xFF00FF88), size: 7),
                            SizedBox(width: 5),
                            Text('AO VIVO', style: TextStyle(
                              fontSize: 9, color: Color(0xFF00FF88),
                              fontWeight: FontWeight.bold, letterSpacing: 1,
                            )),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: Row(
                      children: ['all', 'info', 'success', 'warning', 'error'].map((s) {
                        final selected = _severity == s;
                        final color = s == 'all' ? const Color(0xFF00D4FF) : _severityColor(s);
                        return Padding(
                          padding: const EdgeInsets.only(right: 8),
                          child: GestureDetector(
                            onTap: () {
                              setState(() => _severity = s);
                              _loadLogs();
                            },
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                              decoration: BoxDecoration(
                                color: selected ? color.withOpacity(0.15) : const Color(0xFF0D1117),
                                borderRadius: BorderRadius.circular(20),
                                border: Border.all(
                                  color: selected ? color : Colors.white.withOpacity(0.1),
                                ),
                              ),
                              child: Text(
                                s == 'all' ? 'Todos' : s.toUpperCase(),
                                style: TextStyle(
                                  fontSize: 11,
                                  color: selected ? color : Colors.white.withOpacity(0.4),
                                  fontWeight: selected ? FontWeight.bold : FontWeight.normal,
                                ),
                              ),
                            ),
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: _loading
                  ? const Center(child: CircularProgressIndicator(color: Color(0xFF00D4FF)))
                  : _logs.isEmpty
                      ? Center(child: Text('Nenhum log encontrado',
                          style: TextStyle(color: Colors.white.withOpacity(0.3))))
                      : RefreshIndicator(
                          onRefresh: _loadLogs,
                          color: const Color(0xFF00D4FF),
                          child: ListView.builder(
                            padding: const EdgeInsets.symmetric(horizontal: 16),
                            itemCount: _logs.length,
                            itemBuilder: (ctx, i) {
                              final log = _logs[i];
                              final sev = log['severity'] as String? ?? 'info';
                              final color = _severityColor(sev);
                              return Container(
                                margin: const EdgeInsets.only(bottom: 8),
                                padding: const EdgeInsets.all(12),
                                decoration: BoxDecoration(
                                  color: const Color(0xFF0D1117),
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(color: color.withOpacity(0.12)),
                                ),
                                child: Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Container(
                                      padding: const EdgeInsets.all(5),
                                      decoration: BoxDecoration(
                                        color: color.withOpacity(0.1),
                                        borderRadius: BorderRadius.circular(7),
                                      ),
                                      child: Icon(_severityIcon(sev), color: color, size: 13),
                                    ),
                                    const SizedBox(width: 10),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Text(log['message'] ?? '', style: const TextStyle(
                                            color: Colors.white, fontSize: 12,
                                          )),
                                          const SizedBox(height: 3),
                                          Row(
                                            children: [
                                              if (log['nodeName'] != null) ...[
                                                Text(log['nodeName'], style: TextStyle(
                                                  color: const Color(0xFF00D4FF).withOpacity(0.7),
                                                  fontSize: 10,
                                                )),
                                                Text(' · ', style: TextStyle(
                                                  color: Colors.white.withOpacity(0.2), fontSize: 10,
                                                )),
                                              ],
                                              Text(log['eventType'] ?? '', style: TextStyle(
                                                color: Colors.white.withOpacity(0.3), fontSize: 10,
                                              )),
                                            ],
                                          ),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                              );
                            },
                          ),
                        ),
            ),
          ],
        ),
      ),
    );
  }
}
