import 'package:flutter/material.dart';
import 'package:clinicos_mobile/core/localization/app_localizations.dart';
import 'package:clinicos_mobile/core/network/api_exception.dart';

class AsyncBody<T> extends StatelessWidget {
  const AsyncBody({
    super.key,
    required this.value,
    required this.builder,
    this.onRetry,
    this.emptyMessage,
  });

  final AsyncValueLike<T> value;
  final Widget Function(T data) builder;
  final VoidCallback? onRetry;
  final String? emptyMessage;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    if (value.isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (value.error != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(_mapError(l10n, value.error!), textAlign: TextAlign.center),
              if (onRetry != null) ...[
                const SizedBox(height: 12),
                FilledButton(onPressed: onRetry, child: Text(l10n.retry)),
              ],
            ],
          ),
        ),
      );
    }
    final data = value.data;
    if (data == null || (data is List && (data as List).isEmpty)) {
      return Center(child: Text(emptyMessage ?? l10n.empty));
    }
    return builder(data as T);
  }

  String _mapError(AppLocalizations l10n, Object error) {
    if (error is NetworkException) return l10n.errorNetwork;
    if (error is TimeoutApiException) return l10n.errorTimeout;
    if (error is UnauthorizedException) return l10n.errorUnauthorized;
    if (error is ForbiddenException) return l10n.errorForbidden;
    if (error is NotFoundException) return l10n.errorNotFound;
    if (error is ValidationException) return l10n.errorValidation;
    if (error is ServerException) return l10n.errorServer;
    return l10n.errorGeneric;
  }
}

class AsyncValueLike<T> {
  const AsyncValueLike({this.data, this.error, this.isLoading = false});

  final T? data;
  final Object? error;
  final bool isLoading;

  factory AsyncValueLike.loading() => const AsyncValueLike(isLoading: true);
  factory AsyncValueLike.data(T data) => AsyncValueLike(data: data);
  factory AsyncValueLike.error(Object error) => AsyncValueLike(error: error);
}
