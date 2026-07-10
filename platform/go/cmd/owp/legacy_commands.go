// legacy_commands.go — deprecated CLI commands backed by internal/legacy/application.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/legacy/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/legacysearch"
)

func runOptimise() {
	warnDeprecated("owp optimise", "domain-specific solvers: solve-cvrp, solve-vrptw, solve-jobshop, tune-pfrs")

	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	path := os.Args[2]
	algorithm := parseAlgorithm(os.Args[3:])
	weightsProfile := parseWeights(os.Args[3:])
	profileName := parseProfile(os.Args[3:])

	if _, ok := legacysearch.GetWeightProfile(weightsProfile); !ok {
		fmt.Fprintf(os.Stderr, "Unknown weights profile: %s\n", weightsProfile)
		os.Exit(1)
	}

	algProfile, ok := legacysearch.GetProfile(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown algorithm profile: %s\n", profileName)
		os.Exit(1)
	}
	algProfile = applyProfileOverrides(os.Args[3:], algProfile)

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

	resp := application.BuildResponse(
		result,
		algorithm,
		buildCapacityLookup(dataset.Resources),
		buildDurationLookup(dataset.Events),
		buildItemLocationLookup(dataset.Events),
		buildResourceLocationLookup(dataset.Resources),
		buildTravelDisplayLookup(dataset.TravelMatrix),
	)

	fmt.Println("=== Optimised Plan ===")
	fmt.Println()
	fmt.Printf("Algorithm: %s\n", resp.Algorithm)
	fmt.Printf("Profile: %s\n", profileName)
	displayEffectiveConfig(algorithm, algProfile)
	fmt.Printf("Assignment Score: %d\n", resp.AssignmentScore)
	fmt.Printf("Objective Score:  %d\n", resp.ObjectiveScore)
	fmt.Println()
	fmt.Println("Objective Breakdown:")
	for _, entry := range resp.ObjectiveBreakdown {
		fmt.Printf("  %s: %d\n", entry.Name, entry.Score)
	}
	fmt.Println()

	fmt.Println("Constraints:")
	fmt.Printf("  Hard: %d\n", resp.Constraints.HardCount)
	fmt.Printf("  Soft: %d\n", resp.Constraints.SoftCount)
	fmt.Printf("  Penalty: %d\n", resp.Constraints.TotalPenalty)
	if len(resp.Constraints.Summary) > 0 {
		fmt.Println("  Breakdown:")
		for _, s := range resp.Constraints.Summary {
			fmt.Printf("    %s: %d\n", s.Constraint, s.Count)
		}
	}
	fmt.Println()

	fmt.Printf("Resources: %d\n", len(resp.Resources))
	fmt.Printf("Capacity:  %d\n", resp.TotalCapacity)
	fmt.Println()
	fmt.Println("Assignments:")
	fmt.Println()

	for _, res := range resp.Resources {
		fmt.Printf("  %s\n", res.ResourceID)
		fmt.Printf("    Used: %d / %d mins\n", res.UsedMins, res.CapacityMins)
		fmt.Println("    Work Items:")
		for _, itemID := range res.WorkItems {
			fmt.Printf("      - %s\n", itemID)
		}
		fmt.Println()
	}

	if len(resp.Unassigned) > 0 {
		fmt.Println("Unassigned:")
		fmt.Println()
		for _, item := range resp.Unassigned {
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

	if resp.Constraints.HardCount > 0 {
		fmt.Println("Hard Violations:")
		fmt.Println()
		for _, m := range resp.Constraints.Matches {
			if m.Severity == "hard" {
				fmt.Printf("  [%s] %s\n", m.Constraint, m.Description)
			}
		}
		fmt.Println()
	}

	fmt.Println("Travel:")
	fmt.Println()
	for _, rt := range resp.Travel {
		fmt.Printf("  %s\n", rt.ResourceID)
		for _, leg := range rt.Legs {
			fmt.Printf("    %s -> %s: %d mins\n", leg.From, leg.To, leg.Minutes)
		}
		fmt.Printf("    Total: %d mins\n", rt.TotalMins)
		fmt.Println()
	}

	fmt.Println("Optimisation Statistics:")
	fmt.Printf("  Algorithm: %s\n", resp.Statistics.Algorithm)
	fmt.Printf("  Duration: %dms\n", resp.Statistics.DurationMs)
	fmt.Printf("  Iterations: %d\n", resp.Statistics.Iterations)
	fmt.Printf("  Candidates Evaluated: %d\n", resp.Statistics.CandidatesEvaluated)
	fmt.Printf("  Improvements Accepted: %d\n", resp.Statistics.ImprovementsAccepted)
	fmt.Printf("  Final Objective Score: %d\n", resp.Statistics.FinalObjectiveScore)
	fmt.Println()
	fmt.Println("Done.")
}

func runBenchmark() {
	warnDeprecated("owp benchmark", "domain-specific commands: solve-cvrp, tune-pfrs, benchmark-inrc2")

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

	algs := legacysearch.Available()
	sort.Strings(algs)

	fmt.Printf("%-28s %-26s %7s %11s %8s %9s %10s %6s %6s %10s %12s\n",
		"Dataset", "Algorithm", "Score", "Objective", "Delta", "Delta %", "Assigned", "Hard", "Soft", "Duration", "Candidates")
	fmt.Println(strings.Repeat("-", 145))

	type benchResult struct {
		alg        string
		score      int
		objective  int
		assigned   int
		hard       int
		soft       int
		duration   int64
		candidates int
	}
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
				hard:       result.HardConstraintCount(),
				soft:       result.SoftConstraintCount(),
				duration:   stats.DurationMs,
				candidates: stats.CandidatesEvaluated,
			}
			results = append(results, br)
			if alg == "constructive" {
				baseline = br.objective
			}
		}

		for _, br := range results {
			delta := br.objective - baseline
			deltaStr, pctStr := formatObjectiveDelta(delta, baseline)
			fmt.Printf("%-28s %-26s %7d %11d %8s %9s %10d %6d %6d %8dms %12d\n",
				name, br.alg, br.score, br.objective, deltaStr, pctStr,
				br.assigned, br.hard, br.soft, br.duration, br.candidates)

			if summary[br.alg] == nil {
				summary[br.alg] = &algStats{}
			}
			s := summary[br.alg]
			s.count++
			s.totalObj += br.objective
			s.totalDelta += br.objective - baseline
			s.totalCands += br.candidates
		}
	}

	fmt.Println()
	fmt.Println("Benchmark Summary:")
	fmt.Println()
	fmt.Printf("%-28s %10s %15s %11s %13s %12s\n",
		"Algorithm", "Datasets", "Avg Objective", "Avg Delta", "Avg Delta %", "Candidates")
	fmt.Println(strings.Repeat("-", 92))

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
		avgDeltaStr, pctStr := formatObjectiveDelta(avgDelta, constructiveAvgObj)
		if constructiveAvgObj == 0 {
			pctStr = "0.0%"
		}
		fmt.Printf("%-28s %10d %15d %11s %13s %12d\n",
			alg, s.count, avgObj, avgDeltaStr, pctStr, s.totalCands)
	}
}
