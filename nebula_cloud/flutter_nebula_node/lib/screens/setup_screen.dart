import 'package:flutter/material.dart';
import 'package:uuid/uuid.dart';
import '../services/node_service.dart';
import '../main.dart';

class SetupScreen extends StatefulWidget {
  const SetupScreen({super.key});

  @override
  State<SetupScreen> createState() => _SetupScreenState();
}

class _SetupScreenState extends State<SetupScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nodeNameCtrl = TextEditingController(text: 'nebula_node_01');
  final _coreUrlCtrl = TextEditingController();
  final _apiKeyCtrl = TextEditingController(text: 'nebula-node-secret-2026');
  double _storageCapacity = 50.0;
  bool _isTesting = false;
  bool _testPassed = false;
  String _testMessage = '';
  int _step = 0;

  @override
  void dispose() {
    _nodeNameCtrl.dispose();
    _coreUrlCtrl.dispose();
    _apiKeyCtrl.dispose();
    super.dispose();
  }

  Future<void> _testConnection() async {
    setState(() {
      _isTesting = true;
      _testMessage = 'Testando conexão...';
    });
    final ok = await NodeService.testConnection(
      _coreUrlCtrl.text.trim(),
      'test',
      _apiKeyCtrl.text.trim(),
    );
    setState(() {
      _isTesting = false;
      _testPassed = ok;
      _testMessage = ok
          ? '✅ Conexão com Nébula Core estabelecida!'
          : '❌ Falha na conexão. Verifique a URL e a chave API.';
    });
  }

  Future<void> _saveAndStart() async {
    if (!_formKey.currentState!.validate()) return;
    final nodeId = 'no_${const Uuid().v4().substring(0, 8)}';
    await NodeService.saveConfig(
      nodeId: nodeId,
      nodeName: _nodeNameCtrl.text.trim(),
      coreUrl: _coreUrlCtrl.text.trim(),
      apiKey: _apiKeyCtrl.text.trim(),
      storageCapacityGb: _storageCapacity,
    );
    NodeService.startDaemon();
    if (mounted) {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => const MainScreen()),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: 20),
                // Logo
                Center(
                  child: Container(
                    width: 80,
                    height: 80,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      gradient: const LinearGradient(
                        colors: [Color(0xFF00D4FF), Color(0xFF7B2FFF)],
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF00D4FF).withOpacity(0.3),
                          blurRadius: 20,
                        ),
                      ],
                    ),
                    child: const Icon(Icons.hub, color: Colors.white, size: 40),
                  ),
                ),
                const SizedBox(height: 20),
                const Center(
                  child: Text(
                    'Configurar Nébula Node',
                    style: TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ),
                Center(
                  child: Text(
                    'Configure seu smartphone como nó da rede',
                    style: TextStyle(
                      fontSize: 13,
                      color: Colors.white.withOpacity(0.5),
                    ),
                  ),
                ),
                const SizedBox(height: 36),

                // Step indicators
                Row(
                  children: List.generate(3, (i) => Expanded(
                    child: Container(
                      margin: const EdgeInsets.symmetric(horizontal: 4),
                      height: 4,
                      decoration: BoxDecoration(
                        color: i <= _step
                            ? const Color(0xFF00D4FF)
                            : Colors.white.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                  )),
                ),
                const SizedBox(height: 28),

                // Step 1: Nome do nó
                _SectionTitle(
                  number: '1',
                  title: 'Identidade do Nó',
                  active: _step >= 0,
                ),
                const SizedBox(height: 12),
                _buildTextField(
                  controller: _nodeNameCtrl,
                  label: 'Nome do Nó',
                  hint: 'ex: nebula_infinix_01',
                  icon: Icons.label_outline,
                  validator: (v) => v?.isEmpty == true ? 'Obrigatório' : null,
                  onChanged: (_) => setState(() => _step = 0),
                ),
                const SizedBox(height: 24),

                // Step 2: Conexão
                _SectionTitle(
                  number: '2',
                  title: 'Conexão com Nébula Core',
                  active: _step >= 1,
                ),
                const SizedBox(height: 12),
                _buildTextField(
                  controller: _coreUrlCtrl,
                  label: 'URL do Nébula Core',
                  hint: 'ex: https://seu-dominio.manus.space',
                  icon: Icons.link,
                  keyboardType: TextInputType.url,
                  validator: (v) {
                    if (v?.isEmpty == true) return 'Obrigatório';
                    if (!v!.startsWith('http')) return 'URL inválida';
                    return null;
                  },
                  onChanged: (_) => setState(() {
                    _step = 1;
                    _testPassed = false;
                    _testMessage = '';
                  }),
                ),
                const SizedBox(height: 12),
                _buildTextField(
                  controller: _apiKeyCtrl,
                  label: 'Chave API',
                  hint: 'nebula-node-secret-2026',
                  icon: Icons.key_outlined,
                  obscureText: true,
                  validator: (v) => v?.isEmpty == true ? 'Obrigatório' : null,
                ),
                const SizedBox(height: 12),
                // Test button
                OutlinedButton.icon(
                  onPressed: _isTesting ? null : _testConnection,
                  icon: _isTesting
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(Icons.wifi_tethering),
                  label: Text(_isTesting ? 'Testando...' : 'Testar Conexão'),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: const Color(0xFF00D4FF),
                    side: const BorderSide(color: Color(0xFF00D4FF)),
                    minimumSize: const Size(double.infinity, 48),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
                if (_testMessage.isNotEmpty) ...[
                  const SizedBox(height: 10),
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: _testPassed
                          ? const Color(0xFF00FF88).withOpacity(0.08)
                          : Colors.red.withOpacity(0.08),
                      borderRadius: BorderRadius.circular(10),
                      border: Border.all(
                        color: _testPassed
                            ? const Color(0xFF00FF88).withOpacity(0.3)
                            : Colors.red.withOpacity(0.3),
                      ),
                    ),
                    child: Text(
                      _testMessage,
                      style: TextStyle(
                        color: _testPassed ? const Color(0xFF00FF88) : Colors.red,
                        fontSize: 13,
                      ),
                    ),
                  ),
                ],
                const SizedBox(height: 24),

                // Step 3: Storage
                _SectionTitle(
                  number: '3',
                  title: 'Capacidade de Armazenamento',
                  active: _step >= 2,
                ),
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
                            'Espaço para a rede Nébula',
                            style: TextStyle(
                              color: Colors.white.withOpacity(0.7),
                              fontSize: 13,
                            ),
                          ),
                          Text(
                            '${_storageCapacity.toStringAsFixed(0)} GB',
                            style: const TextStyle(
                              color: Color(0xFF00D4FF),
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ],
                      ),
                      Slider(
                        value: _storageCapacity,
                        min: 10,
                        max: 500,
                        divisions: 49,
                        activeColor: const Color(0xFF00D4FF),
                        inactiveColor: Colors.white.withOpacity(0.1),
                        onChanged: (v) => setState(() {
                          _storageCapacity = v;
                          _step = 2;
                        }),
                      ),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text('10 GB', style: TextStyle(color: Colors.white.withOpacity(0.3), fontSize: 11)),
                          Text(
                            'Ganho estimado: R\$ ${(_storageCapacity * 0.0115 * 30).toStringAsFixed(2)}/mês',
                            style: const TextStyle(
                              color: Color(0xFF00FF88),
                              fontSize: 12,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          Text('500 GB', style: TextStyle(color: Colors.white.withOpacity(0.3), fontSize: 11)),
                        ],
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 36),

                // Save button
                GestureDetector(
                  onTap: _saveAndStart,
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(vertical: 18),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [Color(0xFF00D4FF), Color(0xFF7B2FFF)],
                      ),
                      borderRadius: BorderRadius.circular(16),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF00D4FF).withOpacity(0.3),
                          blurRadius: 20,
                        ),
                      ],
                    ),
                    child: const Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.rocket_launch, color: Colors.white),
                        SizedBox(width: 10),
                        Text(
                          'ATIVAR NÓ NÉBULA',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                            letterSpacing: 2,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 20),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required String hint,
    required IconData icon,
    String? Function(String?)? validator,
    void Function(String)? onChanged,
    bool obscureText = false,
    TextInputType? keyboardType,
  }) {
    return TextFormField(
      controller: controller,
      validator: validator,
      onChanged: onChanged,
      obscureText: obscureText,
      keyboardType: keyboardType,
      style: const TextStyle(color: Colors.white),
      decoration: InputDecoration(
        labelText: label,
        hintText: hint,
        prefixIcon: Icon(icon, color: const Color(0xFF00D4FF), size: 20),
        labelStyle: TextStyle(color: Colors.white.withOpacity(0.5)),
        hintStyle: TextStyle(color: Colors.white.withOpacity(0.2)),
        filled: true,
        fillColor: const Color(0xFF0D1117),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.white.withOpacity(0.08)),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.white.withOpacity(0.08)),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFF00D4FF)),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Colors.red),
        ),
      ),
    );
  }
}

class _SectionTitle extends StatelessWidget {
  final String number;
  final String title;
  final bool active;

  const _SectionTitle({
    required this.number,
    required this.title,
    required this.active,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Container(
          width: 26,
          height: 26,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            color: active
                ? const Color(0xFF00D4FF)
                : Colors.white.withOpacity(0.1),
          ),
          child: Center(
            child: Text(
              number,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.bold,
                color: active ? Colors.black : Colors.white.withOpacity(0.3),
              ),
            ),
          ),
        ),
        const SizedBox(width: 10),
        Text(
          title,
          style: TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.w600,
            color: active ? Colors.white : Colors.white.withOpacity(0.4),
          ),
        ),
      ],
    );
  }
}
