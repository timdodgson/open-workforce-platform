// policy_executor.go — PolicySearchHookRunner (SI 2.0 search-level policy execution).
//
// Note: there is no type named PolicyExecutor; this file implements
// PolicySearchHookRunner, which extends SearchHookRunner with learned policies.
//
// Integrates learned policies into the search loop via SearchHookRunner.
// When wired, replaces rule-only decisions with policy decisions where confidence is sufficient.
// Today: unit-tested scaffold; search.go still uses SearchHookRunner (v1 path).
//
// CLI: --policy-mode rules|hybrid|learned (orthogonal to --worker-decision-mode)
//
//	rules: RulePolicy / SearchHookRunner behaviour
//	hybrid: learned when confident, rules as fallback
//	learned: learned policy for all decisions (rules only for safety)
//
// Every decision records: policy_used, fallback_reason, confidence,
// safety_override, decision_source.
package optimisation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ───────────────────────────────────────────────────────────────
// Policy-Aware Search Hook Runner
// ───────────────────────────────────────────────────────────────

// PolicySearchHookRunner extends SearchHookRunner with policy-based decisions.
// In "rules" mode: delegates to the standard rule-based hook runner.
// In "hybrid" mode: uses learned policy when confident, falls back to rules.
// In "learned" mode: uses learned policy for all decisions (rules only for safety).
type PolicySearchHookRunner struct {
	*SearchHookRunner

	policyMode string // "rules", "hybrid", "learned"
	stagnation *LearnedStagnationDetector
	restart    *RestartPolicy
	confidence float64 // threshold for hybrid fallback

	// Telemetry.
	decisions []PolicySearchDecision
}

// PolicySearchDecision records one policy-aware decision.
type PolicySearchDecision struct {
	Checkpoint     int
	Candidates     int
	PolicyMode     string // which mode was active
	PolicyUsed     string // "rule", "learned", "hybrid_learned", "hybrid_rule"
	Action         string
	Confidence     float64
	FallbackReason string
	SafetyOverride bool
}

// PolicySearchConfig configures policy-based search hooks.
type PolicySearchConfig struct {
	PolicyMode          string  // "rules", "hybrid", "learned"
	PolicyDir           string  // directory containing policy JSON files
	ConfidenceThreshold float64 // below this → fallback to rules (default 0.60)
}

// NewPolicySearchHookRunner creates a policy-aware hook runner.
// Returns nil if assist mode is "off".
func NewPolicySearchHookRunner(assistMode string, assistConfig SearchAssistConfig, iterations int, policyConfig PolicySearchConfig) *PolicySearchHookRunner {
	base := NewSearchHookRunner(assistMode, assistConfig, iterations)
	if base == nil {
		return nil
	}

	policyMode := policyConfig.PolicyMode
	if policyMode == "" {
		policyMode = "rules"
	}

	threshold := policyConfig.ConfidenceThreshold
	if threshold <= 0 {
		threshold = 0.60
	}

	runner := &PolicySearchHookRunner{
		SearchHookRunner: base,
		policyMode:       policyMode,
		confidence:       threshold,
	}

	// Load learned policies if available.
	if policyMode != "rules" && policyConfig.PolicyDir != "" {
		runner.loadPolicies(policyConfig.PolicyDir)
	}

	return runner
}

// Policy JSON filenames loaded by PolicySearchHookRunner.loadPolicies.
const (
	stagnationPolicyFile = "stagnation_policy.json"
	restartPolicyFile    = "restart_policy.json"
)

func (r *PolicySearchHookRunner) loadPolicies(dir string) {
	stagnationPath := filepath.Join(dir, stagnationPolicyFile)
	if model, err := LoadImprovementCurveModel(stagnationPath); err == nil {
		r.stagnation = NewLearnedStagnationDetector(model, DefaultStagnationPolicyConfig())
	}

	restartPath := filepath.Join(dir, restartPolicyFile)
	if model, err := LoadRestartModel(restartPath); err == nil {
		r.restart = NewRestartPolicy(model, DefaultRestartPolicyConfig())
	}
}

