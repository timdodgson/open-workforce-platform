package optimisation

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// --- Portfolio Assist ---
//
// AI-advised budget allocation for portfolio-mode solvers (CVRP, JSS, VRPTW).
// In shadow mode: records what it would have done without changing behaviour.
// In assist mode: adjusts iteration budgets with safety overrides.
// In off mode: zero overhead, existing behaviour unchanged.

// PortfolioAssistAction is the action recommended for a strategy.
type PortfolioAssistAction string

const (
	PortfolioActionRun          PortfolioAssistAction = "run"
	PortfolioActionSkip         PortfolioAssistAction = "skip"
	PortfolioActionReduceBudget PortfolioAssistAction = "reduce_budget"
	PortfolioActionBoostBudget  PortfolioAssistAction = "boost_budget"
)

// PortfolioAssistRecord captures one strategy decision within a portfolio run.
type PortfolioAssistRecord struct {
	// Context.
	Domain   string
	Instance string
	Strategy string
	Seed     int64

	// Budgets.
	OriginalBudget    int
	RecommendedBudget int
	FinalBudget       int

	// Recommendation.
	Recommendation PortfolioAssistAction
	Confidence     float64
	ReasonCodes    string

	// Outcome.
	Accepted       bool
	SafetyRejected bool
	SafetyRule     string

	// Result (filled after strategy completes).
	ResultObjective int
	StrategyWon     bool
	RuntimeMs       int64
}

// PortfolioAssistRecorder collects all portfolio assist decisions.
type PortfolioAssistRecorder struct {
	mu      sync.Mutex
	records []PortfolioAssistRecord
}

// NewPortfolioAssistRecorder creates a new recorder.
func NewPortfolioAssistRecorder() *PortfolioAssistRecorder {
	return &PortfolioAssistRecorder{}
}

// Record adds a strategy decision.
func (r *PortfolioAssistRecorder) Record(rec PortfolioAssistRecord) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	idx := len(r.records)
	r.records = append(r.records, rec)
	return idx
}

// RecordOutcome updates a record with the final result.
func (r *PortfolioAssistRecorder) RecordOutcome(idx int, objective int, won bool, runtimeMs int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if idx < 0 || idx >= len(r.records) {
		return
	}
	r.records[idx].ResultObjective = objective
	r.records[idx].StrategyWon = won
	r.records[idx].RuntimeMs = runtimeMs
}

// Records returns all collected records.
func (r *PortfolioAssistRecorder) Records() []PortfolioAssistRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]PortfolioAssistRecord{}, r.records...)
}

// --- Rule-Based Portfolio Advisor ---

// RuleBasedPortfolioAdvisor recommends budget adjustments based on simple heuristics.
type RuleBasedPortfolioAdvisor struct{}

// NewRuleBasedPortfolioAdvisor creates the advisor.
func NewRuleBasedPortfolioAdvisor() *RuleBasedPortfolioAdvisor {
	return &RuleBasedPortfolioAdvisor{}
}

// StrategyAdvice is the recommendation for one strategy.
type StrategyAdvice struct {
	Strategy   string
	Action     PortfolioAssistAction
	BudgetMult float64 // multiplier on original budget (1.0 = no change)
	Confidence float64
	Reasons    []string
}

// Advise evaluates all strategies and returns per-strategy recommendations.
// Uses simple heuristics based on strategy type and domain knowledge.
func (a *RuleBasedPortfolioAdvisor) Advise(strategies []string, baseBudget int, domain string) []StrategyAdvice {
	advice := make([]StrategyAdvice, len(strategies))

	for i, strat := range strategies {
		budgetMult, confidence, reason := portfolioBudgetHeuristic(strat, domain, len(strategies))
		advice[i] = StrategyAdvice{
			Strategy:   strat,
			Action:     PortfolioActionRun,
			BudgetMult: budgetMult,
			Confidence: confidence,
			Reasons:    []string{reason},
		}
	}

	return advice
}

// --- Portfolio Safety ---

// EvaluatePortfolioSafety checks whether a portfolio recommendation is safe.
func EvaluatePortfolioSafety(strategies []string, advice []StrategyAdvice) (safe bool, rule string, fixedAdvice []StrategyAdvice) {
	// Safety 1: Never skip all strategies.
	runCount := 0
	for _, a := range advice {
		if a.Action != PortfolioActionSkip {
			runCount++
		}
	}
	if runCount == 0 {
		// Override: run all with default budget.
		fixed := make([]StrategyAdvice, len(advice))
		for i, a := range advice {
			fixed[i] = a
			fixed[i].Action = PortfolioActionRun
			fixed[i].BudgetMult = 1.0
		}
		return false, "cannot_skip_all", fixed
	}

	// Safety 2: At least 2 strategies must run if portfolio has 3+.
	if len(strategies) >= 3 && runCount < 2 {
		fixed := make([]StrategyAdvice, len(advice))
		copy(fixed, advice)
		// Un-skip the one with highest confidence for skip.
		for i := range fixed {
			if fixed[i].Action == PortfolioActionSkip {
				fixed[i].Action = PortfolioActionReduceBudget
				fixed[i].BudgetMult = 0.5
				runCount++
				if runCount >= 2 {
					break
				}
			}
		}
		return false, "min_two_strategies", fixed
	}

	// Safety 3: No single strategy gets more than 2x boost.
	for i, a := range advice {
		if a.BudgetMult > 2.0 {
			advice[i].BudgetMult = 2.0
		}
		if a.BudgetMult < 0.25 {
			advice[i].BudgetMult = 0.25
		}
	}

	return true, "", advice
}

