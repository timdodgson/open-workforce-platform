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

	runLabel := parseRunLabelFlag(args, false)
	storage := parseStorageConfig(args, false)

	mode := parseSearchMode(args, "sa")
	iterations := parseSearchIterations(args, 500000)
	temperature := parseSearchTemperature(args, 100.0)
	seed := parseSearchSeed(args, 42)

	lateAcceptanceLength := parseIntFlag(args, "--late-acceptance-length")
	if lateAcceptanceLength <= 0 {
		lateAcceptanceLength = 1000
	}

	tabuTenure := parseIntFlag(args, "--tabu-tenure")
	if tabuTenure <= 0 {
		tabuTenure = 7
	}

	tabuNeighbourhood := parseIntFlag(args, "--tabu-neighbourhood")
	if tabuNeighbourhood <= 0 {
		tabuNeighbourhood = 100
	}

	disp := parseDisplayOptions(args)

	modeLabel := searchModeLabel(mode)

	// Parse portfolio strategies.
	portfolioStr := parseStringFlag(args, "--portfolio")
	var portfolio []string
	if portfolioStr != "" {
		portfolio = strings.Split(portfolioStr, ",")
	}

	fmt.Println(disp.Heading(cli.EmojiConfig, "CVRP Solver ("+modeLabel+")"))
	fmt.Println()

	// Load dataset.
	ds, err := cvrp.LoadDataset(instancePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading instance: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  Instance:   %s\n", disp.Bold(ds.Name))
	fmt.Printf("  Customers:  %d\n", len(ds.Customers))
	fmt.Printf("  Capacity:   %d\n", ds.Capacity)
	fmt.Printf("  Mode:       %s\n", modeLabel)
	fmt.Printf("  Iterations: %dK\n", iterations/1000)
	if mode == "sa" {
		fmt.Printf("  Temperature: %.1f\n", temperature)
	} else if mode == "lahc" {
		fmt.Printf("  LAHC Length: %d\n", lateAcceptanceLength)
	} else if mode == "tabu" {
		fmt.Printf("  Tabu Tenure: %d\n", tabuTenure)
		fmt.Printf("  Neighbourhood: %d\n", tabuNeighbourhood)
	} else if mode == "portfolio" {
		if len(portfolio) > 0 {
			fmt.Printf("  Strategies: %s\n", strings.Join(portfolio, ", "))
		} else {
			fmt.Printf("  Strategies: sa, lahc, tabu (default)\n")
		}
	}
	fmt.Printf("  Seed:       %d\n", seed)
	fmt.Println()

	// Create problem instance.
	problem := cvrp.NewCVRPProblem(ds)

	// Get constructive baseline for comparison.
	baselineSol, _ := problem.CreateInitialSolution()
	baselineCost := problem.Evaluate(baselineSol)
	fmt.Printf("  Constructive baseline: %d\n", baselineCost)

	// Run search via generic search engine.
	config := optimisation.SearchConfig{
		Mode:                 mode,
		Iterations:           iterations,
		InitialTemperature:   temperature,
		MinTemperature:       0.0001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: lateAcceptanceLength,
		TabuTenure:           tabuTenure,
		TabuNeighbourhood:    tabuNeighbourhood,
		Portfolio:            portfolio,
		Seed:                 seed,
		PolicyDomain:         "cvrp",
		PolicyInstance:       ds.Name,
	}

	// Worker decision mode (search intelligence).
	workerDecisionMode := applySearchIntelligenceFlags(args, &config, searchIntelligenceOpts{PrintPolicyDir: true})

	// Run search and capture result for both display and file output.
	var searchResult optimisation.SearchResult
	var winnerMode string
	var portfolioRecorder *optimisation.PortfolioAssistRecorder

	if mode == "portfolio" {
		fmt.Print("  Running portfolio... ")
		os.Stdout.Sync()

		assistConfig := optimisation.PortfolioAssistConfig{
			Mode:     workerDecisionMode,
			Domain:   "cvrp",
			Instance: ds.Name,
			ModelPath: parseStringFlag(args, "--portfolio-model"),
		}
		pr, recorder := optimisation.RunPortfolioWithAssist(problem, config, assistConfig)
		portfolioRecorder = recorder

		fmt.Println("done.")
		fmt.Println()

		searchResult = pr.BestResult
		winnerMode = pr.Winner

		// Per-strategy table.
		fmt.Println(disp.Heading(cli.EmojiValid, "Per-Strategy Results"))
		fmt.Println()
		fmt.Printf("  %-8s %10s %10s %10s %8s\n", "Mode", "Best", "Improve%", "Candidates", "Runtime")
		fmt.Printf("  %-8s %10s %10s %10s %8s\n", "────────", "──────────", "──────────", "──────────", "────────")
		for _, e := range pr.Entries {
			impPct := float64(baselineCost-e.Result.BestPenalty) / float64(baselineCost) * 100
			winner := " "
			if e.Mode == pr.Winner {
				winner = "★"
			}
			fmt.Printf(" %s%-8s %10d %9.1f%% %10d %6dms\n",
				winner, strings.ToUpper(e.Mode), e.Result.BestPenalty, impPct, e.Result.Candidates, e.Result.DurationMs)
		}
		fmt.Println()

		// Winner summary.
		finalCost := problem.Evaluate(pr.BestResult.BestSolution)
		feasible := finalCost == pr.BestResult.BestPenalty
		fmt.Println(disp.Heading(cli.EmojiValid, "Winner: "+strings.ToUpper(pr.Winner)))
		fmt.Println()
		fmt.Printf("  Distance:        %d\n", finalCost)
		fmt.Printf("  Feasible:        %v\n", feasible)
		printImprovementPct(baselineCost, finalCost)
		fmt.Println()
	} else {
		fmt.Printf("  Running %s... ", modeLabel)
		os.Stdout.Sync()

		searchResult = optimisation.RunSearch(problem, config)
		winnerMode = mode

		fmt.Println("done.")
		fmt.Println()

		finalCost := problem.Evaluate(searchResult.BestSolution)
		feasible := finalCost == searchResult.BestPenalty

		fmt.Println(disp.Heading(cli.EmojiValid, "Result"))
		fmt.Println()
		fmt.Printf("  Distance:        %d\n", finalCost)
		fmt.Printf("  Feasible:        %v\n", feasible)
		printImprovementPct(baselineCost, finalCost)
		fmt.Printf("  Initial:         %d\n", searchResult.InitialPenalty)
		fmt.Printf("  Final:           %d\n", finalCost)
		fmt.Printf("  Runtime:         %dms\n", searchResult.DurationMs)
		fmt.Printf("  Candidates:      %d\n", searchResult.Candidates)
		fmt.Printf("  Accepted:        %d (%.1f%%)\n", searchResult.Accepted,
			float64(searchResult.Accepted)/float64(searchResult.Candidates)*100)
		fmt.Printf("  Hard rejected:   %d\n", searchResult.Rejected)
		fmt.Printf("  Improvements:    %d\n", searchResult.Improved)
		fmt.Println()
	}

	// Write output files if --run-label is specified.
	if runLabel != "" {
		outputDir := ensureRunOutputDir(runLabel)
		solJSON, _ := problem.SerializeSolution(searchResult.BestSolution)

		finalizeGenericSolverRun(genericSolverRunOutput{
			OutputDir: outputDir,
			RunMeta: map[string]interface{}{
				"problemType":     "cvrp",
				"mode":            mode,
				"winnerStrategy":  winnerMode,
				"instance":        ds.Name,
				"customers":       len(ds.Customers),
				"capacity":        ds.Capacity,
				"iterations":      iterations,
				"seed":            seed,
				"runLabel":        runLabel,
				"bestObjective":   searchResult.BestPenalty,
				"bestDistance":    searchResult.BestPenalty,
				"initialDistance": searchResult.InitialPenalty,
				"runtimeMs":       searchResult.DurationMs,
				"feasible":        searchResult.BestPenalty == problem.Evaluate(searchResult.BestSolution),
			},
			SolutionJSON: solJSON,
			ExtraFiles: map[string][]byte{
				"results.csv": cvrp.BuildResultsCSV(cvrp.ResultsCSVParams{
					Instance: ds.Name, Seed: seed, WinnerMode: winnerMode,
					Iterations: iterations, Temperature: temperature, SearchResult: searchResult,
				}),
				"discoveries.csv": cvrp.BuildDiscoveriesCSV(cvrp.DiscoveriesCSVParams{
					RunLabel: runLabel, Instance: ds.Name, Seed: seed,
					Iterations: iterations, Temperature: temperature, Discoveries: searchResult.Discoveries,
				}),
			},
			Telemetry: solverTelemetryInput{
				OutputDir: outputDir, ProblemType: "cvrp", Instance: ds.Name, Algorithm: mode,
				Seed: seed, Temperature: temperature, Iterations: iterations,
				Result: searchResult, PortfolioRecorder: portfolioRecorder,
				PolicyMode: config.PolicyMode, PolicyDir: config.PolicyDir,
			},
			Storage: storage, RunLabel: runLabel, Algorithm: mode, Penalty: searchResult.BestPenalty,
		})

		fmt.Printf("  Output: %s/ (run.json, solution.json, results.csv, discoveries.csv)\n", outputDir)
	}

	fmt.Println("Done.")
}
