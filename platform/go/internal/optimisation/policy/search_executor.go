package policy

import (
	"path/filepath"
)

const (
	// StagnationPolicyFile is the learned early-stop model filename in --policy-dir.
	StagnationPolicyFile = "stagnation_policy.json"
	// RestartPolicyFile is the learned restart model filename in --policy-dir.
	RestartPolicyFile = "restart_policy.json"
)

// Search hook action strings returned by CheckpointEngine.Evaluate.
const (
	CheckpointContinue  = "continue"
	CheckpointEarlyStop = "early_stop"
	CheckpointRestart   = "restart"
)

// SearchHookSnapshot captures parent hook state for one checkpoint evaluation.
type SearchHookSnapshot struct {
	ShadowMode        bool
	CheckpointNum     int
	IterationsTotal   int
	LastImproveAt     int
	MinBudgetFraction float64
}

// CheckpointInput is one search checkpoint passed to the policy engine.
type CheckpointInput struct {
	Algorithm      string
	Candidates     int
	CurrentPenalty int
	BestPenalty    int
	InitialPenalty int
	Temperature    float64
	BaseAction     string
}

// CheckpointEngine evaluates SI 2.0 stagnation and restart policies at search checkpoints.
type CheckpointEngine struct {
	policyMode     string
	policyDomain   string
	policyInstance string
	confidence     float64
	stagnation     *LearnedStagnationDetector
	restart        *RestartPolicy
	decisions      []PolicySearchDecision
}

// NewCheckpointEngine creates an engine from policy search config.
func NewCheckpointEngine(cfg PolicySearchConfig) *CheckpointEngine {
	policyMode := cfg.PolicyMode
	if policyMode == "" {
		policyMode = "rules"
	}
	threshold := cfg.ConfidenceThreshold
	if threshold <= 0 {
		threshold = 0.60
	}
	e := &CheckpointEngine{
		policyMode:     policyMode,
		policyDomain:   cfg.Domain,
		policyInstance: cfg.Instance,
		confidence:     threshold,
	}
	if policyMode != "rules" && cfg.PolicyDir != "" {
		e.loadPolicies(cfg.PolicyDir)
	}
	return e
}

func (e *CheckpointEngine) loadPolicies(dir string) {
	stagnationPath := filepath.Join(dir, StagnationPolicyFile)
	if model, err := LoadImprovementCurveModel(stagnationPath); err == nil {
		e.stagnation = NewLearnedStagnationDetector(model, DefaultStagnationPolicyConfig())
	}

	restartPath := filepath.Join(dir, RestartPolicyFile)
	if model, err := LoadRestartModel(restartPath); err == nil {
		e.restart = NewRestartPolicy(model, DefaultRestartPolicyConfig())
	}
}

