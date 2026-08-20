# ONA VA BOLA KLINIKASI — дорожная карта

## Этап 1 — Platform foundation (завершён)

- Монорепозиторий, документация, Docker Compose.
- PostgreSQL schema и миграции.
- Go API: login/refresh/logout, permission RBAC, сотрудники, филиалы, audit.
- Управление филиалами, сотрудниками, специальностями и настраиваемыми ролями через реальную PostgreSQL.
- Web: локализованная панель, пациенты и расписание.
- Unit/API smoke tests, lint, typecheck, build.

## Этап 2 — Пациенты и запись (в работе)

- Единая карточка пациента, контакты и согласия.
- Реестр пациентов, создание, поиск и защита чувствительных полей реализованы; далее — детальная карточка и согласия.
- Реализована первая версия детальной медицинской карты: группа крови, аллергии, повторные визиты и защищённый экспорт архива.
- Поиск, фильтрация, duplicate detection и merge workflow.
- Реализованы расписание и запись: недельные графики врачей, смены и перерывы, календарная дата, длительность, статусы приёма, отмена, защита от пересечений и записи вне смены; далее — перенос и waiting list.
- Call-center задачи и уведомления о записи.

## Этап 3 — Клинический контур

- Электронная медицинская карта, визиты, диагнозы, назначения, шаблоны протоколов.
- Private file service, антивирусное сканирование, consent management.
- BullMQ worker, SMS/email/Telegram notifications, outbox pattern.

## Этап 4 — Операционный контур

- Касса, счета, платежи, возвраты, услуги и прайс-листы.
- Начата реализация: рабочий справочник услуг и прайс-листы с ценами по филиалам и индивидуальными ценами врачей.
- Склад, партии, сроки годности, лаборатория.
- Аналитика, экспорт и регламентные отчёты.

## Этап 5 — Масштабирование

- Полный multi-branch UX и политики доступа по филиалам.
- Patient mobile app, doctor interface, Telegram Mini App.
- Tauri packaging, offline-safe read cache where legally permitted.
- HA deployment, observability, disaster-recovery exercises.

## Production gate

До обработки реальных медицинских данных обязательны threat model review, DPIA/локальная правовая экспертиза, backup restore drill, penetration test, incident response drill и user acceptance testing.
