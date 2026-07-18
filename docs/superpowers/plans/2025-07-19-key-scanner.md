# Key Scanner Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI tool that searches public GitHub repos for exposed AI subscription API keys and optionally validates them.

**Architecture:** Single Go binary with two subcommands (`scan`, `test`). The scanner uses GitHub's REST code search API with text-match enrichment and the commits API for dates. Key patterns and validation endpoints are defined in a built-in registry. Output is a formatted table.

**Tech Stack:** Go 1.21+, `github.com/jedib0t/go-pretty/v6/table`, GitHub REST API, stdlib `net/http`

---

### Task 1: Initialize Go module and directory structure

**Files:**
- Create: `go.mod`
- Create: `main.go` (stub)
- Create: `cmd/scan.go` (stub)
- Create: `cmd/test.go` (stub)
- Create: `internal/keys/patterns.go` (stub)
- Create: `internal/keys/validate.go` (stub)
- Create: `internal/scanner/scanner.go` (stub)
- Create: `internal/output/table.go` (stub)

- [ ] **Step 1: Initialize Go module**

Run:
```bash
go mod init github.com/krzysztofciepka/key-scanner
```

Expected: `go.mod` created.

- [ ] **Step 2: Create directory structure**

Run:
```bash
mkdir -p cmd internal/keys internal/scanner internal/output
```

- [ ] **Step 3: Install table library dependency**

Run:
```bash
go get github.com/jedib0t/go-pretty/v6
```

- [ ] **Step 4: Write stub main.go**

Write to `main.go`:

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: key-scanner <scan|test> [flags]\n")
		os.Exit(1)
	}
	fmt.Println("key-scanner v0.1.0")
}
```

- [ ] **Step 5: Write stub files for all packages**

Write to `cmd/scan.go`:

```go
package cmd
```

Write to `cmd/test.go`:

```go
package cmd
```

Write to `internal/keys/patterns.go`:

```go
package keys
```

Write to `internal/keys/validate.go`:

```go
package keys
```

Write to `internal/scanner/scanner.go`:

```go
package scanner
```

Write to `internal/output/table.go`:

```go
package output
```

- [ ] **Step 6: Verify it builds**

Run:
```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "chore: initialize Go module and project structure"
```

---

### Task 2: Define key patterns

**Files:**
- Modify: `internal/keys/patterns.go`

- [ ] **Step 1: Write key patterns registry**

Write to `internal/keys/patterns.go`:

```go
package keys

type Pattern struct {
	EnvVar   string
	Provider string
}

var BuiltinPatterns = []Pattern{
	{"OPENAI_API_KEY", "OpenAI"},
	{"ANTHROPIC_API_KEY", "Anthropic"},
	{"OPENROUTER_API_KEY", "OpenRouter"},
	{"OPENCODE_API_KEY", "OpenCode"},
	{"GOOGLE_API_KEY", "Google AI"},
	{"GEMINI_API_KEY", "Google AI"},
	{"GROQ_API_KEY", "Groq"},
	{"TOGETHER_API_KEY", "Together AI"},
	{"DEEPSEEK_API_KEY", "DeepSeek"},
	{"COHERE_API_KEY", "Cohere"},
	{"REPLICATE_API_KEY", "Replicate"},
	{"HF_API_KEY", "HuggingFace"},
	{"HUGGINGFACE_API_KEY", "HuggingFace"},
	{"MISTRAL_API_KEY", "Mistral"},
	{"PERPLEXITY_API_KEY", "Perplexity"},
	{"BLACKBOX_AI_KEY", "Blackbox AI"},
	{"XAI_API_KEY", "xAI"},
	{"FIREWORKS_API_KEY", "Fireworks"},
	{"GITHUB_TOKEN", "GitHub"},
	{"GH_TOKEN", "GitHub"},
}

func FilterByEnvVar(envVar string) []Pattern {
	if envVar == "" {
		return BuiltinPatterns
	}
	for _, p := range BuiltinPatterns {
		if p.EnvVar == envVar {
			return []Pattern{p}
		}
	}
	return nil
}
```

- [ ] **Step 2: Verify build and commit**

Run:
```bash
go build ./...
```

```bash
git add internal/keys/patterns.go && git commit -m "feat: add built-in key pattern registry"
```

---

### Task 3: Implement table output

**Files:**
- Modify: `internal/output/table.go`

- [ ] **Step 1: Write table output using go-pretty**

Write to `internal/output/table.go`:

```go
package output

