# Сводка кодовой базы

**Проект**: ONA VA BOLA KLINIKASI (`clinicos`)
**Обновлено**: 2026-08-21

## Обзор

`clinicos` — монорепозиторий на pnpm@10 + Turborepo. Содержит три приложения:
Go-бэкенд (`apps/api`), веб-панель на Next.js (`apps/web`) и мобильное приложение на
Flutter (`apps/mobile`). Инфраструктура описана в `compose.yaml` и `infra/`.

## Структура репозитория

```
MED-CRM-ERP/
├── apps/
│   ├── api/                  # Backend на Go
│   │   ├── cmd/
│   │   │   ├── server/       # HTTP-сервер (точка входа API)
│   │   │   ├── seed/         # Идемпотентный локальный seed организации и владельца
│   │   │   └── backupcheck/  # Проверка целостности архива .ovbk
│   │   ├── internal/
│   │   │   ├── app/          # Конфиг, роутер, middleware и все доменные хендлеры
│   │   │   └── auth/         # Argon2id-пароли, JWT, principal/permissions
│   │   ├── migrations/       # SQL-миграции up/down (000001…000038)
│   │   ├── Dockerfile
│   │   └── go.mod            # module clinicos/api, go 1.24
│   ├── web/                  # Веб-панель (@clinicos/web)
│   │   ├── src/
│   │   │   ├── app/          # Next.js App Router: страницы, CSS, BFF-роуты
│   │   │   │   ├── api/      # BFF: session/*, clinic/*, display/[token]
│   │   │   │   └── display/  # Публичный маршрут ТВ-очереди
│   │   │   ├── components/   # UI-компоненты
│   │   │   └── lib/          # Клиентские утилиты, доступ к API
│   │   └── package.json
│   └── mobile/               # Flutter-приложение (clinicos_mobile)
│       ├── lib/
│       │   ├── core/         # auth, config, localization, network, routing, storage, theme
│       │   ├── features/     # Функциональные модули (feature-based)
│       │   └── shared/       # models, widgets, demo
│       └── pubspec.yaml
├── infra/
│   └── nginx/default.conf    # Reverse proxy для web + api
├── docs/                     # Документация проекта
│   └── deployment/           # Эксплуатационные доки (очередь, backup/restore)
├── compose.yaml              # Docker Compose: postgres, redis, migrate, minio, api, web, nginx
├── pnpm-workspace.yaml       # Воркспейсы: apps/*, packages/*
├── turbo.json                # Задачи Turborepo: dev, build, lint, typecheck, test
├── package.json              # Корневой (clinicos), pnpm@10.11.0
├── README.md
├── ARCHITECTURE.md / SECURITY.md / ROADMAP.md / TASKS.md
└── MOBILE_INTEGRATION_PLAN.md
```

Примечание: `pnpm-workspace.yaml` объявляет `packages/*`, но каталог `packages/`
сейчас пуст (общих TS-пакетов нет). См. открытые вопросы.

## Backend (`apps/api`)

### Точки входа (`cmd/`)
- `server` — HTTP API, слушает порт `4000` (по умолчанию), префикс `/api/v1`.
- `seed` — создаёт организацию, филиалы, владельца и системные роли. Идемпотентен,
  запрещён при `NODE_ENV=production`, требует пароль ≥ 12 символов.
- `backupcheck` — проверяет целостность и расшифровку архива `.ovbk`
  (пароль передаётся через `BACKUP_PASSWORD`).

### Внутренние пакеты (`internal/`)
- `app/` — один пакет со всей HTTP-логикой:
  - `app.go` — сборка приложения (пул `pgx`, клиент Redis, auth-сервис).
  - `config.go` — загрузка конфигурации из окружения.
  - `http.go` — роутер `chi`, middleware, health-эндпоинты, регистрация всех маршрутов.
  - `handlers.go`, `config.go` и доменные файлы: `patients.go`, `appointments.go`,
    `clinical.go`, `doctor_schedules.go`, `laboratory.go`, `diagnoses.go`,
    `diagnosis_catalog.go`, `services.go`, `specialists.go`, `inpatient.go`,
    `inventory.go`, `inventory_accountability.go`, `cash.go`, `accounting.go`,
    `billing.go`, `reports.go`, `messages.go`, `queue.go`, `queue_media.go`,
    `document_templates.go`, `profile_photos.go`, `backup.go`, `management.go`.
  - Тесты: `doctor_schedules_test.go`, `queue_test.go`.
- `auth/` — `password.go` (Argon2id) + `password_test.go`, `service.go`
  (JWT, principal, проверка permissions).

