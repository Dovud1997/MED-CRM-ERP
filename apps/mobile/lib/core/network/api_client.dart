import 'dart:async';

import 'package:dio/dio.dart';
import 'package:clinicos_mobile/core/config/app_config.dart';
import 'package:clinicos_mobile/core/network/error_mapper.dart';
import 'package:clinicos_mobile/core/storage/token_storage.dart';

typedef OnSessionExpired = FutureOr<void> Function();

class ApiClient {
  ApiClient({
    required AppConfig config,
    required TokenStorage tokenStorage,
    OnSessionExpired? onSessionExpired,
  })  : _tokenStorage = tokenStorage,
        _onSessionExpired = onSessionExpired {
    _dio = Dio(
      BaseOptions(
        baseUrl: config.apiBaseUrl,
        connectTimeout: const Duration(seconds: 20),
        receiveTimeout: const Duration(seconds: 30),
        headers: const {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        },
      ),
    );

    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = await _tokenStorage.readAccessToken();
          if (token != null && token.isNotEmpty) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          handler.next(options);
        },
        onError: (error, handler) async {
          if (error.response?.statusCode == 401 &&
              !_isAuthPath(error.requestOptions.path)) {
            final refreshed = await _refreshTokens();
            if (refreshed) {
              final req = error.requestOptions;
              final token = await _tokenStorage.readAccessToken();
              req.headers['Authorization'] = 'Bearer $token';
              try {
                final response = await _dio.fetch(req);
                return handler.resolve(response);
              } catch (e) {
                return handler.next(error);
              }
            }
            await _tokenStorage.clear();
            await _onSessionExpired?.call();
          }
          handler.next(error);
        },
      ),
    );

    if (config.enableLogging) {
      _dio.interceptors.add(
        LogInterceptor(
          requestBody: true,
          responseBody: false,
          error: true,
          requestHeader: false,
          responseHeader: false,
        ),
      );
    }
  }

  late final Dio _dio;
  final TokenStorage _tokenStorage;
  final OnSessionExpired? _onSessionExpired;
  bool _refreshing = false;

  Dio get dio => _dio;

  bool _isAuthPath(String path) =>
      path.contains('/auth/login') ||
      path.contains('/auth/refresh') ||
      path.contains('/auth/logout');

  Future<bool> _refreshTokens() async {
    if (_refreshing) return false;
    _refreshing = true;
    try {
      final refresh = await _tokenStorage.readRefreshToken();
      if (refresh == null || refresh.isEmpty) return false;
      final response = await Dio(
        BaseOptions(baseUrl: _dio.options.baseUrl),
      ).post<Map<String, dynamic>>(
        '/auth/refresh',
        data: {'refreshToken': refresh},
      );
      final data = response.data;
      final access = data?['accessToken'] as String?;
      final nextRefresh = data?['refreshToken'] as String?;
      if (access == null || nextRefresh == null) return false;
      await _tokenStorage.saveTokens(
        accessToken: access,
        refreshToken: nextRefresh,
      );
      return true;
    } catch (_) {
      return false;
    } finally {
      _refreshing = false;
    }
  }

  Future<T> get<T>(
    String path, {
    Map<String, dynamic>? query,
    T Function(dynamic data)? parser,
  }) async {
    try {
      final response = await _dio.get<dynamic>(path, queryParameters: query);
      return parser != null ? parser(response.data) : response.data as T;
    } catch (e) {
      throw mapDioError(e);
    }
  }

  Future<T> post<T>(
    String path, {
    Object? data,
    T Function(dynamic data)? parser,
  }) async {
    try {
      final response = await _dio.post<dynamic>(path, data: data);
      return parser != null ? parser(response.data) : response.data as T;
    } catch (e) {
      throw mapDioError(e);
    }
  }

  Future<T> patch<T>(
    String path, {
    Object? data,
    T Function(dynamic data)? parser,
  }) async {
    try {
      final response = await _dio.patch<dynamic>(path, data: data);
      return parser != null ? parser(response.data) : response.data as T;
    } catch (e) {
      throw mapDioError(e);
    }
  }
}
