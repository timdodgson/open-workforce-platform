// --- Portfolio Assist Runner ---
//
// RunPortfolioWithAssist stays in the parent package because it calls RunSearch.
// Types, advisor, safety, and CSV live in optimisation/assist.

package optimisation

import (
	"strings"
	"sync"
)

// RunPortfolioWithAssist runs a portfolio with optional AI assistance.
func RunPortfolioWithAssist(problem Problem, config SearchConfig, assistConfig PortfolioAssistConfig) (PortfolioResult, *PortfolioAssistRecorder) {
	mode := assistConfig.Mode
	if mode == "" && config.AssistMode != "" {
		mode = config.AssistMode
	}
	if mode == "" && config.PolicyMode != "" {
		mode = "shadow"
	}
	if mode == "" || mode == "off" {
		return RunPortfolio(problem, config), nil
	}
	assistConfig.Mode = mode

	strategies := config.Portfolio
	if len(strategies) == 0 {
		strategies = []string{"sa", "lahc", "tabu"}
	}

	recorder := NewPortfolioAssistRecorder()

	var advice []StrategyAdvice
	var sources []AdviceSource

	if config.PolicyMode != "" {
		modelPath := ""
		if config.PolicyMode != "rules" {
			modelPath = ResolveBudgetPolicyPath(config.PolicyDir, assistConfig.ModelPath, config.PolicyMode)
		}
		policy := NewPortfolioBudgetPolicy(PortfolioBudgetPolicyConfig{
			Domain:    assistConfig.Domain,
			ModelPath: modelPath,
		})
		allocations := AllocateBudgetsViaPolicy(
			policy, strategies, config.Iterations,
			assistConfig.Domain, assistConfig.Instance, "", NewFeatureExtractor(),
		)
		advice = make([]StrategyAdvice, len(allocations))
		sources = make([]AdviceSource, len(strategies))
		for i, alloc := range allocations {
			advice[i] = StrategyAdvice{
				Strategy:   alloc.Strategy,
				Action:     PortfolioActionRun,
				BudgetMult: alloc.BudgetMult,
				Confidence: alloc.Decision.Confidence,
				Reasons:    []string{alloc.Decision.Reason},
			}
			sources[i] = AdviceSource{
				Strategy:       alloc.Strategy,
				FallbackReason: alloc.Decision.FallbackReason,
				RuleBased:      advice[i],
			}
		}
	} else {
		var learnedAdvisor *LearnedPortfolioAdvisor
		if assistConfig.ModelPath != "" {
			model, err := LoadPortfolioBudgetModel(assistConfig.ModelPath)
			if err == nil {
				learnedAdvisor = NewLearnedPortfolioAdvisor(model)
			}
		}

		if learnedAdvisor != nil {
			result := learnedAdvisor.Advise(strategies, config.Iterations, assistConfig.Domain, assistConfig.Instance)
			advice = result.Advice
			sources = result.Source
		} else {
			ruleAdvisor := NewRuleBasedPortfolioAdvisor()
			advice = ruleAdvisor.Advise(strategies, config.Iterations, assistConfig.Domain)
			sources = make([]AdviceSource, len(strategies))
			for i, strat := range strategies {
				sources[i] = AdviceSource{
					Strategy:       strat,
					FallbackReason: "no_model_path",
					RuleBased:      advice[i],
				}
			}
		}
	}

	return runPortfolioWithAdvice(problem, config, assistConfig, strategies, advice, sources, recorder)
}

func runPortfolioWithAdvice(
	problem Problem,
	config SearchConfig,
	assistConfig PortfolioAssistConfig,
	strategies []string,
	advice []StrategyAdvice,
	sources []AdviceSource,
	recorder *PortfolioAssistRecorder,
) (PortfolioResult, *PortfolioAssistRecorder) {
	safe, safetyRule, fixedAdvice := EvaluatePortfolioSafety(strategies, advice)
	if !safe {
		advice = fixedAdvice
	}

	type stratRun struct {
		mode      string
		budget    int
		seed      int64
		recordIdx int
		skip      bool
	}

	runs := make([]stratRun, len(strategies))
	for i, strat := range strategies {
		a := advice[i]
		originalBudget := config.Iterations
		recommendedBudget := int(float64(originalBudget) * a.BudgetMult)
		finalBudget := originalBudget

		reasonCodes := strings.Join(a.Reasons, ";")
		if i < len(sources) && sources[i].FallbackReason != "" {
			reasonCodes += ";fallback:" + sources[i].FallbackReason
		}

		rec := PortfolioAssistRecord{
			Domain:            assistConfig.Domain,
			Instance:          assistConfig.Instance,
			Strategy:          strat,
			Seed:              config.Seed + int64(i)*7919,
			OriginalBudget:    originalBudget,
			RecommendedBudget: recommendedBudget,
			Recommendation:    a.Action,
			Confidence:        a.Confidence,
			ReasonCodes:       reasonCodes,
		}

		if (assistConfig.Mode == "assist" || assistConfig.Mode == "adaptive") && safe {
			finalBudget = recommendedBudget
			rec.Accepted = true
			rec.FinalBudget = finalBudget
		} else if (assistConfig.Mode == "assist" || assistConfig.Mode == "adaptive") && !safe {
			rec.SafetyRejected = true
			rec.SafetyRule = safetyRule
			rec.Accepted = false
			rec.FinalBudget = originalBudget
		} else {
			rec.Accepted = false
			rec.FinalBudget = originalBudget
		}

		skip := a.Action == PortfolioActionSkip && (assistConfig.Mode == "assist" || assistConfig.Mode == "adaptive") && safe
		if skip {
			rec.FinalBudget = 0
		}

		idx := recorder.Record(rec)
		runs[i] = stratRun{
			mode:      strat,
			budget:    rec.FinalBudget,
			seed:      rec.Seed,
			recordIdx: idx,
			skip:      skip,
		}
	}

	type indexedResult struct {
		idx   int
		entry PortfolioEntry
	}

	results := make(chan indexedResult, len(strategies))
	var wg sync.WaitGroup

	for i, run := range runs {
		if run.skip {
			results <- indexedResult{
				idx: i,
				entry: PortfolioEntry{
					Mode:   run.mode,
					Seed:   run.seed,
					Result: SearchResult{BestPenalty: int(^uint(0) >> 1)},
				},
			}
			continue
		}

		wg.Add(1)
		go func(idx int, r stratRun) {
			defer wg.Done()

			stratConfig := config
			stratConfig.Mode = r.mode
			stratConfig.Seed = r.seed
			stratConfig.Iterations = r.budget

			result := RunSearch(problem, stratConfig)

			results <- indexedResult{
				idx: idx,
				entry: PortfolioEntry{
					Mode:   r.mode,
					Seed:   r.seed,
					Result: result,
				},
			}
		}(i, run)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	entries := make([]PortfolioEntry, len(strategies))
	for r := range results {
		entries[r.idx] = r.entry
	}

	bestIdx := 0
	bestPenalty := int(^uint(0) >> 1)
	for i, e := range entries {
		if e.Result.BestPenalty < bestPenalty {
			bestPenalty = e.Result.BestPenalty
			bestIdx = i
		}
	}

	for i, e := range entries {
		won := i == bestIdx
		recorder.RecordOutcome(runs[i].recordIdx, e.Result.BestPenalty, won, e.Result.DurationMs)
	}

	return PortfolioResult{
		Winner:     entries[bestIdx].Mode,
		Entries:    entries,
		BestResult: entries[bestIdx].Result,
	}, recorder
}
