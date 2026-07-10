package optimisation

import (
	"encoding/json"
	"fmt"
	"os"
)

// --- Learned Portfolio Budget Allocation ---
//
// Replaces fixed rule-based heuristics with a data-driven model trained on
// historical portfolio telemetry. Falls back to rule-based allocation when:
//   - Model file is missing or unloadable
//   - Model confidence is below threshold
//   - No data exists for the requested domain/instance/strategy combination
//
// Model source: platform/ml/worker_model/portfolio_budget_model.json
// Training data: worker_learning.csv, portfolio_assist.csv, run.json

// MinLearnedConfidence is the minimum confidence required to use learned allocation
// instead of falling back to rule-based. Below this threshold, the rule-based
// allocator is used instead.
const MinLearnedConfidence = 0.60

// PortfolioBudgetModel holds the trained budget allocation model.
// Each entry maps a (domain, strategy) pair to learned performance statistics.
type PortfolioBudgetModel struct {
	// Version tracks the model format for forward compatibility.
	Version string `json:"version"`

	// TrainedOn records how many historical runs were used for training.
	TrainedOn int `json:"trained_on"`

	// Entries holds per-domain-strategy performance data.
	Entries []StrategyPerformanceEntry `json:"entries"`
}

// StrategyPerformanceEntry captures learned performance for one strategy
// within one domain (and optionally one instance).
type StrategyPerformanceEntry struct {
	Domain   string `json:"domain"`
	Instance string `json:"instance,omitempty"` // empty = domain-wide
	Strategy string `json:"strategy"`

	// Learned statistics.
	WinRate         float64 `json:"win_rate"`         // fraction of portfolio runs where this strategy won
	MeanImprovement float64 `json:"mean_improvement"` // mean objective improvement from initial
	MeanROI         float64 `json:"mean_roi"`         // mean improvement per 1K candidates
	SampleCount     int     `json:"sample_count"`     // number of historical observations

	// Derived recommendation.
	RecommendedMult float64 `json:"recommended_mult"` // budget multiplier (1.0 = no change)
	Confidence      float64 `json:"confidence"`       // model confidence in this recommendation
}

// LearnedPortfolioAdvisor uses historical performance data to recommend budget allocation.
type LearnedPortfolioAdvisor struct {
	model    *PortfolioBudgetModel
	fallback *RuleBasedPortfolioAdvisor
}

// LoadPortfolioBudgetModel loads the model from a JSON file.
// Returns nil and an error if the file cannot be read or parsed.
func LoadPortfolioBudgetModel(path string) (*PortfolioBudgetModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load portfolio budget model: %w", err)
	}

	var model PortfolioBudgetModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("parse portfolio budget model: %w", err)
	}

	if model.Version == "" {
		return nil, fmt.Errorf("parse portfolio budget model: missing version field")
	}

	if len(model.Entries) == 0 {
		return nil, fmt.Errorf("parse portfolio budget model: no entries")
	}

	return &model, nil
}

// NewLearnedPortfolioAdvisor creates an advisor that uses a learned model.
// If model is nil, all recommendations fall back to rule-based.
func NewLearnedPortfolioAdvisor(model *PortfolioBudgetModel) *LearnedPortfolioAdvisor {
	return &LearnedPortfolioAdvisor{
		model:    model,
		fallback: NewRuleBasedPortfolioAdvisor(),
	}
}

// Advise returns budget recommendations using the learned model where confident,
// falling back to rule-based allocation otherwise.
func (a *LearnedPortfolioAdvisor) Advise(strategies []string, baseBudget int, domain string, instance string) LearnedAdviceResult {
	// Always compute rule-based as baseline.
	ruleAdvice := a.fallback.Advise(strategies, baseBudget, domain)

	result := LearnedAdviceResult{
		Advice: make([]StrategyAdvice, len(strategies)),
		Source: make([]AdviceSource, len(strategies)),
	}

	for i, strat := range strategies {
		result.Source[i] = AdviceSource{
			Strategy:  strat,
			RuleBased: ruleAdvice[i],
		}

		// Attempt learned recommendation.
		if a.model == nil {
			result.Advice[i] = ruleAdvice[i]
			result.Source[i].FallbackReason = "no_model_loaded"
			continue
		}

		entry := a.findEntry(domain, instance, strat)
		if entry == nil {
			result.Advice[i] = ruleAdvice[i]
			result.Source[i].FallbackReason = "no_data_for_strategy"
			continue
		}

		if entry.Confidence < MinLearnedConfidence {
			result.Advice[i] = ruleAdvice[i]
			result.Source[i].FallbackReason = fmt.Sprintf("confidence_%.2f_below_threshold", entry.Confidence)
			result.Source[i].LearnedConf = entry.Confidence
			result.Source[i].Learned = a.buildAdvice(strat, entry)
			continue
		}

		if entry.SampleCount < 3 {
			result.Advice[i] = ruleAdvice[i]
			result.Source[i].FallbackReason = fmt.Sprintf("insufficient_samples_%d", entry.SampleCount)
			result.Source[i].LearnedConf = entry.Confidence
			result.Source[i].Learned = a.buildAdvice(strat, entry)
			continue
		}

		// Use learned recommendation.
		learned := a.buildAdvice(strat, entry)
		result.Advice[i] = learned
		result.Source[i].UsedLearned = true
		result.Source[i].LearnedConf = entry.Confidence
		result.Source[i].Learned = learned
	}

	return result
}

// findEntry looks up the best matching entry for a domain/instance/strategy combination.
// Prefers instance-specific entries over domain-wide entries.
func (a *LearnedPortfolioAdvisor) findEntry(domain, instance, strategy string) *StrategyPerformanceEntry {
	var domainMatch *StrategyPerformanceEntry

	for i := range a.model.Entries {
		e := &a.model.Entries[i]
		if e.Domain != domain || e.Strategy != strategy {
			continue
		}

		// Exact instance match takes priority.
		if e.Instance != "" && e.Instance == instance {
			return e
		}

		// Domain-wide match (no instance specified).
		if e.Instance == "" {
			domainMatch = e
		}
	}

	return domainMatch
}

// buildAdvice converts a model entry into a StrategyAdvice.
func (a *LearnedPortfolioAdvisor) buildAdvice(strategy string, entry *StrategyPerformanceEntry) StrategyAdvice {
	mult := entry.RecommendedMult

	// Clamp to safety bounds (2x max, 0.25x min).
	if mult > 2.0 {
		mult = 2.0
	}
	if mult < 0.25 {
		mult = 0.25
	}

	action := PortfolioActionRun
	if mult > 1.05 {
		action = PortfolioActionBoostBudget
	} else if mult < 0.95 {
		action = PortfolioActionReduceBudget
	}

	reasons := []string{
		fmt.Sprintf("learned_win_rate_%.0f_pct", entry.WinRate*100),
		fmt.Sprintf("samples_%d", entry.SampleCount),
	}

	return StrategyAdvice{
		Strategy:   strategy,
		Action:     action,
		BudgetMult: mult,
		Confidence: entry.Confidence,
		Reasons:    reasons,
	}
}
