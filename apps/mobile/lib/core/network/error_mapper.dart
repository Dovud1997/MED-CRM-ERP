import 'package:dio/dio.dart';
import 'package:clinicos_mobile/core/network/api_exception.dart';

ApiException mapDioError(Object error) {
  if (error is DioException) {
    switch (error.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return const TimeoutApiException();
      case DioExceptionType.connectionError:
        return const NetworkException();
      case DioExceptionType.badResponse:
        final status = error.response?.statusCode;
        final code = _errorCode(error.response?.data);
        final message = _errorMessage(error.response?.data) ?? 'error';
        return switch (status) {
          401 => UnauthorizedException(message),
          403 => ForbiddenException(message),
          404 => NotFoundException(message),
          422 => ValidationException(message),
          final s when s != null && s >= 500 => ServerException(message),
          _ => UnknownApiException(code ?? message),
        };
      default:
        return UnknownApiException(error.message ?? 'unknown');
    }
  }
  return UnknownApiException(error.toString());
}

String? _errorCode(dynamic data) {
  if (data is Map && data['error'] is Map) {
    return data['error']['code']?.toString();
  }
  return null;
}

String? _errorMessage(dynamic data) {
  if (data is Map && data['error'] is Map) {
    return data['error']['message']?.toString();
  }
  return null;
}
