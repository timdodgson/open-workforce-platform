// stagnation_policy.go — Learned stagnation detection via improvement curve models.
//
// Replaces fixed stagnation windows with learned probability of future improvement.
// The model learns expected improvement curves per domain/algorithm/instance from
// historical telemetry.
//
// Flow:
//   Historical runs → ImprovementCurveModel → P(improve | state)
//   At each checkpoint: if P(improve) < threshold → recommend early stop
//
// Key metrics:
//   - ExpectedRemainingValue: estimated objective improvement still possible
//   - StagnationConfidence: how certain we are that search has stagnated
//   - PolicyConfidence: overall confidence in the recommendation
package optimisation

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

// ───────────────────────────────────────────────────────────────
// Improvement Curve Model
// ───────────────────────────────────────────────────────────────

// ImprovementCurveModel captures learned improvement behaviour for a
// specific domain/algorithm combination. Built from historical telemetry.
type ImprovementCurveModel struct {
	Version   string                    `json:"version"`
	TrainedOn int                       `json:"trained_on"`
	Curves    []ImprovementCurveEntry   `json:"curves"`
}

// ImprovementCurveEntry is a learned curve for one domain/algorithm/instance.
type ImprovementCurveEntry struct {
	Domain    string `json:"domain"`
	Algorithm string `json:"algorithm"`
	Instance  string `json:"instance,omitempty"` // empty = domain-wide

	// Learned parameters describing the improvement curve.
	// The model uses an exponential decay: P(improve) = A * exp(-λ * plateau_ratio)
	// where plateau_ratio = plateau_length / budget.
	DecayRate   float64 `json:"decay_rate"`   // λ — higher = faster stagnation
	Amplitude   float64 `json:"amplitude"`    // A — initial improvement probability
	HalfLife    float64 `json:"half_life"`    // plateau_ratio at which P(improve) = 0.5

	// Historical statistics for calibration.
	MeanImprovements    float64 `json:"mean_improvements"`     // mean improvement count per run
	MeanLastImproveAt   float64 `json:"mean_last_improve_at"`  // mean budget fraction of last improvement
	StdLastImproveAt    float64 `json:"std_last_improve_at"`   // std of above

	SampleCount int     `json:"sample_count"`
	Confidence  float64 `json:"confidence"` // model confidence for this curve
}

// LoadImprovementCurveModel loads the model from JSON.
func LoadImprovementCurveModel(path string) (*ImprovementCurveModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load improvement curve model: %w", err)
	}
	var model ImprovementCurveModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("parse improvement curve model: %w", err)
	}
	if model.Version == "" || len(model.Curves) == 0 {
		return nil, fmt.Errorf("improvement curve model: missing version or curves")
	}
	return &model, nil
}

// ───────────────────────────────────────────────────────────────
// Stagnation Assessment
// ───────────────────────────────────────────────────────────────

// StagnationAssessment is the output of stagnation analysis at a decision point.
type StagnationAssessment struct {
	// Probability that search will find another improvement given current state.
	ProbImprove float64

	// Estimated remaining objective improvement (expected value).
	ExpectedRemainingValue float64

	// How confident the model is that search has stagnated (0 = uncertain, 1 = certain).
	StagnationConfidence float64

	// Overall confidence in this assessment.
	PolicyConfidence float64

	// Whether the model recommends early stop.
	RecommendEarlyStop bool

	// Reason for recommendation.
	Reason string
}

// ───────────────────────────────────────────────────────────────
// Stagnation Policy (Learned)
// ───────────────────────────────────────────────────────────────

// StagnationPolicyConfig configures the learned stagnation detector.
type StagnationPolicyConfig struct {
	// Threshold below which P(improve) triggers early stop recommendation.
	// Default: 0.10 (10% probability of future improvement).
	StopThreshold float64

	// Minimum budget fraction consumed before policy can recommend stop.
	// Default: 0.20.
	MinBudgetFraction float64

	// Minimum confidence in the curve model before applying learned decision.
	// Default: 0.60.
	MinConfidence float64
}

// DefaultStagnationPolicyConfig returns sensible defaults.
func DefaultStagnationPolicyConfig() StagnationPolicyConfig {
	return StagnationPolicyConfig{
		StopThreshold:     0.10,
		MinBudgetFraction: 0.20,
		MinConfidence:     0.60,
	}
}

