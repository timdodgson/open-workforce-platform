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
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
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
	case "validate-inrc2":
		runValidateINRC2()
	case "solve-inrc2":
		runSolveINRC2()
	case "benchmark-inrc2":
		runBenchmarkINRC2()
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
	fmt.Fprintln(os.Stderr, "  owp validate-inrc2 <scenario-file> <week-file> <history-file> <solution-file>")
	fmt.Fprintln(os.Stderr, "  owp solve-inrc2 <scenario-file> <week-file> <history-file> <solution-output-file> [--algorithm tabu-search] [--profile default]")
	fmt.Fprintln(os.Stderr, "  owp benchmark-inrc2 <instance-directory> [--profile research]")
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

	result, err := application.OptimiseWithNRP(dataset.Events, dataset.Resources, convertTravel(dataset.TravelMatrix), dataset.NRPContext, algorithm, algProfile)
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

	// Hard violations.
	if result.HasHardViolations() {
		fmt.Println("Hard Violations:")
		fmt.Println()
		for _, v := range result.HardViolations() {
			fmt.Printf("  [%s] %s\n", v.Code, v.Message)
		}
		fmt.Println()
	}

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
			result, err := application.OptimiseWithNRP(dataset.Events, dataset.Resources, travel, dataset.NRPContext, alg)
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

func runValidateINRC2() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "Usage: owp validate-inrc2 <scenario-file> <week-file> <history-file> <solution-file>")
		os.Exit(1)
	}

	scenarioPath := os.Args[2]
	weekPath := os.Args[3]
	historyPath := os.Args[4]
	solutionPath := os.Args[5]

	sc, err := inrc2.LoadScenario(scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading scenario: %v\n", err)
		os.Exit(1)
	}

	wd, err := inrc2.LoadWeekData(weekPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading week data: %v\n", err)
		os.Exit(1)
	}

	hist, err := inrc2.LoadHistory(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
		os.Exit(1)
	}

	sol, err := inrc2.LoadSolution(solutionPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading solution: %v\n", err)
		os.Exit(1)
	}

	result := inrc2.Score(sc, wd, hist, sol)

	fmt.Println("=== INRC-II Validation ===")
	fmt.Println()
	fmt.Printf("Scenario: %s\n", sc.ID)
	fmt.Printf("Week: %d\n", sol.Week)
	fmt.Printf("Assignments: %d\n", len(sol.Assignments))
	fmt.Println()

	fmt.Printf("Hard Violations: %d\n", result.HardViolations)
	if result.HardViolations > 0 {
		for _, v := range result.HardDetails {
			fmt.Printf("  [%s] %s (nurse=%s, day=%s)\n", v.Code, v.Message, v.Nurse, inrc2.DayName(v.Day))
		}
	}
	fmt.Println()

	fmt.Printf("Soft Penalty: %d\n", result.SoftPenalty)
	if len(result.SoftDetails) > 0 {
		fmt.Println("  Breakdown:")
		for _, d := range result.SoftDetails {
			if d.Nurse != "" {
				fmt.Printf("    [%s] nurse=%s penalty=%d\n", d.Constraint, d.Nurse, d.Penalty)
			} else {
				fmt.Printf("    [%s] penalty=%d\n", d.Constraint, d.Penalty)
			}
		}
	}
	fmt.Println()

	fmt.Printf("Total Objective: %d\n", result.TotalObjective)
	fmt.Println()
	fmt.Println("Done.")
}

