package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk"
)

type solveDomainHooks struct {
	title              string
	policyDomain       string
	portfolioInstance  func(meta sdk.InstanceMeta, instancePath string) string
	extraPortfolio     []string
	intelOpts          searchIntelligenceOpts
	configureSearch    func(cfg *optimisation.SearchConfig)
	printHeader        func(disp cli.Options, meta sdk.InstanceMeta, opts SearchSolveOptions, modeLabel string)
	printConstructive  func(problem optimisation.Problem, meta sdk.InstanceMeta, baseline int)
	printResults       func(disp cli.Options, problem optimisation.Problem, meta sdk.InstanceMeta, opts SearchSolveOptions, outcome searchRunOutcome, baseline int)
	finalize           func(opts SearchSolveOptions, config optimisation.SearchConfig, workerDecisionMode string, instancePath string, meta sdk.InstanceMeta, problem optimisation.Problem, outcome searchRunOutcome)
}

var solveHooks = map[string]solveDomainHooks{}

func runSolve() {
	if len(os.Args) < 3 {
		printSolveUsage()
		os.Exit(1)
	}
	runSolveDomain(strings.ToLower(os.Args[2]), os.Args[3:])
}

func runSolveDomain(domain string, args []string) {
	desc, ok := sdk.GetProblem(domain)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown domain %q. Try: owp list-solvers\n", domain)
		os.Exit(1)
	}
	hooks, ok := solveHooks[domain]
	if !ok {
		hooks = defaultSolveHooks(desc)
	}

	usage := fmt.Sprintf("  owp solve %s --instance <path> [--mode sa|lahc|tabu|portfolio]", domain)
	if desc.Usage != "" {
		usage = "  " + desc.Usage
	}
	instancePath := requireInstanceFlag(args, usage)

	d := desc.Defaults
	opts := parseSearchSolveOptions(args, d.Mode, d.Iterations, d.Temperature, d.Seed)
	disp := parseDisplayOptions(args)
	modeLabel := searchModeLabel(opts.Mode)

	fmt.Println(disp.Heading(cli.EmojiConfig, hooks.title+" ("+modeLabel+")"))
	fmt.Println()

	problem, meta, err := desc.Load(instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}
	meta.InstancePath = instancePath

	hooks.printHeader(disp, meta, opts, modeLabel)

	baselineSol, _ := problem.CreateInitialSolution()
	baseline := problem.Evaluate(baselineSol)
	hooks.printConstructive(problem, meta, baseline)

	config := opts.BuildSearchConfig(hooks.policyDomain, meta.Name, hooks.configureSearch)
	workerDecisionMode := applySearchIntelligenceFlags(args, &config, hooks.intelOpts)

	portfolioInstance := meta.Name
	if hooks.portfolioInstance != nil {
		portfolioInstance = hooks.portfolioInstance(meta, instancePath)
	}

	fmt.Printf("  Running %s... ", modeLabel)
	os.Stdout.Sync()

	outcome := runSearchSolve(opts.Mode, hooks.extraPortfolio, portfolioRunParams{
		Problem: problem, Config: config, WorkerDecisionMode: workerDecisionMode,
		Domain: hooks.policyDomain, Instance: portfolioInstance, PortfolioModelPath: opts.PortfolioModelPath,
	})

	fmt.Println("done.")
	fmt.Println()

	hooks.printResults(disp, problem, meta, opts, outcome, baseline)

	if opts.RunLabel != "" {
		if hooks.finalize != nil {
			hooks.finalize(opts, config, workerDecisionMode, instancePath, meta, problem, outcome)
		} else {
			defaultFinalizeRun(opts, config, workerDecisionMode, instancePath, meta, problem, outcome, hooks.policyDomain)
		}
	}

	fmt.Println("Done.")
}

func printSolveUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  owp solve <domain> --instance <path> [search flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Domains:")
	for _, name := range sdk.Problems() {
		if desc, ok := sdk.GetProblem(name); ok {
			fmt.Fprintf(os.Stderr, "  %s", name)
			if desc.Command != "" {
				fmt.Fprintf(os.Stderr, " (alias: %s)", desc.Command)
			}
			fmt.Fprintln(os.Stderr)
		}
	}
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Run owp list-solvers for full details.")
}

func printGenericMetaHeader(disp cli.Options, meta sdk.InstanceMeta, opts SearchSolveOptions, modeLabel string) {
	fmt.Printf("  Instance:   %s\n", disp.Bold(meta.Name))
	for _, key := range sortedMetaKeys(meta.Fields) {
		fmt.Printf("  %s: %s\n", strings.Title(key), meta.Fields[key])
	}
	fmt.Printf("  Mode:       %s\n", modeLabel)
	fmt.Printf("  Iterations: %dK\n", opts.Iterations/1000)
	fmt.Printf("  Seed:       %d\n", opts.Seed)
	fmt.Println()
}

