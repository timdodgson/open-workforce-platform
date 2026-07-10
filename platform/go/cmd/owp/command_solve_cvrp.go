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
	runSolveDomain("cvrp", os.Args[2:])
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
