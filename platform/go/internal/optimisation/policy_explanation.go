// policy_explanation.go — Explainable Policy Decisions.
//
// Every policy decision produces a structured explanation:
//   - Which features contributed to the decision
//   - How much each feature contributed (SHAP-like attribution)
//   - A natural language summary
//
// This makes policy behaviour transparent and auditable.
// Explanations are recorded alongside decisions for dashboard display.
package optimisation

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ───────────────────────────────────────────────────────────────
// Feature Contribution
// ───────────────────────────────────────────────────────────────

// FeatureContribution describes how one feature influenced the decision.
type FeatureContribution struct {
	// Feature name (e.g. "plateau_length", "improvement_rate").
	Feature string `json:"feature"`

	// Raw value of the feature at decision time.
	Value float64 `json:"value"`

	// Contribution to the decision (positive = pushed toward action, negative = against).
	Contribution float64 `json:"contribution"`

	// Direction: "for" or "against" the recommended action.
	Direction string `json:"direction"`

	// Human-readable interpretation.
	Interpretation string `json:"interpretation"`
}

// ───────────────────────────────────────────────────────────────
// Policy Explanation
// ───────────────────────────────────────────────────────────────

// PolicyExplanation is the full explanation for one policy decision.
type PolicyExplanation struct {
	// The action being explained.
	Action string `json:"action"`

	// Confidence in the decision.
	Confidence float64 `json:"confidence"`

	// Feature contributions ranked by absolute importance.
	Contributions []FeatureContribution `json:"contributions"`

	// Natural language summary.
	Summary string `json:"summary"`

	// One-line reason (for CSV/logs).
	ReasonCode string `json:"reason_code"`
}

// ───────────────────────────────────────────────────────────────
// Explanation Builder
// ───────────────────────────────────────────────────────────────

// ExplanationBuilder constructs explanations from feature vectors and decisions.
type ExplanationBuilder struct{}

// NewExplanationBuilder creates a builder.
func NewExplanationBuilder() *ExplanationBuilder {
	return &ExplanationBuilder{}
}

// Explain produces a full explanation for a policy decision given the features.
func (b *ExplanationBuilder) Explain(features FeatureVector, decision PolicyDecision) PolicyExplanation {
	contributions := b.computeContributions(features, decision)

	// Sort by absolute contribution (most important first).
	sort.Slice(contributions, func(i, j int) bool {
		return math.Abs(contributions[i].Contribution) > math.Abs(contributions[j].Contribution)
	})

	// Keep top contributions (max 6).
	if len(contributions) > 6 {
		contributions = contributions[:6]
	}

	summary := b.buildSummary(decision, contributions)
	reasonCode := b.buildReasonCode(contributions)

	return PolicyExplanation{
		Action:        decision.Action,
		Confidence:    decision.Confidence,
		Contributions: contributions,
		Summary:       summary,
		ReasonCode:    reasonCode,
	}
}

func (b *ExplanationBuilder) computeContributions(f FeatureVector, d PolicyDecision) []FeatureContribution {
	var contributions []FeatureContribution

	// Budget consumed.
	if f.BudgetConsumed > 0 {
		contrib := b.budgetContribution(f, d)
		if contrib.Contribution != 0 {
			contributions = append(contributions, contrib)
		}
	}

	// Plateau length.
	if f.PlateauLength > 0 {
		contrib := b.plateauContribution(f, d)
		contributions = append(contributions, contrib)
	}

	// Improvement rate.
	contrib := b.improvementRateContribution(f, d)
	if contrib.Contribution != 0 {
		contributions = append(contributions, contrib)
	}

	// Distance from best.
	if f.DistanceFromBest > 0 {
		contrib := b.distanceContribution(f, d)
		contributions = append(contributions, contrib)
	}

	// Temperature (SA only).
	if f.Temperature > 0 {
		contrib := b.temperatureContribution(f, d)
		contributions = append(contributions, contrib)
	}

	// Entropy (beam search).
	if f.Entropy > 0 {
		contrib := b.entropyContribution(f, d)
		contributions = append(contributions, contrib)
	}

	// Beam health.
	if f.BeamHealth > 0 {
		contrib := b.beamHealthContribution(f, d)
		contributions = append(contributions, contrib)
	}

	// Acceptance rate.
	if f.AcceptanceRate > 0 {
		contrib := b.acceptanceContribution(f, d)
		if contrib.Contribution != 0 {
			contributions = append(contributions, contrib)
		}
	}

	return contributions
}

