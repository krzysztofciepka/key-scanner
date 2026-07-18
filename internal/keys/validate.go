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
		patterns: []*regexp.Regexp{regexp.MustCompile(`^sk[-_]`)},
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
		patterns: []*regexp.Regexp{regexp.MustCompile(`^bb_`)},
		endpoint: "https://api.blackbox.ai/v1/models",
		header:   "Bearer",
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

func LookupProvider(name string) (*providerValidator, bool) {
	for i := range validators {
		if validators[i].name == name {
			v := validators[i]
			return &v, true
		}
	}
	return nil, false
}

func ValidateKey(ctx context.Context, provider, value string) (bool, error) {
	pv, ok := LookupProvider(provider)
	if !ok {
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