// Evaluate applies policy logic and returns the final checkpoint action string.
func (e *CheckpointEngine) Evaluate(in CheckpointInput, snap SearchHookSnapshot) string {
	if e == nil {
		return CheckpointContinue
	}

	if e.policyMode == "rules" {
		e.recordDecision(snap, in.Candidates, "rule", in.BaseAction, snap.MinBudgetFraction, "", false)
		return in.BaseAction
	}

	features := e.buildFeatures(in, snap)
	var policyAction string

	if e.stagnation != nil {
		assessment := e.stagnation.Assess(features)

		if e.policyMode == "learned" {
			if assessment.RecommendEarlyStop && assessment.PolicyConfidence >= e.confidence {
				e.recordDecision(snap, in.Candidates, "learned", CheckpointEarlyStop, assessment.PolicyConfidence, "", false)
				policyAction = CheckpointEarlyStop
			} else {
				e.recordDecision(snap, in.Candidates, "learned", CheckpointContinue, assessment.PolicyConfidence, "", false)
				policyAction = CheckpointContinue
			}
		} else if assessment.RecommendEarlyStop && assessment.PolicyConfidence >= e.confidence {
			e.recordDecision(snap, in.Candidates, "hybrid_learned", CheckpointEarlyStop, assessment.PolicyConfidence, "", false)
			policyAction = CheckpointEarlyStop
		} else if assessment.PolicyConfidence < e.confidence && in.BaseAction == CheckpointEarlyStop {
			e.recordDecision(snap, in.Candidates, "hybrid_rule", in.BaseAction, assessment.PolicyConfidence, "learned_low_confidence", false)
			policyAction = in.BaseAction
		} else if !assessment.RecommendEarlyStop && assessment.PolicyConfidence >= e.confidence && in.BaseAction == CheckpointEarlyStop {
			e.recordDecision(snap, in.Candidates, "hybrid_learned", CheckpointContinue, assessment.PolicyConfidence, "learned_overrides_rule", false)
			policyAction = CheckpointContinue
		}
	}

	if policyAction == "" {
		e.recordDecision(snap, in.Candidates, "rule", in.BaseAction, 0.5, "no_learned_assessment", false)
		policyAction = in.BaseAction
	}

	if policyAction == CheckpointEarlyStop {
		if snap.ShadowMode {
			return CheckpointContinue
		}
		return CheckpointEarlyStop
	}

	if e.restart != nil {
		restartDec := e.restart.Evaluate(features)
		if restartDec.ShouldRestart && restartDec.Confidence >= e.confidence {
			e.recordDecision(snap, in.Candidates, "restart", CheckpointRestart, restartDec.Confidence, restartDec.Reason, false)
			if snap.ShadowMode {
				return CheckpointContinue
			}
			return CheckpointRestart
		}
	}

	if snap.ShadowMode {
		return CheckpointContinue
	}
	return policyAction
}

// BuildFeatures constructs the feature vector for a checkpoint (exported for tests).
func (e *CheckpointEngine) BuildFeatures(in CheckpointInput, snap SearchHookSnapshot) FeatureVector {
	return e.buildFeatures(in, snap)
}

func (e *CheckpointEngine) buildFeatures(in CheckpointInput, snap SearchHookSnapshot) FeatureVector {
	budgetConsumed := 0.0
	if snap.IterationsTotal > 0 {
		budgetConsumed = float64(in.Candidates) / float64(snap.IterationsTotal)
	}
	return FeatureVector{
		Problem:            e.policyDomain,
		Instance:           e.policyInstance,
		Algorithm:          in.Algorithm,
		IterationBudget:    snap.IterationsTotal,
		IterationsComplete: in.Candidates,
		BudgetConsumed:     budgetConsumed,
		Temperature:        in.Temperature,
		CurrentObjective:   in.CurrentPenalty,
		BestObjective:      in.BestPenalty,
		ParentObjective:    in.InitialPenalty,
		DistanceFromBest:   in.CurrentPenalty - in.BestPenalty,
		PlateauLength:      in.Candidates - snap.LastImproveAt,
		AcceptanceRate:     0,
	}
}

func (e *CheckpointEngine) recordDecision(snap SearchHookSnapshot, candidates int, policyUsed string, action string, confidence float64, fallbackReason string, safetyOverride bool) {
	e.decisions = append(e.decisions, PolicySearchDecision{
		Checkpoint:     snap.CheckpointNum,
		Candidates:     candidates,
		PolicyMode:     e.policyMode,
		PolicyUsed:     policyUsed,
		Action:         action,
		Confidence:     confidence,
		FallbackReason: fallbackReason,
		SafetyOverride: safetyOverride,
	})
}

// Decisions returns all policy decisions recorded during the search.
func (e *CheckpointEngine) Decisions() []PolicySearchDecision {
	if e == nil {
		return nil
	}
	return e.decisions
}

// StagnationDetector exposes the loaded stagnation model for integration tests.
func (e *CheckpointEngine) StagnationDetector() *LearnedStagnationDetector {
	if e == nil {
		return nil
	}
	return e.stagnation
}
