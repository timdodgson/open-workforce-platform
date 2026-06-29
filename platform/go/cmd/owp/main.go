package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "optimise" {
		fmt.Fprintln(os.Stderr, "Usage: owp optimise <dataset-path> [--algorithm constructive|hill-climbing]")
		os.Exit(1)
	}

	path := os.Args[2]
	algorithm := parseAlgorithm(os.Args[3:])

	dataset, err := loader.LoadDataset(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading dataset: %v\n", err)
		os.Exit(1)
	}

	result, err := application.Optimise(dataset.Events, dataset.Resources, convertTravel(dataset.TravelMatrix), algorithm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during optimisation: %v\n", err)
		os.Exit(1)
	}

	// Build lookups for display.
	capacityOf := buildCapacityLookup(dataset.Resources)
	durationOf := buildDurationLookup(dataset.Events)

	fmt.Println("=== Optimised Plan ===")
	fmt.Println()
	fmt.Printf("Algorithm: %s\n", algorithm)
	fmt.Printf("Assignment Score: %d\n", result.Score())
	fmt.Printf("Objective Score:  %d\n", result.ObjectiveScore())
	fmt.Println()
	fmt.Println("Objective Breakdown:")
	for _, entry := range result.ObjectiveBreakdown() {
		fmt.Printf("  %s: %d\n", entry.Name, entry.Score)
	}
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
		usedMins := 0
		for _, itemID := range g.items {
			usedMins += durationOf[itemID]
		}
		fmt.Printf("  %s\n", resID)
		fmt.Printf("    Used: %d / %d mins\n", usedMins, capacityOf[resID])
		fmt.Println("    Work Items:")
		for _, itemID := range g.items {
			fmt.Printf("      - %s\n", itemID)
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

	// Travel breakdown.
	travelLookup := buildTravelDisplayLookup(dataset.TravelMatrix)
	resourceLocations := buildResourceLocationLookup(dataset.Resources)
	itemLocations := buildItemLocationLookup(dataset.Events)

	fmt.Println("Travel:")
	fmt.Println()
	for _, resID := range order {
		g := groups[resID]
		current := resourceLocations[resID]
		total := 0
		var legs []string

		for _, itemID := range g.items {
			dest := itemLocations[itemID]
			if dest != "" && current != "" && dest != current {
				mins := travelLookup[current+"|"+dest]
				if mins > 0 {
					legs = append(legs, fmt.Sprintf("    %s -> %s: %d mins", current, dest, mins))
					total += mins
				}
			}
			if dest != "" {
				current = dest
			}
		}

		fmt.Printf("  %s\n", resID)
		if len(legs) > 0 {
			for _, leg := range legs {
				fmt.Println(leg)
			}
		}
		fmt.Printf("    Total: %d mins\n", total)
		fmt.Println()
	}
	fmt.Println("Done.")
}

// parseAlgorithm reads the --algorithm flag from remaining args.
// Defaults to "constructive".
func parseAlgorithm(args []string) string {
	for i, arg := range args {
		if arg == "--algorithm" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, "--algorithm=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, "--algorithm="))
		}
	}
	return "constructive"
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

// buildDurationLookup reads duration from each event's details for display.
// Work item IDs are "WI-" + event ID.
func buildDurationLookup(events []event.BusinessEvent) map[string]int {
	lookup := make(map[string]int, len(events))
	for _, evt := range events {
		var details struct {
			Duration int `json:"duration"`
		}
		if err := json.Unmarshal(evt.Details(), &details); err == nil {
			dur := details.Duration
			if dur <= 0 {
				dur = 1
			}
			lookup["WI-"+evt.ID()] = dur
		}
	}
	return lookup
}

// convertTravel converts loader travel entries to optimisation travel entries.
func convertTravel(entries []loader.TravelEntry) []optimisation.TravelEntry {
	result := make([]optimisation.TravelEntry, len(entries))
	for i, e := range entries {
		result[i] = optimisation.TravelEntry{From: e.From, To: e.To, Minutes: e.Minutes}
	}
	return result
}

// buildTravelDisplayLookup creates a map for travel time display.
func buildTravelDisplayLookup(entries []loader.TravelEntry) map[string]int {
	lookup := make(map[string]int, len(entries))
	for _, e := range entries {
		lookup[e.From+"|"+e.To] = e.Minutes
	}
	return lookup
}

// buildResourceLocationLookup reads starting location from each resource's details.
func buildResourceLocationLookup(resources []resource.Resource) map[string]string {
	lookup := make(map[string]string, len(resources))
	for _, res := range resources {
		var details struct {
			Location string `json:"location"`
		}
		if err := json.Unmarshal(res.Details(), &details); err == nil {
			lookup[res.ID()] = details.Location
		}
	}
	return lookup
}

// buildItemLocationLookup reads location from each event's details.
// Work item IDs are "WI-" + event ID.
func buildItemLocationLookup(events []event.BusinessEvent) map[string]string {
	lookup := make(map[string]string, len(events))
	for _, evt := range events {
		var details struct {
			Location string `json:"location"`
		}
		if err := json.Unmarshal(evt.Details(), &details); err == nil {
			lookup["WI-"+evt.ID()] = details.Location
		}
	}
	return lookup
}
