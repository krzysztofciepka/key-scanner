package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/krzysztofciepka/key-scanner/internal/keys"
	"github.com/krzysztofciepka/key-scanner/internal/output"
	"github.com/krzysztofciepka/key-scanner/internal/scanner"
)

func RunScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	limit := fs.Int("limit", 10, "max results per pattern")
	patternFilter := fs.String("pattern", "", "search only a specific key pattern")
	outputPath := fs.String("output", "", "output file path (default: key-scanner-YYYY-MM-DD-HHmmss.txt)")
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

	results = topPerKey(results, *limit)

	output.PrintTable(results)

	path := *outputPath
	if path == "" {
		path = fmt.Sprintf("key-scanner-%s.txt", time.Now().Format("2006-01-02-150405"))
	}
	if err := output.WriteFile(path, results); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Saved to %s\n", path)

	return nil
}

func topPerKey(results []scanner.Result, n int) []scanner.Result {
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].CommitDate > results[j].CommitDate
	})

	grouped := make(map[string][]scanner.Result)
	var order []string
	for _, r := range results {
		if _, ok := grouped[r.Key]; !ok {
			order = append(order, r.Key)
		}
		grouped[r.Key] = append(grouped[r.Key], r)
	}

	var top []scanner.Result
	for _, key := range order {
		g := grouped[key]
		if len(g) > n {
			g = g[:n]
		}
		top = append(top, g...)
	}

	sort.SliceStable(top, func(i, j int) bool {
		return top[i].CommitDate > top[j].CommitDate
	})

	return top
}
