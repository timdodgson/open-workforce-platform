package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
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

	// Build capacity lookup for display.
	capacityOf := buildCapacityLookup(dataset.Resources)

	fmt.Println("=== Optimised Plan ===")
	fmt.Println()
	fmt.Printf("Score: %d\n", result.Score())
	fmt.Println()
	fmt.Printf("Resources: %d\n", len(dataset.Resources))
	fmt.Printf("Capacity:  %d\n", result.TotalCapacity())
	fmt.Println()
	fmt.Println("Assignments:")
	fmt.Println()

	// Group assignments by resource, preserving order.
	type group struct {
		items []string
	}
	groups := make(map[string]*group)
	var order []string

	for _, a := range result.Assignments() {
		g, exists := groups[a.ResourceID()]
		if !exists {
			g = &group{}
			groups[a.ResourceID()] = g
			order = append(order, a.ResourceID())
		}
		g.items = append(g.items, a.WorkItemID())
	}

	for _, resID := range order {
		g := groups[resID]
		fmt.Printf("  %s (%d/%d)\n", resID, len(g.items), capacityOf[resID])
		for _, itemID := range g.items {
			fmt.Printf("    - %s\n", itemID)
		}
		fmt.Println()
	}

	if result.UnassignedCount() > 0 {
		fmt.Println("Unassigned:")
		fmt.Println()
		for _, id := range result.Unassigned() {
			fmt.Printf("    - %s\n", id)
		}
	} else {
		fmt.Println("Unassigned: None")
	}

	fmt.Println()
	fmt.Println("Done.")
}

// buildCapacityLookup reads capacity from each resource's details for display.
func buildCapacityLookup(resources []resource.Resource) map[string]int {
	lookup := make(map[string]int, len(resources))
	for _, res := range resources {
		var details struct {
			Capacity int `json:"capacity"`
		}
		if err := json.Unmarshal(res.Details(), &details); err == nil {
			lookup[res.ID()] = details.Capacity
		}
	}
	return lookup
}
