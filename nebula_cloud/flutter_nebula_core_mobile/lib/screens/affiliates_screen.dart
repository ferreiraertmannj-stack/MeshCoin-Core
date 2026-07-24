import 'dart:async';
import 'package:flutter/material.dart';
import '../services/api_service.dart';

class AffiliatesScreen extends StatefulWidget {
  const AffiliatesScreen({super.key});

  @override
  State<AffiliatesScreen> createState() => _AffiliatesScreenState();
}

class _AffiliatesScreenState extends State<AffiliatesScreen> {
  Timer? _timer;
  List<dynamic> _affiliates = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadAffiliates();
    _timer = Timer.periodic(const Duration(seconds: 30), (_) => _loadAffiliates());
  }

  Future<void> _loadAffiliates() async {
    final affiliates = await ApiService.getAffiliates();
    if (mounted) {
      setState(() { _affiliates = affiliates; _loading = false; });
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final totalEarnings = _affiliates.fold<double>(
      0.0, (sum, a) => sum + ((a['totalEarnings'] ?? 0.0) as num).toDouble(),
    );

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
                  const Text('Afiliados', style: TextStyle(
                    fontSize: 22, fontWeight: FontWeight.bold, color: Colors.white,
                  )),
                  Text('${_affiliates.length} cadastrados · R\$ ${totalEarnings.toStringAsFixed(2)} total',
                    style: TextStyle(fontSize: 12, color: Colors.white.withOpacity(0.4))),
                ],
              ),
            ),
            Expanded(
              child: _loading
                  ? const Center(child: CircularProgressIndicator(color: Color(0xFF00D4FF)))
                  : _affiliates.isEmpty
                      ? Center(
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(Icons.people_outline, size: 48,
                                  color: Colors.white.withOpacity(0.2)),
                              const SizedBox(height: 12),
                              Text('Nenhum afiliado cadastrado',
                                style: TextStyle(color: Colors.white.withOpacity(0.3))),
                            ],
                          ),
                        )
                      : RefreshIndicator(
                          onRefresh: _loadAffiliates,
                          color: const Color(0xFF00D4FF),
                          child: ListView.builder(
                            padding: const EdgeInsets.symmetric(horizontal: 20),
                            itemCount: _affiliates.length,
                            itemBuilder: (ctx, i) => _AffiliateCard(affiliate: _affiliates[i]),
                          ),
                        ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AffiliateCard extends StatelessWidget {
  final dynamic affiliate;
  const _AffiliateCard({required this.affiliate});

  @override
  Widget build(BuildContext context) {
    final earnings = (affiliate['totalEarnings'] ?? 0.0 as num).toDouble();
    final storageGB = (affiliate['storageUsedGB'] ?? 0.0 as num).toDouble();
    final tier = affiliate['tier'] as String? ?? 'BASIC';
    final tierColor = tier == 'PLATINUM' ? const Color(0xFF00D4FF)
        : tier == 'GOLD' ? const Color(0xFFFFB800)
        : Colors.grey;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF0D1117),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.white.withOpacity(0.08)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    width: 36, height: 36,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: const LinearGradient(
                        colors: [Color(0xFF00D4FF), Color(0xFF7B2FFF)],
                      ),
                    ),
                    child: Center(
                      child: Text(
                        (affiliate['userName'] as String? ?? 'A').substring(0, 1).toUpperCase(),
                        style: const TextStyle(
                          color: Colors.white, fontWeight: FontWeight.bold, fontSize: 16,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(affiliate['userName'] ?? 'Afiliado', style: const TextStyle(
                        color: Colors.white, fontSize: 14, fontWeight: FontWeight.w600,
                      )),
                      Text(affiliate['userEmail'] ?? '', style: TextStyle(
                        color: Colors.white.withOpacity(0.4), fontSize: 11,
                      )),
                    ],
                  ),
                ],
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: tierColor.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: tierColor.withOpacity(0.3)),
                ),
                child: Text(tier, style: TextStyle(
                  fontSize: 10, color: tierColor, fontWeight: FontWeight.bold,
                )),
              ),
            ],
          ),
          const SizedBox(height: 14),
          Row(
            children: [
              Expanded(child: _AffStat(
                label: 'Ganhos', value: 'R\$ ${earnings.toStringAsFixed(2)}',
                color: const Color(0xFF00FF88),
              )),
              Expanded(child: _AffStat(
                label: 'Storage', value: '${storageGB.toStringAsFixed(1)} GB',
                color: const Color(0xFF00D4FF),
              )),
              Expanded(child: _AffStat(
                label: 'Nós', value: '${affiliate['nodeCount'] ?? 0}',
                color: const Color(0xFF7B2FFF),
              )),
            ],
          ),
        ],
      ),
    );
  }
}

class _AffStat extends StatelessWidget {
  final String label;
  final String value;
  final Color color;
  const _AffStat({required this.label, required this.value, required this.color});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(value, style: TextStyle(
          fontSize: 14, fontWeight: FontWeight.bold, color: color,
        )),
        Text(label, style: TextStyle(
          fontSize: 10, color: Colors.white.withOpacity(0.4),
        )),
      ],
    );
  }
}
