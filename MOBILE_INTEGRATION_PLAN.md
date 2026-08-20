# Mobile Integration Plan — ONA VA BOLA / Clinicos

Документ описывает, как подключить Flutter-приложение к существующему backend без дублирования БД и без переписывания рабочей веб-системы.

## 1. Backend stack

| Слой | Технология |
|------|------------|
| API | Go 1.24, `chi` router, `pgx`, JWT, Argon2id |
| Prefix | `/api/v1` |
| Auth | Access JWT (~15m) + Refresh JWT (ротация, SHA-256 hash в `sessions`) |
| RBAC | Permission-based (`branches:read`, `clinical:write`, …), deny-by-default |
| Cache / rate limit | Redis (login lockout и т.п.) |
| Files | MinIO / S3 (private), выдача через backend |
| Realtime сейчас | **Нет WebSocket/Socket.IO** — веб использует HTTP polling |

## 2. Frontend stack (web)

- Next.js 15 App Router + React 19
- BFF через `/api/session/*` и `/api/clinic/*` (HttpOnly cookies)
- Локализация RU / UZ / EN
- React Query, Zod

## 3. Database

- **PostgreSQL 17** — единственный источник истины
- Tenant: `organization_id` на сущностях
- Multi-branch: `branches`, `branch_id` на приёмах/кассах/стационаре
- Шифрование чувствительных полей: `pgcrypto` (`phone_encrypted`, паспорт, адрес)
- Audit: `audit_logs` (append-only)
- Migrations: `apps/api/migrations` (000001…000034+)

Ключевые домены уже в БД: organizations, users, employees, roles, patients, appointments, doctor_schedules, clinical profiles/allergies/history, lab panels + attachments, imaging, diagnoses + clinical orders (в т.ч. MEDICATION), laboratory workflow, services/prices, inpatient, inventory, cash desk, accounting, patient billing, internal_messages.

## 4. Authentication (как есть)

```
POST /api/v1/auth/login
Body: { organizationId, login, password }

POST /api/v1/auth/refresh
Body: { refreshToken }

POST /api/v1/auth/logout
GET  /api/v1/auth/me  → { id, organizationId, permissions[] }
```

- Логин сотрудника: **login + password + organizationId** (не телефон/OTP).
- Пациенты в `patients` — **клинические карточки без `user_id`**; роли `PATIENT` нет.
- Системные роли seed: `OWNER`, `DOCTOR_*`, `ACCOUNTANT`, `MANAGER`, `CASHIER`, `RECEPTIONIST`, `SPEECH_THERAPIST` (+ `ADMIN`/`NURSE` упоминаются в docs, в seed частично иначе).
- Врачи = `DOCTOR_%`, не единый `DOCTOR`.

## 5. Existing APIs (использовать, не дублировать)

### Auth / identity
- `/auth/login|refresh|logout|me`

### Org / staff
- `/branches`, `/employees`, `/roles`, `/permissions`, `/specialists`

### Patients & clinical
- `/patients`, `/patients/{id}/clinical`, clinical profile/blood/allergies/vaccinations
- lab-results / urine-panel / blood-panel / imaging / diagnoses / orders / history
- `/lab-attachments/{id}`, `/imaging/{id}/file`
- `/laboratory/tests|orders`, `/diagnosis-catalog`

### Appointments & schedule
- `/appointments`, `/appointments/dashboard`
- `/appointments` POST, `/appointments/{id}/status` PATCH
- `/doctor-schedules/{employeeId}` GET/PUT
- Overlap запрещён DB exclusion constraint

### Messaging (staff only)
- `/messages/contacts`, `/messages/{userId}` GET/POST
- `/messages-archive` (owner/audit)

### Finance / ops
- `/cash`, cash shifts/transactions, patient-account, debts
- `/accounting`, entries, obligations, payroll
- `/reports/summary`
- `/services`, prices, providers
- `/inpatient/*`, `/inventory/*`, `/backups/export`, `/audit`

## 6. Gaps for mobile (что нужно добавить аккуратно)

| Gap | Почему нужно | Предлагаемое решение |
|-----|--------------|----------------------|
| Patient identity | Нет входа пациента | `patients.user_id` nullable + роль `PATIENT` + узкие permissions (`mobile:patient:*`) |
| Phone / OTP login | Требование ТЗ | Опционально поверх users; SMS-провайдер позже; MVP: phone-as-login + password |
| `/auth/me` без ролей/профиля | Mobile UI routing | Расширить `me`: roles, employee/patient profile, displayName |
| Public doctors catalog | Staff API слишком широкий | `GET /api/v1/mobile/doctors` — только публичные поля + availability |
| Availability slots | Есть schedule + appointments, нет slot API | `GET .../doctors/{id}/availability?date=` на базе schedule − busy |
| Doctor live status | Нет сущности статуса | Таблица/Redis presence + endpoint (+ позже WS) |
| Patient-scoped appointments/clinical | Сейчас org-wide staff access | Mobile handlers: всегда `patient_id` из principal |
| Patient↔doctor chat | Есть только staff internal chat | Новая таблица `patient_conversations` + policy; не смешивать с `internal_messages` |
| Push / FCM | Нет | `device_tokens` + worker/outbox (этап позже) |
| WebSocket | Нет | Сначала polling/SSE; WS поверх Redis pub/sub без ломки web |
| White-label branding | Только org.name | `organizations` branding JSON или `organization_settings` |
| Online payments UZ | Только касса (cash/card/transfer/debt) | Не подключать фиктивно; адаптер под Click/Payme позже |
| Cabinets / rating / languages / photo | Частично отсутствуют | Добавлять поля только при необходимости UI |

