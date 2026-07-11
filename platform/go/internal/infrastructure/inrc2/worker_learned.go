package inrc2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

const workerPolicyFilename = "worker_policy.json"

// WorkerPolicyModel is the trained worker spawn policy (metadata + optional tree).
type WorkerPolicyModel struct {
	Version       string       `json:"version"`
	Status        string       `json:"status"`
	Domain        string       `json:"domain"`
	DecisionType  string       `json:"decision_type"`
	Accuracy      float64      `json:"accuracy"`
	PositiveRate  float64      `json:"positive_rate"`
	FeaturesUsed  []string     `json:"features_used"`
	LabelColumn   string       `json:"label_column"`
	Tree          *optimisation.SklearnTree `json:"tree,omitempty"`
	Bandit        *optimisation.BanditPolicy `json:"bandit,omitempty"`
}

// LoadWorkerPolicy loads worker_policy.json from a policy directory.
func LoadWorkerPolicy(policyDir string) (*WorkerPolicyModel, error) {
	path := filepath.Join(policyDir, workerPolicyFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load worker policy: %w", err)
	}
	var model WorkerPolicyModel
	if err := json.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("parse worker policy: %w", err)
	}
	if model.Status != "trained" {
		return nil, fmt.Errorf("worker policy status: %s", model.Status)
	}
	if model.Tree == nil {
		return nil, fmt.Errorf("worker policy: no tree exported (retrain with latest train_policies.py)")
	}
	return &model, nil
}

// LearnedWorkerDecisionEngine uses a trained decision tree at worker spawn time.
type LearnedWorkerDecisionEngine struct {
	model     *WorkerPolicyModel
	threshold float64
}

// NewLearnedWorkerDecisionEngine creates a learned worker engine.
func NewLearnedWorkerDecisionEngine(model *WorkerPolicyModel, threshold float64) *LearnedWorkerDecisionEngine {
	if threshold <= 0 {
		threshold = 0.60
	}
	return &LearnedWorkerDecisionEngine{model: model, threshold: threshold}
}

// Evaluate predicts whether a worker is worth running.
func (e *LearnedWorkerDecisionEngine) Evaluate(input WorkerDecisionInput) WorkerDecision {
	if e == nil || e.model == nil {
		return WorkerDecision{
			Recommendation: RecRun,
			Confidence:     0.5,
			ReasonCodes:    []string{"no_learned_model"},
		}
	}

	if e.model.Bandit != nil {
		ctx := optimisation.WorkerContextKey(input.Week, input.Depth, float64(input.DistanceFromBest))
		if arm, mult, ok := optimisation.WorkerBanditDecision(e.model.Bandit, ctx); ok {
			switch arm {
			case "skip":
				return WorkerDecision{
					Recommendation:  RecSkip,
					Confidence:      0.7,
					ReasonCodes:     []string{"bandit_skip"},
					SuggestedBudget: 0,
				}
			case "boost":
				budget := input.AllocatedIters
				if mult > 1.0 {
					budget = int(float64(input.AllocatedIters) * mult)
				}
				return WorkerDecision{
					Recommendation:  RecRun,
					Confidence:      0.7,
					ReasonCodes:     []string{"bandit_boost"},
					SuggestedBudget: budget,
				}
			default:
				return WorkerDecision{
					Recommendation:  RecRun,
					Confidence:      0.65,
					ReasonCodes:     []string{"bandit_default"},
					SuggestedBudget: input.AllocatedIters,
				}
			}
		}
	}

	if e.model.Tree == nil {
		return WorkerDecision{
			Recommendation: RecRun,
			Confidence:     0.5,
			ReasonCodes:    []string{"no_learned_model"},
		}
	}

	features := workerFeatures(input, e.model.FeaturesUsed)
	prob := e.model.Tree.PositiveClassProbability(features)
	confidence := prob
	if prob < 0.5 {
		confidence = 1 - prob
	}

	if prob >= 0.5 {
		return WorkerDecision{
			Recommendation:  RecRun,
			Confidence:      confidence,
			ReasonCodes:     []string{"learned_run"},
			SuggestedBudget: input.AllocatedIters,
		}
	}

	rec := RecReduceBudget
	if prob < 0.35 {
		rec = RecSkip
	}
	return WorkerDecision{
		Recommendation:  rec,
		Confidence:      confidence,
		ReasonCodes:     []string{"learned_low_value"},
		SuggestedBudget: input.AllocatedIters / 2,
	}
}

// HybridWorkerDecisionEngine combines learned and rule-based worker decisions.
type HybridWorkerDecisionEngine struct {
	learned  *LearnedWorkerDecisionEngine
	rules    *RuleBasedWorkerDecisionEngine
	mode     string
	threshold float64
}

// NewHybridWorkerDecisionEngine creates a hybrid worker engine.
func NewHybridWorkerDecisionEngine(mode string, policyDir string) WorkerDecisionEngine {
	rules := NewRuleBasedEngine()
	if mode == "rules" || policyDir == "" {
		return rules
	}

	model, err := LoadWorkerPolicy(policyDir)
	if err != nil {
		return rules
	}

	learned := NewLearnedWorkerDecisionEngine(model, 0.60)
	if mode == "learned" {
		return learned
	}

	return &HybridWorkerDecisionEngine{
		learned:   learned,
		rules:     rules,
		mode:      mode,
		threshold: 0.60,
	}
}

// Evaluate uses learned policy when confident, otherwise falls back to rules.
func (e *HybridWorkerDecisionEngine) Evaluate(input WorkerDecisionInput) WorkerDecision {
	if e == nil {
		return NewRuleBasedEngine().Evaluate(input)
	}
	if e.learned == nil {
		return e.rules.Evaluate(input)
	}

	learned := e.learned.Evaluate(input)
	if learned.Confidence >= e.threshold {
		learned.ReasonCodes = append(learned.ReasonCodes, "hybrid_learned")
		return learned
	}

	rules := e.rules.Evaluate(input)
	rules.ReasonCodes = append(rules.ReasonCodes, "hybrid_rules_fallback")
	return rules
}

func workerFeatures(input WorkerDecisionInput, names []string) []float64 {
	lookup := map[string]float64{
		"week":                 float64(input.Week),
		"depth":                float64(input.Depth),
		"parent_objective":       float64(input.ParentObjective),
		"global_best":          float64(input.GlobalBest),
		"distance_from_best":   float64(input.DistanceFromBest),
		"beam_rank":            float64(input.BeamRank),
		"entropy":              input.Entropy,
		"beam_health":          input.BeamHealth,
		"recent_improv_rate":   input.RecentImprovRate,
		"worker_count":         float64(input.WorkerCount),
		"active_families":      float64(input.ActiveFamilies),
		"iterations_alloc":     float64(input.AllocatedIters),
		"final_budget":         float64(input.AllocatedIters),
		"allocated_iters":      float64(input.AllocatedIters),
		"family_id":            float64(input.FamilyID),
		"generations_since_gb": float64(input.GenerationsSinceGlobalBest),
	}
	if input.IsGlobalBestLineage {
		lookup["is_global_best_lineage"] = 1
	}
	if input.ParentProducedGlobalBest {
		lookup["parent_produced_global_best"] = 1
	}

	out := make([]float64, len(names))
	for i, name := range names {
		out[i] = lookup[name]
	}
	return out
}
