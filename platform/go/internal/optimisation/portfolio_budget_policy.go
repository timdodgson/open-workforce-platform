// portfolio_budget_policy.go — Policy-based portfolio budget allocation (v2).
//
// Replaces heuristic "SA +10%, LAHC -10%" rules with learned policies that
// allocate budgets based on historical telemetry.
//
// Flow:
//   Historical telemetry → FeatureExtractor → FeatureVector → Policy → Budget
//
// The policy architecture provides:
//   - LearnedPolicy: data-driven allocation from portfolio_budget_model_v2.json
//   - RulePolicy: existing heuristics as fallback
//   - HybridPolicy: learned when confident, rules otherwise
//
// Safety:
//   - Confidence threshold (0.60) before learned policy applies
//   - Hard caps: 0.25× min, 2.0× max budget multiplier
//   - Never skip all strategies
//   - Minimum 2 strategies must run in 3+ portfolio
//   - Policy versioning for audit trail
package optimisation

import (
	"fmt"
	"time"
)

// ───────────────────────────────────────────────────────────────
// Portfolio Budget Policy (Rule-based)
// ───────────────────────────────────────────────────────────────

// NewPortfolioBudgetRulePolicy creates a RulePolicy implementing the
// existing v1 heuristics (SA boost, LAHC reduce, Tabu boost on constrained).
func NewPortfolioBudgetRulePolicy(domain string) *RulePolicy {
	rules := []Rule{
		{
			Name: "sa_generally_strong",
			Matches: func(ctx PolicyContext) bool {
				return ctx.Features.Algorithm == "sa"
			},
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{
					Action:     "allocate",
					Confidence: 0.55,
					Reason:     "rule:sa_generally_strong",
					Parameters: map[string]any{"budget_mult": 1.1},
				}
			},
		},
		{
			Name: "lahc_slower_in_portfolio",
			Matches: func(ctx PolicyContext) bool {
				return ctx.Features.Algorithm == "lahc" && ctx.Features.WorkerCount >= 3
			},
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{
					Action:     "allocate",
					Confidence: 0.50,
					Reason:     "rule:lahc_slower_convergence_in_portfolio",
					Parameters: map[string]any{"budget_mult": 0.9},
				}
			},
		},
		{
			Name: "tabu_strong_on_constrained",
			Matches: func(ctx PolicyContext) bool {
				return ctx.Features.Algorithm == "tabu" &&
					(ctx.Domain == "jss" || ctx.Domain == "nrp")
			},
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{
					Action:     "allocate",
					Confidence: 0.55,
					Reason:     "rule:tabu_strong_on_constrained",
					Parameters: map[string]any{"budget_mult": 1.15},
				}
			},
		},
	}

	return NewRulePolicy(
		fmt.Sprintf("portfolio-budget-rules-%s", domain),
		"1.0.0", domain, "portfolio", rules,
	)
}

// ───────────────────────────────────────────────────────────────
// Portfolio Budget Policy (Learned — v2 model)
// ───────────────────────────────────────────────────────────────

// PortfolioBudgetModelV2 wraps the trained model as a PolicyModel interface.
type PortfolioBudgetModelV2 struct {
	model *PortfolioBudgetModel
}

// NewPortfolioBudgetModelV2 creates a learned model adapter.
func NewPortfolioBudgetModelV2(model *PortfolioBudgetModel) *PortfolioBudgetModelV2 {
	return &PortfolioBudgetModelV2{model: model}
}

// Predict returns a budget allocation recommendation for a given strategy.
func (m *PortfolioBudgetModelV2) Predict(features FeatureVector) ModelPrediction {
	if m.model == nil {
		return ModelPrediction{Action: "defer", Confidence: 0}
	}

	entry := m.findEntry(features.Problem, features.Instance, features.Algorithm)
	if entry == nil {
		return ModelPrediction{
			Action:     "defer",
			Confidence: 0,
			Reason:     "no_data_for_strategy",
		}
	}

	if entry.SampleCount < 3 {
		return ModelPrediction{
			Action:     "defer",
			Confidence: entry.Confidence * 0.5, // penalise low samples
			Reason:     fmt.Sprintf("insufficient_samples_%d", entry.SampleCount),
		}
	}

	return ModelPrediction{
		Action:        "allocate",
		Confidence:    entry.Confidence,
		ExpectedValue: entry.MeanImprovement,
		Reason: fmt.Sprintf("learned_win_rate_%.0f_pct_samples_%d",
			entry.WinRate*100, entry.SampleCount),
	}
}

