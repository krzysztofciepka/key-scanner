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

func TestExtractValueWithHTMLTags(t *testing.T) {
	fragment := "DB_HOST=localhost\n<em>OPENAI_API_KEY</em>=sk-proj-abc123xyz\nDB_PORT=5432"
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
