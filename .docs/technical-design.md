# Technical Design

## Стек

| Компонент | Выбор | Обоснование |
|-----------|--------|-------------|
| Язык | Go 1.22+ | Конкурентность, один бинарник |
| CLI | [cobra](https://github.com/spf13/cobra) | Подкоманды, флаги |
| Конфиг | [viper](https://github.com/spf13/viper) | defaults → `config.yaml` → `config.local.yaml` → env |
| Логи | [zerolog](https://github.com/rs/zerolog) | Структурированные логи |
| Progress | [schollz/progressbar](https://github.com/schollz/progressbar) | CLI progress |
| PDF (MVP) | [razvandimescu/gopdf](https://github.com/razvandimescu/gopdf) | MIT, ToUnicode/encoding, reflow параграфов |
| Контекст (future) | [tqbf/contextwindow](https://github.com/tqbf/contextwindow) | Автосжатие контекста |

### Выбор PDF-библиотеки

| Библиотека | Stars (≈) | Лицензия | Текст для книг | Роль в проекте |
|------------|-----------|----------|----------------|----------------|
| **razvandimescu/gopdf** | новая | MIT | Positioned + line rebuild | **MVP-адаптер** |
| **ledongthuc/pdf** | ~585 | BSD-3 | Plain text по страницам | Заменён в MVP |

**Решение MVP:** `razvandimescu/gopdf` за портом `TextExtractor`. Постобработка: reflow переносов → нормализация параграфов.

**Ограничения:** сканы без текстового слоя и сложная вёрстка — вне MVP; OCR — non-goal.

---

## Архитектурный стиль

**Standard Go Project Layout** + **Clean Architecture (ports & adapters)**:

- **Domain** — сущности и интерфейсы (ports); без зависимостей от фреймворков.
- **Application** — use cases; оркестрация domain + ports.
- **Infrastructure** — adapters (LLM, PDF, storage, context memory).
- **Interfaces** — driving adapters (CLI сейчас, HTTP позже).
- **Composition root** — `cmd/translator/main.go`: wiring, DI.

**Правило зависимостей:** `interfaces` → `application` → `domain` ← `infrastructure`.

---

## Структура проекта

```
book-translator/
├── cmd/
│   └── translator/
│       └── main.go                 # composition root, DI
│
├── configs/
│   ├── config.yaml                 # defaults (в git)
│   └── config.local.yaml.example   # шаблон локальных overrides
│
├── internal/
│   ├── domain/                     # entities + ports
│   │   ├── translation.go
│   │   ├── chunk.go
│   │   ├── paragraph.go
│   │   ├── context_memory.go
│   │   ├── errors.go
│   │   └── ports/
│   │       ├── text_extractor.go   # PDF/EPUB/TXT
│   │       ├── llm_provider.go
│   │       ├── context_manager.go
│   │       ├── translation_store.go
│   │       ├── prompt_renderer.go
│   │       └── token_counter.go      # future: tiktoken
│   │
│   ├── application/                # use cases
│   │   ├── translate/
│   │   │   ├── start_translation.go
│   │   │   ├── process_chunk.go
│   │   │   └── finalize_translation.go
│   │   ├── resume/
│   │   │   └── resume_translation.go
│   │   ├── query/
│   │   │   ├── get_status.go
│   │   │   └── list_translations.go
│   │   └── pipeline/
│   │       ├── chunker.go            # paragraphs → chunks
│   │       └── paragraph_normalizer.go
│   │
│   ├── infrastructure/
│   │   ├── config/
│   │   │   └── loader.go             # viper
│   │   ├── llm/
│   │   │   ├── openai_compatible.go
│   │   │   ├── rate_limiter.go
│   │   │   └── retry.go
│   │   ├── extract/
│   │   │   ├── pdf_extractor.go      # gopdf adapter
│   │   │   ├── paragraph_reflow.go
│   │   │   └── registry.go
│   │   ├── context/
│   │   │   └── fixed_window.go       # LLM-consolidated context store
│   │   ├── store/
│   │   │   └── filesystem.go
│   │   ├── prompt/
│   │   │   └── yaml_renderer.go
│   │   └── logging/
│   │       └── zerolog.go
│   │
│   └── interfaces/
│       └── cli/
│           ├── root.go
│           ├── translate.go
│           ├── extract.go
│           ├── resume.go
│           ├── status.go
│           ├── list.go
│           ├── interactive.go
│           └── progress.go
│
├── translations/                     # runtime (gitignore)
├── .docs/
├── .env.example
├── .gitignore
├── go.mod
├── Makefile
└── README.md
```

**Расширение без переделки ядра:**
- `internal/interfaces/http/` — REST handlers, те же use cases.
- `internal/infrastructure/storage/postgres_translation_store.go` — новый adapter.
- `internal/infrastructure/extract/` — дополнительные `TextExtractor` (OCR, EPUB, …).

---

## Ключевые решения

### 1. Translation — центральная сущность

`ID` — **UUID v4**. Одна книга → несколько переводов (`--to`, `prompt-type`).

### 2. Исходный язык

Не передаётся в CLI. Модель определяет из текста. Whitelist: `allowed_languages` только для целевого языка.

### 3. Конфигурация

| Слой | Содержимое | Git |
|------|------------|-----|
| `.env` | API keys, base URL | нет |
| `config.yaml` | chunk, prompts, llm, languages, delays | да |
| `config.local.yaml` | overrides | нет |

```yaml
# фрагмент config.yaml
chunk:
  size_paragraphs: 10
  overlap_paragraphs: 2

llm:
  translation:
    model: gpt-4o-mini
    temperature: 0.3
    max_tokens: 32768
  context:
    model: gpt-4o-mini
    temperature: 0.2
    max_tokens: 8192

request_delay_ms: 1000

allowed_languages:
  - ru
  - en
  - de

prompts:
  nonfiction:
    system: "..."
    translation: "..."
    context_extraction: "..."
```

### 4. Ports (domain)

```go
// TextExtractor — входной файл → []Paragraph
type TextExtractor interface {
    Extract(ctx context.Context, path string) ([]Paragraph, error)
}

// LLMProvider — OpenAI-compatible chat
type LLMProvider interface {
    Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// ContextManager — память перевода
type ContextManager interface {
    AddExtracted(ctx context.Context, chunkIndex int, data map[string]any) error
    BuildPromptContext(ctx context.Context) (string, error)
    Save(ctx context.Context, translationID string) error
    Load(ctx context.Context, translationID string) error
}

// TranslationStore — персистентность процесса
type TranslationStore interface {
    Create(ctx context.Context, t *Translation) error
    Load(ctx context.Context, id string) (*TranslationState, error)
    SaveChunk(ctx context.Context, id string, chunk Chunk) error
    List(ctx context.Context) ([]TranslationSummary, error)
}

// PromptRenderer — шаблоны из config по prompt-type
type PromptRenderer interface {
    Render(promptType, templateKey string, data PromptData) (string, error)
}
```

### 5. Token usage и стоимость

```go
type Usage struct {
    PromptTokens     int
    CompletionTokens int
    TotalTokens      int
    EstimatedCost    *float64 // optional
}
```

- Заполнять из `response.Usage` провайдера, если есть.
- Иначе — WARN в лог, нули в state; интерфейс `TokenCounter` (tiktoken) — заглушка / no-op в MVP.

### 6. Context memory (MVP)

`FixedWindow` (`infrastructure/context/fixed_window.go`) хранит один consolidated-блок `ContextSummary`, который возвращает контекстная LLM. Клиент не обрезает память — лимит задаётся через `llm.context.max_tokens` (API + промпт `{{.MaxContextTokens}}`).

Поток на чанк:
1. `BuildPromptContext` — контекст из предыдущих чанков → в промпт перевода.
2. Параллельно: перевод + context extraction (previous context + новый excerpt).
3. `AddExtracted` — заменяет сохранённый контекст ответом модели.

| Стратегия (future) | Реализация |
|--------------------|------------|
| LLM consolidation | `fixed_window.go` (MVP) |
| `contextwindow` | v0.2+ |

### 7. Файловое хранилище

```
translations/<uuid>/
  state.json
  source.meta.json
  chunks/0001.md
  output.md              # финальный склей после complete
```

### 8. Параллелизм

- Цикл по чанкам — **последовательно**.
- `ProcessChunk` — `errgroup`: translate + context extraction (параллельно внутри чанка).

### 9. LLM обёртки

Декоратор над `LLMProvider`: `RetryLLM` → `RateLimitedLLM` → concrete client.

---

## Основные сущности

### Translation

| Поле | Описание |
|------|----------|
| ID | UUID v4 |
| SourcePath | Путь к PDF |
| TargetLang | Из `allowed_languages` |
| PromptType | Ключ в `prompts` |
| Status | pending \| running \| paused \| completed \| failed |
| LastCompletedChunk | int |
| TotalChunks | int |

### Paragraph / Chunk

- `Paragraph`: index, text.
- `Chunk`: index, `ParagraphRange [start,end)`, source text, translated text, overlap text.

### TranslationState (`state.json`)

Агрегат: прогресс, `contextSummary` (consolidated LLM memory), cumulative `Usage`.

---

## Инженерные правила

1. Интерфейсы — только в `domain/ports`; реализации — в `infrastructure`.
2. Use cases не импортируют `cobra`, `viper`, HTTP.
3. Каждый успешный чанк — `SaveChunk` до следующего.
4. `resume` идемпотентен: не переводить `index <= LastCompletedChunk`.
5. Новый adapter = новый файл в `infrastructure`, без изменения `application` (кроме wiring в `main`).
6. Комментарии в коде — English; `.docs/` — Russian.

---

## Интеграции

| Интеграция | MVP | Позже |
|------------|-----|-------|
| OpenAI-compatible API | да | — |
| razvandimescu/gopdf | да | — |
| OCR / scanned PDF | нет | отдельная задача |
| HTTP API | нет | v0.3 |
| PostgreSQL store | нет | v0.4 |
| tiktoken TokenCounter | stub | v0.2 |
| Grafana | нет | после стабилизации |

---

## Риски

| Риск | Митигация |
|------|-----------|
| Плохой PDF extract (layout) | Документировать; адаптер gopdf; OCR — non-goal |
| Потеря длинного контекста | LLM consolidation + `llm.context.max_tokens`; стратегии позже |
| Нет usage в ответе API | WARN + tiktoken позже |
| Rate limit 429 | retry + `request_delay_ms` |
| ledongthuc/pdf устарел / заменён | Порт `TextExtractor` изолирует замену |
