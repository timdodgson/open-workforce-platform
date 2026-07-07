// restart_policy.go — Learned restart decision policy.
//
// Answers four questions at each decision point:
//  1. Should restart? (yes/no based on P(improvement | restart))
//  2. When? (optimal timing based on budget consumed and plateau depth)
//  3. How many? (restart budget: how many iterations to allocate)
//  4. Which algorithm? (which algorithm to restart with)
//
// The policy learns restart effectiveness from historical telemetry:
//   - Restarts that found improvements (positive signal)
//   - Restarts that wasted compute (negative signal)
//   - Optimal restart timing (when in the budget was a restart most effective)
//
// Supports: SA, LAHC, Tabu, Portfolio (each may have different restart behaviour).
//
// Output: restart decisions recorded to counterfactual_learning.csv via the
// standard CounterfactualRecorder.
package optimisation

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"
)

// ───────────────────────────────────────────────────────────────
// Restart Model
// ───────────────────────────────────────────────────────────────

// RestartModel captures learned restart behaviour per domain/algorithm.
type RestartModel struct {
	Version   string              `json:"version"`
	TrainedOn int                 `json:"trained_on"`
	Entries   []RestartModelEntry `json:"entries"`
}

// RestartModelEntry holds learned restart statistics for one configuration.
type RestartModelEntry struct {
	Domain    string `json:"domain"`
	Algorithm string `json:"algorithm"`
	Instance  string `json:"instance,omitempty"`

	// Learned restart timing.
	OptimalBudgetFraction float64 `json:"optimal_budget_fraction"` // when restarts are most effective
	OptimalPlateauRatio   float64 `json:"optimal_plateau_ratio"`   // plateau/budget ratio for restart

	// Effectiveness statistics.
	RestartSuccessRate     float64 `json:"restart_success_rate"`      // fraction of restarts that improved
	MeanImprovAfterRestart float64 `json:"mean_improv_after_restart"` // mean improvement after restart
	MeanWasteIfFailed      float64 `json:"mean_waste_if_failed"`      // mean compute wasted on failed restart

	// Algorithm selection.
	BestRestartAlgorithm  string  `json:"best_restart_algorithm"`   // which algorithm works best for restart
	SameAlgoSuccessRate   float64 `json:"same_algo_success_rate"`   // success rate restarting same algorithm
	SwitchAlgoSuccessRate float64 `json:"switch_algo_success_rate"` // success rate switching algorithm

	// Budget allocation.
	OptimalRestartBudget float64 `json:"optimal_restart_budget"` // fraction of remaining budget

	SampleCount int     `json:"sample_count"`
	Confidence  float64 `json:"confidence"`
}

// LoadRestartModel loads the model from JSON.
func LoadRestartModel(path string) (*RestartModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load restart model: %w", err)
	}
	var model RestartModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("parse restart model: %w", err)
	}
	if model.Version == "" || len(model.Entries) == 0 {
		return nil, fmt.Errorf("restart model: missing version or entries")
	}
	return &model, nil
}

// ───────────────────────────────────────────────────────────────
// Restart Decision
// ───────────────────────────────────────────────────────────────

// RestartDecision is the output of the restart policy.
type RestartDecision struct {
	// Should we restart?
	ShouldRestart bool

	// When — timing assessment.
	TimingScore float64 // 0–1, how appropriate the current moment is for restart

	// How many iterations to allocate to the restart.
	RestartBudget int

	// Which algorithm to restart with.
	RestartAlgorithm string

	// Confidence in this recommendation.
	Confidence float64

	// Expected improvement if restart is taken.
	ExpectedImprovement float64

	// Reason for the decision.
	Reason string

	// Policy provenance.
	PolicyID      string
	PolicyVersion string
}

// ───────────────────────────────────────────────────────────────
// Restart Policy
// ───────────────────────────────────────────────────────────────

