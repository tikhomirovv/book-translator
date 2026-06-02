---
name: github-project-ops
description: Uses GitHub for project management in the current repository via gh CLI — issues, labels, milestones, projects, and pull requests. Use when planning work, decomposing tasks, creating or updating issues, managing milestones, or operating GitHub Projects.
---

# GitHub Project Operations

Use GitHub as the operational layer for planning and execution. Product and technical context live in `.docs/` (see the `project-docs` skill). This skill covers how work is tracked in GitHub.

## Principle

All project-management actions go through GitHub in the **current repository**, using the **`gh` CLI**. Do not duplicate roadmaps or backlogs in markdown when GitHub can hold them.

For command syntax, use `gh <command> --help` or `gh help <command>`. Do not rely on memorized flags — discover them at runtime.

## Entity Roles

| Entity | Role |
|--------|------|
| **Issues** | Atomic work items: features, bugs, tasks, spikes. Include clear title, description, and acceptance criteria. |
| **Labels** | Categories and filters: type (`feature`, `bug`, `chore`), area, priority. |
| **Milestones** | Phase or release groupings. Map to project stages (e.g. MVP, v1.1). Track progress toward a goal. |
| **Projects** | Flexible views over issues and PRs — table, board, or roadmap. Use custom fields (status, priority, effort) when labels alone are not enough. |
| **Pull requests** | Code changes linked to issues. Use for review, checks, and merge. |

## Typical Workflow

1. **Plan** — break work from PRD or discussion into issues. Group related issues under a milestone.
2. **Track** — add issues to a Project for board/roadmap views. Update status as work progresses.
3. **Implement** — branch, commit, open a PR referencing the issue (`Fixes #N` or `Closes #N`).
4. **Close** — merge PR, verify linked issues close, update Project status.

## When to Use What

- **Issues only** — sufficient for small projects or early stages.
- **Issues + milestones** — when work is grouped into phases or releases.
- **Issues + milestones + Projects** — when you need visual boards, roadmap views, or custom fields.

Start simple. Add Projects when milestones and labels are no longer enough.

## Decomposition Guidelines

When breaking down work:

- One issue = one deliverable or one clearly bounded task.
- Write acceptance criteria in the issue body.
- **Set native issue dependencies** for every issue that cannot start until another is done (see below).
- Use milestones for phase boundaries, not for every small task.
- Prefer fewer, well-scoped issues over many vague ones.

## Issue Dependencies (required)

Every new issue must declare what it **blocks on** so agents and humans can see execution order without guessing.

### Why

- Agents should pick only issues whose blockers are **closed**.
- Milestones alone do not enforce order; native `blocked by` / `blocking` links do.
- Parallel work stays possible when dependencies are explicit (e.g. two issues both blocked only by #1).

### When creating or splitting issues

1. Identify prerequisites (schema before repos, core before bot UI, MVP before v1 payments, etc.).
2. Add **blocked by** links to every dependent issue immediately — do not leave this for later.
3. Optionally append a short `## Depends on` section in the issue body listing `#N` titles for human readability. Native links remain the source of truth.

### How agents pick work

1. Scope to the **current milestone** (or the phase the user asked for).
2. List open issues in that milestone.
3. **Skip any issue that has an open blocker** (check `blocked_by`; all listed blockers must be closed).
4. Among remaining issues, prefer `priority:p0`, then lowest issue number unless the user specified otherwise.
5. If everything is blocked, report which blockers must close first — do not start out-of-order work unless the user explicitly overrides.

### Setting dependencies via CLI

**Preferred (when available):** discover current syntax with `gh issue edit --help`. Newer `gh` versions support flags like `--add-blocked-by` / `--remove-blocked-by`.

**Fallback — GitHub REST API** (works when `gh issue edit` lacks dependency flags):

```bash
# Get numeric issue id (not the issue number shown in UI)
BLOCKER_ID=$(gh api repos/OWNER/REPO/issues/BLOCKER_NUMBER --jq .id)

# Mark ISSUE_NUMBER as blocked by BLOCKER_NUMBER
gh api repos/OWNER/REPO/issues/ISSUE_NUMBER/dependencies/blocked_by \
  --method POST --input - <<< "{\"issue_id\":${BLOCKER_ID}}"
```

Important:

- `issue_id` in the JSON body must be an **integer**, not a string. Use `--input` with raw JSON or `-F issue_id:=ID` — plain `-f issue_id=ID` sends a string and returns `422`.
- Multiple blockers require **one POST per blocker**.
- Verify: `gh api repos/OWNER/REPO/issues/N/dependencies/blocked_by --jq '.[].number'`
- List what an issue blocks: `gh api repos/OWNER/REPO/issues/N/dependencies/blocking --jq '.[].number'`

### Dependency design rules

- **Root tasks** (e.g. repo scaffold) have no blockers.
- **Infrastructure** (CI, lint) may depend only on scaffold — can run in parallel with domain work once scaffold exists.
- **Domain layers** follow technical design order: schema → persistence → services → adapters → integration/polish.
- **Cross-cutting features** (billing, Telegram UI) depend on the core pieces they integrate with, not the reverse.
- **Next milestone / phase** issues should be blocked by the **last deliverable of the previous phase** (or by all critical path items if parallel entry is unsafe).

### EasyTerms execution order (reference)

MVP (`MVP — Working Bot`):

| Issue | Blocked by | Notes |
|-------|------------|-------|
| #1 Scaffold | — | Start here; includes Dockerfile |
| #2 CI | #1 | GitHub Actions: `go test` + `docker build`; merge gate |
| #3 DB schema | #1 | |
| #4 Repositories | #3 | |
| #5 LLM port | #1 | Parallel with #3→#4 |
| #6 Document service | #4, #5 | |
| #7 Analysis modes | #6 | |
| #9 Billing (stub) | #4 | Parallel with #6→#7 after repos |
| #11 Telegram bot | #6, #7, #9 | Needs core + billing |
| #13 URL ingest | #6 | |
| #15 MVP polish | #7, #11, #13 | Last MVP task |

v1 (`v1 — Payments`):

| Issue | Blocked by |
|-------|------------|
| #8 Cost research | #15 |
| #10 YooKassa provider | #8 |
| #12 Purchase UX | #10 |
| #14 Real billing E2E | #12 |

Update this table when the backlog changes.

## Authentication and Permissions

If `gh` fails with auth, scope, or permission errors:

1. Run `gh auth status` to check the current token.
2. Verify the token has access to **this repository**.
3. For fine-grained PATs, ensure at minimum:
   - **Issues**: Read and write
   - **Pull requests**: Read and write
   - **Contents**: Read and write
   - **Metadata**: Read-only (usually automatic)
4. For GitHub Projects commands, the token may also need the **`project`** scope. Run `gh auth refresh -s project` if suggested.
5. Treat unexpected `404` or GraphQL permission errors as possible auth misconfiguration before assuming a resource is missing.

Tell the user which permission or scope is likely missing and how to fix it. Do not guess silently.

## Boundaries

- Do not store product requirements or technical decisions in GitHub issues long-term — keep `.docs/` as source of truth; issues reference and implement that context.
- Do not embed a full `gh` command reference in responses — discover commands via CLI help.
- Do not create markdown roadmaps that duplicate GitHub milestones or Projects unless the user explicitly asks.
