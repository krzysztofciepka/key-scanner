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
