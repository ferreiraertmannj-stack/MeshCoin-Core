import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:path_provider/path_provider.dart';

class NodeStatus {
  final bool isOnline;
  final int fragmentCount;
  final double storageUsedGB;
  final double storageCapacityGB;
  final double latencyMs;
  final DateTime lastSeen;
  final String nodeId;
  final String nodeName;

  NodeStatus({
    required this.isOnline,
    required this.fragmentCount,
    required this.storageUsedGB,
    required this.storageCapacityGB,
    required this.latencyMs,
    required this.lastSeen,
    required this.nodeId,
    required this.nodeName,
  });
}

class NodeService {
  static const String _keyNodeId = 'node_id';
  static const String _keyNodeName = 'node_name';
  static const String _keyCoreUrl = 'core_url';
  static const String _keyApiKey = 'api_key';
  static const String _keyConfigured = 'node_configured';
  static const String _keyStorageCapacity = 'storage_capacity_gb';

  static Timer? _heartbeatTimer;
  static Timer? _statusTimer;
  static bool _isRunning = false;
  static NodeStatus? _lastStatus;
  static final List<Map<String, dynamic>> _localLogs = [];

  static Future<void> initialize() async {
    final prefs = await SharedPreferences.getInstance();
    if (prefs.getBool(_keyConfigured) ?? false) {
      startDaemon();
    }
  }

  static Future<Map<String, String>> getConfig() async {
    final prefs = await SharedPreferences.getInstance();
    return {
      'nodeId': prefs.getString(_keyNodeId) ?? '',
      'nodeName': prefs.getString(_keyNodeName) ?? '',
      'coreUrl': prefs.getString(_keyCoreUrl) ?? '',
      'apiKey': prefs.getString(_keyApiKey) ?? '',
      'storageCapacityGb': prefs.getDouble(_keyStorageCapacity)?.toString() ?? '50',
    };
  }

