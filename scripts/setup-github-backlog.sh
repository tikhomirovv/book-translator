#!/usr/bin/env bash
# One-time setup: labels, milestones, issues, dependencies, project board.
set -euo pipefail

REPO="tikhomirovv/book-translater"
OWNER="tikhomirovv"
NAME="book-translater"

echo "==> Labels"
for spec in \
  "priority:p0|b60205|Must have for MVP" \
  "priority:p1|fbca04|Post-MVP" \
  "type:feature|1d76db|Feature" \
  "type:chore|c5def5|Chore" \
  "area:domain|5319e7|Domain layer" \
  "area:application|006b75|Application use cases" \
  "area:infra|0e8a16|Infrastructure adapters" \
  "area:cli|d93f0b|CLI interface"; do
  IFS='|' read -r label color desc <<<"$spec"
  gh label create "$label" --repo "$REPO" --color "$color" --description "$desc" 2>/dev/null || true
done

echo "==> Milestones"
MVP_MS=$(gh api "repos/$REPO/milestones" -f title="MVP — CLI Book Translator" \
  -f description="PDF in, Markdown out, resumable CLI translation." \
  -f state=open --jq .number)
V02_MS=$(gh api "repos/$REPO/milestones" -f title="v0.2 — Context & Formats" \
  -f description="auto_summarize, EPUB, tiktoken, gopdf adapter." \
  -f state=open --jq .number)
echo "MVP milestone #$MVP_MS, v0.2 milestone #$V02_MS"

create_issue() {
  local title="$1"
  local body="$2"
  local labels="$3"
  local milestone="$4"
  gh issue create --repo "$REPO" --title "$title" --body "$body" \
    --label "$labels" --milestone "$milestone" | sed -n 's|.*/issues/\([0-9]*\).*|\1|p'
}

block() {
  local blocked="$1"
  local blocker="$2"
  local blocker_id
  blocker_id=$(gh api "repos/$REPO/issues/$blocker" --jq .id)
  gh api "repos/$REPO/issues/$blocked/dependencies/blocked_by" \
    --method POST --input - <<< "{\"issue_id\":${blocker_id}}" >/dev/null
  echo "  #$blocked blocked by #$blocker"
}

echo "==> MVP issues"
# Returns issue numbers via global array
ISSUES=()

n=$(create_issue "Scaffold: Go project layout and module" "$(cat <<'BODY'
## Summary
Bootstrap repository structure per `.docs/technical-design.md`: `cmd/translator`, `internal/{domain,application,infrastructure,interfaces}`, `configs/`, `go.mod`, `Makefile`, `.env.example`.

## Acceptance criteria
- [ ] `go mod init` / module path matches repo
- [ ] Directory tree matches technical design (empty packages OK)
- [ ] `cmd/translator/main.go` compiles (minimal hello or version)
- [ ] `Makefile` targets: `build`, `test`, `lint` (lint optional stub)
- [ ] `configs/config.yaml` and `config.local.yaml.example` present
- [ ] `.env.example` documents required secrets only

## Depends on
—

## Docs
`.docs/technical-design.md` — Structure
BODY
)" "type:chore,area:infra,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n")

n=$(create_issue "CI: GitHub Actions go test" "$(cat <<'BODY'
## Summary
Add `.github/workflows/ci.yml` running `go test ./...` on push/PR.

## Acceptance criteria
- [ ] Workflow runs on `main` and pull requests
- [ ] Uses Go 1.22+
- [ ] Fails PR if tests fail
- [ ] Badge optional in README

## Depends on
#1 (scaffold / go.mod)
BODY
)" "type:chore,area:infra,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); CI=$n

n=$(create_issue "Domain: entities and port interfaces" "$(cat <<'BODY'
## Summary
Implement domain models and ports in `internal/domain` — no external dependencies.

## Acceptance criteria
- [ ] Entities: `Translation`, `Chunk`, `Paragraph`, `TranslationState`, `Usage`
- [ ] Ports: `TextExtractor`, `LLMProvider`, `ContextManager`, `TranslationStore`, `PromptRenderer`, `TokenCounter` (no-op stub OK)
- [ ] Domain errors typed where useful
- [ ] Unit tests for entity helpers / validation (e.g. target lang)

## Depends on
#1
BODY
)" "type:feature,area:domain,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); DOMAIN=$n

