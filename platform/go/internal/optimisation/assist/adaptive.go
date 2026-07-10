package assist

import (
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

// AdaptiveSearchAssist implements searchdef.SearchAssist with live-updating logic.
type AdaptiveSearchAssist struct {
	config searchdef.SearchAssistConfig

	improvementCount  int
	lastImprovementAt int
	peakImprovRate    float64
}

// NewAdaptiveSearchAssist creates an adaptive search assist engine.
func NewAdaptiveSearchAssist(config searchdef.SearchAssistConfig) *AdaptiveSearchAssist {
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
func (a *AdaptiveSearchAssist) Checkpoint(progress searchdef.SearchProgress) *searchdef.SearchRecommendation {
	budgetUsed := float64(progress.IterationsComplete) / float64(progress.IterationsTotal)

	if budgetUsed >= 0.85 && progress.ImprovementRate > 0.5 {
		extensionFactor := 1.25
		if progress.ImprovementRate > 2.0 {
			extensionFactor = 1.5
		}
		newBudget := int(float64(progress.IterationsTotal) * extensionFactor)
		return &searchdef.SearchRecommendation{
			Action:     searchdef.SearchAdjustBudget,
			Confidence: searchdef.Confidence(0.65 + budgetUsed*0.15),
			Reasons:    []string{"adaptive_still_improving", fmt.Sprintf("rate_%.2f", progress.ImprovementRate), fmt.Sprintf("extend_%.0f_pct", (extensionFactor-1)*100)},
			NewBudget:  newBudget,
		}
	}

	adaptiveStagnation := a.config.StagnationWindow
	if a.improvementCount == 0 {
		adaptiveStagnation = int(float64(progress.IterationsTotal) * 0.75)
	} else if a.improvementCount > 0 && a.peakImprovRate > 0 {
		expectedGap := int(10000.0 / a.peakImprovRate)
		adaptiveStagnation = expectedGap * 5
		if adaptiveStagnation < a.config.StagnationWindow {
			adaptiveStagnation = a.config.StagnationWindow
		}
	}

	if budgetUsed >= a.config.MinBudgetFraction && progress.PlateauLength >= adaptiveStagnation {
		return &searchdef.SearchRecommendation{
			Action:     searchdef.SearchEarlyStop,
			Confidence: searchdef.Confidence(0.6 + budgetUsed*0.2),
			Reasons: []string{
				fmt.Sprintf("adaptive_stagnation_%d", progress.PlateauLength),
				fmt.Sprintf("window_%d", adaptiveStagnation),
				fmt.Sprintf("budget_%.0f_pct", budgetUsed*100),
			},
		}
	}

	if budgetUsed >= 0.5 && progress.ImprovementRate < 0.01 && progress.PlateauLength > adaptiveStagnation/2 {
		newBudget := progress.IterationsComplete + (progress.IterationsTotal-progress.IterationsComplete)/2
		return &searchdef.SearchRecommendation{
			Action:     searchdef.SearchAdjustBudget,
			Confidence: 0.55,
			Reasons:    []string{"adaptive_low_rate", "half_remaining"},
			NewBudget:  newBudget,
		}
	}

	return nil
}
