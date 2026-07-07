package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runSolveJobShop() {
	args := os.Args[2:]

	instancePath := requireInstanceFlag(args, "")

	mode := parseSearchMode(args, "sa")
	iterations := parseSearchIterations(args, 500000)
	seed := parseSearchSeed(args, 42)
	temperature := parseSearchTemperature(args, 100.0)

	runLabel := parseRunLabelFlag(args, false)
	storage := parseStorageConfig(args, false)

	disp := parseDisplayOptions(args)

	fmt.Println(disp.Heading(cli.EmojiConfig, "Job Shop Solver ("+strings.ToUpper(mode)+")"))
	fmt.Println()

	// Load dataset.
	ds, err := jobshop.LoadDataset(instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Instance:   %s\n", instancePath)
	fmt.Printf("  Jobs:       %d\n", ds.Jobs)
	fmt.Printf("  Machines:   %d\n", ds.Machines)
	fmt.Printf("  Mode:       %s\n", strings.ToUpper(mode))
	fmt.Printf("  Iterations: %dK\n", iterations/1000)
	fmt.Printf("  Seed:       %d\n", seed)
	fmt.Println()

	// Create problem.
	problem := jobshop.NewJSSProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineMakespan := problem.Evaluate(baselineSol)
	fmt.Printf("  Constructive baseline: %d\n", baselineMakespan)

	// Run search.
	config := optimisation.SearchConfig{
		Mode:                 mode,
		Iterations:           iterations,
		InitialTemperature:   temperature,
		MinTemperature:       0.001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 1000,
		TabuTenure:           7,
		TabuNeighbourhood:    50,
		Portfolio:            []string{"sa", "lahc"},
		Seed:                 seed,
	}

	workerDecisionMode := applySearchIntelligenceFlags(args, &config, searchIntelligenceOpts{})

	fmt.Printf("  Running %s... ", strings.ToUpper(mode))
	os.Stdout.Sync()

	var result optimisation.SearchResult
	var portfolioRecorder *optimisation.PortfolioAssistRecorder
	result, portfolioRecorder = runSearchOrPortfolio(mode, []string{"adaptive"}, portfolioRunParams{
		Problem: problem, Config: config, WorkerDecisionMode: workerDecisionMode,
		Domain: "jss", Instance: instancePath,
		PortfolioModelPath: parseStringFlag(args, "--portfolio-model"),
	})

	fmt.Println("done.")
	fmt.Println()

	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Makespan:    %d\n", result.BestPenalty)
	printImprovementPct(baselineMakespan, result.BestPenalty)
	printSearchResultStats(result)
	fmt.Println()

	// Write output if --run-label specified.
	if runLabel != "" {
		outputDir := ensureRunOutputDir(runLabel)
		solJSON, _ := problem.SerializeSolution(result.BestSolution)

		finalizeGenericSolverRun(genericSolverRunOutput{
			OutputDir: outputDir,
			RunMeta: map[string]interface{}{
				"problemType":     "jss",
				"mode":            mode,
				"instance":        instancePath,
				"jobs":            ds.Jobs,
				"machines":        ds.Machines,
				"iterations":      iterations,
				"seed":            seed,
				"runLabel":        runLabel,
				"bestObjective":   result.BestPenalty,
				"bestMakespan":    result.BestPenalty,
				"initialMakespan": result.InitialPenalty,
				"runtimeMs":       result.DurationMs,
			},
			SolutionJSON: solJSON,
			Telemetry: solverTelemetryInput{
				OutputDir: outputDir, ProblemType: "jss", Instance: instancePath, Algorithm: mode,
				Seed: seed, Temperature: temperature, Iterations: iterations,
				Result: result, PortfolioRecorder: portfolioRecorder,
			},
			Storage: storage, RunLabel: runLabel, Algorithm: mode, Penalty: result.BestPenalty,
		})

		fmt.Printf("  Output: %s/\n", outputDir)
	}

	fmt.Println("Done.")
}

// --- VRPTW Solver ---
