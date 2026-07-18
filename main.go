package main

import (
	"fmt"
	"os"

	"github.com/krzysztofciepka/key-scanner/cmd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: key-scanner <scan|test> [flags]\n")
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  scan   Search GitHub for exposed API keys\n")
		fmt.Fprintf(os.Stderr, "  test   Validate a key against its provider\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		fmt.Fprintf(os.Stderr, "  scan:\n")
		fmt.Fprintf(os.Stderr, "    --limit N      Max results per pattern (default: 100)\n")
		fmt.Fprintf(os.Stderr, "    --pattern NAME Search only a specific key pattern\n")
		fmt.Fprintf(os.Stderr, "  test:\n")
		fmt.Fprintf(os.Stderr, "    --provider NAME  Force a specific provider\n")
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "scan":
		err = cmd.RunScan(os.Args[2:])
	case "test":
		err = cmd.RunTest(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		fmt.Fprintf(os.Stderr, "Usage: key-scanner <scan|test>\n")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
