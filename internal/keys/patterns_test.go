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
