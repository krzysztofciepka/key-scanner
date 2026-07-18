# key-scanner

Scan public GitHub repositories for exposed AI subscription API keys.

## Install

```bash
go install github.com/krzysztofciepka/key-scanner@latest
```

## Usage

**Scan GitHub for exposed keys:**

```bash
export GITHUB_TOKEN=ghp_...
key-scanner scan
```

Output:
```
┌──────────────────────┬──────────────────────┬──────────────────────┬─────────────┬─────────────────────┐
│ KEY                  │ REPO                 │ FILE                 │ COMMIT DATE │ VALUE               │
├──────────────────────┼──────────────────────┼──────────────────────┼─────────────┼─────────────────────┤
│ OPENAI_API_KEY       │ user/my-project      │ .env                 │ 2025-07-14  │ sk-proj-abc...xyz   │
└──────────────────────┴──────────────────────┴──────────────────────┴─────────────┴─────────────────────┘
```

**Scan for a specific key pattern:**

```bash
key-scanner scan --pattern OPENAI_API_KEY --limit 50
```

**Test a found key against its provider:**

```bash
key-scanner test sk-proj-abc123xyz...
```

## Searched Key Patterns

OpenAI, Anthropic, OpenRouter, OpenCode Go, Google AI (Gemini), Groq, Together AI, DeepSeek, Cohere, Replicate, Hugging Face, Mistral, Perplexity, Blackbox AI, xAI, Fireworks, GitHub tokens.

## Authentication

Requires a GitHub personal access token in the `GITHUB_TOKEN` environment variable with no special scopes (public repo access is sufficient for code search).

## License

MIT