func (m *PortfolioBudgetModelV2) findEntry(domain, instance, strategy string) *StrategyPerformanceEntry {
	var domainMatch *StrategyPerformanceEntry
	for i := range m.model.Entries {
		e := &m.model.Entries[i]
		if e.Domain != domain || e.Strategy != strategy {
			continue
		}
		if e.Instance != "" && e.Instance == instance {
			return e
		}
		if e.Instance == "" {
			domainMatch = e
		}
	}
	return domainMatch
}

// BudgetMultiplier extracts the recommended budget multiplier from the model
// for a specific strategy. Returns 1.0 if no entry found.
func (m *PortfolioBudgetModelV2) BudgetMultiplier(domain, instance, strategy string) float64 {
	entry := m.findEntry(domain, instance, strategy)
	if entry == nil {
		return 1.0
	}
	mult := entry.RecommendedMult
	if mult > 2.0 {
		mult = 2.0
	}
	if mult < 0.25 {
		mult = 0.25
	}
	return mult
}

// ───────────────────────────────────────────────────────────────
// Portfolio Budget Policy Factory
// ───────────────────────────────────────────────────────────────

// PortfolioBudgetPolicyConfig configures the portfolio budget policy.
type PortfolioBudgetPolicyConfig struct {
	Domain    string
	ModelPath string // path to portfolio_budget_model_v2.json (empty = rules only)
}

// NewPortfolioBudgetPolicy creates the appropriate policy for portfolio allocation.
// If a model is loaded and valid, returns HybridPolicy (learned + rule fallback).
// Otherwise returns RulePolicy.
func NewPortfolioBudgetPolicy(cfg PortfolioBudgetPolicyConfig) Policy {
	rulePolicy := NewPortfolioBudgetRulePolicy(cfg.Domain)

	if cfg.ModelPath == "" {
		return rulePolicy
	}

	model, err := LoadPortfolioBudgetModel(cfg.ModelPath)
	if err != nil {
		// Model not available — use rules.
		return rulePolicy
	}

	learnedModel := NewPortfolioBudgetModelV2(model)
	learnedPolicy := NewLearnedPolicy(LearnedPolicyConfig{
		ID:              fmt.Sprintf("portfolio-budget-learned-%s", cfg.Domain),
		Version:         model.Version,
		Domain:          cfg.Domain,
		DecisionType:    "portfolio",
		Model:           learnedModel,
		Threshold:       MinLearnedConfidence,
		TrainedSamples:  model.TrainedOn,
		CreatedAt:       time.Now(),
		ValidationScore: -1,
	})

	return NewHybridPolicy(learnedPolicy, rulePolicy)
}

// ───────────────────────────────────────────────────────────────
// Budget Allocation via Policy
// ───────────────────────────────────────────────────────────────

// BudgetAllocation is the result of policy-based budget allocation.
type BudgetAllocation struct {
	Strategy      string
	BudgetMult    float64
	FinalBudget   int
	Decision      PolicyDecision
}

// AllocateBudgetsViaPolicy uses the policy architecture to allocate budgets.
// Returns per-strategy allocations with full decision provenance.
func AllocateBudgetsViaPolicy(
	policy Policy,
	strategies []string,
	baseBudget int,
	domain string,
	instance string,
	runID string,
	extractor *FeatureExtractor,
) []BudgetAllocation {
	allocations := make([]BudgetAllocation, len(strategies))

	for i, strat := range strategies {
		// Build context via feature extractor.
		ctx := PortfolioContext{
			Strategies:  strategies,
			TotalBudget: baseBudget,
			ProblemType: domain,
			Instance:    instance,
		}
		features := extractor.FromPortfolioContext(ctx, strat, runID)

		policyCtx := PolicyContext{
			DecisionType:      "portfolio",
			Features:          features,
			Domain:            domain,
			Instance:          instance,
			HistoricalSamples: 0,
		}

		decision := policy.Decide(policyCtx)

		// Extract budget multiplier from decision parameters.
		mult := 1.0
		if decision.Parameters != nil {
			if m, ok := decision.Parameters["budget_mult"]; ok {
				if mf, ok := m.(float64); ok {
					mult = mf
				}
			}
		}

		// Safety clamp.
		if mult > 2.0 {
			mult = 2.0
		}
		if mult < 0.25 {
			mult = 0.25
		}

		allocations[i] = BudgetAllocation{
			Strategy:    strat,
			BudgetMult:  mult,
			FinalBudget: int(float64(baseBudget) * mult),
			Decision:    decision,
		}
	}

	return allocations
}
