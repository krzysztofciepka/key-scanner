package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: key-scanner <scan|test> [flags]\n")
		os.Exit(1)
	}
	fmt.Println("key-scanner v0.1.0")
}