n=$(create_issue "Config: Viper loader and default config.yaml" "$(cat <<'BODY'
## Summary
Load layered config: defaults → `configs/config.yaml` → optional `config.local.yaml` → env for secrets.

## Acceptance criteria
- [ ] `internal/infrastructure/config/loader.go` with Viper
- [ ] `chunk.size_paragraphs: 10`, `overlap_paragraphs: 2`, `context.*`, `llm.*`, `request_delay_ms`, `allowed_languages`
- [ ] Sample `prompts` entries: `nonfiction`, `fiction` (system, translation, context_extraction)
- [ ] Tests: merge local override, env for API key
- [ ] Secrets only from `.env`, not committed

## Depends on
#1
BODY
)" "type:feature,area:infra,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); CONFIG=$n

n=$(create_issue "Infrastructure: filesystem TranslationStore" "$(cat <<'BODY'
## Summary
Persist translation progress under `translations/<uuid>/`.

## Acceptance criteria
- [ ] Implements `TranslationStore` port
- [ ] Layout: `state.json`, `source.meta.json`, `chunks/NNNN.md`
- [ ] `Create`, `Load`, `SaveChunk`, `List` with summaries
- [ ] UUID v4 for new translation IDs
- [ ] Unit tests with temp dir

## Depends on
Domain ports (#3)
BODY
)" "type:feature,area:infra,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); STORE=$n

n=$(create_issue "Infrastructure: PDF extract and chunk pipeline" "$(cat <<'BODY'
## Summary
`TextExtractor` using `github.com/ledongthuc/pdf`; normalize paragraphs; chunk by config size with overlap.

## Acceptance criteria
- [ ] `pdf_ledongthuc.go` adapter
- [ ] `paragraph_normalizer.go`: split on `\n\n`, trim
- [ ] `chunker.go`: 10 paragraphs default, overlap 2–3 from config
- [ ] Registry by file extension (PDF only for MVP)
- [ ] Unit tests with small fixture PDF or mocked reader

## Depends on
Domain (#3)
BODY
)" "type:feature,area:infra,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); EXTRACT=$n

n=$(create_issue "Infrastructure: LLM OpenAI-compatible adapter" "$(cat <<'BODY'
## Summary
HTTP client for OpenAI-compatible chat API with retry (exponential backoff) and rate limit (delay).

## Acceptance criteria
- [ ] Implements `LLMProvider`
- [ ] Reads model, temperature, max_tokens from config; API key/base URL from env
- [ ] Parses `usage` from response when present; WARN if missing
- [ ] `RetryLLM` and `RateLimitedLLM` decorators
- [ ] Unit tests with httptest mock server

## Depends on
Domain (#3), Config (#4)
BODY
)" "type:feature,area:infra,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); LLM=$n

n=$(create_issue "Infrastructure: ContextManager fixed_window" "$(cat <<'BODY'
## Summary
MVP context strategy: glossary + rolling summary + overlap, ~2000 token budget (approximate).

## Acceptance criteria
- [ ] `fixed_window.go` implements `ContextManager`
- [ ] `factory.go` selects strategy from `context.strategy` (only fixed_window for now)
- [ ] `AddExtracted`, `BuildPromptContext`, `Save`/`Load` integrate with store state
- [ ] Unit tests for eviction / budget behavior

## Depends on
Domain (#3), TranslationStore (#5)
BODY
)" "type:feature,area:infra,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); CTX=$n

