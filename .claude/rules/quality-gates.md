# Quality Gate Rules

Rules for contributors and AI agents working on the claudekit-engineer repo. These are NOT shipped to end users.

## Metadata Deletions (MANDATORY)

When renaming or deleting ANY file under `claude/` directory (skills, hooks, agents, scripts), you MUST add the old relative path to `claude/metadata.json` `deletions[]` array. This tells the CLI installer to remove stale files from user machines during upgrade. Forgetting this leaves orphaned files that cause conflicts.

**Affected files:** `claude/metadata.json`

## Skill Registry Contract

Canonical skill names live in each `claude/skills/*/SKILL.md` frontmatter `name:` field. All cross-references in markdown files MUST use exact registered names with `/ck:` prefix. Before adding a new skill name, check for collisions with Claude Code built-in commands (`/help`, `/clear`, `/debug`, `/plan`, `/compact`, `/review`, `/search`). When renaming a skill, update ALL cross-references in the same PR.

## Skill Cross-Reference Integrity (CI-enforced)

CI runs `node scripts/check-skill-cross-refs.js` on every push. It builds a registry from all `claude/skills/*/SKILL.md` frontmatter `name:` fields and checks every `/ck:` reference in `claude/**/*.md` resolves to a registered name.

**Critical rule: use REGISTERED names, not directory names.**

The registry normalizes names by stripping the `ck:` prefix from frontmatter. So `name: ck:debug` registers as `debug`. The reference `/ck:debug` captures `debug` → match. But `/ck:ck-debug` captures `ck-debug` → **no match, CI fails**.

| Directory Name | Frontmatter `name:` | Registered As | Correct Reference |
|----------------|---------------------|---------------|-------------------|
| `ck-debug` | `ck:debug` | `debug` | `/ck:debug` |
| `ck-plan` | `ck:plan` | `plan` | `/ck:plan` |
| `ck-security` | `ck:security` | `security` | `/ck:security` |
| `cook` | `ck:cook` | `cook` | `/ck:cook` |
| `brainstorm` | `ck:brainstorm` | `brainstorm` | `/ck:brainstorm` |

**Before committing any `/ck:` reference**, verify the name exists in the registry:
```bash
node scripts/check-skill-cross-refs.js
```

**When modifying workflow routing rules** (`claude/rules/skill-workflow-routing.md`, `claude/rules/skill-domain-routing.md`):
- These rules are shipped to end users — they guide Claude's skill suggestions
- Every `/ck:` reference in these files is CI-checked
- Run `~/.claude/skills/.venv/bin/python3 claude/scripts/validate-skill-crossrefs.py claude/skills/` to verify workflow chain integrity
- When adding a new skill to a workflow chain, update BOTH the routing rule AND the skill's `## Workflow Position` section

**Affected files:** `claude/rules/skill-*.md`, `claude/skills/*/SKILL.md`, `scripts/check-skill-cross-refs.js`

## Statusline Changes

Changes to `statusline*.cjs` or `statusline-*.cjs` MUST update snapshot tests. Run test suite with all config variants: minimal config, full config, custom lines, no quota, 1M context window. ANSI escape sequences and special characters (NBSP) must be explicitly tested.