import (
	"os"

	"github.com/krzysztofciepka/key-scanner/internal/scanner"
	"github.com/jedib0t/go-pretty/v6/table"
)

func PrintTable(results []scanner.Result) {
	if len(results) == 0 {
		os.Stdout.WriteString("No exposed keys found.\n")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"KEY", "REPO", "FILE", "COMMIT DATE", "VALUE"})

	for _, r := range results {
		value := r.Value
		if len(value) > 30 {
			value = value[:27] + "..."
		}
		t.AppendRow(table.Row{r.Key, r.Repo, r.File, r.CommitDate, value})
	}

	t.SetStyle(table.StyleLight)
	t.Render()
}
```

- [ ] **Step 2: Verify build**

Run:
```bash
go build ./...
```

Expected: no errors (depends on `scanner.Result` type from Task 4 — may fail until Task 4 is done; if so, this is fine, proceed).

- [ ] **Step 3: Commit**

```bash
git add internal/output/table.go && git commit -m "feat: add table output formatter"
```

---

### Task 4: Implement GitHub scanner

**Files:**
- Modify: `internal/scanner/scanner.go`

- [ ] **Step 1: Write scanner with search and commit date fetching**

Write to `internal/scanner/scanner.go`:

```go
package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SearchItem struct {
	Repo     string
	Path     string
	Fragment string
}

type Result struct {
	Key        string
	Repo       string
	File       string
	CommitDate string
	Value      string
}

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		token:      token,
	}
}

