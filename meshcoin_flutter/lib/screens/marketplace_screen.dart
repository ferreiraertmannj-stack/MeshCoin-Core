import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../theme/app_theme.dart';
import '../mesh_node.dart';
import '../blockchain/ledger.dart';

class MarketplaceScreen extends StatelessWidget {
  final MeshNode meshNode;
  final Ledger ledger;
  final String address;
  
  const MarketplaceScreen({
    super.key,
    required this.meshNode,
    required this.ledger,
    required this.address,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.storefront, size: 80, color: MeshColors.neonCyan),
          const SizedBox(height: 20),
          Text(
            'Marketplace P2P',
            style: GoogleFonts.inter(
              color: MeshColors.textPrimary,
              fontSize: 24,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 10),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 40),
            child: Text(
              'Troque NBL por BTC ou XMR de forma anônima via contratos de Escrow (Fiador). Em breve.',
              textAlign: TextAlign.center,
              style: GoogleFonts.inter(
                color: MeshColors.textSecondary,
                fontSize: 14,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
