---
name: github-issue-implementer
description: Implements GitHub issues end-to-end — reads project context, picks or accepts a task, creates a feature branch, codes with minimal questions, runs tests, and opens a pull request. Use when the user asks to implement an issue, execute a task, take the next unblocked issue, build a feature from the backlog, or act as the coding/implementer agent.
---

# GitHub Issue Implementer

Execution workflow for the **implementer agent**. This skill is about **doing the work**, not planning or organizing the backlog.

For issue selection rules and dependency graphs, see the `github-project-ops` skill. For product and technical context, see the `project-docs` skill.

## Role

You are the implementer:

- Read context, pick or accept **one issue at a time**, implement it on a **dedicated branch** (`issue/<N>-<slug>`).
- Ask **only blocking questions** — things you cannot infer from docs, the issue, or the codebase.
- Stop and notify the user when work is **done** or **paused** (blocked, scope change, or awaiting human input).
- Leave **brief issue comments** at key stages so progress is visible in GitHub, not only in chat.
- End with a **pull request** (`Closes #N`). Then follow the **merge policy** (below) before starting the next issue.

Do not reorganize the backlog, create new milestones, or rewrite `.docs/` unless the issue requires it.

## Sequential execution (default)

When the user asks to implement tasks **in order**, **one after another**, **sequentially**, or does not specify batching — use this loop for **each** issue:

```
issue #N → branch issue/N-slug → implement → test → PR → merge & close #N → update main → issue #N+1
```

**Rules:**

| Rule | Requirement |
|------|-------------|
| One issue | Exactly **one** issue per branch and per PR (`Closes #N` for a single N). |
| Branch base | Always branch from **updated `main`** (or default branch) after the previous issue is merged. |
| Next issue | Start issue #N+1 only when #N is **merged** and closed (or user explicitly overrides). |
| No batching | Do **not** combine multiple issues in one branch/PR unless the user **explicitly** asks (e.g. «в одном PR», «батчем»). |
| Dependencies | Never start an issue whose `blocked_by` issues are still open (see `github-project-ops`). |

**Anti-pattern:** `issue/1-4-foundation` with `Closes #1, #2, #3, #4` — only when the user explicitly requested a batch.

## Merge policy (ask if unclear)

At the **start** of an implementation session (before the first issue), if the user has **not** said whether to merge automatically, **ask once**:

> После каждого PR: **(A)** мержить автоматически и сразу брать следующую задачу, или **(B)** остановиться и ждать вашего ревью?

| Mode | Behavior after PR is opened and CI is green |
|------|---------------------------------------------|
| **A — auto-merge** | `gh pr merge` (use repo default: merge/squash as appropriate), `git checkout main && git pull`, then continue to the next issue in the same session. |
| **B — wait for review** (default if user does not choose) | Stop. Notify user with PR link. Do **not** merge. Do **not** start the next issue until the user merges or asks to continue. |

If the user already stated their preference in the current conversation, do not ask again.

**Explicit overrides:**

- User says «мержи сам», «auto-merge», «не жди ревью» → mode A for this session.
- User says «жди ревью», «не мержи», «stop after PR» → mode B.

## Workflow Overview

```
Orient → Merge policy? → Select ONE issue → Branch from main → Implement → Verify → PR → Merge gate → (next issue)
```

Track progress with this checklist:

