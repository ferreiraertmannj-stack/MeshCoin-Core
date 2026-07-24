import 'dart:async';
import 'package:flutter/material.dart';
import 'package:fl_chart/fl_chart.dart';
import '../services/node_service.dart';

class EarningsScreen extends StatefulWidget {
  const EarningsScreen({super.key});

  @override
  State<EarningsScreen> createState() => _EarningsScreenState();
}

class _EarningsScreenState extends State<EarningsScreen> {
  Timer? _timer;
  Map<String, String> _config = {};
  double _storageUsedGB = 0;
  double _storageCapacityGB = 50;
  double _egressGB = 0;
  double _totalEarnings = 0;
  double _monthlyEstimate = 0;

  // Tabela de preços Nébula (50% afiliado)
  static const double _storageRatePerGB = 0.0115; // R$/GB/mês (50% de 0.023)
  static const double _egressRatePerGB = 0.045;   // R$/GB (50% de 0.09)

  @override
  void initState() {
    super.initState();
    _loadData();
    _timer = Timer.periodic(const Duration(seconds: 15), (_) => _loadData());
  }

  Future<void> _loadData() async {
    final config = await NodeService.getConfig();
    final status = NodeService.lastStatus;
    final capacity = double.tryParse(config['storageCapacityGb'] ?? '50') ?? 50.0;
    final used = status?.storageUsedGB ?? 0.0;

    // Estimativa de egress: 20% do storage por mês (estimativa conservadora)
    final estimatedEgress = used * 0.2;
    final storageEarnings = used * _storageRatePerGB;
    final egressEarnings = estimatedEgress * _egressRatePerGB;
    final monthly = storageEarnings + egressEarnings;

    if (mounted) {
      setState(() {
        _config = config;
        _storageUsedGB = used;
        _storageCapacityGB = capacity;
        _egressGB = estimatedEgress;
        _monthlyEstimate = monthly;
        _totalEarnings = monthly * 0.3; // simulação de 30% do mês atual
      });
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final potentialMonthly = _storageCapacityGB * _storageRatePerGB;

    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Seus Ganhos',
                style: TextStyle(
                  fontSize: 22,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
              Text(
                'Divisão 50/50 — Ertmann Tech / Afiliado',
                style: TextStyle(
                  fontSize: 12,
                  color: Colors.white.withOpacity(0.4),
                ),
              ),
              const SizedBox(height: 24),

              // Total card
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(24),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [Color(0xFF00D4FF), Color(0xFF7B2FFF)],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(20),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF00D4FF).withOpacity(0.25),
                      blurRadius: 20,
                      spreadRadius: 2,
                    ),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Ganhos do Mês Atual',
                      style: TextStyle(
                        color: Colors.white.withOpacity(0.8),
                        fontSize: 13,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'R\$ ${_totalEarnings.toStringAsFixed(2)}',
                      style: const TextStyle(
                        fontSize: 36,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Estimativa mensal completa: R\$ ${_monthlyEstimate.toStringAsFixed(2)}',
                      style: TextStyle(
                        color: Colors.white.withOpacity(0.7),
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),

              // Breakdown
              _buildCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Detalhamento de Ganhos',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: Colors.white,
                      ),
                    ),
                    const SizedBox(height: 16),
                    _EarningsRow(
                      label: 'Storage (${_storageUsedGB.toStringAsFixed(2)} GB)',
                      rate: 'R\$ ${_storageRatePerGB.toStringAsFixed(4)}/GB/mês',
                      value: _storageUsedGB * _storageRatePerGB,
                      color: const Color(0xFF00D4FF),
                    ),
                    const Divider(color: Colors.white12, height: 20),
                    _EarningsRow(
                      label: 'Egress estimado (${_egressGB.toStringAsFixed(2)} GB)',
                      rate: 'R\$ ${_egressRatePerGB.toStringAsFixed(3)}/GB',
                      value: _egressGB * _egressRatePerGB,
                      color: const Color(0xFF7B2FFF),
                    ),
                    const Divider(color: Colors.white12, height: 20),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'TOTAL ESTIMADO/MÊS',
                          style: TextStyle(
                            fontSize: 12,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                            letterSpacing: 1,
                          ),
                        ),
                        Text(
                          'R\$ ${_monthlyEstimate.toStringAsFixed(2)}',
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: Color(0xFF00FF88),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),