func (b *ExplanationBuilder) budgetContribution(f FeatureVector, d PolicyDecision) FeatureContribution {
	// High budget consumed → pushes toward stop/reduce. Low → pushes toward continue/extend.
	isStopAction := d.Action == "early_stop" || d.Action == "skip" || d.Action == "reduce_budget"
	contribution := (f.BudgetConsumed - 0.5) * 2.0 // normalize around midpoint
	if !isStopAction {
		contribution = -contribution
	}

	direction := "for"
	interp := fmt.Sprintf("%.0f%% budget consumed", f.BudgetConsumed*100)
	if contribution < 0 {
		direction = "against"
	}

	return FeatureContribution{
		Feature: "budget_consumed", Value: f.BudgetConsumed,
		Contribution: contribution, Direction: direction, Interpretation: interp,
	}
}

func (b *ExplanationBuilder) plateauContribution(f FeatureVector, d PolicyDecision) FeatureContribution {
	// Long plateau → pushes toward stop/restart. Short → pushes toward continue.
	plateauRatio := 0.0
	if f.IterationBudget > 0 {
		plateauRatio = float64(f.PlateauLength) / float64(f.IterationBudget)
	}

	isStopAction := d.Action == "early_stop" || d.Action == "restart" || d.Action == "skip"
	contribution := plateauRatio * 3.0 // plateau is strong signal
	if !isStopAction {
		contribution = -contribution
	}

	direction := "for"
	interp := fmt.Sprintf("plateau %d iterations (%.0f%% of budget)", f.PlateauLength, plateauRatio*100)
	if contribution < 0 {
		direction = "against"
	}

	return FeatureContribution{
		Feature: "plateau_length", Value: float64(f.PlateauLength),
		Contribution: contribution, Direction: direction, Interpretation: interp,
	}
}

func (b *ExplanationBuilder) improvementRateContribution(f FeatureVector, d PolicyDecision) FeatureContribution {
	// High improvement rate → pushes toward continue/extend. Low → pushes toward stop.
	isExtendAction := d.Action == "continue" || d.Action == "extend" || d.Action == "run" || d.Action == "allocate"
	contribution := f.ImprovementRate * 0.5
	if !isExtendAction {
		contribution = -contribution
	}

	direction := "for"
	interp := fmt.Sprintf("improvement rate %.2f per 10K", f.ImprovementRate)
	if f.ImprovementRate == 0 {
		interp = "no recent improvements"
	}
	if contribution < 0 {
		direction = "against"
	}

	return FeatureContribution{
		Feature: "improvement_rate", Value: f.ImprovementRate,
		Contribution: contribution, Direction: direction, Interpretation: interp,
	}
}

func (b *ExplanationBuilder) distanceContribution(f FeatureVector, d PolicyDecision) FeatureContribution {
	// Large distance from best → pushes toward skip (low value worker).
	isSkipAction := d.Action == "skip" || d.Action == "reduce_budget"
	normalised := math.Min(float64(f.DistanceFromBest)/1000.0, 2.0)
	contribution := normalised
	if !isSkipAction {
		contribution = -contribution
	}

	direction := "for"
	interp := fmt.Sprintf("distance from best: %d", f.DistanceFromBest)
	if contribution < 0 {
		direction = "against"
	}

	return FeatureContribution{
		Feature: "distance_from_best", Value: float64(f.DistanceFromBest),
		Contribution: contribution, Direction: direction, Interpretation: interp,
	}
}

func (b *ExplanationBuilder) temperatureContribution(f FeatureVector, d PolicyDecision) FeatureContribution {
	// Low temperature → search is converging → pushes toward stop if stagnating.
	isStopAction := d.Action == "early_stop" || d.Action == "restart"
	// Normalize: assume initial temp ~100, low is < 1.
	normalised := math.Max(0, 1.0-math.Log10(f.Temperature+1)/2.0)
	contribution := normalised * 0.5
	if !isStopAction {
		contribution = -contribution
	}

	direction := "for"
	interp := fmt.Sprintf("temperature %.4f", f.Temperature)
	if f.Temperature < 0.01 {
		interp = "temperature near zero (converged)"
	}
	if contribution < 0 {
		direction = "against"
	}

	return FeatureContribution{
		Feature: "temperature", Value: f.Temperature,
		Contribution: contribution, Direction: direction, Interpretation: interp,
	}
}