func runSolveINRC2() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "Usage: owp solve-inrc2 <scenario-file> <week-file> <history-file> <solution-output-file> [--algorithm tabu-search] [--profile default]")
		os.Exit(1)
	}

	scenarioPath := os.Args[2]
	weekPath := os.Args[3]
	historyPath := os.Args[4]
	outputPath := os.Args[5]
	algorithm := parseAlgorithm(os.Args[5:])
	profileName := parseProfile(os.Args[5:])

	algProfile, ok := optimisation.GetProfile(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown profile: %s\n", profileName)
		os.Exit(1)
	}

	sc, err := inrc2.LoadScenario(scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading scenario: %v\n", err)
		os.Exit(1)
	}

	wd, err := inrc2.LoadWeekData(weekPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading week data: %v\n", err)
		os.Exit(1)
	}

	hist, err := inrc2.LoadHistory(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
		os.Exit(1)
	}

	sol, _, err := inrc2.SolveWeek(sc, wd, hist, algorithm, algProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error solving: %v\n", err)
		os.Exit(1)
	}

	if err := inrc2.WriteSolution(sol, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing solution: %v\n", err)
		os.Exit(1)
	}

	// Score and display.
	result := inrc2.Score(sc, wd, hist, sol)

	fmt.Printf("=== INRC-II Solve ===\n\n")
	fmt.Printf("Scenario: %s\n", sc.ID)
	fmt.Printf("Week: %d\n", sol.Week)
	fmt.Printf("Algorithm: %s\n", algorithm)
	fmt.Printf("Assignments: %d\n", len(sol.Assignments))
	fmt.Printf("Hard Violations: %d\n", result.HardViolations)
	fmt.Printf("Soft Penalty: %d\n", result.SoftPenalty)
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Println("\nDone.")
}

func runBenchmarkINRC2() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: owp benchmark-inrc2 <instance-directory> [--profile research]")
		os.Exit(1)
	}

	dir := os.Args[2]
	profileName := parseProfile(os.Args[2:])
	algProfile, ok := optimisation.GetProfile(profileName)
	if !ok {
		algProfile = optimisation.DefaultProfile()
	}

	// Find scenario file.
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory: %v\n", err)
		os.Exit(1)
	}

	var scenarioFile string
	var weekFiles []string
	var histFiles []string

	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, "Sc-") && strings.HasSuffix(name, ".json") {
			scenarioFile = filepath.Join(dir, name)
		} else if strings.HasPrefix(name, "WD-") && strings.HasSuffix(name, ".json") {
			weekFiles = append(weekFiles, filepath.Join(dir, name))
		} else if strings.HasPrefix(name, "H0-") && strings.HasSuffix(name, ".json") {
			histFiles = append(histFiles, filepath.Join(dir, name))
		}
	}

	if scenarioFile == "" {
		fmt.Fprintln(os.Stderr, "No scenario file found")
		os.Exit(1)
	}
	if len(histFiles) == 0 || len(weekFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No history or week files found")
		os.Exit(1)
	}

	sort.Strings(weekFiles)
	sort.Strings(histFiles)

	sc, err := inrc2.LoadScenario(scenarioFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Get available algorithms.
	algs := optimisation.Available()
	sort.Strings(algs)

	// Use first history file.
	hist, err := inrc2.LoadHistory(histFiles[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Solve first N weeks (up to scenario.NumberOfWeeks).
	numWeeks := sc.NumberOfWeeks
	if numWeeks > len(weekFiles) {
		numWeeks = len(weekFiles)
	}

	fmt.Printf("=== INRC-II Benchmark: %s ===\n\n", sc.ID)
	fmt.Printf("%-8s %-28s %8s %8s %12s %10s\n",
		"Week", "Algorithm", "Hard", "Soft", "Assignments", "Duration")
	fmt.Println(strings.Repeat("-", 80))

	for w := 0; w < numWeeks; w++ {
		wd, err := inrc2.LoadWeekData(weekFiles[w])
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error loading week %d: %v\n", w, err)
			continue
		}

		for _, alg := range algs {
			sol, planResult, err := inrc2.SolveWeek(sc, wd, hist, alg, algProfile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Error %s week %d: %v\n", alg, w, err)
				continue
			}

			score := inrc2.Score(sc, wd, hist, sol)
			stats := planResult.Statistics()

			fmt.Printf("%-8d %-28s %8d %8d %12d %8dms\n",
				w, alg, score.HardViolations, score.SoftPenalty,
				len(sol.Assignments), stats.DurationMs)
		}

		// Update history using constructive solution for next week.
		sol, _, _ := inrc2.SolveWeek(sc, wd, hist, "constructive", algProfile)
		hist = inrc2.UpdateHistory(sc, hist, sol)
	}

	fmt.Println("\nDone.")
}
