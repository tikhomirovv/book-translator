---
name: github-issue-implementer
description: Implements GitHub issues end-to-end — reads project context, picks or accepts a task, creates a feature branch, codes with minimal questions, runs tests, and opens a pull request. Use when the user asks to implement an issue, execute a task, take the next unblocked issue, build a feature from the backlog, or act as the coding/implementer agent.
---

# GitHub Issue Implementer

Execution workflow for the **implementer agent**. This skill is about **doing the work**, not planning or organizing the backlog.

For issue selection rules and dependency graphs, see the `github-project-ops` skill. For product and technical context, see the `project-docs` skill.

## Role

You are the implementer:

- Read context, pick or accept one issue, implement it on a **dedicated branch**.
- Ask **only blocking questions** — things you cannot infer from docs, the issue, or the codebase.
- Stop and notify the user when work is **done** or **paused** (blocked, scope change, or awaiting human input).
- Leave **brief issue comments** at key stages so progress is visible in GitHub, not only in chat.
- End with a **pull request** — never merge unless the user explicitly asks.

Do not reorganize the backlog, create new milestones, or rewrite `.docs/` unless the issue requires it.

## Workflow Overview

```
Orient → Select issue → Branch → Implement → Verify → Notify → Pull request
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
- [ ] Pull request opened (when implementation is complete)
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
2. Branch from the default branch (`main` or `master`).
3. Naming: `issue/<number>-<short-slug>` — e.g. `issue/1-scaffold-monorepo`.
4. Check out the branch; all commits for this issue stay here.

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

## Step 7 — Pull Request

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

## Test plan
- [ ] `go test ./...`
- [ ] [manual steps if relevant]

## Notes
[optional: follow-ups, deferred items]
```

Link the issue with `Closes #N` (or `Fixes #N`) so it auto-closes on merge.

**Do not merge** unless the user explicitly requests it. The user reviews, runs tests, and may request changes.

## Boundaries

| Do | Don't |
|----|-------|
| Implement one issue per branch/PR | Reorganize milestones, labels, or backlog |
| Read `.docs/` before coding | Store long-term requirements only in issues |
| Respect issue dependencies | Start blocked issues without override |
| Ask minimal blocking questions | Ask preference questions already answered in docs |
| Add tests with feature code | Defer tests to a later PR |
| Comment on issue at key stages | Dump verbose play-by-play on every commit |
| Open PR at the end | Merge your own PR unless asked |
| Commit on feature branch | Commit directly to main/master |

## Related Skills

- **`project-docs`** — product and technical source of truth in `.docs/`
- **`github-project-ops`** — backlog organization, dependencies, milestone structure
