package optimisation

import (
	"fmt"
)

// --- Adaptive Search Assist ---
//
// Extends the rule-based assist with live-updating decisions based on
// observed search progress. Unlike static assist which uses fixed thresholds,
// adaptive mode:
//   - Extends budget if improvement rate remains strong near the end
//   - Early stops if plateau persists after safe minimum budget
//   - Never stops immediately after an improvement
//   - Adjusts stagnation thresholds based on observed improvement curve
//
// All decisions are deterministic given the same search progress state.

// AdaptiveSearchAssist implements SearchAssist with live-updating logic.
type AdaptiveSearchAssist struct {
	config SearchAssistConfig

	// Live state tracking for adaptive decisions.
	improvementCount  int // total improvements observed so far
	lastImprovementAt int // candidate number of most recent improvement
	peakImprovRate    float64
}

// NewAdaptiveSearchAssist creates an adaptive search assist engine.
func NewAdaptiveSearchAssist(config SearchAssistConfig) *AdaptiveSearchAssist {
	return &AdaptiveSearchAssist{config: config}
}

// RecordImprovement updates the adaptive state when an improvement is observed.
func (a *AdaptiveSearchAssist) RecordImprovement(candidates int, improvRate float64) {
	a.improvementCount++
	a.lastImprovementAt = candidates
	if improvRate > a.peakImprovRate {
		a.peakImprovRate = improvRate
	}
}

// Checkpoint evaluates search progress with adaptive logic.
func (a *AdaptiveSearchAssist) Checkpoint(progress SearchProgress) *SearchRecommendation {
	budgetUsed := float64(progress.IterationsComplete) / float64(progress.IterationsTotal)

	// Adaptive Rule 1: Extend budget if still improving near the end.
	// More aggressive than static assist — triggers at 85% budget with lower rate threshold.
	if budgetUsed >= 0.85 && progress.ImprovementRate > 0.5 {
		// Extend proportionally to how productive the search still is.
		extensionFactor := 1.25
		if progress.ImprovementRate > 2.0 {
			extensionFactor = 1.5
		}
		newBudget := int(float64(progress.IterationsTotal) * extensionFactor)
		return &SearchRecommendation{
			Action:     SearchAdjustBudget,
			Confidence: Confidence(0.65 + budgetUsed*0.15),
			Reasons:    []string{"adaptive_still_improving", fmt.Sprintf("rate_%.2f", progress.ImprovementRate), fmt.Sprintf("extend_%.0f_pct", (extensionFactor-1)*100)},
			NewBudget:  newBudget,
		}
	}

	// Adaptive Rule 2: Early stop with adaptive stagnation window.
	// The stagnation window scales based on observed improvement pattern:
	// - If search has found improvements, use the gap since last improvement.
	// - If search has never improved, require longer patience (75% of budget).
	adaptiveStagnation := a.config.StagnationWindow
	if a.improvementCount == 0 {
		// No improvements yet — be more patient (75% of total budget).
		adaptiveStagnation = int(float64(progress.IterationsTotal) * 0.75)
	} else if a.improvementCount > 0 && a.peakImprovRate > 0 {
		// Scale stagnation window based on historical improvement frequency.
		// If improvements were frequent, allow less stagnation time.
		expectedGap := int(10000.0 / a.peakImprovRate) // expected candidates between improvements
		adaptiveStagnation = expectedGap * 5           // allow 5× the expected gap
		if adaptiveStagnation < a.config.StagnationWindow {
			adaptiveStagnation = a.config.StagnationWindow
		}
	}

	if budgetUsed >= a.config.MinBudgetFraction && progress.PlateauLength >= adaptiveStagnation {
		return &SearchRecommendation{
			Action:     SearchEarlyStop,
			Confidence: Confidence(0.6 + budgetUsed*0.2),
			Reasons: []string{
				fmt.Sprintf("adaptive_stagnation_%d", progress.PlateauLength),
				fmt.Sprintf("window_%d", adaptiveStagnation),
				fmt.Sprintf("budget_%.0f_pct", budgetUsed*100),
			},
		}
	}

	// Adaptive Rule 3: Budget reduce if completely stagnant past halfway.
	if budgetUsed >= 0.5 && progress.ImprovementRate < 0.01 && progress.PlateauLength > adaptiveStagnation/2 {
		newBudget := progress.IterationsComplete + (progress.IterationsTotal-progress.IterationsComplete)/2
		return &SearchRecommendation{
			Action:     SearchAdjustBudget,
			Confidence: 0.55,
			Reasons:    []string{"adaptive_low_rate", "half_remaining"},
			NewBudget:  newBudget,
		}
	}

	return nil // continue
}

// --- Adaptive Search Hook Runner ---
//
// The standard NewSearchHookRunner handles adaptive mode by creating an
// AdaptiveSearchAssist engine internally. No separate runner is needed.