**Правило:** не создавать второй набор CRUD для staff-функций. Mobile либо вызывает существующие endpoints (doctor/owner/accountant), либо тонкий `/api/v1/mobile/*` facade с жёстким scoping.

## 7. Mobile architecture

```
Flutter (Riverpod + Dio + go_router)
        │ HTTPS Bearer JWT
        ▼
Existing Go API (/api/v1 + /api/v1/mobile/*)
        │
        ▼
PostgreSQL (та же БД) + Redis + MinIO
```

- Приложение живёт в `apps/mobile` (монорепо) — одна кодовая база Android/iOS.
- Feature-based `lib/`, RBAC guards на router **и** на backend.
- Secure Storage для tokens.
- Offline: только нечувствительный кэш (профиль публичный, список врачей); clinical — минимально и шифрованно/с TTL.

### Role → UI map

| Backend role codes | Mobile shell |
|--------------------|--------------|
| `PATIENT` | Patient nav |
| `DOCTOR_%` (+ later `DOCTOR`) | Doctor nav |
| `OWNER`, `ADMIN` | Owner nav |
| `ACCOUNTANT` | Accountant nav |
| `RECEPTIONIST`, `NURSE`, … | Зарезервировано (extensible) |

## 8. Security

- HTTPS only in production
- JWT + refresh rotation (уже есть)
- Server-side permission checks (уже есть) + patient self-scope
- Rate limit login (уже Redis)
- Audit на клинические/финансовые изменения (уже паттерн)
- Не логировать PII; не отдавать чужие patient records
- Accountant: finance endpoints only — без clinical dump

## 9. Realtime plan

1. **Phase A:** polling / short intervals (как web) для chat unread, appointment status, doctor presence.
2. **Phase B:** Redis pub/sub + SSE or WebSocket channel `/api/v1/realtime` — без изменения бизнес-логики.
3. Push: FCM/APNs после device registration API.

## 10. Backend changes (минимальный набор для MVP patient)

1. Migration: `PATIENT` role + permissions; `patients.user_id`; optional branding columns.
2. Extend `GET /auth/me` with roles + linked patient/employee.
3. Mobile patient endpoints (facade, reuse queries):
   - doctors list/detail/availability
   - my appointments CRUD (create uses existing appointment rules)
   - my clinical record / labs / orders / timeline
4. Doctor status MVP (optional Redis key).
5. CORS: добавить mobile scheme / deep link origins при необходимости (mobile использует Bearer напрямую, не cookies).

**Не трогать** без нужды: cash, inventory, inpatient, backup export, web BFF cookies.

## 11. Flutter delivery stages

| Stage | Scope |
|-------|--------|
| 1 | Analysis + this plan |
| 2 | Flutter scaffold: theme, l10n, Dio, secure storage, go_router, auth |
| 3 | Patient: login → home → doctors → book → my appointments |
| 4 | Medical record / labs / prescriptions / documents |
| 5 | Chat (patient↔doctor) |
| 6 | Doctor role |
| 7 | Owner dashboard (reuse reports/cash summaries) |
| 8 | Accountant |
| 9 | Push |
| 10 | QA Android/iOS |

## 12. Risks / problems found

1. **Нет patient accounts** — главный блокер patient app.
2. **Auth model = orgId + login**, не phone+OTP.
3. **Нет WebSocket** — realtime нужно строить отдельно.
4. **Internal chat ≠ patient chat**.
5. **`/auth/me` не отдаёт roles/name** — неудобно для mobile shell.
6. **Врачебные роли дробные** (`DOCTOR_DENTIST`…) — mobile должен маппить `role.code.startsWith('DOCTOR')`.
7. **ARCHITECTURE.md** упоминает NestJS в диаграмме — фактически API на **Go**; ориентироваться на код.
8. **packages/shared** в docs — папка packages пуста; DTO держать в mobile + OpenAPI/ручные models.
9. **Онлайн-платежи Узбекистана отсутствуют** — только офлайн-касса.
10. Workspace `D:\Medmobill` был пуст; рабочая система — **`D:\med`**.

## 13. Decision: where mobile code lives

- **`D:\med\apps\mobile`** — Flutter app внутри монорепо clinicos (предпочтительно).
- Backend изменения только по отдельному согласованию владельца.
- Документ: этот файл в корне `D:\med`.

## 14. Constraint (2026-08-09)

По запросу владельца: **не изменять существующий backend/web**.  
Mobile использует только текущие endpoints. Недостающие API (PATIENT login, presence, patient chat, push) откладываются и не эмулируются как «готовая интеграция».