// LearnedStagnationDetector uses improvement curve models to assess stagnation.
type LearnedStagnationDetector struct {
	model  *ImprovementCurveModel
	config StagnationPolicyConfig
}

// NewLearnedStagnationDetector creates the detector.
// If model is nil, Assess always returns a low-confidence "continue" assessment.
func NewLearnedStagnationDetector(model *ImprovementCurveModel, config StagnationPolicyConfig) *LearnedStagnationDetector {
	if config.StopThreshold <= 0 {
		config.StopThreshold = 0.10
	}
	if config.MinBudgetFraction <= 0 {
		config.MinBudgetFraction = 0.20
	}
	if config.MinConfidence <= 0 {
		config.MinConfidence = 0.60
	}
	return &LearnedStagnationDetector{model: model, config: config}
}

// Assess evaluates current search state and returns a stagnation assessment.
func (d *LearnedStagnationDetector) Assess(features FeatureVector) StagnationAssessment {
	if d.model == nil {
		return StagnationAssessment{
			ProbImprove:          0.5,
			StagnationConfidence: 0.0,
			PolicyConfidence:     0.0,
			Reason:               "no_model_loaded",
		}
	}

	curve := d.findCurve(features.Problem, features.Algorithm, features.Instance)
	if curve == nil {
		return StagnationAssessment{
			ProbImprove:          0.5,
			StagnationConfidence: 0.0,
			PolicyConfidence:     0.0,
			Reason:               "no_curve_for_config",
		}
	}

	// Compute plateau ratio (how much of the budget has been spent in a plateau).
	budgetTotal := features.IterationBudget
	if budgetTotal <= 0 {
		budgetTotal = 1
	}
	plateauRatio := float64(features.PlateauLength) / float64(budgetTotal)

	// Compute P(improve) using exponential decay model.
	probImprove := curve.Amplitude * math.Exp(-curve.DecayRate*plateauRatio)
	if probImprove > 1.0 {
		probImprove = 1.0
	}
	if probImprove < 0.0 {
		probImprove = 0.0
	}

	// Stagnation confidence = 1 - P(improve), scaled by model confidence.
	stagnationConf := (1.0 - probImprove) * curve.Confidence

	// Expected remaining value: if we were to continue, how much improvement is likely.
	// Use historical mean improvement rate scaled by P(improve).
	expectedRemaining := probImprove * curve.MeanImprovements * features.AcceptanceRate * 100

	// Policy confidence is model confidence, penalised if few samples.
	policyConf := curve.Confidence
	if curve.SampleCount < 10 {
		policyConf *= float64(curve.SampleCount) / 10.0
	}

	// Decide whether to recommend early stop.
	budgetConsumed := features.BudgetConsumed
	recommend := false
	reason := "continue"

	if budgetConsumed < d.config.MinBudgetFraction {
		reason = "below_min_budget"
	} else if policyConf < d.config.MinConfidence {
		reason = fmt.Sprintf("low_confidence_%.2f", policyConf)
	} else if probImprove < d.config.StopThreshold {
		recommend = true
		reason = fmt.Sprintf("p_improve_%.3f_below_threshold_%.2f", probImprove, d.config.StopThreshold)
	} else {
		reason = fmt.Sprintf("p_improve_%.3f_above_threshold", probImprove)
	}

	return StagnationAssessment{
		ProbImprove:            probImprove,
		ExpectedRemainingValue: expectedRemaining,
		StagnationConfidence:   stagnationConf,
		PolicyConfidence:       policyConf,
		RecommendEarlyStop:     recommend,
		Reason:                 reason,
	}
}

func (d *LearnedStagnationDetector) findCurve(domain, algorithm, instance string) *ImprovementCurveEntry {
	var domainMatch *ImprovementCurveEntry
	for i := range d.model.Curves {
		c := &d.model.Curves[i]
		if c.Domain != domain || c.Algorithm != algorithm {
			continue
		}
		if c.Instance != "" && c.Instance == instance {
			return c
		}
		if c.Instance == "" {
			domainMatch = c
		}
	}
	return domainMatch
}
