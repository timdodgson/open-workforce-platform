package assist

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

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
	Domain   string
	Instance string
	Strategy string
	Seed     int64

	OriginalBudget    int
	RecommendedBudget int
	FinalBudget       int

	Recommendation PortfolioAssistAction
	Confidence     float64
	ReasonCodes    string

	Accepted       bool
	SafetyRejected bool
	SafetyRule     string

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
	BudgetMult float64
	Confidence float64
	Reasons    []string
}

// Advise evaluates all strategies and returns per-strategy recommendations.
func (a *RuleBasedPortfolioAdvisor) Advise(strategies []string, baseBudget int, domain string) []StrategyAdvice {
	advice := make([]StrategyAdvice, len(strategies))
	for i, strat := range strategies {
		budgetMult, confidence, reason := PortfolioBudgetHeuristic(strat, domain, len(strategies))
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

// EvaluatePortfolioSafety checks whether a portfolio recommendation is safe.
func EvaluatePortfolioSafety(strategies []string, advice []StrategyAdvice) (safe bool, rule string, fixedAdvice []StrategyAdvice) {
	runCount := 0
	for _, a := range advice {
		if a.Action != PortfolioActionSkip {
			runCount++
		}
	}
	if runCount == 0 {
		fixed := make([]StrategyAdvice, len(advice))
		for i, a := range advice {
			fixed[i] = a
			fixed[i].Action = PortfolioActionRun
			fixed[i].BudgetMult = 1.0
		}
		return false, "cannot_skip_all", fixed
	}

	if len(strategies) >= 3 && runCount < 2 {
		fixed := make([]StrategyAdvice, len(advice))
		copy(fixed, advice)
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

// PortfolioAssistConfig holds configuration for portfolio assist.
type PortfolioAssistConfig struct {
	Mode     string
	Domain   string
	Instance string
	ModelPath string
}

// AdviceSource indicates whether a recommendation came from the learned model
// or from rule-based fallback.
type AdviceSource struct {
	Strategy       string
	UsedLearned    bool
	FallbackReason string
	LearnedConf    float64
	RuleBased      StrategyAdvice
	Learned        StrategyAdvice
}

// LearnedAdviceResult extends StrategyAdvice with learned-vs-fallback provenance.
type LearnedAdviceResult struct {
	Advice []StrategyAdvice
	Source []AdviceSource
}

// StrategyRunResult is the opaque outcome of one strategy run inside a portfolio.
type StrategyRunResult struct {
	BestPenalty int
	DurationMs  int64
}

// StrategyRunner runs one portfolio strategy. The parent package injects RunSearch.
type StrategyRunner func(strategyIndex int, mode string, seed int64, iterations int) StrategyRunResult

// PortfolioExecuteInput describes one assisted portfolio run.
type PortfolioExecuteInput struct {
	AssistMode string
	Domain     string
	Instance   string
	BaseSeed   int64
	Iterations int
}

// PortfolioExecuteEntry is one strategy outcome from ExecutePortfolio.
type PortfolioExecuteEntry struct {
	Mode        string
	Seed        int64
	BestPenalty int
	DurationMs  int64
}

// PortfolioExecuteResult is the aggregate outcome of ExecutePortfolio.
type PortfolioExecuteResult struct {
	Winner  string
	Entries []PortfolioExecuteEntry
}

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