```
- [ ] Context read (.docs/ + issue + repo state)
- [ ] Issue selected (specified or auto-picked)
- [ ] Feature branch created and checked out
- [ ] Acceptance criteria implemented
- [ ] Tests added/updated (same change set)
- [ ] `go test ./...` passes locally **or** CI checks green after push
- [ ] `docker build` passes locally **or** CI docker job green after push
- [ ] Key stages commented on the issue (see below)
- [ ] User notified (done or paused)
- [ ] Pull request opened (when implementation is complete; `Closes #N` — single issue)
- [ ] Merge policy applied (merged + main pulled, or stopped for user review)
- [ ] Only then: next issue (if sequential run continues)
```

## Step 1 — Orient

Before writing code:

1. Read `.docs/` in order: `project-overview.md` → `prd.md` → `technical-design.md`.
2. Inspect the repository — layout, existing packages, conventions, test patterns.
3. Use `gh` to understand backlog state:
   - Open issues for the active milestone (default: earliest incomplete milestone, usually MVP first).
   - Read the target issue body and acceptance criteria.
   - Check blockers: `gh api repos/OWNER/REPO/issues/N/dependencies/blocked_by --jq '.[].number'`

If the user gave no issue number, auto-pick (Step 2). If they named `#N`, use that issue after verifying it is not blocked unless they explicitly override.

## Step 2 — Select Issue

**User specified `#N`:** use it. If it has open blockers, warn once and stop unless the user overrides.

**User did not specify:** pick the next executable issue:

1. Scope to the current milestone (or the phase the user named).
2. List open issues in that milestone.
3. Exclude any issue with **open** blockers (all `blocked_by` issues must be closed).
4. Prefer `priority:p0`, then lowest issue number.
5. If everything is blocked, stop and report which blockers must close first — do not pick out-of-order work.

Discover listing/filter syntax via `gh issue list --help` at runtime.

## Step 3 — Branch

**Always** create a new branch before implementation. Never commit implementation work directly on `main` / `master`.

1. Ensure a clean working tree (or stash only with user awareness).
2. Branch from the default branch (`main` or `master`) — **must be up to date** (previous issue merged and `git pull` done when working sequentially).
3. Naming: `issue/<number>-<short-slug>` — e.g. `issue/1-scaffold-monorepo` (**one issue number per branch**).
4. Check out the branch; **all** commits for **this issue only** stay here.

```bash
git fetch origin
git checkout main   # or master
git pull --ff-only
git checkout -b issue/1-scaffold-monorepo
```

If a branch for this issue already exists and has WIP the user wants continued, check it out instead of creating a duplicate — confirm with the user only if ambiguous.

## Step 4 — Implement

Follow the issue acceptance criteria and `.docs/technical-design.md`.

### Coding rules

- Match existing project conventions (structure, naming, error handling).
- Keep changes scoped to the issue — no drive-by refactors.
- Core business logic stays in `internal/core`; clients stay thin.
- Use ports/interfaces for external dependencies (LLM, payments, storage) so core stays testable.
- Tests are **mandatory** for changed business logic — include them in the same change set, not a follow-up PR.

### Questions policy

- **Do not ask** for decisions already documented in `.docs/` or the issue.
- **Do not ask** for permission to proceed with the obvious implementation path.
- **Do ask** only when missing information **blocks** progress — e.g. missing API keys with no stub path, contradictory acceptance criteria, destructive choice with no default.
- Ask **one focused question** at a time. While waiting, stop work and report paused state.
- **Also post the question on the issue** — chat alone is not enough when work is blocked (see Issue comments).

### Issue comments (key stages)

Keep a lightweight paper trail on the issue via `gh issue comment N --body "..."`. Comments should be **short** — a few lines, not a full log. Prefer bullets over prose.

**When to comment:**

| Stage | Comment? | Example |
|-------|----------|---------|
| Started work / branch created | Yes | Branch name, brief plan |
| Major milestone reached | Yes, if non-obvious | «Schema migration added», «LLM port wired» |
| Blocked — need human input | **Required** | Question + what is already done + branch |
| Done — PR opened | **Required** | Summary, PR link, test status |

**When one comment is enough:** small, linear tasks — a single **final comment** with branch, PR link, and 2–4 bullets is fine.

**When to add mid-task comments:** long or multi-step issues, blocked work, or after a milestone that would be hard to infer from the PR alone.

**Blocked comment template:**

```markdown
⏸ **Paused** — need input

**Branch:** `issue/N-slug`
**Done so far:** [1–2 bullets]
**Blocker:** [one focused question]
```