func (c *Client) Search(ctx context.Context, query string, limit int) ([]SearchItem, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 100 {
		limit = 100
	}

	apiURL := fmt.Sprintf(
		"https://api.github.com/search/code?q=%s&per_page=%d",
		url.QueryEscape(query), limit,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github.v3.text-match+json")
	req.Header.Set("User-Agent", "key-scanner")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 429 {
		return nil, fmt.Errorf("rate limited (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid GitHub token")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("search failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			Path       string `json:"path"`
			Repository struct {
				FullName string `json:"full_name"`
			} `json:"repository"`
			TextMatches []struct {
				Fragment string `json:"fragment"`
			} `json:"text_matches"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var items []SearchItem
	for _, item := range result.Items {
		fragment := ""
		if len(item.TextMatches) > 0 {
			fragment = item.TextMatches[0].Fragment
		}
		items = append(items, SearchItem{
			Repo:     item.Repository.FullName,
			Path:     item.Path,
			Fragment: fragment,
		})
	}

	return items, nil
}

func (c *Client) GetCommitDate(ctx context.Context, repo, path string) (string, error) {
	apiURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/commits?path=%s&per_page=1",
		repo, url.QueryEscape(path),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "key-scanner")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("commits request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "unknown", nil
	}

	var commits []struct {
		Commit struct {
			Committer struct {
				Date string `json:"date"`
			} `json:"committer"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return "unknown", nil
	}

	if len(commits) == 0 {
		return "unknown", nil
	}

	date := commits[0].Commit.Committer.Date
	if len(date) >= 10 {
		return date[:10], nil
	}
	return date, nil
}

func ExtractValue(fragment, envVar string) string {
	patterns := []string{
		envVar + "=",
		envVar + " =",
		envVar + " :",
		envVar + ":",
	}

	for _, prefix := range patterns {
		idx := strings.Index(fragment, prefix)
		if idx == -1 {
			continue
		}
		value := fragment[idx+len(prefix):]
		value = strings.TrimLeft(value, " ")
		value = strings.TrimLeft(value, ":")
		value = strings.TrimLeft(value, " ")

		if newlineIdx := strings.IndexAny(value, "\r\n"); newlineIdx != -1 {
			value = value[:newlineIdx]
		}
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if value != "" {
			return value
		}
	}

	return ""
}
```

- [ ] **Step 2: Verify build**

Run:
```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/scanner/scanner.go && git commit -m "feat: implement GitHub code search scanner"
```

---

### Task 5: Implement scan subcommand

**Files:**
- Modify: `cmd/scan.go`

- [ ] **Step 1: Write scan command**

Write to `cmd/scan.go`:

```go
package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/krzysztofciepka/key-scanner/internal/keys"
	"github.com/krzysztofciepka/key-scanner/internal/output"
	"github.com/krzysztofciepka/key-scanner/internal/scanner"
)

func RunScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	limit := fs.Int("limit", 100, "max results per pattern")
	patternFilter := fs.String("pattern", "", "search only a specific key pattern")
	fs.Parse(args)

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable not set")
	}

	patterns := keys.FilterByEnvVar(*patternFilter)
	if len(patterns) == 0 {
		return fmt.Errorf("unknown key pattern: %s", *patternFilter)
	}

	client := scanner.NewClient(token)
	ctx := context.Background()

	seen := make(map[string]struct{})
	var results []scanner.Result

	for _, pattern := range patterns {
		query := fmt.Sprintf(`"%s="`, pattern.EnvVar)
		fmt.Fprintf(os.Stderr, "Searching for %s...\n", pattern.EnvVar)

		items, err := client.Search(ctx, query, *limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  warning: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, item := range items {
			value := scanner.ExtractValue(item.Fragment, pattern.EnvVar)
			if value == "" {
				continue
			}

			dedupKey := fmt.Sprintf("%s:%s:%s", value, item.Repo, item.Path)
			if _, ok := seen[dedupKey]; ok {
				continue
			}
			seen[dedupKey] = struct{}{}

			fmt.Fprintf(os.Stderr, "  found in %s/%s\n", item.Repo, item.Path)

			commitDate, err := client.GetCommitDate(ctx, item.Repo, item.Path)
			if err != nil {
				commitDate = "unknown"
			}

			results = append(results, scanner.Result{
				Key:        pattern.EnvVar,
				Repo:       item.Repo,
				File:       item.Path,
				CommitDate: commitDate,
				Value:      value,
			})

			time.Sleep(500 * time.Millisecond)
		}

		time.Sleep(2 * time.Second)
	}

	output.PrintTable(results)
	return nil
}
```

- [ ] **Step 2: Verify build**

Run:
```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add cmd/scan.go && git commit -m "feat: implement scan subcommand"
```

---

### Task 6: Implement key validation and provider detection

**Files:**
- Modify: `internal/keys/validate.go`

- [ ] **Step 1: Write provider detection and validation**

Write to `internal/keys/validate.go`:

```go
package keys

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

type providerValidator struct {
	name     string
	patterns []*regexp.Regexp
	endpoint string
	header   string
}

