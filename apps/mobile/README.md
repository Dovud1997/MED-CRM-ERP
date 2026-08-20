# Clinicos Mobile (Flutter)

Мобильный клиент для существующего Go API (`/api/v1`). Отдельной БД нет.

## Требования

1. Установить [Flutter SDK](https://docs.flutter.dev/get-started/install) (3.24+).
2. Запущенный backend из корня монорепо (`docker compose up` или `pnpm dev`).

## Первый запуск

```powershell
cd D:\med\apps\mobile
flutter create . --project-name clinicos_mobile --org uz.clinicos
flutter pub get
flutter run
```

`flutter create .` добавит платформенные папки `android/` / `ios/` поверх уже готового `lib/`.

### API URL

- Android emulator → `http://10.0.2.2/api/v1` (через nginx) или `http://10.0.2.2:4000/api/v1`
- iOS simulator → `http://127.0.0.1/api/v1`
- Device → IP машины в LAN

Переопределение:

```powershell
flutter run --dart-define=API_BASE_URL=http://192.168.1.10/api/v1 --dart-define=DEFAULT_ORGANIZATION_ID=...
```

## Auth

Использует существующие endpoints:

- `POST /auth/login` `{ organizationId, login, password }`
- `POST /auth/refresh`
- `GET /auth/me` (расширен: `roles`, `displayName`, `employeeId`, `patientId`)

Токены в Flutter Secure Storage.

## Статус

См. `../../MOBILE_INTEGRATION_PLAN.md`.

**Backend / web не изменяются** — mobile только читает существующие `/api/v1/*`.

Сейчас подключено к реальному API:

- login / refresh / me
- врачи ← `/employees` + `/specialists` + `/doctor-schedules/{id}`
- слоты (клиентский расчёт) ← schedule − `/appointments?date=`
- запись ← `POST /appointments` (нужен patientId; без PATIENT-аккаунта — поиск пациента)
- мои/активные записи ← `/appointments`
- doctor/owner dashboard ← `/appointments/dashboard`, `/reports/summary`

Пока нет в backend (не мокаем как готовое): live-статус врача, patient↔doctor chat, patient self-login, FCM, online payments.

