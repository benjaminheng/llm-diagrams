Generate diagrams from text descriptions using LLMs. Only PlantUML diagrams are
supported. Only Anthropic models are supported.

The `anthropic` package is lifted from another project of mine. The rest of the
code is mostly generated using Claude 3.5 Sonnet, with some manual tweaks.

Usage:

1. Ensure you have plantuml installed locally.
2. Set `ANTHROPIC_API_KEY` environment variable.
3. Execute `go run main.go`.
