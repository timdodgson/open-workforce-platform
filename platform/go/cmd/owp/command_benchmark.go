package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	cvrpilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	jssilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/s3upload"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
	vrptwilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

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
	fmt.Printf("%-28s %-26s %7s %11s %8s %9s %10s %6s %6s %10s %12s\n",
		"Dataset", "Algorithm", "Score", "Objective", "Delta", "Delta %", "Assigned", "Hard", "Soft", "Duration", "Candidates")
	fmt.Println(strings.Repeat("-", 145))

	// Run each combination.
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

		// Print results with delta.
		for _, br := range results {
			delta := br.objective - baseline
			deltaStr, pctStr := formatObjectiveDelta(delta, baseline)

			fmt.Printf("%-28s %-26s %7d %11d %8s %9s %10d %6d %6d %8dms %12d\n",
				name, br.alg, br.score, br.objective, deltaStr, pctStr,
				br.assigned, br.hard, br.soft, br.duration, br.candidates)

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

		avgDeltaStr, pctStr := formatObjectiveDelta(avgDelta, constructiveAvgObj)
		if constructiveAvgObj == 0 {
			pctStr = "0.0%"
		}

		fmt.Printf("%-28s %10d %15d %11s %13s %12d\n",
			alg, s.count, avgObj, avgDeltaStr, pctStr, s.totalCands)
	}
}