func sortedMetaKeys(fields map[string]string) []string {
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	// stable display order for known keys
	order := []string{"customers", "capacity", "vehicles", "horizon", "jobs", "machines"}
	ordered := make([]string, 0, len(fields))
	seen := map[string]bool{}
	for _, k := range order {
		if _, ok := fields[k]; ok {
			ordered = append(ordered, k)
			seen[k] = true
		}
	}
	for _, k := range keys {
		if !seen[k] {
			ordered = append(ordered, k)
		}
	}
	return ordered
}

func printGenericSearchResults(disp cli.Options, problem optimisation.Problem, outcome searchRunOutcome, baseline int, objectiveLabel string) {
	if outcome.Portfolio != nil {
		printPortfolioResults(disp, *outcome.Portfolio, baseline, "Best")
		printPortfolioWinner(disp, *outcome.Portfolio, problem, baseline, objectiveLabel)
		return
	}
	finalCost := problem.Evaluate(outcome.Result.BestSolution)
	fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
	fmt.Println()
	fmt.Printf("  %s:        %d\n", objectiveLabel, finalCost)
	fmt.Printf("  Feasible:        %v\n", finalCost == outcome.Result.BestPenalty)
	printImprovementPct(baseline, finalCost)
	printSearchResultStats(outcome.Result)
	fmt.Println()
}

func defaultSolveHooks(desc sdk.ProblemDescriptor) solveDomainHooks {
	title := desc.Title
	if title == "" {
		title = strings.ToUpper(desc.Name) + " Solver"
	}
	policyDomain := desc.PolicyDomain
	if policyDomain == "" {
		policyDomain = desc.Name
	}
	objective := desc.ObjectiveLabel
	if objective == "" {
		objective = "Objective"
	}
	return solveDomainHooks{
		title:        title,
		policyDomain: policyDomain,
		printHeader:  printGenericMetaHeader,
		printConstructive: func(_ optimisation.Problem, _ sdk.InstanceMeta, baseline int) {
			fmt.Printf("  Constructive baseline: %d\n", baseline)
		},
		printResults: func(disp cli.Options, problem optimisation.Problem, _ sdk.InstanceMeta, _ SearchSolveOptions, outcome searchRunOutcome, baseline int) {
			printGenericSearchResults(disp, problem, outcome, baseline, objective)
		},
	}
}

func defaultFinalizeRun(opts SearchSolveOptions, config optimisation.SearchConfig, workerDecisionMode string, instancePath string, meta sdk.InstanceMeta, problem optimisation.Problem, outcome searchRunOutcome, problemType string) {
	outputDir := ensureRunOutputDir(opts.RunLabel)
	solJSON, _ := problem.SerializeSolution(outcome.Result.BestSolution)
	finalCost := problem.Evaluate(outcome.Result.BestSolution)

	finalizeGenericSolverRun(genericSolverRunOutput{
		OutputDir: outputDir,
		RunMeta: map[string]interface{}{
			"problemType":      problemType,
			"mode":             opts.Mode,
			"winnerStrategy":   outcome.WinnerMode,
			"instance":         meta.Name,
			"instancePath":     instancePath,
			"iterations":       opts.Iterations,
			"seed":             opts.Seed,
			"runLabel":         opts.RunLabel,
			"bestObjective":    outcome.Result.BestPenalty,
			"initialObjective": outcome.Result.InitialPenalty,
			"runtimeMs":        outcome.Result.DurationMs,
			"feasible":         finalCost == outcome.Result.BestPenalty,
			"policyMode":       config.PolicyMode,
			"policyDir":        config.PolicyDir,
		},
		SolutionJSON: solJSON,
		Telemetry: solverTelemetryInput{
			OutputDir: outputDir, ProblemType: problemType, Instance: meta.Name, Algorithm: opts.Mode,
			Seed: opts.Seed, Temperature: opts.Temperature, Iterations: opts.Iterations,
			AssistMode: workerDecisionMode,
			Result: outcome.Result, PortfolioRecorder: outcome.Recorder,
			PolicyMode: config.PolicyMode, PolicyDir: config.PolicyDir,
		},
		Storage: opts.Storage, RunLabel: opts.RunLabel, Algorithm: opts.Mode, Penalty: outcome.Result.BestPenalty,
	})
	fmt.Printf("  Output: %s/ (run.json, solution.json)\n", outputDir)
}
