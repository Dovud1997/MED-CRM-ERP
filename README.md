# ONA VA BOLA KLINIKASI

API-first система управления частной клиникой. Backend написан на Go. Текущий этап реализует Foundation: аутентификацию, permission-based роли, сотрудников, филиалы, сессии и аудит на PostgreSQL.

## Запуск на Windows

Требования: Git, Node.js 22 LTS, pnpm и Docker Desktop. Go локально не обязателен при запуске через Docker.

1. Скопируйте `.env.example` в `.env`.
2. Замените все значения с `change-me`; для JWT используйте случайные строки не короче 32 байт.
3. Выполните `docker compose up --build`.
4. Compose автоматически применит SQL migrations отдельным сервисом `migrate` до запуска API.
5. Откройте `http://localhost`; Swagger (development) — `http://localhost/api/docs`.

Создание локальной организации и владельца:

```powershell
docker compose build api
docker compose --profile tools run --rm seed
```

Логин владельца задаётся в `SEED_OWNER_LOGIN`, пароль — в `SEED_OWNER_PASSWORD`. Seed запрещён при `NODE_ENV=production`, требует пароль не короче 12 символов и безопасно запускается повторно без дублирования организации, филиалов и владельца.

Локальный `.env`, создаваемый для тестирования, содержит только development-секреты. Никогда не переносите его в production; `.gitignore` исключает файл из репозитория.

Web-аутентификация использует `HttpOnly`, `SameSite=Strict` cookies. Для HTTPS production обязательно установите `COOKIE_SECURE=true`.

Локальная разработка без контейнеров приложений:

```powershell
pnpm install
docker compose up -d postgres redis minio
pnpm dev
```

Frontend quality gate: `pnpm lint`, `pnpm typecheck`, `pnpm test`, `pnpm build`. Backend: `cd apps/api`, затем `go vet ./...`, `go test ./...`, `go build ./cmd/server`.

Проект не содержит production-паролей или тестовых медицинских записей. Первого владельца следует создавать отдельной bootstrap-командой/администратором перед развёртыванием (следующая задача hardening).