n=$(create_issue "Infrastructure: YAML PromptRenderer" "$(cat <<'BODY'
## Summary
Render prompts by `prompt-type` key from config with template data (target lang, context block, chunk text).

## Acceptance criteria
- [ ] Implements `PromptRenderer`
- [ ] Templates: `system`, `translation`, `context_extraction`
- [ ] Unknown prompt-type returns clear error
- [ ] Unit tests for rendering

## Depends on
Config (#4)
BODY
)" "type:feature,area:infra,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); PROMPT=$n

n=$(create_issue "Application: process chunk (translate + extract context)" "$(cat <<'BODY'
## Summary
Core loop for one chunk: build prompts, call LLM (parallel translate + context extraction inside chunk via errgroup), update ContextManager, persist chunk.

## Acceptance criteria
- [ ] `process_chunk.go` in `internal/application/translate`
- [ ] Sequential chunk index enforced by caller
- [ ] Adaptive context extraction prompt (not hardcoded to characters)
- [ ] Saves chunk to store on success
- [ ] Unit tests with mock LLM and mock store

## Depends on
Extract (#6), LLM (#7), Context (#8), Prompt (#9)
BODY
)" "type:feature,area:application,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); PROCESS=$n

n=$(create_issue "Application: start translation and finalize Markdown" "$(cat <<'BODY'
## Summary
Orchestrate full new translation: extract → chunk → loop process_chunk → write output Markdown with YAML frontmatter.

## Acceptance criteria
- [ ] `start_translation.go`, `finalize_translation.go`
- [ ] Frontmatter: target lang, date, model, translation-id, token usage
- [ ] Validates `--to` against `allowed_languages`
- [ ] Does not require `--from` (model detects source)
- [ ] Integration-ready structure (tested via mocks in separate issue)

## Depends on
Process chunk (#10), TranslationStore (#5)
BODY
)" "type:feature,area:application,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); START=$n

n=$(create_issue "Application: resume, status, list translations" "$(cat <<'BODY'
## Summary
Use cases for interrupted runs and visibility.

## Acceptance criteria
- [ ] `resume_translation.go`: idempotent, skip completed chunks
- [ ] `get_status.go`: N/M chunks, usage, status, last error
- [ ] `list_translations.go`: table fields per PRD
- [ ] Unit tests with fixture state on disk

## Depends on
Process chunk (#10), TranslationStore (#5)
BODY
)" "type:feature,area:application,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); QUERY=$n

n=$(create_issue "CLI: cobra commands, logging, progress bar" "$(cat <<'BODY'
## Summary
User-facing CLI: `translate`, `resume`, `status`, `list` with flags and interactive mode.

## Acceptance criteria
- [ ] Cobra root + subcommands per PRD
- [ ] Flag mode: `--input`, `--output`, `--to`, `--prompt-type`
- [ ] Interactive mode when flags omitted
- [ ] Zerolog to stderr; progress bar compatible with logs
- [ ] `main.go` wires DI (config, adapters, use cases)

## Depends on
Start (#11), Query (#12)
BODY
)" "type:feature,area:cli,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); CLI_ISSUE=$n

n=$(create_issue "Integration tests: end-to-end with mock LLM" "$(cat <<'BODY'
## Summary
E2E test: small PDF/text fixture → mock LLM → assert chunks and output file without real API.

## Acceptance criteria
- [ ] `internal/..._test.go` or `tests/integration/` per project convention
- [ ] No network in CI
- [ ] Covers translate start + resume path minimally
- [ ] `go test ./...` green

## Depends on
CLI (#13)
BODY
)" "type:chore,area:application,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); INTEG=$n