func (b *ExplanationBuilder) entropyContribution(f FeatureVector, d PolicyDecision) FeatureContribution {
	// Low entropy → beam collapse → pushes toward diversification/increase.
	isSkipAction := d.Action == "skip" || d.Action == "reduce_budget"
	contribution := (1.0 - f.Entropy) * 0.8 // low entropy = high contribution to skip
	if !isSkipAction {
		contribution = -contribution
	}

	direction := "for"
	interp := fmt.Sprintf("lineage entropy %.2f", f.Entropy)
	if f.Entropy < 0.3 {
		interp = "low diversity (beam collapse risk)"
	} else if f.Entropy > 0.7 {
		interp = "healthy diversity"
	}
	if contribution < 0 {
		direction = "against"
	}

	return FeatureContribution{
		Feature: "entropy", Value: f.Entropy,
		Contribution: contribution, Direction: direction, Interpretation: interp,
	}
}

func (b *ExplanationBuilder) beamHealthContribution(f FeatureVector, d PolicyDecision) FeatureContribution {
	// High beam health → run workers. Low → skip.
	isRunAction := d.Action == "run" || d.Action == "increase_budget"
	normalised := f.BeamHealth / 100.0
	contribution := normalised * 0.6
	if !isRunAction {
		contribution = -contribution
	}

	direction := "for"
	interp := fmt.Sprintf("beam health %.0f/100", f.BeamHealth)
	if contribution < 0 {
		direction = "against"
	}

	return FeatureContribution{
		Feature: "beam_health", Value: f.BeamHealth,
		Contribution: contribution, Direction: direction, Interpretation: interp,
	}
}

func (b *ExplanationBuilder) acceptanceContribution(f FeatureVector, d PolicyDecision) FeatureContribution {
	// High acceptance rate → search is productive. Low → stagnating.
	isExtendAction := d.Action == "continue" || d.Action == "extend" || d.Action == "run"
	contribution := f.AcceptanceRate * 5.0 // amplify small rates
	if !isExtendAction {
		contribution = -contribution
	}

	direction := "for"
	interp := fmt.Sprintf("acceptance rate %.2f%%", f.AcceptanceRate*100)
	if contribution < 0 {
		direction = "against"
	}

	return FeatureContribution{
		Feature: "acceptance_rate", Value: f.AcceptanceRate,
		Contribution: contribution, Direction: direction, Interpretation: interp,
	}
}

// ───────────────────────────────────────────────────────────────
// Natural Language Summary
// ───────────────────────────────────────────────────────────────

func (b *ExplanationBuilder) buildSummary(d PolicyDecision, contributions []FeatureContribution) string {
	if len(contributions) == 0 {
		return fmt.Sprintf("%s (no feature contributions)", d.Action)
	}

	actionDesc := describeAction(d.Action)

	// Build "because" clause from top contributions.
	forReasons := []string{}
	againstReasons := []string{}
	for _, c := range contributions {
		if c.Direction == "for" {
			forReasons = append(forReasons, c.Interpretation)
		} else {
			againstReasons = append(againstReasons, c.Interpretation)
		}
	}

	var parts []string
	parts = append(parts, actionDesc)

	if len(forReasons) > 0 {
		parts = append(parts, "because: "+strings.Join(forReasons, ", "))
	}

	if len(againstReasons) > 0 && len(againstReasons) <= 2 {
		parts = append(parts, "despite: "+strings.Join(againstReasons, ", "))
	}

	parts = append(parts, fmt.Sprintf("(confidence %.2f)", d.Confidence))

	return strings.Join(parts, ". ")
}

func describeAction(action string) string {
	descriptions := map[string]string{
		"run":              "Worker submitted",
		"skip":             "Worker skipped",
		"early_stop":       "Search stopped early",
		"continue":         "Search continues",
		"extend":           "Budget extended",
		"restart":          "Search restarted",
		"allocate":         "Budget allocated",
		"reduce_budget":    "Budget reduced",
		"increase_budget":  "Budget increased",
		"change_algorithm": "Algorithm changed",
	}
	if desc, ok := descriptions[action]; ok {
		return desc
	}
	return action
}

func (b *ExplanationBuilder) buildReasonCode(contributions []FeatureContribution) string {
	if len(contributions) == 0 {
		return "no_signal"
	}
	// Take top 3 features as reason code.
	codes := []string{}
	for i, c := range contributions {
		if i >= 3 {
			break
		}
		codes = append(codes, fmt.Sprintf("%s:%s", c.Feature, c.Direction))
	}
	return strings.Join(codes, ";")
}
