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
