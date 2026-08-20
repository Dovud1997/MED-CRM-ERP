import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:clinicos_mobile/main.dart';

void main() {
  testWidgets('App boots to splash/login shell', (tester) async {
    await tester.pumpWidget(const ProviderScope(child: ClinicosApp()));
    await tester.pump();
    expect(find.byType(ClinicosApp), findsOneWidget);
  });
}
