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

	dataset, err := loader.LoadDataset(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading dataset: %v\n", err)
		os.Exit(1)
	}

	result, err := application.Optimise(dataset.Events, dataset.Resources)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during optimisation: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Optimised Plan ===")
	fmt.Printf("Assignments: %d\n\n", result.Size())

	for i, a := range result.Assignments() {
		fmt.Printf("  %d. Resource [%s] → Work Item [%s]\n", i+1, a.ResourceID(), a.WorkItemID())
	}

	fmt.Println("\nDone.")
}
