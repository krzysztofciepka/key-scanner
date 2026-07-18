package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/krzysztofciepka/key-scanner/internal/keys"
)

func RunTest(args []string) error {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	provider := fs.String("provider", "", "force a specific provider for validation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() < 1 {
		return fmt.Errorf("usage: key-scanner test [--provider NAME] <key-value>")
	}

	value := fs.Arg(0)
	ctx := context.Background()

	detectedProvider := keys.DetectProvider(value)
	useProvider := *provider
	if useProvider == "" {
		if detectedProvider == "" {
			providers := strings.Join(keys.ProviderNames(), ", ")
			return fmt.Errorf("could not auto-detect provider from key format; use --provider with one of: %s", providers)
		}
		useProvider = detectedProvider
	} else if detectedProvider != "" && detectedProvider != useProvider {
		fmt.Fprintf(os.Stderr, "Note: detected prefix matches %s, but using --provider %s\n", detectedProvider, useProvider)
	}

	fmt.Fprintf(os.Stderr, "Testing key against %s...\n", useProvider)

	valid, err := keys.ValidateKey(ctx, useProvider, value)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if valid {
		fmt.Printf("Valid %s (%s)\n", "\u2713", useProvider)
	} else {
		fmt.Printf("Invalid %s (%s)\n", "\u2717", useProvider)
	}

	return nil
}