// RestartPolicyConfig configures the restart policy.
type RestartPolicyConfig struct {
	// Minimum success probability to recommend restart. Default: 0.30.
	MinSuccessProb float64

	// Minimum budget fraction consumed before restart is considered. Default: 0.25.
	MinBudgetFraction float64

	// Minimum confidence in model before applying learned decision. Default: 0.60.
	MinConfidence float64

	// Maximum fraction of remaining budget to allocate to restart. Default: 0.50.
	MaxRestartBudgetFraction float64
}

// DefaultRestartPolicyConfig returns sensible defaults.
func DefaultRestartPolicyConfig() RestartPolicyConfig {
	return RestartPolicyConfig{
		MinSuccessProb:           0.30,
		MinBudgetFraction:        0.25,
		MinConfidence:            0.60,
		MaxRestartBudgetFraction: 0.50,
	}
}

// RestartPolicy evaluates whether a restart is warranted and with what parameters.
type RestartPolicy struct {
	model  *RestartModel
	config RestartPolicyConfig
	id     string
}

// NewRestartPolicy creates a learned restart policy.
// If model is nil, always returns ShouldRestart=false with low confidence.
func NewRestartPolicy(model *RestartModel, config RestartPolicyConfig) *RestartPolicy {
	if config.MinSuccessProb <= 0 {
		config.MinSuccessProb = 0.30
	}
	if config.MinBudgetFraction <= 0 {
		config.MinBudgetFraction = 0.25
	}
	if config.MinConfidence <= 0 {
		config.MinConfidence = 0.60
	}
	if config.MaxRestartBudgetFraction <= 0 {
		config.MaxRestartBudgetFraction = 0.50
	}

	id := "restart-rules"
	if model != nil {
		id = "restart-learned"
	}

	return &RestartPolicy{model: model, config: config, id: id}
}

// Evaluate assesses restart decision for the current search state.
func (p *RestartPolicy) Evaluate(features FeatureVector) RestartDecision {
	if p.model == nil {
		return p.ruleBasedDecision(features)
	}

	entry := p.findEntry(features.Problem, features.Algorithm, features.Instance)
	if entry == nil {
		return p.ruleBasedDecision(features)
	}

	if entry.Confidence < p.config.MinConfidence {
		d := p.ruleBasedDecision(features)
		d.Reason = fmt.Sprintf("learned_low_confidence_%.2f: %s", entry.Confidence, d.Reason)
		return d
	}

	return p.learnedDecision(features, entry)
}

func (p *RestartPolicy) learnedDecision(features FeatureVector, entry *RestartModelEntry) RestartDecision {
	budgetConsumed := features.BudgetConsumed
	plateauRatio := 0.0
	if features.IterationBudget > 0 {
		plateauRatio = float64(features.PlateauLength) / float64(features.IterationBudget)
	}

	// Safety: don't restart too early.
	if budgetConsumed < p.config.MinBudgetFraction {
		return RestartDecision{
			ShouldRestart: false,
			Confidence:    entry.Confidence,
			Reason:        "below_min_budget",
			PolicyID:      p.id,
			PolicyVersion: p.model.Version,
		}
	}

	// Compute timing score: how close are we to the optimal restart point?
	timingDelta := math.Abs(budgetConsumed - entry.OptimalBudgetFraction)
	timingScore := math.Max(0, 1.0-timingDelta*3.0) // peaks at optimal, decays

	// Compute restart probability based on plateau depth relative to learned optimal.
	plateauScore := 0.0
	if entry.OptimalPlateauRatio > 0 {
		plateauScore = math.Min(1.0, plateauRatio/entry.OptimalPlateauRatio)
	}

	// Combined probability of restart success.
	successProb := entry.RestartSuccessRate * timingScore * math.Max(0.5, plateauScore)

	// Should restart?
	shouldRestart := successProb >= p.config.MinSuccessProb

	// Which algorithm?
	restartAlgo := features.Algorithm // default: same
	if entry.SwitchAlgoSuccessRate > entry.SameAlgoSuccessRate && entry.BestRestartAlgorithm != "" {
		restartAlgo = entry.BestRestartAlgorithm
	}

	// How many iterations?
	remainingBudget := features.IterationBudget - features.IterationsComplete
	restartFraction := entry.OptimalRestartBudget
	if restartFraction > p.config.MaxRestartBudgetFraction {
		restartFraction = p.config.MaxRestartBudgetFraction
	}
	restartBudget := int(float64(remainingBudget) * restartFraction)
	if restartBudget < 1000 {
		restartBudget = 1000
	}

	reason := fmt.Sprintf("success_prob_%.3f_timing_%.2f_plateau_%.2f",
		successProb, timingScore, plateauScore)
	if !shouldRestart {
		reason = "below_threshold: " + reason
	}

	return RestartDecision{
		ShouldRestart:       shouldRestart,
		TimingScore:         timingScore,
		RestartBudget:       restartBudget,
		RestartAlgorithm:    restartAlgo,
		Confidence:          entry.Confidence * timingScore,
		ExpectedImprovement: entry.MeanImprovAfterRestart * successProb,
		Reason:              reason,
		PolicyID:            p.id,
		PolicyVersion:       p.model.Version,
	}
}