// RunPolicyCheckpoint evaluates search state using the configured policy.
// assistMode shadow/assist/adaptive: delegates to SearchHookRunner first.
// policyMode rules/hybrid/learned: layers policy decisions on top (when models loaded).
func (r *PolicySearchHookRunner) RunPolicyCheckpoint(algorithm string, candidates int, currentPenalty int, bestPenalty int, initialPenalty int, temperature float64) SearchAction {
	if r == nil {
		return SearchContinue
	}

	// Always run the base checkpoint (records telemetry regardless of policy).
	baseAction := r.RunCheckpoint(algorithm, candidates, currentPenalty, bestPenalty, initialPenalty, temperature)

	// If policy mode is "rules", the base action IS the final action.
	if r.policyMode == "rules" {
		r.recordDecision(candidates, "rule", string(baseAction), float64(r.assist.config.MinBudgetFraction), "", false)
		return baseAction
	}

	// Build feature vector for policy evaluation.
	budgetConsumed := 0.0
	if r.iterationsTotal > 0 {
		budgetConsumed = float64(candidates) / float64(r.iterationsTotal)
	}
	plateauLength := candidates - r.lastImproveAt

	features := FeatureVector{
		Problem:            "", // filled by caller if needed
		Algorithm:          algorithm,
		IterationBudget:    r.iterationsTotal,
		IterationsComplete: candidates,
		BudgetConsumed:     budgetConsumed,
		Temperature:        temperature,
		CurrentObjective:   currentPenalty,
		BestObjective:      bestPenalty,
		ParentObjective:    initialPenalty,
		DistanceFromBest:   currentPenalty - bestPenalty,
		PlateauLength:      plateauLength,
		AcceptanceRate:     0, // not available at checkpoint
	}

	// Evaluate learned stagnation policy.
	if r.stagnation != nil {
		assessment := r.stagnation.Assess(features)

		if r.policyMode == "learned" {
			// Learned mode: policy decides directly.
			if assessment.RecommendEarlyStop && assessment.PolicyConfidence >= r.confidence {
				r.recordDecision(candidates, "learned", "early_stop", assessment.PolicyConfidence, "", false)
				return SearchEarlyStop
			}
			// Learned says continue.
			r.recordDecision(candidates, "learned", "continue", assessment.PolicyConfidence, "", false)
			return SearchContinue
		}

		// Hybrid mode: use learned if confident, else fall back to rule.
		if assessment.RecommendEarlyStop && assessment.PolicyConfidence >= r.confidence {
			r.recordDecision(candidates, "hybrid_learned", "early_stop", assessment.PolicyConfidence, "", false)
			return SearchEarlyStop
		}

		if assessment.PolicyConfidence < r.confidence && baseAction == SearchEarlyStop {
			// Low confidence learned but rule says stop — use rule with fallback note.
			r.recordDecision(candidates, "hybrid_rule", string(baseAction), assessment.PolicyConfidence, "learned_low_confidence", false)
			return baseAction
		}

		// Learned says continue with sufficient confidence — override rule if rule wanted stop.
		if !assessment.RecommendEarlyStop && assessment.PolicyConfidence >= r.confidence && baseAction == SearchEarlyStop {
			r.recordDecision(candidates, "hybrid_learned", "continue", assessment.PolicyConfidence, "learned_overrides_rule", false)
			return SearchContinue
		}
	}

	// Default: use base (rule) action.
	r.recordDecision(candidates, "rule", string(baseAction), 0.5, "no_learned_assessment", false)
	return baseAction
}

func (r *PolicySearchHookRunner) recordDecision(candidates int, policyUsed string, action string, confidence float64, fallbackReason string, safetyOverride bool) {
	r.decisions = append(r.decisions, PolicySearchDecision{
		Checkpoint:     r.checkpointNum,
		Candidates:     candidates,
		PolicyMode:     r.policyMode,
		PolicyUsed:     policyUsed,
		Action:         action,
		Confidence:     confidence,
		FallbackReason: fallbackReason,
		SafetyOverride: safetyOverride,
	})
}

// Decisions returns all policy decisions made during this search.
func (r *PolicySearchHookRunner) Decisions() []PolicySearchDecision {
	if r == nil {
		return nil
	}
	return r.decisions
}

// WritePolicyDecisionsCSV writes the policy decisions to a CSV file.
func WritePolicyDecisionsCSV(path string, decisions []PolicySearchDecision) error {
	if len(decisions) == 0 {
		return nil
	}

	header := "checkpoint,candidates,policy_mode,policy_used,action,confidence,fallback_reason,safety_override\n"
	var rows string
	for _, d := range decisions {
		safety := "0"
		if d.SafetyOverride {
			safety = "1"
		}
		rows += fmt.Sprintf("%d,%d,%s,%s,%s,%.4f,%s,%s\n",
			d.Checkpoint, d.Candidates, d.PolicyMode, d.PolicyUsed,
			d.Action, d.Confidence, d.FallbackReason, safety)
	}

	return os.WriteFile(path, []byte(header+rows), 0644)
}

// ───────────────────────────────────────────────────────────────
// Policy Loading Utilities
// ───────────────────────────────────────────────────────────────

// LoadPolicyRegistryFile loads policy_v1.json (runtime type→file map).
// This is separate from LoadPolicyRegistry which reads policy_registry.json
// (lifecycle version records). The two JSON schemas are intentionally different.
func LoadPolicyRegistryFile(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var reg struct {
		Policies []struct {
			Type   string `json:"type"`
			File   string `json:"file"`
			Status string `json:"status"`
		} `json:"policies"`
	}
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, p := range reg.Policies {
		result[p.Type] = p.File
	}
	return result, nil
}