// --- Portfolio Runner with Assist ---

// PortfolioAssistConfig holds configuration for portfolio assist.
type PortfolioAssistConfig struct {
	Mode     string // "off", "shadow", "assist"
	Domain   string // "cvrp", "jss", "vrptw", "nrp"
	Instance string

	// ModelPath is the path to portfolio_budget_model.json.
	// If empty or file not found, falls back to rule-based allocation.
	ModelPath string
}

// RunPortfolioWithAssist runs a portfolio with optional AI assistance.
// When mode is "off": calls RunPortfolio directly (zero overhead).
// When mode is "shadow": evaluates and records, runs with original budgets.
// When mode is "assist": applies safe budget adjustments.
//
// Budget allocation uses a learned model if available and confident.
// Falls back to rule-based allocation otherwise.
func RunPortfolioWithAssist(problem Problem, config SearchConfig, assistConfig PortfolioAssistConfig) (PortfolioResult, *PortfolioAssistRecorder) {
	if assistConfig.Mode == "" || assistConfig.Mode == "off" {
		return RunPortfolio(problem, config), nil
	}

	strategies := config.Portfolio
	if len(strategies) == 0 {
		strategies = []string{"sa", "lahc", "tabu"}
	}

	recorder := NewPortfolioAssistRecorder()

	// Load learned model (graceful fallback if not available).
	var learnedAdvisor *LearnedPortfolioAdvisor
	if assistConfig.ModelPath != "" {
		model, err := LoadPortfolioBudgetModel(assistConfig.ModelPath)
		if err == nil {
			learnedAdvisor = NewLearnedPortfolioAdvisor(model)
		}
	}

	// Get advice — use learned if available, otherwise rule-based.
	var advice []StrategyAdvice
	var sources []AdviceSource

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

	// Safety check.
	safe, safetyRule, fixedAdvice := EvaluatePortfolioSafety(strategies, advice)
	if !safe {
		advice = fixedAdvice
	}

	// Determine final budgets per strategy.
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
		finalBudget := originalBudget // default: no change

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
			// Shadow: record but use original budget.
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

	// Execute strategies (parallel for 2+ non-skipped).
	type indexedResult struct {
		idx   int
		entry PortfolioEntry
	}

	results := make(chan indexedResult, len(strategies))
	var wg sync.WaitGroup

	for i, run := range runs {
		if run.skip {
			// Record a zero-result for skipped strategies.
			results <- indexedResult{
				idx: i,
				entry: PortfolioEntry{
					Mode:   run.mode,
					Seed:   run.seed,
					Result: SearchResult{BestPenalty: int(^uint(0) >> 1)}, // max int = worst
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

	// Collect results.
	entries := make([]PortfolioEntry, len(strategies))
	for r := range results {
		entries[r.idx] = r.entry
	}

	// Find best.
	bestIdx := 0
	bestPenalty := int(^uint(0) >> 1)
	for i, e := range entries {
		if e.Result.BestPenalty < bestPenalty {
			bestPenalty = e.Result.BestPenalty
			bestIdx = i
		}
	}

	// Update recorder with outcomes.
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

// --- CSV Output ---

// WritePortfolioAssistCSV writes portfolio_assist.csv.
func WritePortfolioAssistCSV(path string, records []PortfolioAssistRecord) error {
	if len(records) == 0 {
		return nil
	}

	header := strings.Join([]string{
		"domain", "instance", "strategy", "seed",
		"original_budget", "recommended_budget", "final_budget",
		"recommendation", "confidence", "reason_codes",
		"accepted", "safety_rejected", "safety_rule",
		"result_objective", "strategy_won", "runtime_ms",
	}, ",")

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")

	for _, r := range records {
		accepted := 0
		if r.Accepted {
			accepted = 1
		}
		safetyRejected := 0
		if r.SafetyRejected {
			safetyRejected = 1
		}
		won := 0
		if r.StrategyWon {
			won = 1
		}

		row := fmt.Sprintf("%s,%s,%s,%d,%d,%d,%d,%s,%.2f,%s,%d,%d,%s,%d,%d,%d",
			r.Domain, r.Instance, r.Strategy, r.Seed,
			r.OriginalBudget, r.RecommendedBudget, r.FinalBudget,
			r.Recommendation, r.Confidence, r.ReasonCodes,
			accepted, safetyRejected, r.SafetyRule,
			r.ResultObjective, won, r.RuntimeMs,
		)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
