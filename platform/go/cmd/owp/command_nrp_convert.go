package main

import (
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/nrp"
)

func runConvertNRP() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: owp convert-nrp <nrp-input> <output-dataset>")
		os.Exit(1)
	}

	inputPath := os.Args[2]
	outputPath := os.Args[3]

	input, err := nrp.LoadNRP(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading NRP file: %v\n", err)
		os.Exit(1)
	}

	dataset := nrp.Convert(input)
	if err := nrp.WriteDataset(dataset, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing dataset: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Converted %d nurses and %d shift demands into OWP dataset.\n", len(input.Nurses), len(dataset.BusinessEvents))
	fmt.Printf("Output: %s\n", outputPath)
}