              // Potential earnings
              _buildCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.trending_up, color: Color(0xFF00FF88), size: 18),
                        const SizedBox(width: 8),
                        const Text(
                          'Potencial com Capacidade Total',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Colors.white,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        _PotentialCard(
                          label: 'Capacidade\nDisponível',
                          value: '${_storageCapacityGB.toStringAsFixed(0)} GB',
                          color: const Color(0xFF00D4FF),
                        ),
                        const Icon(Icons.arrow_forward, color: Colors.white24),
                        _PotentialCard(
                          label: 'Potencial\nMensal',
                          value: 'R\$ ${potentialMonthly.toStringAsFixed(2)}',
                          color: const Color(0xFF00FF88),
                        ),
                        const Icon(Icons.arrow_forward, color: Colors.white24),
                        _PotentialCard(
                          label: 'Potencial\nAnual',
                          value: 'R\$ ${(potentialMonthly * 12).toStringAsFixed(0)}',
                          color: const Color(0xFF7B2FFF),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: const Color(0xFF00FF88).withOpacity(0.05),
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(
                          color: const Color(0xFF00FF88).withOpacity(0.2),
                        ),
                      ),
                      child: Row(
                        children: [
                          const Icon(Icons.info_outline,
                              color: Color(0xFF00FF88), size: 14),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              'Quanto mais fragmentos armazenados, maiores os ganhos. '
                              'Mantenha o nó online 24/7 para maximizar a receita.',
                              style: TextStyle(
                                fontSize: 11,
                                color: Colors.white.withOpacity(0.6),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),

              // Divisão visual
              _buildCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Modelo de Divisão',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: Colors.white,
                      ),
                    ),
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Expanded(
                          child: Column(
                            children: [
                              Container(
                                height: 8,
                                decoration: BoxDecoration(
                                  color: const Color(0xFF00D4FF),
                                  borderRadius: BorderRadius.circular(4),
                                ),
                              ),
                              const SizedBox(height: 8),
                              const Text(
                                '50%',
                                style: TextStyle(
                                  fontSize: 20,
                                  fontWeight: FontWeight.bold,
                                  color: Color(0xFF00D4FF),
                                ),
                              ),
                              Text(
                                'Você (Afiliado)',
                                style: TextStyle(
                                  fontSize: 11,
                                  color: Colors.white.withOpacity(0.5),
                                ),
                              ),
                            ],
                          ),
                        ),
                        Container(
                          width: 1,
                          height: 50,
                          color: Colors.white.withOpacity(0.1),
                          margin: const EdgeInsets.symmetric(horizontal: 16),
                        ),
                        Expanded(
                          child: Column(
                            children: [
                              Container(
                                height: 8,
                                decoration: BoxDecoration(
                                  color: const Color(0xFF7B2FFF),
                                  borderRadius: BorderRadius.circular(4),
                                ),
                              ),
                              const SizedBox(height: 8),
                              const Text(
                                '50%',
                                style: TextStyle(
                                  fontSize: 20,
                                  fontWeight: FontWeight.bold,
                                  color: Color(0xFF7B2FFF),
                                ),
                              ),
                              Text(
                                'Ertmann Tech',
                                style: TextStyle(
                                  fontSize: 11,
                                  color: Colors.white.withOpacity(0.5),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
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
}

class _EarningsRow extends StatelessWidget {
  final String label;
  final String rate;
  final double value;
  final Color color;

  const _EarningsRow({
    required this.label,
    required this.rate,
    required this.value,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: const TextStyle(color: Colors.white, fontSize: 13)),
              Text(rate, style: TextStyle(color: Colors.white.withOpacity(0.4), fontSize: 11)),
            ],
          ),
        ),
        Text(
          'R\$ ${value.toStringAsFixed(4)}',
          style: TextStyle(
            color: color,
            fontWeight: FontWeight.bold,
            fontSize: 14,
          ),
        ),
      ],
    );
  }
}

class _PotentialCard extends StatelessWidget {
  final String label;
  final String value;
  final Color color;

  const _PotentialCard({
    required this.label,
    required this.value,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(
          value,
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.bold,
            color: color,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          textAlign: TextAlign.center,
          style: TextStyle(
            fontSize: 10,
            color: Colors.white.withOpacity(0.4),
          ),
        ),
      ],
    );
  }
}
