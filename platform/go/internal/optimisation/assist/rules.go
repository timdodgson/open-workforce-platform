package assist

import (
	"fmt"
	"os"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

// RuleBasedSearchAssist implements searchdef.SearchAssist using simple heuristic rules.
type RuleBasedSearchAssist struct {
	config searchdef.SearchAssistConfig
}

// NewRuleBasedSearchAssist creates a rule-based search assist engine.
func NewRuleBasedSearchAssist(config searchdef.SearchAssistConfig) *RuleBasedSearchAssist {
	return &RuleBasedSearchAssist{config: config}
}

// Checkpoint evaluates search progress and returns a recommendation.
func (a *RuleBasedSearchAssist) Checkpoint(progress searchdef.SearchProgress) *searchdef.SearchRecommendation {
	budgetUsed := float64(progress.IterationsComplete) / float64(progress.IterationsTotal)
	if budgetUsed >= a.config.MinBudgetFraction && progress.PlateauLength >= a.config.StagnationWindow {
		return &searchdef.SearchRecommendation{
			Action:     searchdef.SearchEarlyStop,
			Confidence: searchdef.Confidence(0.6 + budgetUsed*0.2),
			Reasons:    []string{fmt.Sprintf("stagnation_%d_cands", progress.PlateauLength), fmt.Sprintf("budget_%.0f_pct", budgetUsed*100)},
		}
	}

	if budgetUsed >= 0.5 && progress.ImprovementRate < 0.01 && progress.PlateauLength > a.config.StagnationWindow/2 {
		newBudget := progress.IterationsComplete + (progress.IterationsTotal-progress.IterationsComplete)/2
		return &searchdef.SearchRecommendation{
			Action:     searchdef.SearchAdjustBudget,
			Confidence: 0.55,
			Reasons:    []string{"low_improvement_rate", "half_remaining"},
			NewBudget:  newBudget,
		}
	}

	if budgetUsed >= 0.9 && progress.ImprovementRate > 1.0 {
		newBudget := int(float64(progress.IterationsTotal) * 1.25)
		return &searchdef.SearchRecommendation{
			Action:     searchdef.SearchAdjustBudget,
			Confidence: 0.6,
			Reasons:    []string{"still_improving_near_end", "extend_25_pct"},
			NewBudget:  newBudget,
		}
	}

	return nil
}

// EvaluateSearchSafety checks whether a search assist recommendation is safe.
func EvaluateSearchSafety(progress searchdef.SearchProgress, rec *searchdef.SearchRecommendation, config searchdef.SearchAssistConfig) (safe bool, rule string) {
	if rec == nil {
		return true, ""
	}

	if rec.Action == searchdef.SearchEarlyStop {
		budgetUsed := float64(progress.IterationsComplete) / float64(progress.IterationsTotal)
		if budgetUsed < config.MinBudgetFraction {
			return false, "below_min_budget"
		}
	}

	if rec.Action == searchdef.SearchEarlyStop && progress.PlateauLength < config.RecentImprovWindow {
		return false, "recent_improvement"
	}

	if rec.Action == searchdef.SearchAdjustBudget && rec.NewBudget > 0 {
		minAllowed := int(float64(progress.IterationsTotal) * config.MinBudgetFraction)
		if rec.NewBudget < minAllowed {
			return false, "budget_below_minimum"
		}
	}

	return true, ""
}

// WriteSearchAssistCSV writes generic_search_assist.csv.
func WriteSearchAssistCSV(path string, records []searchdef.SearchAssistRecord) error {
	if len(records) == 0 {
		return nil
	}

	header := strings.Join([]string{
		"algorithm", "checkpoint", "candidates", "iterations_total",
		"current_penalty", "best_penalty", "initial_penalty",
		"temperature", "plateau_length", "improvement_rate",
		"recommended_action", "confidence", "reasons",
		"safety_triggered", "safety_rule",
		"accepted", "final_action",
		"final_best_penalty", "total_candidates", "runtime_ms",
	}, ",")

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")

	for _, r := range records {
		accepted := 0
		if r.Accepted {
			accepted = 1
		}
		safety := 0
		if r.SafetyTriggered {
			safety = 1
		}

		row := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%.6f,%d,%.4f,%s,%.2f,%s,%d,%s,%d,%s,%d,%d,%d",
			r.Algorithm, r.Checkpoint, r.Candidates, r.IterationsTotal,
			r.CurrentPenalty, r.BestPenalty, r.InitialPenalty,
			r.Temperature, r.PlateauLength, r.ImprovementRate,
			r.RecommendedAction, r.Confidence, r.Reasons,
			safety, r.SafetyRule,
			accepted, r.FinalAction,
			r.FinalBestPenalty, r.TotalCandidates, r.RuntimeMs,
		)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
