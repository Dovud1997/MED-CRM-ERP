# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Context

**ONA VA BOLA KLINIKASI** (`clinicos`) — API-first система управления частной клиникой (MED CRM/ERP). Монорепо на pnpm + Turborepo.

Apps:
- `apps/api` — **Backend на Go** (`cmd/`, `internal/`, SQL `migrations/`). REST API, permission-based роли, PostgreSQL.
- `apps/web` — **Web на Next.js 15 / React 19** (`@clinicos/web`): TanStack Query, react-hook-form, zod, Vitest.
- `apps/mobile` — **Mobile на Flutter/Dart** (`pubspec.yaml`, `lib/`, l10n).

Инфраструктура: PostgreSQL, Redis, MinIO (S3), Docker Compose (`compose.yaml`). Prod-подобный запуск — `docker compose up --build` (сервис `migrate` применяет SQL-миграции до старта API). Swagger (dev): `http://localhost/api/docs`.

## Tech Stack & Commands

Корень монорепо — pnpm@10 + Turbo. Общие команды из корня:
- `pnpm dev` / `pnpm build` / `pnpm lint` / `pnpm typecheck` / `pnpm test` (turbo)
- `pnpm format` (prettier)

Проверки по стекам (запускать перед заявлением «готово»):
- **Go** (`apps/api`): `go build ./...`, `go test ./...`, `go vet ./...`
- **Web** (`apps/web`): `pnpm --filter @clinicos/web typecheck`, `pnpm --filter @clinicos/web test` (vitest), `pnpm --filter @clinicos/web lint`
- **Flutter** (`apps/mobile`): `flutter analyze`, `flutter test`

## Agent Team (по реальному стеку проекта)

Делегируй стек-задачи профильным агентам (строго TDD: RED → GREEN → REFACTOR):
- Backend (Go, `apps/api`) → агент **`go-dev`**
- Web (Next.js/React, `apps/web`) → агент **`nextjs-dev`**
- Mobile (Flutter, `apps/mobile`) → агент **`flutter-dev`**
- Кросс-стек / оркестрация фазами → `fullstack-developer`; ревью → `code-reviewer`; тесты/приёмка → `tester`; отладка → `debugger`.

> Примечание: глобальный `~/.claude` содержит команду под другой стек (`nestjs-dev`, `expo-dev`). В этом проекте они НЕ применяются — используем `go-dev` и `flutter-dev`. `~/.claude` не редактировать (правило проекта).

## Role & Responsibilities

Your role is to analyze user requirements, delegate tasks to appropriate sub-agents, and ensure cohesive delivery of features that meet specifications and architectural standards.

## Workflows

- Primary workflow: `./.claude/rules/primary-workflow.md`
- Development rules: `./.claude/rules/development-rules.md`
- Orchestration protocols: `./.claude/rules/orchestration-protocol.md`
- Documentation management: `./.claude/rules/documentation-management.md`
- And other workflows: `./.claude/rules/*`

**IMPORTANT:** Analyze the skills catalog and activate the skills that are needed for the task during the process.
**IMPORTANT:** DO NOT modify skills in `~/.claude/skills` directory directly. **MUST** modify skills in this current working directory. Unless you are asked to do so.
**IMPORTANT:** You must follow strictly the development rules in `./.claude/rules/development-rules.md` file.
**IMPORTANT:** Before you plan or proceed any implementation, always read the `./README.md` file first to get context.
**IMPORTANT:** Sacrifice grammar for the sake of concision when writing reports.
**IMPORTANT:** In reports, list any unresolved questions at the end, if any.

## Git

**DO NOT** use `chore` and `docs` in commit messages of file changes in `.claude` directory.

## Hook Response Protocol

### Privacy Block Hook (`@@PRIVACY_PROMPT@@`)

When a tool call is blocked by the privacy-block hook, the output contains a JSON marker between `@@PRIVACY_PROMPT_START@@` and `@@PRIVACY_PROMPT_END@@`. **You MUST use the `AskUserQuestion` tool** to get proper user approval.

**Required Flow:**

1. Parse the JSON from the hook output
2. Use `AskUserQuestion` with the question data from the JSON
3. Based on user's selection:
   - **"Yes, approve access"** → Use `bash cat "filepath"` to read the file (bash is auto-approved)
   - **"No, skip this file"** → Continue without accessing the file

**Example AskUserQuestion call:**
```json
{
  "questions": [{
    "question": "I need to read \".env\" which may contain sensitive data. Do you approve?",
    "header": "File Access",
    "options": [
      { "label": "Yes, approve access", "description": "Allow reading .env this time" },
      { "label": "No, skip this file", "description": "Continue without accessing this file" }
    ],
    "multiSelect": false
  }]
}
```

**IMPORTANT:** Always ask the user via `AskUserQuestion` first. Never try to work around the privacy block without explicit user approval.

## Python Scripts (Skills)

When running Python scripts from `.claude/skills/`, use the venv Python interpreter:
- **Linux/macOS:** `.claude/skills/.venv/bin/python3 scripts/xxx.py`
- **Windows:** `.claude\skills\.venv\Scripts\python.exe scripts\xxx.py`

This ensures packages installed by `install.sh` (google-genai, pypdf, etc.) are available.

**IMPORTANT:** When scripts of skills failed, don't stop, try to fix them directly.

## [IMPORTANT] Consider Modularization
- If a code file exceeds 200 lines of code, consider modularizing it
- Check existing modules before creating new
- Analyze logical separation boundaries (functions, classes, concerns)
- Use kebab-case naming with long descriptive names, it's fine if the file name is long because this ensures file names are self-documenting for LLM tools (Grep, Glob, Search)
- Write descriptive code comments
- After modularization, continue with main task
- When not to modularize: Markdown files, plain text files, bash scripts, configuration files, environment variables files, etc.

## Documentation Management

We keep all important docs in `./docs` folder and keep updating them, structure like below:

```
./docs
├── project-overview-pdr.md
├── code-standards.md
├── codebase-summary.md
├── design-guidelines.md
├── deployment-guide.md
├── system-architecture.md
└── project-roadmap.md
```

**IMPORTANT:** *MUST READ* and *MUST COMPLY* all *INSTRUCTIONS* in project `./CLAUDE.md`, especially *WORKFLOWS* section is *CRITICALLY IMPORTANT*, this rule is *MANDATORY. NON-NEGOTIABLE. NO EXCEPTIONS. MUST REMEMBER AT ALL TIMES!!!*