n=$(create_issue "MVP polish: README, examples, manual test checklist" "$(cat <<'BODY'
## Summary
Document how to run first translation; finalize `.env.example`; manual test checklist in PR description template or docs.

## Acceptance criteria
- [ ] README: install, config, example commands
- [ ] Link to `.docs/`
- [ ] Document known PDF limitations
- [ ] All acceptance criteria from PRD MVP verifiable manually

## Depends on
Integration tests (#14)
BODY
)" "type:chore,priority:p0" "MVP — CLI Book Translator")
ISSUES+=("$n"); POLISH=$n

SCAFFOLD=${ISSUES[0]}

echo "==> Dependencies"
block "$CI" "$SCAFFOLD"
block "$DOMAIN" "$SCAFFOLD"
block "$CONFIG" "$SCAFFOLD"
block "$STORE" "$DOMAIN"
block "$EXTRACT" "$DOMAIN"
block "$LLM" "$DOMAIN"
block "$LLM" "$CONFIG"
block "$CTX" "$DOMAIN"
block "$CTX" "$STORE"
block "$PROMPT" "$CONFIG"
block "$PROCESS" "$EXTRACT"
block "$PROCESS" "$LLM"
block "$PROCESS" "$CTX"
block "$PROCESS" "$PROMPT"
block "$START" "$PROCESS"
block "$START" "$STORE"
block "$QUERY" "$PROCESS"
block "$QUERY" "$STORE"
block "$CLI_ISSUE" "$START"
block "$CLI_ISSUE" "$QUERY"
block "$INTEG" "$CLI_ISSUE"
block "$POLISH" "$INTEG"

echo "==> v0.2 issues"
n=$(create_issue "v0.2: ContextManager auto_summarize (contextwindow)" "$(cat <<'BODY'
## Summary
Second context strategy using `github.com/tqbf/contextwindow` or superfly fork.

## Acceptance criteria
- [ ] `auto_summarize.go` adapter
- [ ] Config switch `context.strategy: auto_summarize`
- [ ] Tests with mock summarizer model

## Depends on
MVP complete (#15)
BODY
)" "type:feature,area:infra,priority:p1" "v0.2 — Context & Formats")
V02_1=$n
block "$V02_1" "$POLISH"

n=$(create_issue "v0.2: EPUB TextExtractor adapter" "$(cat <<'BODY'
## Acceptance criteria
- [ ] EPUB support via new adapter behind `TextExtractor`
- [ ] Registry routes by extension
- [ ] Tests with minimal EPUB fixture
BODY
)" "type:feature,area:infra,priority:p1" "v0.2 — Context & Formats")
V02_2=$n
block "$V02_2" "$POLISH"

n=$(create_issue "v0.2: tiktoken TokenCounter fallback" "$(cat <<'BODY'
## Acceptance criteria
- [ ] Implement `TokenCounter` when API usage missing
- [ ] WARN + fallback behavior per PRD
- [ ] Tests with known token counts
BODY
)" "type:feature,area:infra,priority:p1" "v0.2 — Context & Formats")
V02_3=$n
block "$V02_3" "$POLISH"

n=$(create_issue "v0.2: optional gopdf PDF extractor adapter" "$(cat <<'BODY'
## Acceptance criteria
- [ ] `pdf_gopdf.go` behind `TextExtractor`
- [ ] Config/select extractor strategy
- [ ] Document when to use vs ledongthuc
BODY
)" "type:feature,area:infra,priority:p1" "v0.2 — Context & Formats")
V02_4=$n
block "$V02_4" "$POLISH"

echo "==> GitHub Project"
PROJ=$(gh project create --owner "@me" --title "book-translater" --format json)
PROJ_NUM=$(echo "$PROJ" | python3 -c "import sys,json; print(json.load(sys.stdin)['number'])")
PROJ_ID=$(echo "$PROJ" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
echo "Project #$PROJ_NUM id=$PROJ_ID"

echo "==> Add MVP issues to project"
for i in "${ISSUES[@]}"; do
  gh project item-add "$PROJ_NUM" --owner "@me" --url "https://github.com/$REPO/issues/$i" 2>/dev/null || \
    gh api graphql -f query='mutation($p:ID!,$c:ID!){addProjectV2ItemById(input:{projectId:$p,contentId:$c}){item{id}}}' \
      -f p="$PROJ_ID" -f c="$(gh api repos/$REPO/issues/$i --jq .node_id)" >/dev/null
done

echo "Done. MVP issues: ${ISSUES[*]}"
echo "Repo: https://github.com/$REPO"
echo "Project: #$PROJ_NUM"
