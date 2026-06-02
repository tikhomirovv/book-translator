# book-translator

CLI tool to translate books (PDF → Markdown) using LLM with resumable progress and context memory.

## Documentation

Product and technical docs: [`.docs/`](.docs/)

- [Project overview](.docs/project-overview.md)
- [PRD](.docs/prd.md)
- [Technical design](.docs/technical-design.md)

## Status

MVP in development. See [GitHub Issues](https://github.com/tikhomirovv/book-translator/issues) and [Project board](https://github.com/users/tikhomirovv/projects/4).

## Quick start (dev)

```bash
cp .env.example .env   # add OPENAI_API_KEY
make build
./bin/translator version
make test
```

## License

TBD
