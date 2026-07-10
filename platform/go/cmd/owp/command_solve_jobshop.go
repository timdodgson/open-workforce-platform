package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runSolveJobShop() {
	args := os.Args[2:]
	instancePath := requireInstanceFlag(args, "")

	opts := parseSearchSolveOptions(args, "sa", 500000, 100.0, 42)
	disp := parseDisplayOptions(args)

	fmt.Println(disp.Heading(cli.EmojiConfig, "Job Shop Solver ("+strings.ToUpper(opts.Mode)+")"))
	fmt.Println()

	ds, err := jobshop.LoadDataset(instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	instanceName := strings.TrimSuffix(filepath.Base(instancePath), filepath.Ext(instancePath))
	fmt.Printf("  Instance:   %s\n", instancePath)
	fmt.Printf("  Jobs:       %d\n", ds.Jobs)
	fmt.Printf("  Machines:   %d\n", ds.Machines)
	fmt.Printf("  Mode:       %s\n", strings.ToUpper(opts.Mode))
	fmt.Printf("  Iterations: %dK\n", opts.Iterations/1000)
	fmt.Printf("  Seed:       %d\n", opts.Seed)
	fmt.Println()

	problem := jobshop.NewJSSProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineMakespan := problem.Evaluate(baselineSol)
	fmt.Printf("  Constructive baseline: %d\n", baselineMakespan)

	config := opts.BuildSearchConfig("jss", instanceName, func(cfg *optimisation.SearchConfig) {
		cfg.MinTemperature = 0.001
		cfg.TabuNeighbourhood = 50
		cfg.Portfolio = []string{"sa", "lahc"}
	})
	workerDecisionMode := applySearchIntelligenceFlags(args, &config, searchIntelligenceOpts{})

	fmt.Printf("  Running %s... ", strings.ToUpper(opts.Mode))
	os.Stdout.Sync()

	outcome := runSearchSolve(opts.Mode, []string{"adaptive"}, portfolioRunParams{
		Problem: problem, Config: config, WorkerDecisionMode: workerDecisionMode,
		Domain: "jss", Instance: instancePath, PortfolioModelPath: opts.PortfolioModelPath,
	})

	fmt.Println("done.")
	fmt.Println()

	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Makespan:    %d\n", outcome.Result.BestPenalty)
	printImprovementPct(baselineMakespan, outcome.Result.BestPenalty)
	printSearchResultStats(outcome.Result)
	fmt.Println()

	if opts.RunLabel != "" {
		outputDir := ensureRunOutputDir(opts.RunLabel)
		solJSON, _ := problem.SerializeSolution(outcome.Result.BestSolution)

		finalizeGenericSolverRun(genericSolverRunOutput{
			OutputDir: outputDir,
			RunMeta: map[string]interface{}{
				"problemType":     "jss",
				"mode":            opts.Mode,
				"instance":        instancePath,
				"jobs":            ds.Jobs,
				"machines":        ds.Machines,
				"iterations":      opts.Iterations,
				"seed":            opts.Seed,
				"runLabel":        opts.RunLabel,
				"bestObjective":   outcome.Result.BestPenalty,
				"bestMakespan":    outcome.Result.BestPenalty,
				"initialMakespan": outcome.Result.InitialPenalty,
				"runtimeMs":       outcome.Result.DurationMs,
				"policyMode":      config.PolicyMode,
				"policyDir":       config.PolicyDir,
			},
			SolutionJSON: solJSON,
			Telemetry: solverTelemetryInput{
				OutputDir: outputDir, ProblemType: "jss", Instance: instancePath, Algorithm: opts.Mode,
				Seed: opts.Seed, Temperature: opts.Temperature, Iterations: opts.Iterations,
				AssistMode: workerDecisionMode,
				Result: outcome.Result, PortfolioRecorder: outcome.Recorder,
				PolicyMode: config.PolicyMode, PolicyDir: config.PolicyDir,
			},
			Storage: opts.Storage, RunLabel: opts.RunLabel, Algorithm: opts.Mode, Penalty: outcome.Result.BestPenalty,
		})
		fmt.Printf("  Output: %s/\n", outputDir)
	}

	fmt.Println("Done.")
}
