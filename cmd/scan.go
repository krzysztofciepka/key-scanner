package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/krzysztofciepka/key-scanner/internal/keys"
	"github.com/krzysztofciepka/key-scanner/internal/output"
	"github.com/krzysztofciepka/key-scanner/internal/scanner"
)

func RunScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	limit := fs.Int("limit", 100, "max results per pattern")
	patternFilter := fs.String("pattern", "", "search only a specific key pattern")
	if err := fs.Parse(args); err != nil {
		return err
	}

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
			errStr := err.Error()
			if strings.Contains(errStr, "invalid GitHub token") || strings.Contains(errStr, "401") {
				return err
			}
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
