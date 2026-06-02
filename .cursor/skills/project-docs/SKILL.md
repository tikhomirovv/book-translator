---
name: project-docs
description: Maintains project documentation in .docs/ — reads, creates, updates, and validates project-overview, prd, and technical-design. Use when brainstorming, planning, discussing product or technical decisions, onboarding to a project, or when docs may be missing, outdated, or incomplete.
---

# Project Docs

Portable workflow for keeping project documentation accurate. Works at any project stage — not MVP-specific. No Git hosting assumptions.

## Language

All project documentation in `.docs/` must be written in **Russian** — including new sections, updates, and scaffolding. The skill itself stays in English; only the docs content is Russian.

## Document Set

All project docs live in `.docs/` at the repository root.

| File | Purpose |
|------|---------|
| `project-overview.md` | What the project is, who it is for, the problem it solves, current status. Short — readable in 1–2 minutes. |
| `prd.md` | Living product spec: scope, user flows, requirements, constraints, non-goals, open questions, planned expansions. |
| `technical-design.md` | Stack, technical decisions, core entities, project structure, engineering rules. |

Do not create extra markdown files unless the user explicitly asks. Prefer updating these three.

## Before Any Project Work

1. Check whether `.docs/` exists. If not, create it and scaffold the three files with minimal section headers.
2. Read existing docs in order: `project-overview.md` → `prd.md` → `technical-design.md`.
3. Treat these files as source of truth for product and technical context.

## When to Update Docs

Update docs during discussions — not only after implementation.

| Change type | Update |
|-------------|--------|
| Product vision, audience, value | `project-overview.md` |
| Scope, requirements, flows, non-goals | `prd.md` |
| Stack, architecture, entities, tech rules | `technical-design.md` |

After updating, keep sections consistent across files. Remove contradictions inline.

## Gap Detection

After reading docs, check for:

- empty or placeholder sections (`TBD`, `TODO`, `???`)
- missing scope boundaries or non-goals
- undocumented open questions
- conflicts between overview, PRD, and technical design

If gaps exist, briefly tell the user what is missing. Offer to fill gaps in **wizard mode**: one focused question at a time, updating the relevant doc after each answer.

Do not block unrelated work — mention gaps once, then proceed if the user prefers.

## Scaffolding Templates

When creating missing files, use these minimal structures.

### project-overview.md

```markdown
# Обзор проекта

## Что это
## Для кого
## Проблема
## Ценность
## Текущий статус
```

### prd.md

```markdown
# PRD

## Текущий scope
## Пользовательские сценарии
## Требования
## Ограничения
## Non-goals
## Открытые вопросы
## Запланированные расширения
```

### technical-design.md

```markdown
# Technical Design

## Стек
## Ключевые решения
## Основные сущности
## Структура проекта
## Инженерные правила
```

Expand sections only when there is content to add.

## Update Rules

- Edit existing docs in place. Do not create parallel "draft" or "temp" files for the same purpose.
- Keep `project-overview.md` short. Put detailed requirements in `prd.md`.
- Record decisions with brief rationale in `technical-design.md`.
- When a decision replaces an old one, update or remove the old text — do not leave conflicting versions.

## What This Skill Does Not Cover

- Task tracking, roadmaps, issues, milestones, or pull requests — use a separate Git hosting skill if the project uses one.
- Implementation plans or code changes — this skill only maintains documentation.
