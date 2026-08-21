# Как использовать ClaudeKit в проекте clinicos

Практическое руководство по работе с AI-тулингом (ClaudeKit + Claude Code) в репозитории **ONA VA BOLA KLINIKASI** (`clinicos`).

> Обновлено: 2026-08-21. Стек проекта: Go API (`apps/api`) · Next.js web (`apps/web`) · Flutter mobile (`apps/mobile`) · pnpm + Turborepo.

---

## 1. Что это и где что лежит

| Каталог | Роль | Читает Claude Code? |
|---|---|---|
| `.claude/` | **Runtime** — то, что активно в сессии | ✅ да |
| `.claude/agents/` | Профильные агенты проекта | ✅ да |
| `.claude/rules/` | Правила и workflow | ✅ да |
| `claude/` | **Source** ClaudeKit CLI (исходник для синка в runtime) | ❌ нет |
| `CLAUDE.md` | Главные инструкции проекта (контекст, стек, команда, hook-протокол) | ✅ да |
| `docs/` | Проектная документация | по запросу |

**Статус:** полный ClaudeKit **установлен** в runtime `.claude/` (синк из локального source `claude/`, 2026-08-21). Активны: **27 агентов, 84 скилла `/ck:*`, 16 хуков, statusline, правила**. Хуки и скиллы Claude Code подхватывает **со следующей сессии** (settings.json и каталог скиллов читаются при старте сессии).

---

## 2. Что работает прямо сейчас

### 2.1. Команда агентов

Делегирование фразой **«пусть <агент> …»**. Агенты работают строго по TDD.

**Профильные (этот проект, `.claude/agents/`)** — доступны со следующей сессии:

| Агент | Стек | Зона |
|---|---|---|
| `go-dev` | Go (chi + pgx) | `apps/api` — эндпоинты, миграции, RBAC, PostgreSQL |
| `nextjs-dev` | Next.js 15 / React 19 | `apps/web` — страницы, компоненты, формы (zod), TanStack Query |
| `flutter-dev` | Flutter / Dart | `apps/mobile` — экраны, виджеты, l10n |

**Общие (глобальные `~/.claude/agents`)** — доступны всегда:

| Агент | Когда звать |
|---|---|
| `architect` | Спроектировать фичу/систему ДО кода |
| `pm` | Статус, декомпозиция плана на задачи |
| `reviewer` | Ревью изменений перед коммитом/merge |
| `qa` | Тест-кейсы, edge cases, приёмка |
| `debugger` | Корневая причина бага/регрессии |
| `docs` | README, API-доки, changelog, обновление доков |

> Глобальные `nestjs-dev` / `expo-dev` в этом проекте **не используем** — стек другой (Go / Flutter).

### 2.2. Активные скиллы (вызов `/<имя>` или по описанию задачи)

| Скилл | Для чего |
|---|---|
| `mattpocock-skills:tdd` | Red → Green → Refactor, интеграционные тесты |
| `mattpocock-skills:diagnosing-bugs` | Диагностика сложных багов до фикса |
| `mattpocock-skills:code-review` | Ревью изменений по стандартам + спеке |
| `mattpocock-skills:codebase-design` | Дизайн модулей/интерфейсов |
| `mattpocock-skills:domain-modeling` | Доменная модель, единый словарь |
| `mattpocock-skills:research` / `deep-research` | Исследование по первоисточникам |
| `mattpocock-skills:resolving-merge-conflicts` | Разрешение merge/rebase-конфликтов |
| `shadcn` | UI-компоненты для web (`apps/web`) |

---

## 3. Основной рабочий цикл

```
Спека/дизайн  →  Декомпозиция  →  Реализация (TDD)  →  Ревью  →  Приёмка  →  Верификация  →  Доки
 architect       pm                go-dev/nextjs-dev/    reviewer  qa          (см. §4)       docs
 codebase-        (задачи)          flutter-dev +
 design                            mattpocock:tdd
```

1. **Спека/дизайн** — сложную фичу сначала к `architect` (или `mattpocock-skills:codebase-design`). Идею проверить — `mattpocock-skills:prototype`.
2. **Декомпозиция** — `pm` разбивает план на задачи с зависимостями.
3. **Реализация** — берём готовую задачу → профильный агент по стеку, строго TDD (`mattpocock-skills:tdd`): RED (падающий тест) → GREEN (минимум кода) → REFACTOR.
4. **Ревью** — `reviewer` или `mattpocock-skills:code-review`.
5. **Приёмка** — `qa` (тест-кейсы, edge cases, вердикт).
6. **Верификация** — обязательно прогнать проверочные команды (§4) и прочитать вывод, только потом «готово».
7. **Документация** — `docs` обновляет `./docs`, README, changelog.

