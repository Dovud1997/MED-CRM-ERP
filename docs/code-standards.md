# Стандарты кода

**Проект**: ONA VA BOLA KLINIKASI (`clinicos`)
**Обновлено**: 2026-08-21
**Применимо к**: `apps/api` (Go), `apps/web` (TypeScript/React), `apps/mobile` (Dart/Flutter)

## Общие принципы

- **YAGNI** — не строить инфраструктуру под гипотетические требования.
- **KISS** — простое решение предпочтительнее «умного».
- **DRY** — единый источник истины, общая логика выносится.
- **TDD** — production-код появляется под падающий тест (RED → GREEN → REFACTOR).
- **API-first** — бизнес-правила в backend, клиенты не дублируют логику и БД.
- **Безопасность по умолчанию** — deny-by-default RBAC, tenant scope в каждом запросе,
  никакого PII в логах, параметризованные SQL-запросы.

Верификация перед словом «готово»: запустить проверочные команды (тесты, сборка,
линт), прочитать полный вывод и код возврата.

## Именование файлов

- Каталоги и файлы документации/конфигов — **kebab-case** с длинными описательными
  именами (самодокументирование для поиска). Длинное имя — норма.
- Go-файлы — короткие имена по домену (`patients.go`, `cash.go`), тесты — `*_test.go`.
- React/Next: маршруты App Router и CSS — kebab-case (`doctor-schedules.css`).
- Dart: файлы — snake_case (соглашение Flutter).

Файлы длиннее ~200 строк — кандидаты на модуляризацию по логическим границам
(функции, классы, зоны ответственности). Исключения: markdown, plain text, bash,
конфиги, env-файлы.

## Backend — Go (`apps/api`)

### Стиль
- Идиоматичный Go, форматирование `gofmt` (обязательно перед коммитом).
- Пакетная организация: `cmd/*` — точки входа, `internal/app` — HTTP-слой и домены,
  `internal/auth` — аутентификация и permissions.
- Ошибки — явные, оборачиваются с контекстом; без паник в бизнес-логике
  (глобальный `recoverer` middleware — последняя линия защиты).

### HTTP и слои
- Роутер — `chi`. Все маршруты регистрируются в `internal/app/http.go`.
- Middleware цепочка: `recoverer` → `requestID` (`X-Request-ID`) →
  `securityHeaders` (`X-Content-Type-Options: nosniff`, `Referrer-Policy: no-referrer`,
  `Cache-Control: no-store`).
- Защищённые маршруты оборачиваются `authenticate` + `require("<permission>")`.
- Ответы — JSON. Ошибки имеют стабильную форму:
  `{"error":{"code","message","requestId"}}` и корректный HTTP-статус.
- Тело запроса декодируется с `DisallowUnknownFields` и лимитом размера.

### Данные
- Доступ к PostgreSQL — только через `pgx`, запросы параметризованы (`$1`, `$2`).
- Денежные значения — целые в минорных единицах; временные метки — `timestamptz`/UTC.
- Чувствительные поля (паспорт, адрес, приватный телефон) — `pgcrypto`; для
  дедупликации телефона — необратимый хеш нормализованного номера.
- Значимые операции пишутся в append-only `audit_logs` без клинического содержимого.
- Схема меняется только через версионированные миграции в `apps/api/migrations`
  (пара `.up.sql` / `.down.sql`), применяемые сервисом `migrate` до старта API.

### Безопасность
- Пароли — Argon2id (`internal/auth/password.go`).
- JWT: короткоживущий access + ротируемый refresh (в БД только SHA-256 hash).
- Никаких секретов в коде — только через окружение.

### Проверки (quality gate)
Из каталога `apps/api`:

```bash
go vet ./...
go test ./...
go build ./cmd/server
```

`gofmt -l .` не должен выводить файлов. Команды подтверждены по `README.md` и
`TASKS.md` (Go quality gate прогонялся в Go 1.24 контейнере).

## Веб-панель — TypeScript / React (`apps/web`)

### Стиль
- Next.js 15 App Router, React 19, TypeScript strict (`tsc --noEmit`).
- Валидация данных и форм — `zod` + `react-hook-form`.
- Серверное состояние — `@tanstack/react-query` (не хранить серверные данные в
  локальном стейте без нужды).
- BFF-слой (`src/app/api/*`) инкапсулирует доступ к API: `session/*` для
  аутентификации через HttpOnly `SameSite=Strict` cookies, `clinic/*` — прокси к
  защищённым эндпоинтам. Клиентские компоненты не ходят в backend напрямую с токеном.

### Проверки (quality gate)
Из каталога `apps/web` (или через Turborepo из корня):

```bash
pnpm lint       # eslint . --max-warnings=0
pnpm typecheck  # tsc --noEmit
pnpm test       # vitest run
pnpm build      # next build
```

Скрипты подтверждены в `apps/web/package.json`. Линт запрещает предупреждения
(`--max-warnings=0`).

## Мобильное — Dart / Flutter (`apps/mobile`)

### Стиль
- Feature-based структура: `lib/core`, `lib/features`, `lib/shared`.
- Состояние — Riverpod (`flutter_riverpod` + генерация `riverpod_generator`).
- Модели — `freezed` + `json_serializable` (через `build_runner`).
- Навигация — `go_router`; сеть — `dio` c Bearer JWT; токены — `flutter_secure_storage`.
- Линты — `flutter_lints` (`analysis_options.yaml`).
- RBAC-гварды дублируются на роутере И на backend; чужие записи пациентов не отдаются.

### Проверки (quality gate)
Из каталога `apps/mobile` (требуется установленный Flutter SDK):

```bash
flutter pub get
dart run build_runner build --delete-conflicting-outputs
flutter analyze
flutter test
```

Примечание: команды приведены по конфигурации `pubspec.yaml`. В данном окружении
Flutter SDK не проверялся — считать инструкцию непроверенной до наличия SDK.

## Монорепо (корень)

Turborepo прогоняет задачи по всем приложениям:

```bash
pnpm dev        # turbo dev
pnpm build      # turbo build
pnpm lint       # turbo lint
pnpm typecheck  # turbo typecheck
pnpm test       # turbo test
pnpm format     # prettier --write .
```

Подтверждено в корневом `package.json` и `turbo.json`. Пакетный менеджер —
`pnpm@10.11.0`. Хуки — `husky` + `lint-staged` (prettier по `*.{ts,tsx,js,json,md,yml,yaml}`).

## Git

- Осмысленные сообщения коммитов, описывающие «почему».
- Не коммитить секреты; `.env` в `.gitignore`, шаблон — `.env.example`.

## Связанные документы

- [Обзор продукта (PDR)](./project-overview-pdr.md)
- [Сводка кодовой базы](./codebase-summary.md)
- [Системная архитектура](./system-architecture.md)
- Безопасность данных — `SECURITY.md` (корень репозитория)

## Открытые вопросы

1. Каталог `packages/` пуст — общих TS-пакетов и правил для них пока нет.
2. Полноценный CI-пайплайн для Go/Flutter в репозитории не оформлен в едином виде;
   quality gate описан как локальные команды по стекам.
</content>
