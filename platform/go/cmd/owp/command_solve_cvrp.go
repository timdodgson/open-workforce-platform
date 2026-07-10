package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runSolveCVRP() {
	args := os.Args[2:]
	instancePath := requireInstanceFlag(args, "  owp solve-cvrp --instance <path.vrp> [--mode sa|lahc|tabu|portfolio] [--iterations <n>] [--temperature <t>] [--seed <s>] [--run-label <name>]")

	opts := parseSearchSolveOptions(args, "sa", 500000, 100.0, 42)
	disp := parseDisplayOptions(args)
	modeLabel := searchModeLabel(opts.Mode)

	fmt.Println(disp.Heading(cli.EmojiConfig, "CVRP Solver ("+modeLabel+")"))
	fmt.Println()

	ds, err := cvrp.LoadDataset(instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	printCVRPHeader(disp, ds, opts, modeLabel)

	problem := cvrp.NewCVRPProblem(ds)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)
	fmt.Printf("  Constructive baseline: %d\n", baselineCost)

	config := opts.BuildSearchConfig("cvrp", ds.Name, nil)
	workerDecisionMode := applySearchIntelligenceFlags(args, &config, searchIntelligenceOpts{PrintPolicyDir: true})

	fmt.Printf("  Running %s... ", modeLabel)
	os.Stdout.Sync()

	outcome := runSearchSolve(opts.Mode, nil, portfolioRunParams{
		Problem: problem, Config: config, WorkerDecisionMode: workerDecisionMode,
		Domain: "cvrp", Instance: ds.Name, PortfolioModelPath: opts.PortfolioModelPath,
	})

	fmt.Println("done.")
	fmt.Println()

	if outcome.Portfolio != nil {
		printPortfolioResults(disp, *outcome.Portfolio, baselineCost, "Best")
		printPortfolioWinner(disp, *outcome.Portfolio, problem, baselineCost, "Distance")
	} else {
		finalCost := problem.Evaluate(outcome.Result.BestSolution)
		fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
		fmt.Println()
		fmt.Printf("  Distance:        %d\n", finalCost)
		fmt.Printf("  Feasible:        %v\n", finalCost == outcome.Result.BestPenalty)
		printImprovementPct(baselineCost, finalCost)
		printSearchResultStats(outcome.Result)
		fmt.Println()
	}

	if opts.RunLabel != "" {
		finalizeCVRPRun(opts, config, workerDecisionMode, ds, problem, outcome)
	}

	fmt.Println("Done.")
}

func printCVRPHeader(disp cli.Options, ds *cvrp.Dataset, opts SearchSolveOptions, modeLabel string) {
	fmt.Printf("  Instance:   %s\n", disp.Bold(ds.Name))
	fmt.Printf("  Customers:  %d\n", len(ds.Customers))
	fmt.Printf("  Capacity:   %d\n", ds.Capacity)
	fmt.Printf("  Mode:       %s\n", modeLabel)
	fmt.Printf("  Iterations: %dK\n", opts.Iterations/1000)
	switch opts.Mode {
	case "sa":
		fmt.Printf("  Temperature: %.1f\n", opts.Temperature)
	case "lahc":
		fmt.Printf("  LAHC Length: %d\n", opts.LateAcceptanceLength)
	case "tabu":
		fmt.Printf("  Tabu Tenure: %d\n", opts.TabuTenure)
		fmt.Printf("  Neighbourhood: %d\n", opts.TabuNeighbourhood)
	case "portfolio":
		if len(opts.Portfolio) > 0 {
			fmt.Printf("  Strategies: %s\n", strings.Join(opts.Portfolio, ", "))
		} else {
			fmt.Printf("  Strategies: sa, lahc, tabu (default)\n")
		}
	}
	fmt.Printf("  Seed:       %d\n", opts.Seed)
	fmt.Println()
}

func finalizeCVRPRun(opts SearchSolveOptions, config optimisation.SearchConfig, workerDecisionMode string, ds *cvrp.Dataset, problem optimisation.Problem, outcome searchRunOutcome) {
	outputDir := ensureRunOutputDir(opts.RunLabel)
	solJSON, _ := problem.SerializeSolution(outcome.Result.BestSolution)

	finalizeGenericSolverRun(genericSolverRunOutput{
		OutputDir: outputDir,
		RunMeta: map[string]interface{}{
			"problemType":     "cvrp",
			"mode":            opts.Mode,
			"winnerStrategy":  outcome.WinnerMode,
			"instance":        ds.Name,
			"customers":       len(ds.Customers),
			"capacity":        ds.Capacity,
			"iterations":      opts.Iterations,
			"seed":            opts.Seed,
			"runLabel":        opts.RunLabel,
			"bestObjective":   outcome.Result.BestPenalty,
			"bestDistance":    outcome.Result.BestPenalty,
			"initialDistance": outcome.Result.InitialPenalty,
			"runtimeMs":       outcome.Result.DurationMs,
			"feasible":        outcome.Result.BestPenalty == problem.Evaluate(outcome.Result.BestSolution),
			"policyMode":      config.PolicyMode,
			"policyDir":       config.PolicyDir,
		},
		SolutionJSON: solJSON,
		ExtraFiles: map[string][]byte{
			"results.csv": cvrp.BuildResultsCSV(cvrp.ResultsCSVParams{
				Instance: ds.Name, Seed: opts.Seed, WinnerMode: outcome.WinnerMode,
				Iterations: opts.Iterations, Temperature: opts.Temperature, SearchResult: outcome.Result,
			}),
			"discoveries.csv": cvrp.BuildDiscoveriesCSV(cvrp.DiscoveriesCSVParams{
				RunLabel: opts.RunLabel, Instance: ds.Name, Seed: opts.Seed,
				Iterations: opts.Iterations, Temperature: opts.Temperature, Discoveries: outcome.Result.Discoveries,
			}),
		},
		Telemetry: solverTelemetryInput{
			OutputDir: outputDir, ProblemType: "cvrp", Instance: ds.Name, Algorithm: opts.Mode,
			Seed: opts.Seed, Temperature: opts.Temperature, Iterations: opts.Iterations,
			AssistMode: workerDecisionMode,
			Result: outcome.Result, PortfolioRecorder: outcome.Recorder,
			PolicyMode: config.PolicyMode, PolicyDir: config.PolicyDir,
		},
		Storage: opts.Storage, RunLabel: opts.RunLabel, Algorithm: opts.Mode, Penalty: outcome.Result.BestPenalty,
	})

	fmt.Printf("  Output: %s/ (run.json, solution.json, results.csv, discoveries.csv)\n", outputDir)
}
