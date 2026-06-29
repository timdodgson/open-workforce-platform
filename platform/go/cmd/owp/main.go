package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "optimise":
		runOptimise()
	case "benchmark":
		runBenchmark()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  owp optimise <dataset-path> [--algorithm constructive|hill-climbing|simulated-annealing] [--weights default]")
	fmt.Fprintln(os.Stderr, "  owp benchmark <datasets-directory>")
}

func runOptimise() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	path := os.Args[2]
	algorithm := parseAlgorithm(os.Args[3:])
	weightsProfile := parseWeights(os.Args[3:])

	// Validate weights profile.
	if _, ok := optimisation.GetWeightProfile(weightsProfile); !ok {
		fmt.Fprintf(os.Stderr, "Unknown weights profile: %s\n", weightsProfile)
		os.Exit(1)
	}

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
		for _, item := range result.UnassignedDetails() {
			fmt.Printf("    %s\n", item.WorkItemID)
			if len(item.Reasons) > 0 {
				fmt.Println("      Reasons:")
				for _, reason := range item.Reasons {
					fmt.Printf("        - %s\n", reason)
				}
			}
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

	// Statistics.
	s := result.Statistics()
	fmt.Println("Optimisation Statistics:")
	fmt.Printf("  Algorithm: %s\n", s.Algorithm)
	fmt.Printf("  Duration: %dms\n", s.DurationMs)
	fmt.Printf("  Iterations: %d\n", s.Iterations)
	fmt.Printf("  Candidates Evaluated: %d\n", s.CandidatesEvaluated)
	fmt.Printf("  Improvements Accepted: %d\n", s.ImprovementsAccepted)
	fmt.Printf("  Final Objective Score: %d\n", s.FinalObjectiveScore)
	fmt.Println()

	fmt.Println("Done.")
}

func runBenchmark() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	dir := os.Args[2]
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	// Discover dataset files.
	var datasetFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			datasetFiles = append(datasetFiles, e.Name())
		}
	}

	if len(datasetFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No dataset files found in directory")
		os.Exit(1)
	}

	// Get available algorithms.
	algs := optimisation.Available()
	sort.Strings(algs)

	// Print header.
	fmt.Printf("%-28s %-24s %7s %11s %10s %10s %12s\n",
		"Dataset", "Algorithm", "Score", "Objective", "Assigned", "Duration", "Candidates")
	fmt.Println(strings.Repeat("-", 105))

	// Run each combination.
	for _, file := range datasetFiles {
		path := filepath.Join(dir, file)
		dataset, err := loader.LoadDataset(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR loading %s: %v\n", file, err)
			continue
		}

		travel := convertTravel(dataset.TravelMatrix)
		name := strings.TrimSuffix(file, ".json")

		for _, alg := range algs {
			result, err := application.Optimise(dataset.Events, dataset.Resources, travel, alg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ERROR %s/%s: %v\n", name, alg, err)
				continue
			}

			stats := result.Statistics()
			fmt.Printf("%-28s %-24s %7d %11d %10d %8dms %12d\n",
				name, alg, result.Score(), result.ObjectiveScore(),
				result.Size(), stats.DurationMs, stats.CandidatesEvaluated)
		}
	}
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

// parseWeights reads the --weights flag from remaining args.
// Defaults to "default".
func parseWeights(args []string) string {
	for i, arg := range args {
		if arg == "--weights" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, "--weights=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, "--weights="))
		}
	}
	return "default"
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