  static Future<void> saveConfig({
    required String nodeId,
    required String nodeName,
    required String coreUrl,
    required String apiKey,
    required double storageCapacityGb,
  }) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_keyNodeId, nodeId);
    await prefs.setString(_keyNodeName, nodeName);
    await prefs.setString(_keyCoreUrl, coreUrl);
    await prefs.setString(_keyApiKey, apiKey);
    await prefs.setDouble(_keyStorageCapacity, storageCapacityGb);
    await prefs.setBool(_keyConfigured, true);
  }

  static void startDaemon() {
    if (_isRunning) return;
    _isRunning = true;
    _addLog('info', 'Daemon Nébula Node iniciado');
    // Heartbeat a cada 30 segundos
    _heartbeatTimer = Timer.periodic(const Duration(seconds: 30), (_) => _sendHeartbeat());
    // Status local a cada 10 segundos
    _statusTimer = Timer.periodic(const Duration(seconds: 10), (_) => _updateLocalStatus());
    _sendHeartbeat();
  }

  static void stopDaemon() {
    _heartbeatTimer?.cancel();
    _statusTimer?.cancel();
    _isRunning = false;
    _addLog('info', 'Daemon Nébula Node parado');
  }

  static bool get isRunning => _isRunning;
  static NodeStatus? get lastStatus => _lastStatus;
  static List<Map<String, dynamic>> get logs => List.unmodifiable(_localLogs);

  static Future<void> _sendHeartbeat() async {
    try {
      final config = await getConfig();
      final coreUrl = config['coreUrl']!;
      final nodeId = config['nodeId']!;
      final apiKey = config['apiKey']!;
      final capacity = double.tryParse(config['storageCapacityGb'] ?? '50') ?? 50.0;

      if (coreUrl.isEmpty || nodeId.isEmpty) return;

      // Medir latência
      final stopwatch = Stopwatch()..start();
      final response = await http.post(
        Uri.parse('$coreUrl/api/trpc/nodes.heartbeat'),
        headers: {
          'Content-Type': 'application/json',
          'x-nebula-key': apiKey,
        },
        body: jsonEncode({
          'json': {
            'nodeId': nodeId,
            'storageUsedGB': await _getStorageUsed(),
            'storageCapacityGB': capacity,
            'fragmentCount': await _getFragmentCount(),
            'platform': 'android',
            'version': '1.0.0',
          }
        }),
      ).timeout(const Duration(seconds: 10));
      stopwatch.stop();

      if (response.statusCode == 200) {
        _addLog('success', 'Heartbeat enviado — ${stopwatch.elapsedMilliseconds}ms');
        await _sendLog(coreUrl, nodeId, apiKey, 'info', 'heartbeat',
            'Nó online — ${stopwatch.elapsedMilliseconds}ms');
      } else {
        _addLog('warning', 'Heartbeat falhou: ${response.statusCode}');
      }
    } catch (e) {
      _addLog('error', 'Erro no heartbeat: $e');
    }
  }

  static Future<void> _updateLocalStatus() async {
    try {
      final config = await getConfig();
      final storageUsed = await _getStorageUsed();
      final capacity = double.tryParse(config['storageCapacityGb'] ?? '50') ?? 50.0;
      final fragments = await _getFragmentCount();

      _lastStatus = NodeStatus(
        isOnline: _isRunning,
        fragmentCount: fragments,
        storageUsedGB: storageUsed,
        storageCapacityGB: capacity,
        latencyMs: 0,
        lastSeen: DateTime.now(),
        nodeId: config['nodeId'] ?? '',
        nodeName: config['nodeName'] ?? 'Meu Nó',
      );
    } catch (_) {}
  }

  static Future<double> _getStorageUsed() async {
    try {
      final dir = await getApplicationDocumentsDirectory();
      final fragDir = Directory('${dir.path}/nebula_fragments');
      if (!await fragDir.exists()) return 0.0;
      int totalBytes = 0;
      await for (final entity in fragDir.list(recursive: true)) {
        if (entity is File) {
          totalBytes += await entity.length();
        }
      }
      return totalBytes / (1024 * 1024 * 1024);
    } catch (_) {
      return 0.0;
    }
  }

  static Future<int> _getFragmentCount() async {
    try {
      final dir = await getApplicationDocumentsDirectory();
      final fragDir = Directory('${dir.path}/nebula_fragments');
      if (!await fragDir.exists()) return 0;
      int count = 0;
      await for (final entity in fragDir.list()) {
        if (entity is File) count++;
      }
      return count;
    } catch (_) {
      return 0;
    }
  }

  static Future<void> _sendLog(
    String coreUrl, String nodeId, String apiKey,
    String severity, String eventType, String message,
  ) async {
    try {
      await http.post(
        Uri.parse('$coreUrl/api/trpc/logs.create'),
        headers: {
          'Content-Type': 'application/json',
          'x-nebula-key': apiKey,
        },
        body: jsonEncode({
          'json': {
            'nodeId': nodeId,
            'severity': severity,
            'eventType': eventType,
            'message': message,
            'metadata': {'platform': 'android', 'app': 'nebula_node'},
          }
        }),
      ).timeout(const Duration(seconds: 5));
    } catch (_) {}
  }

  static void _addLog(String type, String message) {
    _localLogs.insert(0, {
      'type': type,
      'message': message,
      'timestamp': DateTime.now().toIso8601String(),
    });
    if (_localLogs.length > 200) _localLogs.removeLast();
  }

  static Future<bool> testConnection(String coreUrl, String nodeId, String apiKey) async {
    try {
      final response = await http.get(
        Uri.parse('$coreUrl/api/trpc/nodes.list'),
        headers: {'x-nebula-key': apiKey},
      ).timeout(const Duration(seconds: 8));
      return response.statusCode == 200;
    } catch (_) {
      return false;
    }
  }

  static Future<Map<String, dynamic>> getEarningsFromCore(String coreUrl, String nodeId, String apiKey) async {
    try {
      final response = await http.get(
        Uri.parse('$coreUrl/api/trpc/earnings.myNode?input=${Uri.encodeComponent(jsonEncode({'json': {'nodeId': nodeId}}))}'),
        headers: {'x-nebula-key': apiKey},
      ).timeout(const Duration(seconds: 10));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return data['result']?['data']?['json'] ?? {};
      }
    } catch (_) {}
    return {};
  }
}
