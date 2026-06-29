package main

import (
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "optimise" {
		fmt.Fprintln(os.Stderr, "Usage: owp optimise <dataset-path>")
		os.Exit(1)
	}

	path := os.Args[2]

	events, err := loader.LoadEvents(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading dataset: %v\n", err)
		os.Exit(1)
	}

	result, err := application.Optimise(events)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during optimisation: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Optimised Plan ===")
	fmt.Printf("Work items planned: %d\n\n", result.Size())

	for i, item := range result.Items() {
		fmt.Printf("  %d. [%s] %s\n", i+1, item.Type(), item.ID())
	}

	fmt.Println("\nDone.")
}
