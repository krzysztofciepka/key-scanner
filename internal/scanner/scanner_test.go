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

func TestExtractValueEmptyValue(t *testing.T) {
	fragment := "OPENAI_API_KEY="
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "" {
		t.Errorf("ExtractValue = %q, want empty string", result)
	}
}

func TestExtractValuePlaceholder(t *testing.T) {
	fragment := "OPENAI_API_KEY=your_api_key_here"
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "" {
		t.Errorf("ExtractValue = %q, want empty string for placeholder", result)
	}
}

func TestExtractValueTemplatePlaceholder(t *testing.T) {
	fragment := "OPENAI_API_KEY=<your_openai_api_key>"
	result := ExtractValue(fragment, "OPENAI_API_KEY")
	if result != "" {
		t.Errorf("ExtractValue = %q, want empty string for template placeholder", result)
	}
}

func TestIsPlaceholderTrue(t *testing.T) {
	cases := []string{
		"your_api_key",
		"YOUR-API-KEY",
		"your_key_here_123",
		"<your_secret>",
		"placeholder",
		"changeme",
		"change_this",
		"",
	}
	for _, v := range cases {
		if !IsPlaceholder(v) {
			t.Errorf("IsPlaceholder(%q) = false, want true", v)
		}
	}
}

func TestIsPlaceholderFalse(t *testing.T) {
	cases := []string{
		"sk-proj-abc123xyz",
		"sk-ant-api03-abc123",
		"ghp_abc123def456",
		"hf_abc123xyz",
		"bb_abc123def456",
	}
	for _, v := range cases {
		if IsPlaceholder(v) {
			t.Errorf("IsPlaceholder(%q) = true, want false", v)
		}
	}
}
