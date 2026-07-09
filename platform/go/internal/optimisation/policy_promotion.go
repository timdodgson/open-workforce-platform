// policy_promotion.go — Automatic Policy Promotion.
//
// Moves policies through the lifecycle automatically based on validation:
//
//	Candidate → Shadow → Hybrid → Production
//
// Promotion rules are configurable. Failed policies are never promoted.
// Rollback is always supported.
//
// The promoter evaluates gate conditions and advances policies that pass.
// Policies that fail gates remain at their current stage with a recorded reason.
package optimisation

import (
	"fmt"
	"time"
)

// ───────────────────────────────────────────────────────────────
// Promotion Rules
// ───────────────────────────────────────────────────────────────

// PromotionRules configures automatic promotion thresholds.
type PromotionRules struct {
	// Candidate → Shadow: minimum offline accuracy.
	MinOfflineAccuracy float64 // default 0.65

	// Shadow → Hybrid: minimum shadow runs and accuracy.
	MinShadowRuns     int     // default 20
	MinShadowAccuracy float64 // default 0.60

	// Hybrid → Production: minimum production runs and improvement over rules.
	MinProductionRuns     int     // default 50
	MaxRegretVsRules      float64 // default 0.0 (must be equal or better)
	MinProductionAccuracy float64 // default 0.60

	// Safety: never promote if drift detected.
	BlockOnDrift bool // default true

	// Safety: never promote if safety override rate exceeds this.
	MaxSafetyOverrideRate float64 // default 0.05 (5%)
}

// DefaultPromotionRules returns sensible defaults.
func DefaultPromotionRules() PromotionRules {
	return PromotionRules{
		MinOfflineAccuracy:    MinLearnedPolicyAgreement,
		MinShadowRuns:         20,
		MinShadowAccuracy:     0.60,
		MinProductionRuns:     50,
		MaxRegretVsRules:      0.0,
		MinProductionAccuracy: 0.60,
		BlockOnDrift:          true,
		MaxSafetyOverrideRate: 0.05,
	}
}

// ───────────────────────────────────────────────────────────────
// Promotion Result
// ───────────────────────────────────────────────────────────────

// PromotionResult captures what happened during a promotion evaluation.
type PromotionResult struct {
	PolicyID      string
	Version       string
	FromStatus    PolicyStatus
	ToStatus      PolicyStatus
	Promoted      bool
	BlockedReason string
	Gates         []GateResult
	EvaluatedAt   time.Time
}

// ───────────────────────────────────────────────────────────────
// Policy Promoter
// ───────────────────────────────────────────────────────────────

// PolicyPromoter evaluates and promotes policies automatically.
type PolicyPromoter struct {
	rules    PromotionRules
	registry *PolicyLifecycleRegistry
	history  []PromotionResult
}

// NewPolicyPromoter creates a promoter with the given rules and registry.
func NewPolicyPromoter(rules PromotionRules, registry *PolicyLifecycleRegistry) *PolicyPromoter {
	return &PolicyPromoter{
		rules:    rules,
		registry: registry,
	}
}

// EvaluateAll checks all policies and promotes those that pass their gates.
// Returns a list of promotion results (one per policy evaluated).
func (p *PolicyPromoter) EvaluateAll() []PromotionResult {
	var results []PromotionResult

	for i := range p.registry.Versions {
		v := &p.registry.Versions[i]

		switch v.Status {
		case PolicyStatusTraining:
			result := p.evaluateCandidateToShadow(v)
			results = append(results, result)
			p.history = append(p.history, result)

		case PolicyStatusShadow:
			result := p.evaluateShadowToActive(v)
			results = append(results, result)
			p.history = append(p.history, result)
		}
	}

	return results
}

// EvaluateOne evaluates a single policy for promotion.
func (p *PolicyPromoter) EvaluateOne(policyID string, version string) PromotionResult {
	v := p.registry.FindVersion(policyID, version)
	if v == nil {
		return PromotionResult{
			PolicyID: policyID, Version: version,
			Promoted: false, BlockedReason: "version_not_found",
			EvaluatedAt: time.Now(),
		}
	}

	switch v.Status {
	case PolicyStatusTraining:
		return p.evaluateCandidateToShadow(v)
	case PolicyStatusShadow:
		return p.evaluateShadowToActive(v)
	default:
		return PromotionResult{
			PolicyID: policyID, Version: version,
			FromStatus: v.Status, Promoted: false,
			BlockedReason: fmt.Sprintf("no_promotion_path_from_%s", v.Status),
			EvaluatedAt:   time.Now(),
		}
	}
}