### Ключевые технологии Go (`go.mod`)
`github.com/go-chi/chi/v5`, `github.com/jackc/pgx/v5`, `github.com/redis/go-redis/v9`,
`github.com/golang-jwt/jwt/v5`, `github.com/google/uuid`, `golang.org/x/crypto`.

### Миграции (`apps/api/migrations`)
38 версий (76 файлов `.up.sql` / `.down.sql`), от `000001_foundation` до
`000038_queue_display_brand_settings`. Каждая миграция очерчивает доменную область:
foundation/auth, management permissions, employee profile, patients, clinical records,
appointments, doctor schedules, internal messages, services/prices, specialties,
inpatient, reports, lab attachments, imaging, diagnoses/orders, laboratory workflow,
inventory, cash desk, accounting, patient billing, document templates, profile photos,
electronic queue. Применяются сервисом `migrate` до старта API.

## Веб-панель (`apps/web`)

- Пакет `@clinicos/web`, Next.js 15 (App Router), React 19.
- Данные: `@tanstack/react-query`; формы: `react-hook-form` + `zod`.
- Тесты: `vitest`; линт: `eslint` (`--max-warnings=0`); типы: `tsc --noEmit`.
- BFF-слой в `src/app/api/`:
  - `session/{login,logout,me}` — аутентификация через HttpOnly cookies.
  - `clinic/*` — прокси к защищённым эндпоинтам API (patients, appointments,
    cash, accounting, inventory, queue, reports, services, employees, roles и др.).
  - `display/[token]` — публичный прокси данных ТВ-очереди.
- Страницы и стили — в `src/app/*` (например `dashboard`, `appointments`, `cash`,
  `accounting`, `inpatient`, `inventory`, `laboratory`, `reports`, `messages`,
  `clinical`, `diagnosis-catalog`, `document-builder`); публичный `display/`.

**Скрипты** (`apps/web/package.json`): `dev`, `build`, `start`, `lint`, `typecheck`, `test`.

## Мобильное приложение (`apps/mobile`)

- Пакет `clinicos_mobile` (Flutter, Dart SDK `>=3.5.0 <4.0.0`).
- Состояние: `flutter_riverpod` + `riverpod_annotation`; навигация: `go_router`;
  сеть: `dio`; хранилище токенов: `flutter_secure_storage`; кодогенерация:
  `freezed` + `json_serializable` + `build_runner`.
- Локализация: `flutter_localizations` + `intl` (`generate: true`, папка `l10n`).
- Структура `lib/`: `core/` (auth, config, localization, network, routing, storage,
  theme), `features/` (auth, appointments, doctors, medical_record, patient,
  owner_dashboard, doctor_dashboard, accounting, favorites, search), `shared/`
  (models, widgets, demo).
- Статус: scaffold. Backend/web не меняются под mobile (решение владельца, 2026-08-09).

## Инфраструктура (`compose.yaml`)

| Сервис | Образ / контекст | Назначение |
|--------|------------------|------------|
| `postgres` | `postgres:17-alpine` | Источник истины, healthcheck `pg_isready` |
| `redis` | `redis:8-alpine` | Pub/Sub для очереди, rate-limit; AOF-персистенс |
| `migrate` | `migrate/migrate:v4.18.3` | Применяет SQL-миграции до старта API |
| `minio` | `minio/minio` | Приватное S3-совместимое хранилище |
| `queue-media-init` | `alpine:3.22` | Готовит volume `queue_media` (права/владелец) |
| `api` | build `./apps/api` | Go HTTP API (`:4000`) |
| `seed` | образ `med-api`, профиль `tools` | Разовый запуск локального seed |
| `web` | build (arg `APP=web`) | Next.js веб-панель |
| `nginx` | `nginx:1.29-alpine` | Reverse proxy, публикует порт `80` |

Сеть `private`, тома `postgres_data`, `redis_data`, `minio_data`, `queue_media`.

## Точки входа для новичка

- Запуск и быстрый старт — `README.md`.
- Архитектура и решения — `ARCHITECTURE.md`, `docs/system-architecture.md`.
- Безопасность данных — `SECURITY.md`.
- Регистр реализованных задач — `TASKS.md`.
- Роуты API — `apps/api/internal/app/http.go`.

## Открытые вопросы

1. `packages/` объявлен в воркспейсе, но пуст. `ARCHITECTURE.md` упоминает
   `packages/shared` с общими DTO/схемами — фактически такого пакета нет.
2. Точное число системных ролей в проде определяется seed-скриптом и может
   расходиться с текстами старых доков.
</content>