var validators = []providerValidator{
	{
		name:     "OpenAI",
		patterns: []*regexp.Regexp{regexp.MustCompile(`^sk-(proj|admin|svcacct)-`)},
		endpoint: "https://api.openai.com/v1/models",
		header:   "Bearer",
	},
	{
		name:     "Anthropic",
		patterns: []*regexp.Regexp{regexp.MustCompile(`^sk-ant-api\d+-`)},
		endpoint: "https://api.anthropic.com/v1/models",
		header:   "x-api-key",
	},
	{
		name:     "OpenRouter",
		patterns: []*regexp.Regexp{regexp.MustCompile(`^sk-or-v[12]-`)},
		endpoint: "https://openrouter.ai/api/v1/models",
		header:   "Bearer",
	},
	{
		name:     "GitHub",
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^ghp_`),
			regexp.MustCompile(`^github_pat_`),
			regexp.MustCompile(`^gho_`),
			regexp.MustCompile(`^ghu_`),
			regexp.MustCompile(`^ghs_`),
			regexp.MustCompile(`^ghr_`),
		},
		endpoint: "https://api.github.com/user",
		header:   "Bearer",
	},
	{
		name:     "HuggingFace",
		patterns: []*regexp.Regexp{regexp.MustCompile(`^hf_`)},
		endpoint: "https://huggingface.co/api/whoami-v2",
		header:   "Bearer",
	},
	{
		name:     "Google AI",
		patterns: []*regexp.Regexp{regexp.MustCompile(`^AIza`)},
		endpoint: "https://generativelanguage.googleapis.com/v1beta/models?key=",
		header:   "",
	},
	{
		name:     "Groq",
		patterns: []*regexp.Regexp{regexp.MustCompile(`^gsk_`)},
		endpoint: "https://api.groq.com/openai/v1/models",
		header:   "Bearer",
	},
	{
		name:     "Together AI",
		patterns: []*regexp.Regexp{},
		endpoint: "https://api.together.xyz/v1/models",
		header:   "Bearer",
	},
	{
		name:     "DeepSeek",
		patterns: []*regexp.Regexp{},
		endpoint: "https://api.deepseek.com/user/balance",
		header:   "Bearer",
	},
	{
		name:     "Cohere",
		patterns: []*regexp.Regexp{},
		endpoint: "https://api.cohere.ai/v1/check-api-key",
		header:   "Bearer",
	},
	{
		name:     "Mistral",
		patterns: []*regexp.Regexp{},
		endpoint: "https://api.mistral.ai/v1/models",
		header:   "Bearer",
	},
	{
		name:     "Replicate",
		patterns: []*regexp.Regexp{regexp.MustCompile(`^r8_`)},
		endpoint: "https://api.replicate.com/v1/account",
		header:   "Bearer",
	},
	{
		name:     "Perplexity",
		patterns: []*regexp.Regexp{regexp.MustCompile(`^pplx-`)},
		endpoint: "https://api.perplexity.ai/auth/status",
		header:   "Bearer",
	},
	{
		name:     "xAI",
		patterns: []*regexp.Regexp{regexp.MustCompile(`^xai-`)},
		endpoint: "https://api.x.ai/v1/models",
		header:   "Bearer",
	},
	{
		name:     "Fireworks",
		patterns: []*regexp.Regexp{},
		endpoint: "https://api.fireworks.ai/v1/account",
		header:   "Bearer",
	},
	{
		name:     "Blackbox AI",
		patterns: []*regexp.Regexp{},
		endpoint: "",
		header:   "",
	},
	{
		name:     "OpenCode",
		patterns: []*regexp.Regexp{},
		endpoint: "",
		header:   "",
	},
}

func DetectProvider(value string) string {
	for _, v := range validators {
		for _, p := range v.patterns {
			if p.MatchString(value) {
				return v.name
			}
		}
	}
	return ""
}

func LookupProvider(name string) *providerValidator {
	for _, v := range validators {
		if v.name == name {
			return &v
		}
	}
	return nil
}

func ValidateKey(ctx context.Context, provider, value string) (bool, error) {
	pv := LookupProvider(provider)
	if pv == nil {
		return false, fmt.Errorf("unknown provider: %s", provider)
	}

	client := &http.Client{Timeout: 15 * time.Second}

	if pv.endpoint == "" {
		return false, fmt.Errorf("no validation endpoint available for provider: %s", provider)
	}

	endpoint := pv.endpoint
	method := "GET"

	if provider == "Google AI" {
		endpoint = endpoint + value
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	switch pv.header {
	case "Bearer":
		req.Header.Set("Authorization", "Bearer "+value)
	case "x-api-key":
		req.Header.Set("x-api-key", value)
	}

	req.Header.Set("User-Agent", "key-scanner")

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true, nil
	}
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return false, nil
	}

	return false, fmt.Errorf("unexpected status %d", resp.StatusCode)
}
```

- [ ] **Step 2: Verify build**

Run:
```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/keys/validate.go && git commit -m "feat: implement key validation and provider detection"
```

---

### Task 7: Implement test subcommand

**Files:**
- Modify: `cmd/test.go`

- [ ] **Step 1: Write test command**

Write to `cmd/test.go`:

```go
package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/krzysztofciepka/key-scanner/internal/keys"
)

func RunTest(args []string) error {
	fs := flag.NewFlagSet("test", flag.ExitOnError)
	provider := fs.String("provider", "", "force a specific provider for validation")
	fs.Parse(args)

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: key-scanner test [--provider NAME] <key-value>")
	}

	value := fs.Arg(0)
	ctx := context.Background()

	detectedProvider := keys.DetectProvider(value)
	useProvider := *provider
	if useProvider == "" {
		if detectedProvider == "" {
			return fmt.Errorf("could not auto-detect provider from key format; use --provider to specify one")
		}
		useProvider = detectedProvider
	}

	fmt.Fprintf(os.Stderr, "Testing key against %s...\n", useProvider)

	valid, err := keys.ValidateKey(ctx, useProvider, value)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if valid {
		fmt.Printf("Valid %s (%s)\n", "\u2713", useProvider)
	} else {
		fmt.Printf("Invalid %s (%s)\n", "\u2717", useProvider)
	}

	return nil
}
```

- [ ] **Step 2: Verify build**

Run:
```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add cmd/test.go && git commit -m "feat: implement test subcommand"
```

---

### Task 8: Wire main.go entry point

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Write main routing**

Write to `main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/krzysztofciepka/key-scanner/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: key-scanner <scan|test> [flags]\n")
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  scan   Search GitHub for exposed API keys\n")
		fmt.Fprintf(os.Stderr, "  test   Validate a key against its provider\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		fmt.Fprintf(os.Stderr, "  scan:\n")
		fmt.Fprintf(os.Stderr, "    --limit N      Max results per pattern (default: 100)\n")
		fmt.Fprintf(os.Stderr, "    --pattern NAME Search only a specific key pattern\n")
		fmt.Fprintf(os.Stderr, "  test:\n")
		fmt.Fprintf(os.Stderr, "    --provider NAME  Force a specific provider\n")
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "scan":
		err = cmd.RunScan(os.Args[2:])
	case "test":
		err = cmd.RunTest(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		fmt.Fprintf(os.Stderr, "Usage: key-scanner <scan|test>\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Verify build and run**

Run:
```bash
go build -o key-scanner .
```

Expected: binary produced, no errors.

Run:
```bash
./key-scanner
```

Expected: usage message printed.

- [ ] **Step 3: Commit**

```bash
git add main.go && git commit -m "feat: wire main entry point routing scan and test"
```

---

### Task 9: Add tests

**Files:**
- Create: `internal/keys/patterns_test.go`
- Create: `internal/keys/validate_test.go`
- Create: `internal/scanner/scanner_test.go`
- Create: `internal/output/table_test.go`

- [ ] **Step 1: Write patterns tests**

Write to `internal/keys/patterns_test.go`:

```go
package keys

import (
	"testing"
)

func TestBuiltinPatternsNotEmpty(t *testing.T) {
	if len(BuiltinPatterns) == 0 {
		t.Fatal("BuiltinPatterns is empty")
	}
}

func TestFilterByEnvVarExactMatch(t *testing.T) {
	result := FilterByEnvVar("OPENAI_API_KEY")
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Provider != "OpenAI" {
		t.Errorf("expected provider OpenAI, got %s", result[0].Provider)
	}
}

func TestFilterByEnvVarEmptyReturnsAll(t *testing.T) {
	result := FilterByEnvVar("")
	if len(result) != len(BuiltinPatterns) {
		t.Errorf("expected %d results, got %d", len(BuiltinPatterns), len(result))
	}
}

func TestFilterByEnvVarUnknownReturnsNil(t *testing.T) {
	result := FilterByEnvVar("NONEXISTENT_KEY")
	if result != nil {
		t.Error("expected nil for unknown key")
	}
}

func TestNoDuplicateEnvVars(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range BuiltinPatterns {
		if seen[p.EnvVar] {
			t.Errorf("duplicate env var: %s", p.EnvVar)
		}
		seen[p.EnvVar] = true
	}
}
```

- [ ] **Step 2: Run patterns tests**

Run:
```bash
go test ./internal/keys/ -run TestBuiltinPatterns -v
go test ./internal/keys/ -run TestFilter -v
go test ./internal/keys/ -run TestNoDuplicate -v
```

Expected: all PASS.

- [ ] **Step 3: Write validate tests**

Write to `internal/keys/validate_test.go`:

```go
package keys

import (
	"testing"
)

func TestDetectProviderOpenAI(t *testing.T) {
	cases := []string{
		"sk-proj-abc123xyz",
		"sk-admin-abc123xyz",
		"sk-svcacct-abc123xyz",
	}
	for _, key := range cases {
		result := DetectProvider(key)
		if result != "OpenAI" {
			t.Errorf("DetectProvider(%q) = %q, want OpenAI", key, result)
		}
	}
}

func TestDetectProviderAnthropic(t *testing.T) {
	result := DetectProvider("sk-ant-api03-abc123xyz")
	if result != "Anthropic" {
		t.Errorf("DetectProvider = %q, want Anthropic", result)
	}
}

func TestDetectProviderOpenRouter(t *testing.T) {
	result := DetectProvider("sk-or-v1-abc123xyz")
	if result != "OpenRouter" {
		t.Errorf("DetectProvider = %q, want OpenRouter", result)
	}
}

func TestDetectProviderGitHub(t *testing.T) {
	cases := []string{"ghp_abc123", "github_pat_abc123", "gho_abc", "ghu_abc", "ghs_abc", "ghr_abc"}
	for _, key := range cases {
		result := DetectProvider(key)
		if result != "GitHub" {
			t.Errorf("DetectProvider(%q) = %q, want GitHub", key, result)
		}
	}
}

func TestDetectProviderHuggingFace(t *testing.T) {
	result := DetectProvider("hf_abc123xyz")
	if result != "HuggingFace" {
		t.Errorf("DetectProvider = %q, want HuggingFace", result)
	}
}

func TestDetectProviderGoogleAI(t *testing.T) {
	result := DetectProvider("AIzaSyAbc123xyz")
	if result != "Google AI" {
		t.Errorf("DetectProvider = %q, want Google AI", result)
	}
}

func TestDetectProviderUnknown(t *testing.T) {
	result := DetectProvider("totally-random-string-12345")
	if result != "" {
		t.Errorf("DetectProvider = %q, want empty string", result)
	}
}

func TestLookupProviderKnown(t *testing.T) {
	pv := LookupProvider("OpenAI")
	if pv == nil {
		t.Fatal("LookupProvider(OpenAI) returned nil")
	}
	if pv.name != "OpenAI" {
		t.Errorf("name = %q, want OpenAI", pv.name)
	}
}

func TestLookupProviderUnknown(t *testing.T) {
	pv := LookupProvider("NonExistent")
	if pv != nil {
		t.Error("LookupProvider for unknown provider should return nil")
	}
}
```

- [ ] **Step 4: Run validate tests**

Run:
```bash
go test ./internal/keys/ -run TestDetect -v
go test ./internal/keys/ -run TestLookup -v
```

Expected: all PASS.

- [ ] **Step 5: Write scanner unit tests**

Write to `internal/scanner/scanner_test.go`:

```go
package scanner

import (
	"testing"
)

func TestExtractValueBasic(t *testing.T) {
	fragment := "DB_HOST=localhost\nOPENAI_API_KEY=sk-proj-abc123xyz\nDB_PORT=5432"
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "sk-proj-abc123xyz" {
		t.Errorf("ExtractValue = %q, want sk-proj-abc123xyz", result)
	}
}

func TestExtractValueWithSpaceBeforeEquals(t *testing.T) {
	fragment := "OPENAI_API_KEY =sk-proj-abc123xyz"
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "sk-proj-abc123xyz" {
		t.Errorf("ExtractValue = %q, want sk-proj-abc123xyz", result)
	}
}

func TestExtractValueWithQuotes(t *testing.T) {
	fragment := `OPENAI_API_KEY="sk-proj-abc123xyz"`
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "sk-proj-abc123xyz" {
		t.Errorf("ExtractValue = %q, want sk-proj-abc123xyz", result)
	}
}

func TestExtractValueWithSingleQuotes(t *testing.T) {
	fragment := "OPENAI_API_KEY='sk-proj-abc123xyz'"
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "sk-proj-abc123xyz" {
		t.Errorf("ExtractValue = %q, want sk-proj-abc123xyz", result)
	}
}

func TestExtractValueWithColon(t *testing.T) {
	fragment := "OPENAI_API_KEY: sk-proj-abc123xyz"
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "sk-proj-abc123xyz" {
		t.Errorf("ExtractValue = %q, want sk-proj-abc123xyz", result)
	}
}

func TestExtractValueNotFound(t *testing.T) {
	fragment := "DB_HOST=localhost\nDB_PORT=5432"
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "" {
		t.Errorf("ExtractValue = %q, want empty string", result)
	}
}

func TestExtractValueNewline(t *testing.T) {
	fragment := "OPENAI_API_KEY=sk-proj-abc123xyz\nNEXT_LINE=something"
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "sk-proj-abc123xyz" {
		t.Errorf("ExtractValue = %q, want sk-proj-abc123xyz", result)
	}
}

func TestExtractValueCRLF(t *testing.T) {
	fragment := "OPENAI_API_KEY=sk-proj-abc123xyz\r\nNEXT_LINE=something"
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "sk-proj-abc123xyz" {
		t.Errorf("ExtractValue = %q, want sk-proj-abc123xyz", result)
	}
}
```

- [ ] **Step 6: Run scanner tests**

Run:
```bash
go test ./internal/scanner/ -v
```

Expected: all PASS.

- [ ] **Step 7: Write output tests**

Write to `internal/output/table_test.go`:

```go
package output

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/krzysztofciepka/key-scanner/internal/scanner"
)

func captureOutput(fn func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = stdout
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestPrintTableEmpty(t *testing.T) {
	output := captureOutput(func() {
		PrintTable(nil)
	})
	if !strings.Contains(output, "No exposed keys found") {
		t.Errorf("expected 'No exposed keys found', got: %s", output)
	}
}

func TestPrintTableWithResults(t *testing.T) {
	results := []scanner.Result{
		{
			Key:        "OPENAI_API_KEY",
			Repo:       "user/repo",
			File:       ".env",
			CommitDate: "2025-07-14",
			Value:      "sk-proj-abc123xyz",
		},
	}

	output := captureOutput(func() {
		PrintTable(results)
	})

	if !strings.Contains(output, "OPENAI_API_KEY") {
		t.Errorf("output should contain OPENAI_API_KEY, got: %s", output)
	}
	if !strings.Contains(output, "user/repo") {
		t.Errorf("output should contain user/repo, got: %s", output)
	}
}

func TestPrintTableTruncatesLongValue(t *testing.T) {
	results := []scanner.Result{
		{
			Key:        "OPENAI_API_KEY",
			Repo:       "user/repo",
			File:       ".env",
			CommitDate: "2025-07-14",
			Value:      "sk-proj-this-is-a-very-long-key-value-that-should-be-truncated",
		},
	}

	output := captureOutput(func() {
		PrintTable(results)
	})

	if !strings.Contains(output, "...") {
		t.Errorf("long value should be truncated with ..., got: %s", output)
	}
}
```

- [ ] **Step 8: Run output tests**

Run:
```bash
go test ./internal/output/ -v
```

Expected: all PASS.

- [ ] **Step 9: Run all tests**

Run:
```bash
go test ./... -v
```

Expected: all tests PASS.

- [ ] **Step 10: Commit**

```bash
git add internal/keys/patterns_test.go internal/keys/validate_test.go internal/scanner/scanner_test.go internal/output/table_test.go && git commit -m "test: add unit tests for all packages"
```

---

### Task 10: CI and README

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.gitignore`
- Create: `README.md`

- [ ] **Step 1: Create .gitignore**

Write to `.gitignore`:

```
/key-scanner
*.exe
*.test
*.out
```

- [ ] **Step 2: Create CI workflow**

```bash
mkdir -p .github/workflows
```

Write to `.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [master, main]
  pull_request:
    branches: [master, main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"
      - name: Build
        run: go build ./...
      - name: Test
        run: go test ./... -v
      - name: Vet
        run: go vet ./...
```

- [ ] **Step 3: Create README**

Write to `README.md`:

````markdown
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
````

- [ ] **Step 4: Verify everything**

Run:
```bash
go build -o key-scanner . && go test ./... -v && go vet ./...
```

Expected: build succeeds, all tests pass, no vet warnings.

- [ ] **Step 5: Commit**

```bash
git add .github .gitignore README.md && git commit -m "chore: add CI workflow, gitignore, and README"
```
````
