import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:clinicos_mobile/features/auth/presentation/auth_controller.dart';
import 'package:clinicos_mobile/shared/demo/demo_data.dart';

const _prefsKey = 'favorite_doctor_ids';
const _seededKey = 'favorite_doctor_ids_seeded';

final favoritesControllerProvider =
    StateNotifierProvider<FavoritesController, Set<String>>((ref) {
  return FavoritesController();
});

class FavoritesController extends StateNotifier<Set<String>> {
  FavoritesController() : super(const {}) {
    _restore();
  }

  Future<void> _restore() async {
    final prefs = await SharedPreferences.getInstance();
    var ids = prefs.getStringList(_prefsKey) ?? <String>[];
    final seeded = prefs.getBool(_seededKey) ?? false;
    if (kSkipLogin && !seeded && ids.isEmpty) {
      ids = DemoData.doctors.take(2).map((d) => d.id).toList();
      await prefs.setStringList(_prefsKey, ids);
      await prefs.setBool(_seededKey, true);
    }
    state = ids.toSet();
  }

  Future<void> toggle(String doctorId) async {
    final next = {...state};
    if (next.contains(doctorId)) {
      next.remove(doctorId);
    } else {
      next.add(doctorId);
    }
    state = next;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setStringList(_prefsKey, next.toList());
  }

  bool isFavorite(String doctorId) => state.contains(doctorId);
}
