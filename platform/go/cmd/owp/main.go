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
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/nrp"
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
	case "convert-nrp":
		runConvertNRP()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  owp optimise <dataset-path> [--algorithm constructive|hill-climbing|simulated-annealing|tabu-search|large-neighbourhood-search] [--weights default]")
	fmt.Fprintln(os.Stderr, "  owp benchmark <datasets-directory>")
	fmt.Fprintln(os.Stderr, "  owp convert-nrp <nrp-input> <output-dataset>")
}

func runOptimise() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	path := os.Args[2]
	algorithm := parseAlgorithm(os.Args[3:])
	weightsProfile := parseWeights(os.Args[3:])
	profileName := parseProfile(os.Args[3:])

	// Validate weights profile.
	if _, ok := optimisation.GetWeightProfile(weightsProfile); !ok {
		fmt.Fprintf(os.Stderr, "Unknown weights profile: %s\n", weightsProfile)
		os.Exit(1)
	}

	// Validate algorithm profile.
	algProfile, ok := optimisation.GetProfile(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown algorithm profile: %s\n", profileName)
		os.Exit(1)
	}

	dataset, err := loader.LoadDataset(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading dataset: %v\n", err)
		os.Exit(1)
	}

	result, err := application.Optimise(dataset.Events, dataset.Resources, convertTravel(dataset.TravelMatrix), algorithm, algProfile)
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
	fmt.Printf("%-28s %-26s %7s %11s %8s %9s %10s %10s %12s\n",
		"Dataset", "Algorithm", "Score", "Objective", "Delta", "Delta %", "Assigned", "Duration", "Candidates")
	fmt.Println(strings.Repeat("-", 125))

	// Run each combination.
	type benchResult struct {
		alg       string
		score     int
		objective int
		assigned  int
		duration  int64
		candidates int
	}

	// Aggregate stats per algorithm for summary.
	type algStats struct {
		count      int
		totalObj   int
		totalDelta int
		totalCands int
	}
	summary := make(map[string]*algStats)

	for _, file := range datasetFiles {
		path := filepath.Join(dir, file)
		dataset, err := loader.LoadDataset(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR loading %s: %v\n", file, err)
			continue
		}

		travel := convertTravel(dataset.TravelMatrix)
		name := strings.TrimSuffix(file, ".json")

		// Run all algorithms and collect results.
		var results []benchResult
		baseline := 0

		for _, alg := range algs {
			result, err := application.Optimise(dataset.Events, dataset.Resources, travel, alg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ERROR %s/%s: %v\n", name, alg, err)
				continue
			}

			stats := result.Statistics()
			br := benchResult{
				alg:        alg,
				score:      result.Score(),
				objective:  result.ObjectiveScore(),
				assigned:   result.Size(),
				duration:   stats.DurationMs,
				candidates: stats.CandidatesEvaluated,
			}
			results = append(results, br)

			if alg == "constructive" {
				baseline = br.objective
			}
		}

		// Print results with delta.
		for _, br := range results {
			delta := br.objective - baseline
			deltaStr := "0"
			if delta > 0 {
				deltaStr = fmt.Sprintf("+%d", delta)
			} else if delta < 0 {
				deltaStr = fmt.Sprintf("%d", delta)
			}

			pctStr := "0.0%"
			if baseline == 0 {
				pctStr = "n/a"
			} else if delta != 0 {
				pct := float64(delta) / float64(baseline) * 100
				if pct > 0 {
					pctStr = fmt.Sprintf("+%.1f%%", pct)
				} else {
					pctStr = fmt.Sprintf("%.1f%%", pct)
				}
			}

			fmt.Printf("%-28s %-26s %7d %11d %8s %9s %10d %8dms %12d\n",
				name, br.alg, br.score, br.objective, deltaStr, pctStr,
				br.assigned, br.duration, br.candidates)

			// Accumulate for summary.
			if summary[br.alg] == nil {
				summary[br.alg] = &algStats{}
			}
			s := summary[br.alg]
			s.count++
			s.totalObj += br.objective
			s.totalDelta += (br.objective - baseline)
			s.totalCands += br.candidates
		}
	}

	// Print summary.
	fmt.Println()
	fmt.Println("Benchmark Summary:")
	fmt.Println()
	fmt.Printf("%-28s %10s %15s %11s %13s %12s\n",
		"Algorithm", "Datasets", "Avg Objective", "Avg Delta", "Avg Delta %", "Candidates")
	fmt.Println(strings.Repeat("-", 92))

	// Get constructive average objective for percentage calculation.
	constructiveAvgObj := 0
	if cs, ok := summary["constructive"]; ok && cs.count > 0 {
		constructiveAvgObj = cs.totalObj / cs.count
	}

	for _, alg := range algs {
		s, ok := summary[alg]
		if !ok || s.count == 0 {
			continue
		}

		avgObj := s.totalObj / s.count
		avgDelta := s.totalDelta / s.count

		avgDeltaStr := "0"
		if avgDelta > 0 {
			avgDeltaStr = fmt.Sprintf("+%d", avgDelta)
		} else if avgDelta < 0 {
			avgDeltaStr = fmt.Sprintf("%d", avgDelta)
		}

		pctStr := "0.0%"
		if constructiveAvgObj > 0 && avgDelta != 0 {
			pct := float64(avgDelta) / float64(constructiveAvgObj) * 100
			if pct > 0 {
				pctStr = fmt.Sprintf("+%.1f%%", pct)
			} else {
				pctStr = fmt.Sprintf("%.1f%%", pct)
			}
		}

		fmt.Printf("%-28s %10d %15d %11s %13s %12d\n",
			alg, s.count, avgObj, avgDeltaStr, pctStr, s.totalCands)
	}
}

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

// parseProfile reads the --profile flag from remaining args.
// Defaults to "default".
func parseProfile(args []string) string {
	for i, arg := range args {
		if arg == "--profile" && i+1 < len(args) {
			return strings.TrimSpace(args[i+1])
		}
		if strings.HasPrefix(arg, "--profile=") {
			return strings.TrimSpace(strings.TrimPrefix(arg, "--profile="))
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
