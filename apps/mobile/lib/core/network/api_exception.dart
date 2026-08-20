sealed class ApiException implements Exception {
  const ApiException(this.message, {this.code, this.statusCode});

  final String message;
  final String? code;
  final int? statusCode;
}

class NetworkException extends ApiException {
  const NetworkException([super.message = 'network']);
}

class TimeoutApiException extends ApiException {
  const TimeoutApiException([super.message = 'timeout']);
}

class UnauthorizedException extends ApiException {
  const UnauthorizedException([super.message = 'unauthorized'])
      : super(statusCode: 401, code: 'unauthorized');
}

class ForbiddenException extends ApiException {
  const ForbiddenException([super.message = 'forbidden'])
      : super(statusCode: 403, code: 'forbidden');
}

class NotFoundException extends ApiException {
  const NotFoundException([super.message = 'not_found'])
      : super(statusCode: 404, code: 'not_found');
}

class ValidationException extends ApiException {
  const ValidationException([super.message = 'validation_error'])
      : super(statusCode: 422, code: 'validation_error');
}

class ServerException extends ApiException {
  const ServerException([super.message = 'server'])
      : super(statusCode: 500, code: 'server_error');
}

class UnknownApiException extends ApiException {
  const UnknownApiException([super.message = 'unknown']);
}