func (p *PolicyPromoter) evaluateCandidateToShadow(v *PolicyVersionRecord) PromotionResult {
	result := PromotionResult{
		PolicyID:    v.ID,
		Version:     v.Version,
		FromStatus:  PolicyStatusTraining,
		ToStatus:    PolicyStatusShadow,
		EvaluatedAt: time.Now(),
	}

	// Gate: offline accuracy (outcome-based from validation pipeline).
	accGate := GateResult{
		Gate:      "offline_accuracy",
		Value:     v.OfflineAccuracy,
		Threshold: p.rules.MinOfflineAccuracy,
		Passed:    v.OfflineAccuracy >= p.rules.MinOfflineAccuracy,
	}
	if !accGate.Passed {
		accGate.Reason = fmt.Sprintf("outcome accuracy %.2f below threshold %.2f", v.OfflineAccuracy, p.rules.MinOfflineAccuracy)
	}
	result.Gates = append(result.Gates, accGate)

	// Gate: regret vs rules (learned must be equal or better than rules).
	regretGate := GateResult{
		Gate:      "regret_vs_rules",
		Value:     v.RegretVsRules,
		Threshold: p.rules.MaxRegretVsRules,
		Passed:    v.RegretVsRules <= p.rules.MaxRegretVsRules,
	}
	if !regretGate.Passed {
		regretGate.Reason = fmt.Sprintf("regret %.2f exceeds maximum %.2f", v.RegretVsRules, p.rules.MaxRegretVsRules)
	}
	result.Gates = append(result.Gates, regretGate)

	// Gate: drift.
	if p.rules.BlockOnDrift && v.DriftDetected {
		driftGate := GateResult{Gate: "drift", Passed: false, Reason: "drift_detected"}
		result.Gates = append(result.Gates, driftGate)
	}

	// Evaluate.
	allPassed := true
	for _, g := range result.Gates {
		if !g.Passed {
			allPassed = false
			result.BlockedReason = g.Reason
			break
		}
	}

	if allPassed {
		v.Status = PolicyStatusShadow
		result.Promoted = true
	}

	return result
}

func (p *PolicyPromoter) evaluateShadowToActive(v *PolicyVersionRecord) PromotionResult {
	result := PromotionResult{
		PolicyID:    v.ID,
		Version:     v.Version,
		FromStatus:  PolicyStatusShadow,
		ToStatus:    PolicyStatusActive,
		EvaluatedAt: time.Now(),
	}

	// Gate: shadow runs.
	runsGate := GateResult{
		Gate: "shadow_runs", Value: float64(v.ProductionRuns),
		Threshold: float64(p.rules.MinShadowRuns),
		Passed:    v.ProductionRuns >= p.rules.MinShadowRuns,
	}
	if !runsGate.Passed {
		runsGate.Reason = fmt.Sprintf("runs %d below minimum %d", v.ProductionRuns, p.rules.MinShadowRuns)
	}
	result.Gates = append(result.Gates, runsGate)

	// Gate: shadow accuracy.
	accGate := GateResult{
		Gate: "shadow_accuracy", Value: v.ShadowAccuracy,
		Threshold: p.rules.MinShadowAccuracy,
		Passed:    v.ShadowAccuracy >= p.rules.MinShadowAccuracy,
	}
	if !accGate.Passed {
		accGate.Reason = fmt.Sprintf("accuracy %.2f below threshold %.2f", v.ShadowAccuracy, p.rules.MinShadowAccuracy)
	}
	result.Gates = append(result.Gates, accGate)

	// Gate: regret.
	regretGate := GateResult{
		Gate: "regret_vs_rules", Value: v.RegretVsRules,
		Threshold: p.rules.MaxRegretVsRules,
		Passed:    v.RegretVsRules <= p.rules.MaxRegretVsRules,
	}
	if !regretGate.Passed {
		regretGate.Reason = fmt.Sprintf("regret %.2f exceeds maximum %.2f", v.RegretVsRules, p.rules.MaxRegretVsRules)
	}
	result.Gates = append(result.Gates, regretGate)

	// Gate: drift.
	if p.rules.BlockOnDrift && v.DriftDetected {
		driftGate := GateResult{Gate: "drift", Passed: false, Reason: "drift_detected"}
		result.Gates = append(result.Gates, driftGate)
	}

	// Evaluate all gates.
	allPassed := true
	for _, g := range result.Gates {
		if !g.Passed {
			allPassed = false
			result.BlockedReason = g.Reason
			break
		}
	}

	if allPassed {
		// Retire current active for same domain/decision type.
		current := p.registry.ActiveVersion(v.Domain, v.DecisionType)
		if current != nil {
			now := time.Now()
			current.Status = PolicyStatusRetired
			current.RetiredAt = &now
		}
		now := time.Now()
		v.Status = PolicyStatusActive
		v.PromotedAt = &now
		result.Promoted = true
	}

	return result
}

// Rollback reverts the active policy to a previous version.
func (p *PolicyPromoter) Rollback(domain string, decisionType string, targetVersion string, reason string) (PromotionResult, error) {
	err := p.registry.Rollback(domain, decisionType, targetVersion, reason)
	result := PromotionResult{
		PolicyID:    domain + "/" + decisionType,
		Version:     targetVersion,
		FromStatus:  PolicyStatusActive,
		ToStatus:    PolicyStatusActive,
		Promoted:    err == nil,
		EvaluatedAt: time.Now(),
	}
	if err != nil {
		result.BlockedReason = err.Error()
	}
	p.history = append(p.history, result)
	return result, err
}

// History returns all promotion/rollback events.
func (p *PolicyPromoter) History() []PromotionResult {
	return p.history
}

// Registry returns the underlying registry.
func (p *PolicyPromoter) Registry() *PolicyLifecycleRegistry {
	return p.registry
}
