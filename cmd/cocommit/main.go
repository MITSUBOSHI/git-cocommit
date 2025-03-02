package main

import (
	"fmt"
	"os"

	"github.com/MITSUBOSHI/cocommit/pkg/git"
)

func main() {
	// Get command line arguments
	args := os.Args[1:]

	// Execute git cocommit
	if err := git.Cocommit(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
