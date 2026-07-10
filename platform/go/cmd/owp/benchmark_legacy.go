package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/legacy/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

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

	algs := optimisation.Available()
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
