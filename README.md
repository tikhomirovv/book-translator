# book-translator

CLI tool to translate books (PDF → Markdown) using an LLM, with resumable progress and translation context memory.

## Documentation

Product and technical docs live in [`.docs/`](.docs/):

- [Project overview](.docs/project-overview.md)
- [PRD](.docs/prd.md)
- [Technical design](.docs/technical-design.md)

## Requirements

- Go 1.24+
- OpenAI-compatible API key (or compatible provider base URL)

## Install and build

```bash
git clone https://github.com/tikhomirovv/book-translator.git
cd book-translator
cp .env.example .env   # set OPENAI_API_KEY
make build             # produces bin/translator
make test
```

## Configuration

| Layer | Purpose |
|-------|---------|
| `.env` | Secrets: `OPENAI_API_KEY`, optional `OPENAI_BASE_URL`, `LOG_LEVEL` |
| `configs/config.yaml` | Defaults: chunk size, context strategy, prompts, LLM model |
| `configs/config.local.yaml` | Local overrides (not committed) |

Allowed target languages are listed in `configs/config.yaml` under `allowed_languages`.

To translate only part of a book while tuning prompts (cheap iteration), set in `config.local.yaml`:

```yaml
translation:
  paragraph_from: 30   # inclusive, 0-based paragraph index
  paragraph_to: 70       # inclusive; use -1 for open end
```

Leave both at `-1` (default) for a full-book run.

## Usage

### Translate a new book (flags)

```bash
./bin/translator translate \
  --input book.pdf \
  --output book.ru.md \
  --to ru \
  --prompt-type nonfiction
```

Shorthand flags: `-i` / `-o`.

### Interactive mode

Omit flags to be prompted for input path, output path, target language, and prompt type:

```bash
./bin/translator translate
```

### Resume, status, list

```bash
./bin/translator resume --id <translation-uuid>
./bin/translator status --id <translation-uuid>
./bin/translator list
```

Progress is stored under `translations/<uuid>/`. The translation ID is printed when a job starts and appears in `list` / `status`.

## Output

The output Markdown file includes YAML frontmatter:

- `target_lang`, `date`, `model`, `translation_id`
- Token usage (`prompt_tokens`, `completion_tokens`, `total_tokens`)

The body is the concatenated translated chunks.

## Known PDF limitations (MVP)

- Requires a **text-based PDF** (no OCR for scanned pages).
- Complex layout (multi-column, heavy formatting) may produce noisy paragraph splits.
- MVP uses plain-text extraction via `razvandimescu/gopdf` with paragraph reflow; quality varies by source file.

## Manual test checklist (MVP)

Use this to verify PRD MVP acceptance locally:

- [ ] `make build` and `make test` succeed
- [ ] `./bin/translator version` prints a version string
- [ ] `translate` with flags creates output Markdown and a `translations/<uuid>/` directory
- [ ] Output frontmatter contains target language, model, translation ID, token fields
- [ ] Interrupt a long run (Ctrl+C), then `resume --id …` continues from the last saved chunk
- [ ] `status --id …` shows chunk progress `N/M`, status, and token usage
- [ ] `list` shows id, source path, target language, progress, and updated date
- [ ] Invalid `--to` language is rejected with a clear error

## Development

```bash
make build
make test
make lint   # optional; requires golangci-lint
```

Integration tests (offline, no API calls): `go test ./tests/integration/...`

## License

TBD
