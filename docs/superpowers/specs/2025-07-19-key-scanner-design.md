# Key Scanner — Design Spec

A CLI tool that searches public GitHub repos for exposed AI subscription API keys via GitHub's code search API, displays them in a table, and optionally validates individual keys.

## Overview

- **Language:** Go
- **Type:** Single-binary CLI utility
- **Auth:** `GITHUB_TOKEN` environment variable
- **Output:** Human-readable table (default)

## Key Patterns

Built-in curated list of AI provider API key environment variables. Each maps to a provider name for the `test` subcommand.

| Env Var | Provider |
|---------|----------|
| `OPENAI_API_KEY` | OpenAI |
| `ANTHROPIC_API_KEY` | Anthropic |
| `OPENROUTER_API_KEY` | OpenRouter |
| `OPENCODE_API_KEY` | OpenCode Go |
| `GOOGLE_API_KEY`, `GEMINI_API_KEY` | Google AI |
| `GROQ_API_KEY` | Groq |
| `TOGETHER_API_KEY` | Together AI |
| `DEEPSEEK_API_KEY` | DeepSeek |
| `COHERE_API_KEY` | Cohere |
| `REPLICATE_API_KEY` | Replicate |
| `HF_API_KEY`, `HUGGINGFACE_API_KEY` | Hugging Face |
| `MISTRAL_API_KEY` | Mistral |
| `PERPLEXITY_API_KEY` | Perplexity |
| `BLACKBOX_AI_KEY` | Blackbox AI |
| `XAI_API_KEY` | xAI / Grok |
| `FIREWORKS_API_KEY` | Fireworks |
| `GITHUB_TOKEN`, `GH_TOKEN` | GitHub |

Search query pattern for each: `<ENV_VAR>=` (finds assignments in `.env` files, shell scripts, configs, etc.)

## Architecture

```
key-scanner/
├── main.go                  # Entry point, routes to subcommands
├── cmd/
│   ├── scan.go              # 'scan' subcommand
│   └── test.go              # 'test' subcommand
├── internal/
│   ├── scanner/
│   │   └── scanner.go       # GitHub code search queries, pagination, rate limiting
│   ├── keys/
│   │   ├── patterns.go      # Built-in key list: env var → provider mapping
│   │   └── validate.go      # Key validation logic per provider
│   └── output/
│       └── table.go         # Formatted table output
├── go.mod
└── go.sum
```

**Dependencies:** Standard library `net/http` for GitHub API calls. A lightweight table library (`github.com/jedib0t/go-pretty/v6/table`) for pretty-printed output. No heavy frameworks.

## CLI Interface

### `key-scanner scan`

Searches all built-in patterns via GitHub's code search API.

```
$ key-scanner scan

┌──────────────────────┬──────────────────────┬──────────────────────┬─────────────┬─────────────────────┐
│ KEY                  │ REPO                 │ FILE                 │ COMMIT DATE │ VALUE               │
├──────────────────────┼──────────────────────┼──────────────────────┼─────────────┼─────────────────────┤
│ OPENAI_API_KEY       │ user/my-project      │ .env                 │ 2025-07-14  │ sk-proj-abc...xyz   │
│ ANTHROPIC_API_KEY    │ org/backend          │ config/prod.env      │ 2025-01-03  │ sk-ant-api03-...    │
│ OPENROUTER_API_KEY   │ user/dotfiles        │ .zshrc               │ 2024-11-22  │ sk-or-v1-...        │
└──────────────────────┴──────────────────────┴──────────────────────┴─────────────┴─────────────────────┘

Flags:
  --limit N       Max results per pattern (default: 100)
  --pattern NAME  Search only a specific key pattern (e.g., OPENAI_API_KEY)
```

### `key-scanner test <value>`

Validates a key against its provider's API. Auto-detects the provider from key format prefixes:

- `sk-proj-*`, `sk-admin-*` → OpenAI
- `sk-ant-api03-*` → Anthropic
- `sk-or-v1-*` → OpenRouter
- `ghp_*`, `github_pat_*` → GitHub
- `hf_*` → Hugging Face
- And others based on known key formats.

```
$ key-scanner test sk-proj-abc123xyz...
Valid ✓ (OpenAI)

$ key-scanner test sk-deadbeef...
Invalid ✗ - 401 Unauthorized (OpenAI)

Flags:
  --provider NAME  Force a specific provider for validation
```

## Data Flow

### Scan Flow
1. Read all key patterns from `internal/keys/patterns.go`
2. For each pattern, issue a GitHub code search query: `"<ENV_VAR>="` (with quotes for exact match)
3. Concurrent requests, rate-limited to 2s between calls to stay under GitHub's 30 req/min limit
4. Parse each response: extract repo name, file path, commit date, and key value snippet
5. Deduplicate by (key_value_hash, repo, file) — same key committed to same file multiple times counts once, using most recent commit date
6. Sort results by commit date (most recent first)
7. Print table

### Test Flow
1. Receive key value as argument
2. Auto-detect provider from key prefix (regex matching against known formats)
3. Call the provider's validation endpoint (e.g., `GET https://api.openai.com/v1/models` with the key as Bearer token)
4. Report valid/invalid with HTTP status

## Rate Limiting & Error Handling

- **Rate limit:** 2-second delay between API calls. On HTTP 403/429, exponential backoff (4s, 8s, 16s) with up to 3 retries.
- **Pagination:** GitHub returns up to 100 results per page, max 1000 total. Follow `Link` headers.
- **Network errors:** Retry 3x with 2s delay. If all retries fail, skip that pattern and continue.
- **Auth errors:** If GITHUB_TOKEN is missing or invalid, exit immediately with a clear message.
- **No results:** Print "No exposed keys found." and exit 0.
- **Timeouts:** 30s per HTTP request.

## Provider Validation Endpoints

| Provider | Validation Endpoint | Method |
|----------|-------------------|--------|
| OpenAI | `https://api.openai.com/v1/models` | GET |
| Anthropic | `https://api.anthropic.com/v1/messages` | GET |
| OpenRouter | `https://openrouter.ai/api/v1/models` | GET |
| GitHub | `https://api.github.com/user` | GET |
| Google AI | `https://generativelanguage.googleapis.com/v1/models` | GET |
| Groq | `https://api.groq.com/openai/v1/models` | GET |
| Together AI | `https://api.together.xyz/v1/models` | GET |
| Mistral | `https://api.mistral.ai/v1/models` | GET |
| Perplexity | `https://api.perplexity.ai/auth/status` | GET |
| Hugging Face | `https://huggingface.co/api/whoami-v2` | GET |
| DeepSeek | `https://api.deepseek.com/user/balance` | GET |
| Cohere | `https://api.cohere.ai/v1/check-api-key` | GET |
| Replicate | `https://api.replicate.com/v1/account` | GET |
| xAI | `https://api.x.ai/v1/models` | GET |
| Fireworks | `https://api.fireworks.ai/v1/account` | GET |
| Blackbox AI | (no public validation endpoint known — skip auto-detect, require `--provider` with manual check) | — |
| OpenCode Go | (no public validation endpoint known — skip auto-detect) | — |

## Testing Strategy

- **Unit tests:** Pattern definitions, key prefix parsing, provider detection, deduplication logic, table formatting
- **Integration tests:** Mock HTTP server simulating GitHub code search API and provider validation endpoints
- **CI:** GitHub Actions, run tests on push

## Non-Goals

- No caching of search results (keep it stateless)
- No GitHub secret scanning alerts API (different product)
- No local file system scanning (focus on public GitHub)
- No Slack/email notifications
