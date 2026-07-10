package assist

import (
	"strings"
	"sync"
)

// ExecutePortfolio runs strategies with pre-resolved advice and records assist telemetry.
func ExecutePortfolio(
	input PortfolioExecuteInput,
	advice []StrategyAdvice,
	sources []AdviceSource,
	recorder *PortfolioAssistRecorder,
	run StrategyRunner,
) (PortfolioExecuteResult, *PortfolioAssistRecorder) {
	strategies := make([]string, len(advice))
	for i, a := range advice {
		strategies[i] = a.Strategy
	}

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
		originalBudget := input.Iterations
		recommendedBudget := int(float64(originalBudget) * a.BudgetMult)
		finalBudget := originalBudget

		reasonCodes := strings.Join(a.Reasons, ";")
		if i < len(sources) && sources[i].FallbackReason != "" {
			reasonCodes += ";fallback:" + sources[i].FallbackReason
		}

		rec := PortfolioAssistRecord{
			Domain:            input.Domain,
			Instance:          input.Instance,
			Strategy:          strat,
			Seed:              input.BaseSeed + int64(i)*7919,
			OriginalBudget:    originalBudget,
			RecommendedBudget: recommendedBudget,
			Recommendation:    a.Action,
			Confidence:        a.Confidence,
			ReasonCodes:       reasonCodes,
		}

		if (input.AssistMode == "assist" || input.AssistMode == "adaptive") && safe {
			finalBudget = recommendedBudget
			rec.Accepted = true
			rec.FinalBudget = finalBudget
		} else if (input.AssistMode == "assist" || input.AssistMode == "adaptive") && !safe {
			rec.SafetyRejected = true
			rec.SafetyRule = safetyRule
			rec.Accepted = false
			rec.FinalBudget = originalBudget
		} else {
			rec.Accepted = false
			rec.FinalBudget = originalBudget
		}

		skip := a.Action == PortfolioActionSkip && (input.AssistMode == "assist" || input.AssistMode == "adaptive") && safe
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
		entry PortfolioExecuteEntry
	}

	results := make(chan indexedResult, len(strategies))
	var wg sync.WaitGroup

	for i, runCfg := range runs {
		if runCfg.skip {
			results <- indexedResult{
				idx: i,
				entry: PortfolioExecuteEntry{
					Mode:        runCfg.mode,
					Seed:        runCfg.seed,
					BestPenalty: int(^uint(0) >> 1),
				},
			}
			continue
		}

		wg.Add(1)
		go func(idx int, r stratRun) {
			defer wg.Done()
			outcome := run(idx, r.mode, r.seed, r.budget)
			results <- indexedResult{
				idx: idx,
				entry: PortfolioExecuteEntry{
					Mode:        r.mode,
					Seed:        r.seed,
					BestPenalty: outcome.BestPenalty,
					DurationMs:  outcome.DurationMs,
				},
			}
		}(i, runCfg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	entries := make([]PortfolioExecuteEntry, len(strategies))
	for r := range results {
		entries[r.idx] = r.entry
	}

	bestIdx := 0
	bestPenalty := int(^uint(0) >> 1)
	for i, e := range entries {
		if e.BestPenalty < bestPenalty {
			bestPenalty = e.BestPenalty
			bestIdx = i
		}
	}

	for i, e := range entries {
		won := i == bestIdx
		recorder.RecordOutcome(runs[i].recordIdx, e.BestPenalty, won, e.DurationMs)
	}

	return PortfolioExecuteResult{
		Winner:  entries[bestIdx].Mode,
		Entries: entries,
	}, recorder
}