**Железные правила** (из `CLAUDE.md` и `.claude/rules/`):
- Нет production-кода без падающего теста.
- Нет заявлений «готово» без вывода проверочных команд.
- Файлы > 200 строк — модуляризовать; имена kebab-case, длинные и говорящие.
- Баг — сначала `diagnosing-bugs` (корневая причина), потом фикс через падающий тест.

---

## 4. Верификация по стекам (запускать перед «готово»)

**Монорепо (из корня):**
```bash
pnpm dev | build | lint | typecheck | test    # turbo
pnpm format                                    # prettier
```

**Backend — Go (`apps/api`):**
```bash
cd apps/api
go build ./...
go test ./...
go vet ./...
```

**Web — Next.js (`apps/web`):**
```bash
pnpm --filter @clinicos/web typecheck
pnpm --filter @clinicos/web test     # vitest
pnpm --filter @clinicos/web lint     # --max-warnings=0
```

**Mobile — Flutter (`apps/mobile`):**
```bash
cd apps/mobile
flutter analyze
flutter test
```

**Локальный запуск всей системы:**
```bash
cp .env.example .env    # заменить все change-me
docker compose up --build
# API: http://localhost/api/v1 · Swagger (dev): http://localhost/api/docs
```

---

## 5. Полный ClaudeKit: `/ck:*` скиллы, хуки, statusline

**Уже установлено** (синк `claude/` → `.claude/`, 2026-08-21). Доступные команды со следующей сессии, например:
`/ck:plan`, `/ck:cook`, `/ck:security`, `/ck:debug`, `/ck:code-review`, `/ck:brainstorm` и др. (всего 84 скилла).

### Обслуживание через CLI `ck` (v3.35.0)
```bash
ck doctor            # health-check рантайма
ck skills --list --installed   # что установлено
ck config            # дашборд конфигурации (privacyBlock, statusline, plan-naming)
```

### Пере-синк после правок исходника
Если правишь `claude/` (source), обнови runtime:
```bash
cp -R claude/skills claude/hooks claude/schemas claude/output-styles .claude/
cp claude/settings.json claude/statusline.cjs claude/.ck.json .claude/
```

### Python-скиллы (опционально — только для image/pdf-скиллов)
Тяжёлая установка (ffmpeg, imagemagick, venv), **не обязательна** для основного `/ck:*`-workflow:
```bash
cd .claude/skills && chmod +x install.sh && ./install.sh
```
Запуск Python-скриптов скиллов — через venv: `.claude/skills/.venv/bin/python3 scripts/xxx.py`.

Конфигурация ClaudeKit — `.claude/.ck.json` (privacyBlock, statusline, plan-naming и др.).

> ⚠️ Активные хуки меняют поведение сессии: `privacy-block` (спросит при доступе к секретам), `scout-block` (ограничивает слишком широкий поиск), `dev-rules-reminder`, `simplify-gate`, statusline. Отключить — отредактировать `.claude/settings.json`.
> ⚠️ Удалённый репо `claudekit/claudekit-engineer` недоступен (404), поэтому `ck init`/`ck update` из сети не работают — используем локальный source `claude/`.

---

## 6. Куда смотреть дальше

- `CLAUDE.md` — главные инструкции и hook-протокол (privacy-block).
- `.claude/rules/` — `primary-workflow`, `development-rules`, `orchestration-protocol`, `documentation-management`, `skill-domain-routing`, `team-coordination-rules`.
- `docs/project-overview-pdr.md`, `docs/system-architecture.md`, `docs/code-standards.md`, `docs/codebase-summary.md` — контекст продукта (агенты читают их перед работой).
- `docs/deployment/` — эксплуатация (электронная очередь, backup/restore).
- `README.md` — запуск и обзор проекта.

---

## 7. Быстрая шпаргалка

| Хочу… | Делаю |
|---|---|
| Новая фича на бэкенде | «пусть go-dev …» (TDD) |
| Экран/страница на web | «пусть nextjs-dev …» |
| Экран мобилки | «пусть flutter-dev …» |
| Спроектировать до кода | «пусть architect …» |
| Разбить план на задачи | «пусть pm …» |
| Проверить перед merge | «пусть reviewer …» |
| Найти причину бага | `mattpocock-skills:diagnosing-bugs` |
| Обновить доки | «пусть docs …» |
| Спланировать задачу | `/ck:plan` (со след. сессии) |
| Проверить безопасность | `/ck:security` |
| Health-check тулинга | `ck doctor` |