func (p *RestartPolicy) ruleBasedDecision(features FeatureVector) RestartDecision {
	budgetConsumed := features.BudgetConsumed
	plateauRatio := 0.0
	if features.IterationBudget > 0 {
		plateauRatio = float64(features.PlateauLength) / float64(features.IterationBudget)
	}

	// Simple rule: restart if plateau > 30% of budget and > 25% budget consumed.
	shouldRestart := budgetConsumed >= p.config.MinBudgetFraction && plateauRatio > 0.30

	remainingBudget := features.IterationBudget - features.IterationsComplete
	restartBudget := remainingBudget / 2
	if restartBudget < 1000 {
		restartBudget = 1000
	}

	reason := "rule:continue"
	if shouldRestart {
		reason = fmt.Sprintf("rule:plateau_ratio_%.2f_above_0.30", plateauRatio)
	}

	return RestartDecision{
		ShouldRestart:    shouldRestart,
		TimingScore:      budgetConsumed,
		RestartBudget:    restartBudget,
		RestartAlgorithm: features.Algorithm,
		Confidence:       0.45,
		Reason:           reason,
		PolicyID:         "restart-rules",
		PolicyVersion:    "1.0.0",
	}
}

func (p *RestartPolicy) findEntry(domain, algorithm, instance string) *RestartModelEntry {
	var domainMatch *RestartModelEntry
	for i := range p.model.Entries {
		e := &p.model.Entries[i]
		if e.Domain != domain || e.Algorithm != algorithm {
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

// Metadata returns policy metadata for the restart policy.
func (p *RestartPolicy) Metadata() PolicyMetadata {
	version := "1.0.0"
	pType := "rule"
	trained := 0
	if p.model != nil {
		version = p.model.Version
		pType = "learned"
		trained = p.model.TrainedOn
	}
	return PolicyMetadata{
		ID:              p.id,
		Version:         version,
		Type:            pType,
		Domain:          "*",
		DecisionType:    "restart",
		TrainedSamples:  trained,
		CreatedAt:       time.Time{},
		ValidationScore: -1,
	}
}

// ───────────────────────────────────────────────────────────────
// Restart Effectiveness Record
// ───────────────────────────────────────────────────────────────

// RestartEffectivenessRecord captures one restart decision and its outcome.
// Used to build training data for future model improvement.
type RestartEffectivenessRecord struct {
	// Context at restart decision point.
	RunID           string
	Domain          string
	Instance        string
	Algorithm       string
	BudgetConsumed  float64
	PlateauLength   int
	ObjectiveBefore int

	// Decision.
	Restarted        bool
	RestartAlgorithm string
	RestartBudget    int
	PolicyID         string
	PolicyVersion    string
	Confidence       float64

	// Outcome.
	ObjectiveAfter    int
	Improved          bool
	ImprovementAmount int
	ComputeUsed       int
	RuntimeMs         int64
}
