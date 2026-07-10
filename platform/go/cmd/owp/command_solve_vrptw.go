package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runSolveVRPTW() {
	args := os.Args[2:]
	instancePath := requireInstanceFlag(args, "  owp solve-vrptw --instance <path.txt> [--mode sa|lahc|tabu|portfolio] [--iterations <n>] [--seed <s>] [--run-label <name>] [--worker-decision-mode off|shadow|assist|adaptive]")

	opts := parseSearchSolveOptions(args, "sa", 500000, 100.0, 42)
	disp := parseDisplayOptions(args)

	fmt.Println(disp.Heading(cli.EmojiConfig, "VRPTW Solver ("+strings.ToUpper(opts.Mode)+")"))
	fmt.Println()

	ds, err := vrptw.LoadDataset(instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Instance:   %s\n", disp.Bold(ds.Name))
	fmt.Printf("  Customers:  %d\n", len(ds.Customers))
	fmt.Printf("  Capacity:   %d\n", ds.Capacity)
	fmt.Printf("  Vehicles:   %d\n", ds.Vehicles)
	fmt.Printf("  Horizon:    [%d, %d]\n", ds.Depot.ReadyTime, ds.Depot.DueDate)
	fmt.Printf("  Mode:       %s\n", strings.ToUpper(opts.Mode))
	fmt.Printf("  Iterations: %dK\n", opts.Iterations/1000)
	fmt.Printf("  Seed:       %d\n", opts.Seed)
	fmt.Println()

	problem := vrptw.NewVRPTWProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineDistance := problem.TotalDistance(baselineSol)
	fmt.Printf("  Constructive: %d distance, %d vehicles, feasible=%v\n",
		baselineDistance, problem.RouteCount(baselineSol), problem.IsFeasible(baselineSol))

	config := opts.BuildSearchConfig("vrptw", ds.Name, func(cfg *optimisation.SearchConfig) {
		cfg.MinTemperature = 0.001
		cfg.Portfolio = []string{"sa", "lahc", "tabu"}
	})
	workerDecisionMode := applySearchIntelligenceFlags(args, &config, searchIntelligenceOpts{})

	fmt.Printf("  Running %s... ", strings.ToUpper(opts.Mode))
	os.Stdout.Sync()

	outcome := runSearchSolve(opts.Mode, nil, portfolioRunParams{
		Problem: problem, Config: config, WorkerDecisionMode: workerDecisionMode,
		Domain: "vrptw", Instance: ds.Name, PortfolioModelPath: opts.PortfolioModelPath,
	})

	fmt.Println("done.")
	fmt.Println()

	bestDistance := problem.TotalDistance(outcome.Result.BestSolution)
	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Distance:    %d\n", bestDistance)
	fmt.Printf("  Vehicles:    %d\n", problem.RouteCount(outcome.Result.BestSolution))
	fmt.Printf("  Feasible:    %v\n", problem.IsFeasible(outcome.Result.BestSolution))
	printImprovementPct(baselineDistance, bestDistance)
	printSearchResultStats(outcome.Result)
	fmt.Println()

	if opts.RunLabel != "" {
		finalizeVRPTWRun(opts, config, workerDecisionMode, instancePath, ds, problem, outcome, bestDistance, baselineDistance,
			problem.RouteCount(outcome.Result.BestSolution), problem.IsFeasible(outcome.Result.BestSolution))
	}

	fmt.Println("Done.")
}

func finalizeVRPTWRun(opts SearchSolveOptions, config optimisation.SearchConfig, workerDecisionMode string, instancePath string, ds *vrptw.Dataset, problem *vrptw.VRPTWProblem, outcome searchRunOutcome, bestDistance, baselineDistance, bestVehicles int, bestFeasible bool) {
	outputDir := ensureRunOutputDir(opts.RunLabel)
	instanceName := strings.TrimSuffix(filepath.Base(instancePath), filepath.Ext(instancePath))
	solJSON, _ := problem.SerializeSolution(outcome.Result.BestSolution)

	extra := map[string][]byte{}
	if disc := vrptw.BuildDiscoveriesCSV(outcome.Result.Discoveries); disc != nil {
		extra["discoveries.csv"] = disc
	}

	finalizeGenericSolverRun(genericSolverRunOutput{
		OutputDir: outputDir,
		RunMeta: map[string]interface{}{
			"problemType":     "vrptw",
			"mode":            opts.Mode,
			"instance":        instanceName,
			"customers":       len(ds.Customers),
			"capacity":        ds.Capacity,
			"vehicles":        ds.Vehicles,
			"iterations":      opts.Iterations,
			"seed":            opts.Seed,
			"runLabel":        opts.RunLabel,
			"bestObjective":   bestDistance,
			"bestDistance":    bestDistance,
			"initialDistance": baselineDistance,
			"bestVehicles":    bestVehicles,
			"feasible":        bestFeasible,
			"runtimeMs":       outcome.Result.DurationMs,
			"policyMode":      config.PolicyMode,
			"policyDir":       config.PolicyDir,
		},
		SolutionJSON: solJSON,
		ExtraFiles:   extra,
		Telemetry: solverTelemetryInput{
			OutputDir: outputDir, ProblemType: "vrptw", Instance: instanceName, Algorithm: opts.Mode,
			Seed: opts.Seed, Temperature: opts.Temperature, Iterations: opts.Iterations,
			AssistMode: workerDecisionMode,
			Result: outcome.Result, PortfolioRecorder: outcome.Recorder,
			PolicyMode: config.PolicyMode, PolicyDir: config.PolicyDir,
		},
		Storage: opts.Storage, RunLabel: opts.RunLabel, Algorithm: opts.Mode, Penalty: bestDistance,
	})
	fmt.Printf("  Output: %s/\n", outputDir)
}
