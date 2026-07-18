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
	pv, ok := LookupProvider("OpenAI")
	if !ok {
		t.Fatal("LookupProvider(OpenAI) returned false")
	}
	if pv == nil {
		t.Fatal("LookupProvider(OpenAI) returned nil")
	}
}

func TestLookupProviderUnknown(t *testing.T) {
	_, ok := LookupProvider("NonExistent")
	if ok {
		t.Error("LookupProvider for unknown provider should return false")
	}
}
