package output

import (
	"fmt"
	"os"
	"strings"

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

func WriteFile(path string, results []scanner.Result) error {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-24s %-30s %-40s %-12s %s\n", "KEY", "REPO", "FILE", "COMMIT DATE", "VALUE"))
	b.WriteString(strings.Repeat("-", 150) + "\n")
	for _, r := range results {
		b.WriteString(fmt.Sprintf("%-24s %-30s %-40s %-12s %s\n", r.Key, r.Repo, r.File, r.CommitDate, r.Value))
	}
	b.WriteString(fmt.Sprintf("\nTotal: %d results\n", len(results)))
	return os.WriteFile(path, []byte(b.String()), 0644)
}
