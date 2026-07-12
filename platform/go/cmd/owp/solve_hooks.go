package main

import (
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/cvrp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/jobshop"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/vrptw"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk"
)

func init() {
	solveHooks["cvrp"] = solveDomainHooks{
		title:        "CVRP Solver",
		policyDomain: "cvrp",
		intelOpts:    searchIntelligenceOpts{PrintPolicyDir: true},
		printHeader:  printCVRPHeaderFromMeta,
		printConstructive: func(problem optimisation.Problem, _ sdk.InstanceMeta, baseline int) {
			fmt.Printf("  Constructive baseline: %d\n", baseline)
		},
		printResults: func(disp cli.Options, problem optimisation.Problem, _ sdk.InstanceMeta, _ SearchSolveOptions, outcome searchRunOutcome, baseline int) {
			printGenericSearchResults(disp, problem, outcome, baseline, "Distance")
		},
		finalize: finalizeCVRPFromMeta,
	}

	solveHooks["vrptw"] = solveDomainHooks{
		title:        "VRPTW Solver",
		policyDomain: "vrptw",
		configureSearch: func(cfg *optimisation.SearchConfig) {
			cfg.MinTemperature = 0.001
			cfg.Portfolio = []string{"sa", "lahc", "tabu", "ga"}
		},
		printHeader: printGenericMetaHeader,
		printConstructive: func(problem optimisation.Problem, _ sdk.InstanceMeta, baseline int) {
			p := problem.(*vrptw.VRPTWProblem)
			sol, _ := problem.CreateInitialSolution()
			fmt.Printf("  Constructive: %d distance, %d vehicles, feasible=%v\n",
				baseline, p.RouteCount(sol), p.IsFeasible(sol))
		},
		printResults: printVRPTWResults,
		finalize:     finalizeVRPTWFromMeta,
	}

	solveHooks["jobshop"] = solveDomainHooks{
		title:             "Job Shop Solver",
		policyDomain:      "jss",
		extraPortfolio:    []string{"adaptive"},
		portfolioInstance: func(_ sdk.InstanceMeta, instancePath string) string { return instancePath },
		configureSearch: func(cfg *optimisation.SearchConfig) {
			cfg.MinTemperature = 0.001
			cfg.TabuNeighbourhood = 50
			cfg.Portfolio = []string{"sa", "lahc"}
		},
		printHeader: printGenericMetaHeader,
		printConstructive: func(_ optimisation.Problem, _ sdk.InstanceMeta, baseline int) {
			fmt.Printf("  Constructive baseline: %d\n", baseline)
		},
		printResults: func(disp cli.Options, problem optimisation.Problem, _ sdk.InstanceMeta, _ SearchSolveOptions, outcome searchRunOutcome, baseline int) {
			if outcome.Portfolio != nil {
				printGenericSearchResults(disp, problem, outcome, baseline, "Makespan")
				return
			}
			fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
			fmt.Println()
			fmt.Printf("  Makespan:    %d\n", outcome.Result.BestPenalty)
			printImprovementPct(baseline, outcome.Result.BestPenalty)
			printSearchResultStats(outcome.Result)
			fmt.Println()
		},
		finalize: finalizeJobShopFromMeta,
	}
}

func printCVRPHeaderFromMeta(disp cli.Options, meta sdk.InstanceMeta, opts SearchSolveOptions, modeLabel string) {
	ds, _ := meta.Data.(*cvrp.Dataset)
	if ds != nil {
		printCVRPHeader(disp, ds, opts, modeLabel)
		return
	}
	printGenericMetaHeader(disp, meta, opts, modeLabel)
}

func printVRPTWResults(disp cli.Options, problem optimisation.Problem, meta sdk.InstanceMeta, _ SearchSolveOptions, outcome searchRunOutcome, baseline int) {
	p := problem.(*vrptw.VRPTWProblem)
	if outcome.Portfolio != nil {
		printPortfolioResults(disp, *outcome.Portfolio, baseline, "Best")
		printPortfolioWinner(disp, *outcome.Portfolio, problem, baseline, "Distance")
		return
	}
	bestDistance := p.TotalDistance(outcome.Result.BestSolution)
	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  Distance:    %d\n", bestDistance)
	fmt.Printf("  Vehicles:    %d\n", p.RouteCount(outcome.Result.BestSolution))
	fmt.Printf("  Feasible:    %v\n", p.IsFeasible(outcome.Result.BestSolution))
	printImprovementPct(baseline, bestDistance)
	printSearchResultStats(outcome.Result)
	fmt.Println()
	_ = meta
}

func finalizeCVRPFromMeta(opts SearchSolveOptions, config optimisation.SearchConfig, workerDecisionMode string, _ string, meta sdk.InstanceMeta, problem optimisation.Problem, outcome searchRunOutcome) {
	ds, ok := meta.Data.(*cvrp.Dataset)
	if !ok {
		return
	}
	finalizeCVRPRun(opts, config, workerDecisionMode, ds, problem, outcome)
}

func finalizeVRPTWFromMeta(opts SearchSolveOptions, config optimisation.SearchConfig, workerDecisionMode string, instancePath string, meta sdk.InstanceMeta, problem optimisation.Problem, outcome searchRunOutcome) {
	ds, ok := meta.Data.(*vrptw.Dataset)
	if !ok {
		return
	}
	p := problem.(*vrptw.VRPTWProblem)
	bestDistance := p.TotalDistance(outcome.Result.BestSolution)
	baselineSol, _ := problem.CreateInitialSolution()
	baselineDistance := p.TotalDistance(baselineSol)
	finalizeVRPTWRun(opts, config, workerDecisionMode, instancePath, ds, p, outcome, bestDistance, baselineDistance,
		p.RouteCount(outcome.Result.BestSolution), p.IsFeasible(outcome.Result.BestSolution))
}

func finalizeJobShopFromMeta(opts SearchSolveOptions, config optimisation.SearchConfig, workerDecisionMode string, instancePath string, meta sdk.InstanceMeta, problem optimisation.Problem, outcome searchRunOutcome) {
	ds, ok := meta.Data.(*jobshop.Dataset)
	if !ok {
		return
	}
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
	_ = meta
}