**Final comment template:**

```markdown
✅ **Ready for review**

**Branch:** `issue/N-slug`
**PR:** #M (or full URL)

- [acceptance criterion → what was done]
- Tests: `go test ./...` — pass
```

Do not close the issue manually — let the PR (`Closes #N`) close it on merge.

## Step 5 — Verify

Before notifying the user or opening a PR:

1. Run tests locally **if Go is available**: `go test ./...` (or the project's documented test command).
2. If local Go/Docker are **not** available, rely on **GitHub Actions** after push — CI runs `go test ./...` and `docker build`; wait for checks and report status.
3. Fix failures before proceeding (locally or via follow-up commits until CI is green).
4. Review the diff against acceptance criteria — every criterion met or explicitly deferred with user approval.

After the PR is opened, **always mention CI status** — green checks are the merge gate when local tools are missing.

## Step 6 — Notify User

Always stop and report when implementation is **complete** or **paused**. Mirror the same message on the issue (see Issue comments) — user chat and issue thread should stay in sync for blockers and completion.

### Done template

```markdown
## Issue #N — ready for review

**Branch:** `issue/N-slug`
**PR:** [link]
**Issue:** [title](link)

### Done
- [bullets mapped to acceptance criteria]

### Tests
- `go test ./...` — pass

### Next
- Review the PR and diff
- Run tests locally if you want
- Request changes or merge when satisfied
```

### Paused template

```markdown
## Issue #N — paused

**Branch:** `issue/N-slug` (WIP committed or uncommitted: state which)

### Progress
- [what is done]

### Blocker
- [single blocking question or external dependency]

### Needed from you
- [specific answer or action]
```

When implementation is complete and tests pass, **notify the user and open the PR in the same session** — the PR is the handoff artifact for review.

## Step 7 — Pull Request & merge gate

Open a PR as soon as implementation is complete and tests pass. Do not leave work only on a branch without a PR unless paused or blocked.

1. Commit on the feature branch with clear messages (user may ask for specific commit style).
2. Push the branch: `git push -u origin issue/N-slug`
3. Create the PR via `gh pr create` — discover flags via `--help`.

PR body should include:

```markdown
## Summary
[1–3 bullets: what changed and why]

## Issue
Closes #N

<!-- Exactly one issue per PR unless user explicitly requested a batch. -->

## Test plan
- [ ] `go test ./...`
- [ ] [manual steps if relevant]

## Notes
[optional: follow-ups, deferred items]
```

Link the issue with `Closes #N` (or `Fixes #N`) so it auto-closes on merge.

### After PR

1. Report CI status.
2. Apply **merge policy** (see above):
   - **Mode A:** merge PR, `git checkout main && git pull --ff-only`, comment on issue if helpful, proceed to next issue.
   - **Mode B:** stop; tell user the PR is ready for review; do not start the next issue.

## Boundaries

| Do | Don't |
|----|-------|
| **One issue** per branch and per PR | Pack multiple issues into one PR without explicit user request |
| Work **sequentially** from updated `main` when user asks for ordered execution | Start #N+1 before #N is merged (unless user overrides) |
| Ask **merge policy** once per session if not specified | Assume auto-merge or assume wait-for-review without asking |
| Read `.docs/` before coding | Store long-term requirements only in issues |
| Respect issue dependencies | Start blocked issues without override |
| Ask minimal blocking questions | Ask preference questions already answered in docs |
| Add tests with feature code | Defer tests to a follow-up PR |
| Comment on issue at key stages | Dump verbose play-by-play on every commit |
| Merge only in **mode A** or when user explicitly asks | Merge silently in mode B |
| Commit on feature branch | Commit directly on main/master |

## Related Skills

- **`project-docs`** — product and technical source of truth in `.docs/`
- **`github-project-ops`** — backlog organization, dependencies, milestone structure