func runBenchmarkINRC2() {
	// Default instance and profile.
	defaultInstance := "n012w8"
	defaultProfile := "research"
	defaultBasePath := "../../examples/inrc2/testdatasets_json"

	// Resolve instance directory.
	var dir string
	profileName := defaultProfile

	if len(os.Args) >= 3 && !strings.HasPrefix(os.Args[2], "--") {
		// Explicit instance name or path supplied.
		arg := os.Args[2]
		if _, err := os.Stat(arg); err == nil {
			dir = arg
		} else {
			// Try as instance name under test datasets.
			candidate := filepath.Join(defaultBasePath, arg)
			if _, err := os.Stat(candidate); err == nil {
				dir = candidate
			} else {
				// Try competition datasets.
				candidate = filepath.Join("../../examples/inrc2/datasets_json", arg)
				if _, err := os.Stat(candidate); err == nil {
					dir = candidate
				} else {
					fmt.Fprintf(os.Stderr, "Instance not found: %s\n", arg)
					os.Exit(1)
				}
			}
		}
		profileName = parseProfile(os.Args[2:])
	} else {
		// No instance argument — use default n010w4. No fallback.
		dir = filepath.Join(defaultBasePath, defaultInstance)
		if _, err := os.Stat(dir); err != nil {
			// Also try datasets_json path.
			dir = filepath.Join("../../examples/inrc2/datasets_json", defaultInstance)
			if _, err := os.Stat(dir); err != nil {
				fmt.Fprintln(os.Stderr, "Default INRC-II benchmark instance not found. Please ensure examples/inrc2/testdatasets_json/n012w8 exists.")
				os.Exit(1)
			}
		}
		if len(os.Args) >= 3 {
			profileName = parseProfile(os.Args[2:])
		}
	}

	algProfile, ok := optimisation.GetProfile(profileName)
	if !ok {
		algProfile = optimisation.ResearchProfile()
		profileName = "research"
	}

	// Apply explicit CLI overrides.
	algProfile = applyProfileOverrides(os.Args[1:], algProfile)

	// --time is not supported. Reject clearly.
	timeBudget := parseTimeBudget(os.Args[1:])
	if timeBudget > 0 {
		fmt.Fprintln(os.Stderr, "--time is not supported. Use explicit algorithm tuning flags such as --tabu-max-iterations or --sa-max-iterations.")
		os.Exit(1)
	}

	// Support --algorithm to filter to a single algorithm.
	algorithmFilter := parseStringFlag(os.Args[1:], "--algorithm")

	// Find scenario file.
	scenarioFile, weekFiles, histFiles := scanINRC2Dir(dir)

	if scenarioFile == "" {
		fmt.Fprintln(os.Stderr, "No scenario file found")
		os.Exit(1)
	}
	if len(histFiles) == 0 || len(weekFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No history or week files found")
		os.Exit(1)
	}

	sc, err := inrc2.LoadScenario(scenarioFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	algs := optimisation.Available()
	sort.Strings(algs)

	// Filter to single algorithm if --algorithm supplied in benchmark-inrc2.
	if algorithmFilter != "" {
		algs = []string{strings.TrimSpace(algorithmFilter)}
	}

	hist, err := inrc2.LoadHistory(histFiles[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	numWeeks := sc.NumberOfWeeks
	if numWeeks > len(weekFiles) {
		numWeeks = len(weekFiles)
	}

	// Header.
	fmt.Println("=================================================")
	fmt.Println("  Open Workforce Platform Benchmark")
	fmt.Printf("  Instance: %s\n", sc.ID)
	fmt.Printf("  Profile:  %s\n", profileName)
	fmt.Printf("  Weeks:    %d\n", numWeeks)
	fmt.Println("=================================================")
	fmt.Println()
	if len(algs) == 1 {
		displayEffectiveConfig(algs[0], algProfile)
	}

	// Accumulate results per algorithm across all weeks.
	type algResult struct {
		alg          string
		totalPenalty int
		totalHard    int
		totalSoft    int
		totalAssign  int
		totalMs      int64
		totalCands   int
	}

	results := make(map[string]*algResult)
	for _, alg := range algs {
		results[alg] = &algResult{alg: alg}
	}

	currentHist := hist
	for w := 0; w < numWeeks; w++ {
		wd, err := inrc2.LoadWeekData(weekFiles[w])
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error loading week %d: %v\n", w, err)
			continue
		}

		// Track the solution to use for history advancement when running a single algorithm.
		var weekSolForHistory inrc2.Solution
		hasSolForHistory := false

		for _, alg := range algs {
			fmt.Printf("\r  Solving week %d/%d %s...          ", w+1, numWeeks, alg)

			var sol inrc2.Solution
			var scoreResult inrc2.ScoreResult
			var durationMs int64
			var candidatesEval int

			if alg == "parallel-feasible-roster-search" {
				pfrsConfig := parsePFRSConfig(os.Args[1:])
				pfrsSol, pfrsStats, pfrsScore, err := inrc2.SolveWeekPFRS(sc, wd, currentHist, pfrsConfig)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\n  Error %s week %d: %v\n", alg, w, err)
					continue
				}
				sol = pfrsSol
				scoreResult = pfrsScore
				durationMs = pfrsStats.DurationMs
				candidatesEval = pfrsStats.CandidatesEvaluated
			} else {
				owpSol, planResult, err := inrc2.SolveWeek(sc, wd, currentHist, alg, algProfile)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\n  Error %s week %d: %v\n", alg, w, err)
					continue
				}
				sol = owpSol
				scoreResult = inrc2.Score(sc, wd, currentHist, sol)
				stats := planResult.Statistics()
				durationMs = stats.DurationMs
				candidatesEval = stats.CandidatesEvaluated
			}

			r := results[alg]
			r.totalPenalty += scoreResult.SoftPenalty
			r.totalHard += scoreResult.HardViolations
			r.totalSoft += len(scoreResult.SoftDetails)
			r.totalAssign += len(sol.Assignments)
			r.totalMs += durationMs
			r.totalCands += candidatesEval

			// Capture solution for single-algorithm history advancement.
			if len(algs) == 1 {
				weekSolForHistory = sol
				hasSolForHistory = true
			}
		}

		// Advance history for next week.
		// When running a single algorithm, use its solution for history advancement
		// so the algorithm gets consistent history across weeks.
		// When running multiple algorithms, use constructive for fair comparison.
		if len(algs) == 1 && hasSolForHistory {
			currentHist = inrc2.UpdateHistory(sc, currentHist, weekSolForHistory)
		} else {
			sol, _, _ := inrc2.SolveWeek(sc, wd, currentHist, "constructive", algProfile)
			currentHist = inrc2.UpdateHistory(sc, currentHist, sol)
		}
	}
	fmt.Print("\r                                              \r")

	// Separate valid (Hard == 0) from invalid results.
	var valid []*algResult
	var invalid []*algResult
	for _, alg := range algs {
		r := results[alg]
		if r.totalHard == 0 {
			valid = append(valid, r)
		} else {
			invalid = append(invalid, r)
		}
	}

	// Check --show-invalid flag.
	showInvalid := parseShowInvalidFlag(os.Args[1:])

	// Sort valid: penalty asc, then soft asc, then runtime asc, then name asc.
	sort.Slice(valid, func(i, j int) bool {
		if valid[i].totalPenalty != valid[j].totalPenalty {
			return valid[i].totalPenalty < valid[j].totalPenalty
		}
		if valid[i].totalSoft != valid[j].totalSoft {
			return valid[i].totalSoft < valid[j].totalSoft
		}
		if valid[i].totalMs != valid[j].totalMs {
			return valid[i].totalMs < valid[j].totalMs
		}
		return valid[i].alg < valid[j].alg
	})

	// Sort invalid: hard asc (least invalid first), then penalty asc.
	sort.Slice(invalid, func(i, j int) bool {
		if invalid[i].totalHard != invalid[j].totalHard {
			return invalid[i].totalHard < invalid[j].totalHard
		}
		return invalid[i].totalPenalty < invalid[j].totalPenalty
	})

	// Print valid league table.
	if len(valid) > 0 {
		fmt.Printf("%-6s %-28s %10s %8s %12s %10s %12s\n",
			"Rank", "Algorithm", "Penalty", "Soft", "Assignments", "Runtime", "Candidates")
		fmt.Println(strings.Repeat("-", 92))

		for rank, r := range valid {
			fmt.Printf("%-6d %-28s %10d %8d %12d %8dms %12d\n",
				rank+1, r.alg, r.totalPenalty, r.totalSoft,
				r.totalAssign, r.totalMs, r.totalCands)
		}
	} else {
		fmt.Println("No valid solutions (Hard = 0) found.")
	}

	// Print invalid/rejected section only when --show-invalid is supplied.
	if len(invalid) > 0 && showInvalid {
		fmt.Println()
		fmt.Println("Rejected (Invalid Solutions):")
		fmt.Printf("       %-28s %10s %8s %8s %12s %10s %12s\n",
			"Algorithm", "Penalty", "Hard", "Soft", "Assignments", "Runtime", "Candidates")
		fmt.Println("       " + strings.Repeat("-", 92))

		for _, r := range invalid {
			fmt.Printf("       %-28s %10d %8d %8d %12d %8dms %12d\n",
				r.alg, r.totalPenalty, r.totalHard, r.totalSoft,
				r.totalAssign, r.totalMs, r.totalCands)
		}
	}

	// Summary.
	fmt.Println()
	fmt.Println("Summary:")

	if len(valid) > 0 {
		fmt.Printf("  Best algorithm:    %s (penalty: %d)\n", valid[0].alg, valid[0].totalPenalty)

		fastest := valid[0]
		for _, r := range valid {
			if r.totalMs < fastest.totalMs {
				fastest = r
			}
		}
		fmt.Printf("  Fastest valid:     %s (%dms)\n", fastest.alg, fastest.totalMs)

		totalPenalty := 0
		totalMs := int64(0)
		totalSoft := 0
		for _, r := range valid {
			totalPenalty += r.totalPenalty
			totalMs += r.totalMs
			totalSoft += r.totalSoft
		}
		n := len(valid)
		fmt.Printf("  Average penalty:   %d\n", totalPenalty/n)
		fmt.Printf("  Average runtime:   %dms\n", totalMs/int64(n))
		fmt.Printf("  Average soft:      %d\n", totalSoft/n)
	} else {
		fmt.Println("  No valid solution found.")
		if showInvalid && len(invalid) > 0 {
			fmt.Printf("  Least invalid:     %s (hard: %d, penalty: %d)\n",
				invalid[0].alg, invalid[0].totalHard, invalid[0].totalPenalty)
		}
	}

	fmt.Println()
	fmt.Println("Done.")
}

