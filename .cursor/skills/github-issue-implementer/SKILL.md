---
name: github-issue-implementer
description: Implements GitHub issues end-to-end — reads project context, picks or accepts a task, creates a feature branch, codes with minimal questions, runs tests, and opens a pull request. Use when the user asks to implement an issue, execute a task, take the next unblocked issue, build a feature from the backlog, or act as the coding/implementer agent.
---

# GitHub Issue Implementer

Execution workflow for the **implementer agent**. This skill is about **doing the work**, not planning or organizing the backlog.

## Standalone use

This skill is **self-contained** and must work in **any repository**, with or without other skills installed.

## Optional related skills

If another skill is available in the workspace, **use it** — but **never stop or fail** because it is missing:

| Skill (if present) | Purpose |
|--------------------|---------|
| `github-project-ops` | Backlog decomposition, `blocked by` dependencies, milestones |
| `project-docs` | Product/tech context in `.docs/` |

When a related skill is absent, follow **this document** and discover context from the repo (README, issues, CI config, code layout) directly.

## Role

You are the implementer:

- Read context, pick or accept **one issue at a time**, implement it on a **dedicated branch** (`issue/<N>-<slug>`).
- Ask **only blocking questions** — things you cannot infer from docs, the issue, or the codebase.
- Stop and notify the user when work is **done** or **paused** (blocked, scope change, or awaiting human input).
- Leave **brief issue comments** at key stages so progress is visible in GitHub, not only in chat.
- End with a **pull request** (`Closes #N`). Then follow the **execution mode** (below) before starting the next issue.

Do not reorganize the backlog, create new milestones, or rewrite product docs unless the issue requires it.

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
| Dependencies | Never start an issue whose `blocked_by` issues are still open (use `github-project-ops` if available; otherwise read issue links / milestone order). |

**Anti-pattern:** one branch/PR with `Closes #1, #2, #3` — only when the user **explicitly** requested a batch.

## Execution modes (ask if unclear)

Two workflows. At the **start** of a session (before the first issue), if the user has **not** chosen a mode, **ask once**:

> How should I handle each PR after CI passes?  
> **(A) Autonomous** — merge, close the issue, pull `main`, continue to the next task without waiting.  
> **(B) Review-driven** — stop after opening the PR; you review and merge when ready.

### Mode A — Autonomous (full pipeline, no user gate per issue)

Use when the user wants the agent to run through the backlog with minimal interruption.

After PR is opened and **CI is green**:

1. **Merge** the PR (`gh pr merge` — match repo settings: merge commit / squash / rebase).
2. **Close the issue** — normally automatic via `Closes #N` on merge. **Verify** the issue is closed (`gh issue view N --json state`). If still open, close it (`gh issue close N`) and note why linking failed.
3. **Update local default branch:** `git checkout main && git pull --ff-only` (or the repo’s default branch).
4. **Comment** on the issue briefly (merged, tests passed) if useful.
5. **Continue** to the next issue in the same session (sequential rules above).

Do **not** stop after each PR waiting for the user unless blocked or tests fail.

### Mode B — Review-driven (user reviews code)

Use when the user wants to **review diffs** and stay in the loop technically.

After PR is opened and CI is green:

1. **Notify** the user with PR link and short summary.
2. **Do not merge** the PR.
3. **Do not close** the issue manually — it should close when the **user** merges the PR (`Closes #N`).
4. **Do not start** the next issue until the user merges (or explicitly says to continue).

The user merges when satisfied; issue closure is a consequence of their merge, not agent action.

### Choosing and remembering the mode

| Signal | Mode |
|--------|------|
| User did not specify | Ask once (A or B) |
| «auto-merge», «мержи сам», «без ревью», «продолжай сам» | A |
| «жди ревью», «не мержи», «stop after PR», «хочу ревьюить» | B |
| Already stated earlier in the conversation | Do not ask again |

**Default if user refuses to choose:** Mode B (safer).

## Workflow Overview

