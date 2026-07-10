package main

import (
	"fmt"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/cli"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk"
)

// searchRunOutcome captures a single-mode or portfolio search run.
type searchRunOutcome struct {
	Result     optimisation.SearchResult
	WinnerMode string
	Portfolio  *optimisation.PortfolioResult
	Recorder   *optimisation.PortfolioAssistRecorder
}

type portfolioRunParams struct {
	Problem            optimisation.Problem
	Config             optimisation.SearchConfig
	WorkerDecisionMode string
	Domain             string
	Instance           string
	PortfolioModelPath string
}

// runSearchSolve runs portfolio assist or a single search depending on mode.
func runSearchSolve(mode string, extraPortfolioModes []string, p portfolioRunParams) searchRunOutcome {
	usePortfolio := mode == "portfolio"
	for _, m := range extraPortfolioModes {
		if mode == m {
			usePortfolio = true
			break
		}
	}
	if !usePortfolio {
		return searchRunOutcome{
			Result:     sdk.RunSearch(p.Problem, p.Config),
			WinnerMode: mode,
		}
	}
	assistConfig := optimisation.PortfolioAssistConfig{
		Mode:      p.WorkerDecisionMode,
		Domain:    p.Domain,
		Instance:  p.Instance,
		ModelPath: p.PortfolioModelPath,
	}
	pr, recorder := optimisation.RunPortfolioWithAssist(p.Problem, p.Config, assistConfig)
	return searchRunOutcome{
		Result:     pr.BestResult,
		WinnerMode: pr.Winner,
		Portfolio:  &pr,
		Recorder:   recorder,
	}
}

func printImprovementPct(baseline, final int) {
	delta := baseline - final
	pct := 0.0
	if baseline > 0 {
		pct = float64(delta) / float64(baseline) * 100
	}
	fmt.Printf("  Improvement:     %d → %d (%+.1f%%)\n", baseline, final, pct)
}

func printSearchResultStats(result optimisation.SearchResult) {
	fmt.Printf("  Initial:         %d\n", result.InitialPenalty)
	fmt.Printf("  Final:           %d\n", result.BestPenalty)
	fmt.Printf("  Runtime:         %dms\n", result.DurationMs)
	fmt.Printf("  Candidates:      %d\n", result.Candidates)
	if result.Candidates > 0 {
		fmt.Printf("  Accepted:        %d (%.1f%%)\n", result.Accepted,
			float64(result.Accepted)/float64(result.Candidates)*100)
	}
	fmt.Printf("  Hard rejected:   %d\n", result.Rejected)
	fmt.Printf("  Improvements:    %d\n", result.Improved)
}

func printPortfolioResults(disp cli.Options, pr optimisation.PortfolioResult, baseline int, objectiveLabel string) {
	fmt.Println(disp.Heading(cli.EmojiValid, "Per-Strategy Results"))
	fmt.Println()
	fmt.Printf("  %-8s %10s %10s %10s %8s\n", "Mode", objectiveLabel, "Improve%", "Candidates", "Runtime")
	fmt.Printf("  %-8s %10s %10s %10s %8s\n", "────────", "──────────", "──────────", "──────────", "────────")
	for _, e := range pr.Entries {
		impPct := 0.0
		if baseline > 0 {
			impPct = float64(baseline-e.Result.BestPenalty) / float64(baseline) * 100
		}
		winner := " "
		if e.Mode == pr.Winner {
			winner = "★"
		}
		fmt.Printf(" %s%-8s %10d %9.1f%% %10d %6dms\n",
			winner, strings.ToUpper(e.Mode), e.Result.BestPenalty, impPct, e.Result.Candidates, e.Result.DurationMs)
	}
	fmt.Println()
}

func printPortfolioWinner(disp cli.Options, pr optimisation.PortfolioResult, problem optimisation.Problem, baseline int, objectiveLabel string) {
	finalCost := problem.Evaluate(pr.BestResult.BestSolution)
	feasible := finalCost == pr.BestResult.BestPenalty
	fmt.Println(disp.Heading(cli.EmojiValid, "Winner: "+strings.ToUpper(pr.Winner)))
	fmt.Println()
	fmt.Printf("  %s:        %d\n", objectiveLabel, finalCost)
	fmt.Printf("  Feasible:        %v\n", feasible)
	printImprovementPct(baseline, finalCost)
	fmt.Println()
}