func runBenchmarkILP() {
	args := os.Args[2:]

	instanceName := parseStringFlag(args, "--instance")
	if instanceName == "" {
		instanceName = "n005w4"
	}

	weeks := parseIntFlag(args, "--weeks")
	if weeks <= 0 {
		weeks = 1 // Default: start with 1 week for tractability.
	}

	timeLimitSec := parseIntFlag(args, "--time-limit")
	if timeLimitSec <= 0 {
		timeLimitSec = 300 // Default 5 minutes.
	}

	outputPath := parseStringFlag(args, "--output")
	if outputPath == "" {
		outputPath = "../web/pfrs-lab/data/ilp-benchmark.json"
	}

	solverName := parseStringFlag(args, "--solver")
	if solverName == "" {
		solverName = "highs"
	}

	// S3 upload options.
	storage := parseStorageConfig(args, false)
	runLabel := parseRunLabelFlag(args, false)
	if runLabel == "" {
		runLabel = fmt.Sprintf("ilp-%s-%dw", instanceName, weeks)
	}

	parallel := parseBoolFlag(args, "--parallel") != "false"
	// Optional PFRS comparison.
	comparePFRS := parseIntFlag(args, "--compare-pfrs")
	comparePFRSRuntime := parseFloatFlag(args, "--compare-pfrs-runtime")

	inst := loadINRC2Instance(instanceName)
	sc := inst.Scenario
	hist := inst.History
	weekFiles := inst.WeekFiles

	if weeks > len(weekFiles) {
		weeks = len(weekFiles)
	}
	if weeks > sc.NumberOfWeeks {
		weeks = sc.NumberOfWeeks
	}

	disp := parseDisplayOptions(args)

	fmt.Println(disp.Heading(cli.EmojiConfig, "ILP Benchmark"))
	fmt.Println()
	fmt.Printf("  Instance:   %s\n", disp.Bold(sc.ID))
	fmt.Printf("  Nurses:     %d\n", len(sc.Nurses))
	fmt.Printf("  Weeks:      %d\n", weeks)
	fmt.Printf("  Solver:     %s\n", solverName)
	fmt.Printf("  Time Limit: %ds\n", timeLimitSec)
	fmt.Printf("  Output:     %s\n", outputPath)
	fmt.Println()

	// Check solver availability.
	requireHiGHS("Use the Apache static build for parallel support on Windows.")

	// Build model.
	fmt.Print("  Building LP model... ")
	os.Stdout.Sync()

	config := ilp.BenchmarkConfig{
		Instance:   instanceName,
		Weeks:      weeks,
		TimeLimit:  time.Duration(timeLimitSec) * time.Second,
		SolverName: solverName,
		OutputPath: outputPath,
		Parallel:   parallel,
	}

	result, err := ilp.RunBenchmark(sc, weekFiles[:weeks], hist, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("done.")
	fmt.Println()

	// Display results.
	fmt.Println(disp.Heading(cli.EmojiValid, "Benchmark Result"))
	fmt.Println()
	fmt.Printf("  Status:          %s\n", result.Status)
	fmt.Printf("  Objective:       %d\n", result.Objective)
	if result.LowerBound > 0 {
		fmt.Printf("  Lower Bound:     %d\n", result.LowerBound)
		fmt.Printf("  Gap:             %.2f%%\n", result.GapPercent)
	}
	fmt.Printf("  Hard Violations: %d\n", result.HardViolations)
	fmt.Printf("  Runtime:         %.1fs\n", result.RuntimeSeconds)
	if result.Notes != "" {
		fmt.Printf("  Notes:           %s\n", result.Notes)
	}
	fmt.Println()

	if result.SolutionPath != "" {
		fmt.Printf("  Output written: %s\n", result.SolutionPath)
	}
	if result.ProgressPath != "" {
		fmt.Printf("  Progress CSV:  %s\n", result.ProgressPath)
	}

	// Comparison with PFRS if requested.
	if comparePFRS > 0 {
		comparison := ilp.Compare(result, comparePFRS, comparePFRSRuntime)
		fmt.Println()
		fmt.Println(disp.Heading(cli.EmojiConfig, "PFRS vs ILP Comparison"))
		fmt.Println()
		fmt.Printf("  %-12s %10s %12s %10s %10s\n", "Algorithm", "Penalty", "Gap to ILP", "Gap %%", "Runtime")
		fmt.Printf("  %-12s %10s %12s %10s %10s\n", "─────────", "───────", "──────────", "─────", "───────")
		fmt.Printf("  %-12s %10d %12s %10s %10.1fs\n",
			"ILP", result.Objective, "—", "—", result.RuntimeSeconds)

		gapStr := fmt.Sprintf("+%d", comparison.AbsoluteGap)
		if comparison.AbsoluteGap <= 0 {
			gapStr = fmt.Sprintf("%d", comparison.AbsoluteGap)
		}
		gapPctStr := fmt.Sprintf("+%.1f%%", comparison.GapPercent)
		if comparison.GapPercent <= 0 {
			gapPctStr = fmt.Sprintf("%.1f%%", comparison.GapPercent)
		}
		runtimeStr := "—"
		if comparePFRSRuntime > 0 {
			runtimeStr = fmt.Sprintf("%.1fs", comparePFRSRuntime)
		}
		fmt.Printf("  %-12s %10d %12s %10s %10s\n",
			"PFRS", comparePFRS, gapStr, gapPctStr, runtimeStr)
		fmt.Println()
	}

	// Upload to S3 if requested.
	if storage.Mode == "s3" {
		fmt.Fprintf(os.Stderr, "\n  Uploading ILP results to S3: %s/%s\n", storage.Bucket, runLabel)
		s3Client, err := s3upload.NewClient(storage.Bucket, storage.Region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error creating S3 client: %v\n", err)
		} else {
			// Upload benchmark JSON.
			if result.SolutionPath != "" {
				if err := s3Client.UploadLocalFile(runLabel, "ilp-benchmark.json", result.SolutionPath); err != nil {
					fmt.Fprintf(os.Stderr, "  Error uploading benchmark JSON: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "  ✓ ilp-benchmark.json\n")
				}
			}
			// Upload progress CSV.
			if result.ProgressPath != "" {
				if err := s3Client.UploadLocalFile(runLabel, "ilp-progress.csv", result.ProgressPath); err != nil {
					fmt.Fprintf(os.Stderr, "  Error uploading progress CSV: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "  ✓ ilp-progress.csv\n")
				}
			}
			// Upload roster.json for schedule viewer.
			if result.RosterPath != "" {
				if err := s3Client.UploadLocalFile(runLabel, "roster.json", result.RosterPath); err != nil {
					fmt.Fprintf(os.Stderr, "  Error uploading roster.json: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "  ✓ roster.json\n")
				}
			}
			// Upload run.json metadata for the dashboard.
			runMeta := map[string]interface{}{
				"mode":      "ilp",
				"instance":  instanceName,
				"weeks":     weeks,
				"solver":    solverName,
				"timeLimit": timeLimitSec,
				"parallel":  parallel,
				"objective": result.Objective,
				"bound":     result.LowerBound,
				"gap":       result.GapPercent,
				"status":    result.Status,
				"runtime":   result.RuntimeSeconds,
			}
			metaJSON, _ := json.MarshalIndent(runMeta, "", "  ")
			if err := s3Client.UploadFile(runLabel, "run.json", string(metaJSON)); err != nil {
				fmt.Fprintf(os.Stderr, "  Error uploading run.json: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "  ✓ run.json\n")
			}
			// Update manifest.
			if err := s3Client.UpdateManifest(s3upload.ManifestEntry{
				RunID:        runLabel,
				Label:        runLabel,
				Algorithm:    "ilp",
				Timestamp:    s3upload.Timestamp(),
				TotalPenalty: result.Objective,
			}); err != nil {
				fmt.Fprintf(os.Stderr, "  Error updating manifest: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "  ✓ manifest.json updated\n")
			}
			fmt.Fprintf(os.Stderr, "  S3 upload complete.\n")
		}
	}

	fmt.Println("Done.")
}

// --- CVRP ILP Benchmark ---

func runBenchmarkCVRPILP() {
	args := os.Args[2:]

	instancePath := parseStringFlag(args, "--instance")
	if instancePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --instance <path.vrp> is required")
		os.Exit(1)
	}

	timeLimitSec := parseIntFlag(args, "--time-limit")
	if timeLimitSec <= 0 {
		timeLimitSec = 300 // 5 minutes default
	}

	parallel := parseBoolFlag(args, "--parallel") != "false"
	runLabel := parseRunLabelFlag(args, false)
	storage := parseStorageConfig(args, false)

	disp := parseDisplayOptions(args)

	fmt.Println(disp.Heading(cli.EmojiConfig, "CVRP ILP Benchmark"))
	fmt.Println()

	// Load dataset.
	ds, err := cvrp.LoadDataset(instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Instance:   %s\n", disp.Bold(ds.Name))
	fmt.Printf("  Customers:  %d\n", len(ds.Customers))
	fmt.Printf("  Capacity:   %d\n", ds.Capacity)
	fmt.Printf("  Time Limit: %ds\n", timeLimitSec)
	fmt.Printf("  Parallel:   %v\n", parallel)
	fmt.Println()

	requireHiGHS("")

	// Run benchmark.
	fmt.Print("  Solving... ")
	os.Stdout.Sync()

	config := cvrpilp.BenchmarkConfig{
		Instance:  ds.Name,
		TimeLimit: time.Duration(timeLimitSec) * time.Second,
		Parallel:  parallel,
	}

	result, err := cvrpilp.RunBenchmark(ds, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("done.")
	fmt.Println()

	// Display results.
	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Status:      %s\n", result.Status)
	fmt.Printf("  Objective:   %d\n", result.Objective)
	if result.LowerBound > 0 {
		fmt.Printf("  Lower Bound: %d\n", result.LowerBound)
		fmt.Printf("  Gap:         %.2f%%\n", result.GapPercent)
	}
	fmt.Printf("  Runtime:     %.1fs\n", result.RuntimeSeconds)
	fmt.Printf("  Variables:   %d\n", result.Variables)
	fmt.Printf("  Constraints: %d\n", result.Constraints)
	fmt.Printf("  Vehicles:    %d\n", result.Vehicles)
	if result.Notes != "" {
		fmt.Printf("  Notes:       %s\n", result.Notes)
	}
	fmt.Println()

	// Write output if run-label specified.
	if runLabel != "" {
		outputDir := ensureRunOutputDir(runLabel)

		writeRunMetadata(outputDir, map[string]interface{}{
			"problemType": "cvrp",
			"mode":        "ilp",
			"instance":    ds.Name,
			"customers":   len(ds.Customers),
			"capacity":    ds.Capacity,
			"solver":      "highs",
			"objective":   result.Objective,
			"bound":       result.LowerBound,
			"gap":         result.GapPercent,
			"status":      result.Status,
			"runtime":     result.RuntimeSeconds,
			"timeLimit":   timeLimitSec,
			"vehicles":    result.Vehicles,
			"variables":   result.Variables,
			"constraints": result.Constraints,
			"runLabel":    runLabel,
		})

		benchJSON, _ := json.MarshalIndent(result, "", "  ")
		writeTelemetryFile(outputDir, "ilp-benchmark.json", benchJSON)

		fmt.Printf("  Output: %s/\n", outputDir)

		uploadRunOutput(storage, runLabel, outputDir, "ilp", result.Objective)
	}

	fmt.Println("Done.")
}

func runBenchmarkVRPTWILP() {
	args := os.Args[2:]

	instancePath := parseStringFlag(args, "--instance")
	if instancePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --instance <path.txt> is required")
		os.Exit(1)
	}

	timeLimitSec := parseIntFlag(args, "--time-limit")
	if timeLimitSec <= 0 {
		timeLimitSec = 300
	}

	parallel := parseBoolFlag(args, "--parallel") != "false"
	runLabel := parseRunLabelFlag(args, false)
	storage := parseStorageConfig(args, false)
	disp := parseDisplayOptions(args)

	fmt.Println(disp.Heading(cli.EmojiConfig, "VRPTW ILP Benchmark"))
	fmt.Println()

	ds, err := vrptw.LoadDataset(instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Instance:   %s\n", disp.Bold(ds.Name))
	fmt.Printf("  Customers:  %d\n", len(ds.Customers))
	fmt.Printf("  Capacity:   %d\n", ds.Capacity)
	fmt.Printf("  Time Limit: %ds\n", timeLimitSec)
	fmt.Printf("  Parallel:   %v\n", parallel)
	fmt.Println()

	requireHiGHS("")

	fmt.Print("  Solving... ")
	os.Stdout.Sync()

	result, err := vrptwilp.RunBenchmark(ds, vrptwilp.BenchmarkConfig{
		Instance:  ds.Name,
		TimeLimit: time.Duration(timeLimitSec) * time.Second,
		Parallel:  parallel,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("done.")
	fmt.Println()
	printRoutingILPResult(disp, result.Status, result.Objective, result.LowerBound, result.GapPercent,
		result.RuntimeSeconds, result.Variables, result.Constraints, result.Vehicles, result.Notes)

	if runLabel != "" {
		outputDir := ensureRunOutputDir(runLabel)
		writeRunMetadata(outputDir, map[string]interface{}{
			"problemType": "vrptw",
			"mode":        "ilp",
			"instance":    filepath.Base(ds.Name),
			"customers":   len(ds.Customers),
			"capacity":    ds.Capacity,
			"vehicles":    ds.Vehicles,
			"solver":      "highs",
			"objective":   result.Objective,
			"bound":       result.LowerBound,
			"gap":         result.GapPercent,
			"status":      result.Status,
			"runtime":     result.RuntimeSeconds,
			"timeLimit":   timeLimitSec,
			"variables":   result.Variables,
			"constraints": result.Constraints,
			"runLabel":    runLabel,
		})
		benchJSON, _ := json.MarshalIndent(result, "", "  ")
		writeTelemetryFile(outputDir, "ilp-benchmark.json", benchJSON)
		fmt.Printf("  Output: %s/\n", outputDir)
		uploadRunOutput(storage, runLabel, outputDir, "ilp", result.Objective)
	}

	fmt.Println("Done.")
}

func runBenchmarkJSSILP() {
	args := os.Args[2:]

	instancePath := parseStringFlag(args, "--instance")
	if instancePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --instance <path.txt> is required")
		os.Exit(1)
	}

	timeLimitSec := parseIntFlag(args, "--time-limit")
	if timeLimitSec <= 0 {
		timeLimitSec = 300
	}

	parallel := parseBoolFlag(args, "--parallel") != "false"
	runLabel := parseRunLabelFlag(args, false)
	storage := parseStorageConfig(args, false)
	disp := parseDisplayOptions(args)

	fmt.Println(disp.Heading(cli.EmojiConfig, "JSS ILP Benchmark"))
	fmt.Println()

	ds, err := jobshop.LoadDataset(instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Instance:   %s\n", disp.Bold(ds.Name))
	fmt.Printf("  Jobs:       %d\n", ds.Jobs)
	fmt.Printf("  Machines:   %d\n", ds.Machines)
	fmt.Printf("  Time Limit: %ds\n", timeLimitSec)
	fmt.Printf("  Parallel:   %v\n", parallel)
	fmt.Println()

	requireHiGHS("")

	fmt.Print("  Solving... ")
	os.Stdout.Sync()

	result, err := jssilp.RunBenchmark(ds, jssilp.BenchmarkConfig{
		Instance:  ds.Name,
		TimeLimit: time.Duration(timeLimitSec) * time.Second,
		Parallel:  parallel,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("done.")
	fmt.Println()
	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Status:      %s\n", result.Status)
	fmt.Printf("  Makespan:    %d\n", result.Objective)
	if result.LowerBound > 0 {
		fmt.Printf("  Lower Bound: %d\n", result.LowerBound)
		fmt.Printf("  Gap:         %.2f%%\n", result.GapPercent)
	}
	fmt.Printf("  Runtime:     %.1fs\n", result.RuntimeSeconds)
	fmt.Printf("  Variables:   %d\n", result.Variables)
	fmt.Printf("  Constraints: %d\n", result.Constraints)
	fmt.Printf("  Operations:  %d\n", result.Operations)
	if result.Notes != "" {
		fmt.Printf("  Notes:       %s\n", result.Notes)
	}
	fmt.Println()

	if runLabel != "" {
		outputDir := ensureRunOutputDir(runLabel)
		instanceName := filepath.Base(instancePath)
		writeRunMetadata(outputDir, map[string]interface{}{
			"problemType":  "jss",
			"mode":         "ilp",
			"instance":     instanceName,
			"jobs":         ds.Jobs,
			"machines":     ds.Machines,
			"solver":       "highs",
			"objective":    result.Objective,
			"bestMakespan": result.Objective,
			"bound":        result.LowerBound,
			"gap":          result.GapPercent,
			"status":       result.Status,
			"runtime":      result.RuntimeSeconds,
			"timeLimit":    timeLimitSec,
			"variables":    result.Variables,
			"constraints":  result.Constraints,
			"runLabel":     runLabel,
		})
		benchJSON, _ := json.MarshalIndent(result, "", "  ")
		writeTelemetryFile(outputDir, "ilp-benchmark.json", benchJSON)
		if len(result.SolutionJSON) > 0 {
			writeTelemetryFile(outputDir, "solution.json", result.SolutionJSON)
		}
		fmt.Printf("  Output: %s/\n", outputDir)
		uploadRunOutput(storage, runLabel, outputDir, "ilp", result.Objective)
	}

	fmt.Println("Done.")
}

func printRoutingILPResult(disp cli.Options, status string, objective, lowerBound int, gap, runtime float64, variables, constraints, vehicles int, notes string) {
	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Status:      %s\n", status)
	fmt.Printf("  Objective:   %d\n", objective)
	if lowerBound > 0 {
		fmt.Printf("  Lower Bound: %d\n", lowerBound)
		fmt.Printf("  Gap:         %.2f%%\n", gap)
	}
	fmt.Printf("  Runtime:     %.1fs\n", runtime)
	fmt.Printf("  Variables:   %d\n", variables)
	fmt.Printf("  Constraints: %d\n", constraints)
	fmt.Printf("  Vehicles:    %d\n", vehicles)
	if notes != "" {
		fmt.Printf("  Notes:       %s\n", notes)
	}
	fmt.Println()
}