```
Orient → Execution mode (A/B)? → Select ONE issue → Branch → Implement → Verify → PR → Mode gate → (next issue)
```

Track progress with this checklist:

```
- [ ] Context read (docs if present + issue + repo state)
- [ ] Execution mode A or B confirmed
- [ ] Issue selected (specified or auto-picked)
- [ ] Feature branch created and checked out
- [ ] Acceptance criteria implemented
- [ ] Tests added/updated (same change set)
- [ ] Project tests pass locally **or** CI checks green after push (see Makefile / README / CI workflow)
- [ ] Key stages commented on the issue (see below)
- [ ] User notified (done or paused)
- [ ] Pull request opened (when implementation is complete; `Closes #N` — single issue)
- [ ] Mode gate applied (A: merged + issue closed + main pulled; B: stopped for user review)
- [ ] Only then: next issue (if sequential run continues)
```

## Step 1 — Orient

Before writing code:

1. **Product/tech context** (first match wins):
   - If skill `project-docs` exists or `.docs/` is present: read `project-overview.md` → `prd.md` → `technical-design.md`.
   - Else: README, CONTRIBUTING, and the target issue body.
2. Inspect the repository — layout, conventions, test commands (Makefile, package scripts, CI).
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

Follow the issue acceptance criteria and any technical design doc in the repo (e.g. `.docs/technical-design.md`).

### Coding rules

- Match existing project conventions (structure, naming, error handling).
- Keep changes scoped to the issue — no drive-by refactors.
- Separate domain/business logic from delivery layers (CLI, HTTP, adapters) when the repo already does — follow local patterns.
- Use interfaces/ports for external dependencies when the architecture uses them.
- Tests are **mandatory** for changed business logic — include them in the same change set, not a follow-up PR.

### Questions policy

- **Do not ask** for decisions already documented in repo docs or the issue.
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
- Tests: [project test command] — pass
```

**Issue closure:** In **mode B**, do not close the issue — the user merges when ready. In **mode A**, the issue closes on merge via `Closes #N`; verify it is closed after you merge.

## Step 5 — Verify

Before notifying the user or opening a PR:

1. Run the project's documented test command locally when the toolchain is available (Makefile, README, `package.json`, etc.).
2. If local tools are unavailable, rely on **CI** after push; wait for required checks and report status.
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
- [project test command] — pass

### Next
- **Mode B:** review the PR; merge when satisfied (issue closes on merge).
- **Mode A:** (agent continues to next issue.)
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
- [ ] [project test command from CI/Makefile]
- [ ] [manual steps if relevant]

## Notes
[optional: follow-ups, deferred items]
```

Link the issue with `Closes #N` (or `Fixes #N`) so it auto-closes on merge.

### After PR

1. Report CI status.
2. Apply **execution mode** (see above):
   - **Mode A:** merge PR → verify issue **closed** → `git pull` on default branch → next issue.
   - **Mode B:** stop → PR ready for user review → do not merge or close issue → wait for user.

## Boundaries

| Do | Don't |
|----|-------|
| **One issue** per branch and per PR | Pack multiple issues into one PR without explicit user request |
| Work **sequentially** from updated `main` when user asks for ordered execution | Start #N+1 before #N is merged (unless user overrides) |
| Ask **execution mode (A/B)** once per session if not specified | Assume mode without asking |
| Read repo docs (`.docs/` if present) before coding | Store long-term requirements only in issues |
| **Mode A:** merge, verify issue closed, pull main, continue | Merge or close issues in mode B |
| **Mode B:** stop after PR; user merges and closes | Start next issue before user merges in mode B |
| Respect issue dependencies | Start blocked issues without override |
| Ask minimal blocking questions | Ask preference questions already answered in docs |
| Add tests with feature code | Defer tests to a follow-up PR |
| Comment on issue at key stages | Dump verbose play-by-play on every commit |
| Commit on feature branch | Commit directly on main/master |

## Optional related skills (reference)

See **Optional related skills** at the top. This skill does not require them.
