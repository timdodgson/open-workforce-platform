// portfolio_bridge.go wires assist.ExecutePortfolio to RunSearch.

package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/assist"
)

// RunPortfolioWithAssist runs a portfolio with optional AI assistance.
func RunPortfolioWithAssist(problem Problem, config SearchConfig, assistConfig PortfolioAssistConfig) (PortfolioResult, *PortfolioAssistRecorder) {
	mode := resolvePortfolioAssistMode(assistConfig, config)
	if mode == "" || mode == "off" {
		return RunPortfolio(problem, config), nil
	}
	assistConfig.Mode = mode

	strategies := config.Portfolio
	if len(strategies) == 0 {
		strategies = []string{"sa", "lahc", "tabu"}
	}

	bundle := resolvePortfolioAdvice(config, assistConfig, strategies)
	recorder := NewPortfolioAssistRecorder()

	searchResults := make([]SearchResult, len(strategies))
	execResult, recorder := assist.ExecutePortfolio(
		assist.PortfolioExecuteInput{
			AssistMode: mode,
			Domain:     assistConfig.Domain,
			Instance:   assistConfig.Instance,
			BaseSeed:   config.Seed,
			Iterations: config.Iterations,
		},
		bundle.advice,
		bundle.sources,
		recorder,
		func(strategyIndex int, mode string, seed int64, iterations int) assist.StrategyRunResult {
			stratConfig := config
			stratConfig.Mode = mode
			stratConfig.Seed = seed
			stratConfig.Iterations = iterations
			result := RunSearch(problem, stratConfig)
			searchResults[strategyIndex] = result
			return assist.StrategyRunResult{
				BestPenalty: result.BestPenalty,
				DurationMs:  result.DurationMs,
			}
		},
	)

	entries := make([]PortfolioEntry, len(execResult.Entries))
	bestIdx := 0
	bestPenalty := int(^uint(0) >> 1)
	worstPenalty := int(^uint(0) >> 1)
	for i, e := range execResult.Entries {
		result := searchResults[i]
		if e.BestPenalty == worstPenalty && result.Candidates == 0 {
			result = SearchResult{BestPenalty: worstPenalty}
		}
		entries[i] = PortfolioEntry{
			Mode:   e.Mode,
			Seed:   e.Seed,
			Result: result,
		}
		if e.BestPenalty < bestPenalty {
			bestPenalty = e.BestPenalty
			bestIdx = i
		}
	}

	return PortfolioResult{
		Winner:     execResult.Winner,
		Entries:    entries,
		BestResult: entries[bestIdx].Result,
	}, recorder
}
