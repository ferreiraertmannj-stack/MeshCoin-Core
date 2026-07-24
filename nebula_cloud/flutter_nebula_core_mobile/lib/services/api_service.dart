import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

class ApiService {
  static String? _baseUrl;
  static String? _token;

  static Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _baseUrl = prefs.getString('core_url');
    _token = prefs.getString('admin_token');
  }

  static Future<void> setCredentials(String url, String token) async {
    _baseUrl = url;
    _token = token;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('core_url', url);
    await prefs.setString('admin_token', token);
  }

  static Future<void> logout() async {
    _baseUrl = null;
    _token = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('core_url');
    await prefs.remove('admin_token');
  }

  static Map<String, String> get _headers => {
    'Content-Type': 'application/json',
    if (_token != null) 'Authorization': 'Bearer $_token',
    'x-nebula-key': 'nebula-node-secret-2026',
  };

  static Future<Map<String, dynamic>?> getNetworkStats() async {
    await init();
    if (_baseUrl == null) return null;
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/trpc/admin.networkStats'),
        headers: _headers,
      ).timeout(const Duration(seconds: 10));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return data['result']?['data']?['json'] as Map<String, dynamic>?;
      }
    } catch (_) {}
    return null;
  }

  static Future<List<dynamic>> getNodes() async {
    await init();
    if (_baseUrl == null) return [];
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/trpc/nodes.list'),
        headers: _headers,
      ).timeout(const Duration(seconds: 10));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return data['result']?['data']?['json'] as List<dynamic>? ?? [];
      }
    } catch (_) {}
    return [];
  }

  static Future<List<dynamic>> getLogs({int limit = 50, String? severity}) async {
    await init();
    if (_baseUrl == null) return [];
    try {
      final params = jsonEncode({'json': {'limit': limit, if (severity != null) 'severity': severity}});
      final response = await http.get(
        Uri.parse('$_baseUrl/api/trpc/logs.list?input=${Uri.encodeComponent(params)}'),
        headers: _headers,
      ).timeout(const Duration(seconds: 10));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return data['result']?['data']?['json'] as List<dynamic>? ?? [];
      }
    } catch (_) {}
    return [];
  }

  static Future<List<dynamic>> getAffiliates() async {
    await init();
    if (_baseUrl == null) return [];
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/trpc/admin.listAffiliates'),
        headers: _headers,
      ).timeout(const Duration(seconds: 10));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return data['result']?['data']?['json'] as List<dynamic>? ?? [];
      }
    } catch (_) {}
    return [];
  }

  static Future<List<dynamic>> getNotifications() async {
    await init();
    if (_baseUrl == null) return [];
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/api/trpc/notifications.list'),
        headers: _headers,
      ).timeout(const Duration(seconds: 10));
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return data['result']?['data']?['json'] as List<dynamic>? ?? [];
      }
    } catch (_) {}
    return [];
  }

  static Future<bool> testConnection(String url) async {
    try {
      final response = await http.get(
        Uri.parse('$url/api/trpc/nodes.list'),
        headers: {'x-nebula-key': 'nebula-node-secret-2026'},
      ).timeout(const Duration(seconds: 8));
      return response.statusCode == 200;
    } catch (_) {
      return false;
    }
  }
}
