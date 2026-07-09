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

	mode := parseSearchMode(args, "sa")
	iterations := parseSearchIterations(args, 500000)
	seed := parseSearchSeed(args, 42)
	temperature := parseSearchTemperature(args, 100.0)

	runLabel := parseRunLabelFlag(args, false)
	storage := parseStorageConfig(args, false)

	disp := parseDisplayOptions(args)

	fmt.Println(disp.Heading(cli.EmojiConfig, "VRPTW Solver ("+strings.ToUpper(mode)+")"))
	fmt.Println()

	// Load dataset.
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
	fmt.Printf("  Mode:       %s\n", strings.ToUpper(mode))
	fmt.Printf("  Iterations: %dK\n", iterations/1000)
	fmt.Printf("  Seed:       %d\n", seed)
	fmt.Println()

	// Create problem.
	problem := vrptw.NewVRPTWProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineDistance := problem.TotalDistance(baselineSol)
	baselineVehicles := problem.RouteCount(baselineSol)
	baselineFeasible := problem.IsFeasible(baselineSol)

	fmt.Printf("  Constructive: %d distance, %d vehicles, feasible=%v\n", baselineDistance, baselineVehicles, baselineFeasible)

	// Run search.
	config := optimisation.SearchConfig{
		Mode:                 mode,
		Iterations:           iterations,
		InitialTemperature:   temperature,
		MinTemperature:       0.001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 1000,
		TabuTenure:           7,
		TabuNeighbourhood:    100,
		Portfolio:            []string{"sa", "lahc", "tabu"},
		Seed:                 seed,
		PolicyDomain:         "vrptw",
		PolicyInstance:       ds.Name,
	}

	workerDecisionMode := applySearchIntelligenceFlags(args, &config, searchIntelligenceOpts{})

	fmt.Printf("  Running %s... ", strings.ToUpper(mode))
	os.Stdout.Sync()

	var result optimisation.SearchResult
	var portfolioRecorder *optimisation.PortfolioAssistRecorder
	result, portfolioRecorder = runSearchOrPortfolio(mode, nil, portfolioRunParams{
		Problem: problem, Config: config, WorkerDecisionMode: workerDecisionMode,
		Domain: "vrptw", Instance: ds.Name,
		PortfolioModelPath: parseStringFlag(args, "--portfolio-model"),
	})

	fmt.Println("done.")
	fmt.Println()

	bestDistance := problem.TotalDistance(result.BestSolution)
	bestVehicles := problem.RouteCount(result.BestSolution)
	bestFeasible := problem.IsFeasible(result.BestSolution)

	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Distance:    %d\n", bestDistance)
	fmt.Printf("  Vehicles:    %d\n", bestVehicles)
	fmt.Printf("  Feasible:    %v\n", bestFeasible)
	printImprovementPct(baselineDistance, bestDistance)
	printSearchResultStats(result)
	fmt.Println()

	// Write output.
	if runLabel != "" {
		outputDir := ensureRunOutputDir(runLabel)

		instanceName := filepath.Base(instancePath)
		instanceName = strings.TrimSuffix(instanceName, filepath.Ext(instanceName))
		solJSON, _ := problem.SerializeSolution(result.BestSolution)

		extra := map[string][]byte{}
		if disc := vrptw.BuildDiscoveriesCSV(result.Discoveries); disc != nil {
			extra["discoveries.csv"] = disc
		}

		finalizeGenericSolverRun(genericSolverRunOutput{
			OutputDir: outputDir,
			RunMeta: map[string]interface{}{
				"problemType":     "vrptw",
				"mode":            mode,
				"instance":        instanceName,
				"customers":       len(ds.Customers),
				"capacity":        ds.Capacity,
				"vehicles":        ds.Vehicles,
				"iterations":      iterations,
				"seed":            seed,
				"runLabel":        runLabel,
				"bestObjective":   bestDistance,
				"bestDistance":    bestDistance,
				"initialDistance": baselineDistance,
				"bestVehicles":    bestVehicles,
				"feasible":        bestFeasible,
				"runtimeMs":       result.DurationMs,
				"policyMode":      config.PolicyMode,
				"policyDir":       config.PolicyDir,
			},
			SolutionJSON: solJSON,
			ExtraFiles:   extra,
			Telemetry: solverTelemetryInput{
				OutputDir: outputDir, ProblemType: "vrptw", Instance: instanceName, Algorithm: mode,
				Seed: seed, Temperature: temperature, Iterations: iterations,
				AssistMode: workerDecisionMode,
				Result: result, PortfolioRecorder: portfolioRecorder,
				PolicyMode: config.PolicyMode, PolicyDir: config.PolicyDir,
			},
			Storage: storage, RunLabel: runLabel, Algorithm: mode, Penalty: bestDistance,
		})

		fmt.Printf("  Output: %s/\n", outputDir)
	}

	fmt.Println("Done.")
}
