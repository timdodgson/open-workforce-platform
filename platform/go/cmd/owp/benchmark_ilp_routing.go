package main

import (
	"fmt"
	"os"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	cvrpilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	jssilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
	vrptwilp "github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw/ilp"
)

type routingILPFlags struct {
	instancePath string
	timeLimitSec int
	parallel     bool
	runLabel     string
	storage      storageConfig
	disp         cli.Options
}

func parseRoutingILPFlags(args []string) routingILPFlags {
	timeLimitSec := parseIntFlag(args, "--time-limit")
	if timeLimitSec <= 0 {
		timeLimitSec = 300
	}
	return routingILPFlags{
		instancePath: parseStringFlag(args, "--instance"),
		timeLimitSec: timeLimitSec,
		parallel:     parseParallelFlag(args),
		runLabel:     parseRunLabelFlag(args, false),
		storage:      parseStorageConfig(args, false),
		disp:         parseDisplayOptions(args),
	}
}

func runBenchmarkCVRPILP() {
	flags := parseRoutingILPFlags(os.Args[2:])
	if flags.instancePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --instance <path.vrp> is required")
		os.Exit(1)
	}

	fmt.Println(flags.disp.Heading(cli.EmojiConfig, "CVRP ILP Benchmark"))
	fmt.Println()

	ds, err := cvrp.LoadDataset(flags.instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Instance:   %s\n", flags.disp.Bold(ds.Name))
	fmt.Printf("  Customers:  %d\n", len(ds.Customers))
	fmt.Printf("  Capacity:   %d\n", ds.Capacity)
	fmt.Printf("  Time Limit: %ds\n", flags.timeLimitSec)
	fmt.Printf("  Parallel:   %v\n", flags.parallel)
	fmt.Println()

	requireHiGHS("")
	fmt.Print("  Solving... ")
	os.Stdout.Sync()

	result, err := cvrpilp.RunBenchmark(ds, cvrpilp.BenchmarkConfig{
		Instance:  ds.Name,
		TimeLimit: time.Duration(flags.timeLimitSec) * time.Second,
		Parallel:  flags.parallel,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark failed: %v\n", err)
		os.Exit(1)
	}

	outputDir, finalizeErr := cvrpilp.FinalizeBenchmark(ds, result, cvrpilp.FinalizeOptions{
		RunLabel:     flags.runLabel,
		Storage:      runoutputStorage(flags.storage),
		TimeLimitSec: flags.timeLimitSec,
	})
	if finalizeErr != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark output failed: %v\n", finalizeErr)
		os.Exit(1)
	}

	fmt.Println("done.")
	fmt.Println()
	printRoutingILPResult(flags.disp, result.Status, result.Objective, result.LowerBound, result.GapPercent,
		result.RuntimeSeconds, result.Variables, result.Constraints, result.Vehicles, result.Notes)
	finishRoutingILP(outputDir)
}

func runBenchmarkVRPTWILP() {
	flags := parseRoutingILPFlags(os.Args[2:])
	if flags.instancePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --instance <path.txt> is required")
		os.Exit(1)
	}

	fmt.Println(flags.disp.Heading(cli.EmojiConfig, "VRPTW ILP Benchmark"))
	fmt.Println()

	ds, err := vrptw.LoadDataset(flags.instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Instance:   %s\n", flags.disp.Bold(ds.Name))
	fmt.Printf("  Customers:  %d\n", len(ds.Customers))
	fmt.Printf("  Capacity:   %d\n", ds.Capacity)
	fmt.Printf("  Time Limit: %ds\n", flags.timeLimitSec)
	fmt.Printf("  Parallel:   %v\n", flags.parallel)
	fmt.Println()

	requireHiGHS("")
	fmt.Print("  Solving... ")
	os.Stdout.Sync()

	result, err := vrptwilp.RunBenchmark(ds, vrptwilp.BenchmarkConfig{
		Instance:  ds.Name,
		TimeLimit: time.Duration(flags.timeLimitSec) * time.Second,
		Parallel:  flags.parallel,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark failed: %v\n", err)
		os.Exit(1)
	}

	outputDir, finalizeErr := vrptwilp.FinalizeBenchmark(ds, result, vrptwilp.FinalizeOptions{
		RunLabel:     flags.runLabel,
		Storage:      runoutputStorage(flags.storage),
		TimeLimitSec: flags.timeLimitSec,
	})
	if finalizeErr != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark output failed: %v\n", finalizeErr)
		os.Exit(1)
	}

	fmt.Println("done.")
	fmt.Println()
	printRoutingILPResult(flags.disp, result.Status, result.Objective, result.LowerBound, result.GapPercent,
		result.RuntimeSeconds, result.Variables, result.Constraints, result.Vehicles, result.Notes)
	finishRoutingILP(outputDir)
}

func runBenchmarkJSSILP() {
	flags := parseRoutingILPFlags(os.Args[2:])
	if flags.instancePath == "" {
		fmt.Fprintln(os.Stderr, "Error: --instance <path.txt> is required")
		os.Exit(1)
	}

	fmt.Println(flags.disp.Heading(cli.EmojiConfig, "JSS ILP Benchmark"))
	fmt.Println()

	ds, err := jobshop.LoadDataset(flags.instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Instance:   %s\n", flags.disp.Bold(ds.Name))
	fmt.Printf("  Jobs:       %d\n", ds.Jobs)
	fmt.Printf("  Machines:   %d\n", ds.Machines)
	fmt.Printf("  Time Limit: %ds\n", flags.timeLimitSec)
	fmt.Printf("  Parallel:   %v\n", flags.parallel)
	fmt.Println()

	requireHiGHS("")
	fmt.Print("  Solving... ")
	os.Stdout.Sync()

	result, err := jssilp.RunBenchmark(ds, jssilp.BenchmarkConfig{
		Instance:  ds.Name,
		TimeLimit: time.Duration(flags.timeLimitSec) * time.Second,
		Parallel:  flags.parallel,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark failed: %v\n", err)
		os.Exit(1)
	}

	outputDir, finalizeErr := jssilp.FinalizeBenchmark(ds, result, jssilp.FinalizeOptions{
		RunLabel:     flags.runLabel,
		InstancePath: flags.instancePath,
		Storage:      runoutputStorage(flags.storage),
		TimeLimitSec: flags.timeLimitSec,
	})
	if finalizeErr != nil {
		fmt.Fprintf(os.Stderr, "\nBenchmark output failed: %v\n", finalizeErr)
		os.Exit(1)
	}

	fmt.Println("done.")
	fmt.Println()
	printJSSILPResult(flags.disp, result.Status, result.Objective, result.LowerBound, result.GapPercent,
		result.RuntimeSeconds, result.Variables, result.Constraints, result.Operations, result.Notes)
	finishRoutingILP(outputDir)
}

func finishRoutingILP(outputDir string) {
	if outputDir != "" {
		fmt.Printf("  Output: %s/\n", outputDir)
	}
	fmt.Println("Done.")
}
