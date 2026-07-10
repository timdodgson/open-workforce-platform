package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/runoutput"
)

func runBenchmarkILP() {
	args := os.Args[2:]

	instanceName := parseStringFlag(args, "--instance")
	if instanceName == "" {
		instanceName = "n005w4"
	}

	weeks := parseIntFlag(args, "--weeks")
	if weeks <= 0 {
		weeks = 1
	}

	timeLimitSec := parseIntFlag(args, "--time-limit")
	if timeLimitSec <= 0 {
		timeLimitSec = 300
	}

	outputPath := parseStringFlag(args, "--output")
	runLabel := parseRunLabelFlag(args, false)
	if runLabel == "" {
		runLabel = fmt.Sprintf("ilp-%s-%dw", instanceName, weeks)
	}
	if runLabel != "" && outputPath == "" {
		outputPath = filepath.Join(runoutput.EnsureDir(runLabel), "ilp-benchmark.json")
	} else if outputPath == "" {
		outputPath = "../web/pfrs-lab/data/ilp-benchmark.json"
	}

	parallel := parseParallelFlag(args)
	solverName := parseStringFlag(args, "--solver")
	if solverName == "" {
		solverName = "highs"
	}
	storage := parseStorageConfig(args, false)
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
	fmt.Printf("  Parallel:   %v\n", parallel)
	fmt.Printf("  Output:     %s\n", outputPath)
	fmt.Println()

	requireHiGHS("")

	fmt.Print("  Building LP model... ")
	os.Stdout.Sync()

	benchOut, err := ilp.RunAndFinalize(sc, weekFiles[:weeks], hist, ilp.BenchmarkOptions{
		Config: ilp.BenchmarkConfig{
			Instance:   instanceName,
			Weeks:      weeks,
			TimeLimit:  time.Duration(timeLimitSec) * time.Second,
			SolverName: solverName,
			OutputPath: outputPath,
			Parallel:   parallel,
		},
		RunLabel:     runLabel,
		Storage:      runoutputStorage(storage),
		TimeLimitSec: timeLimitSec,
		Parallel:     parallel,
		SolverName:   solverName,
	})
	result := benchOut.Result
	outputDir := benchOut.OutputDir
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("done.")
	fmt.Println()

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

	if outputDir != "" {
		fmt.Printf("  Output: %s/\n", outputDir)
	}

	fmt.Println("Done.")
}
