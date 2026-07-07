// policy_hybrid.go — HybridPolicy implementation.
//
// Uses learned policy when confident, falls back to rules otherwise.
// Guarantees never worse than the rule policy alone.
// This is the recommended production default.
package optimisation

// HybridPolicy combines a LearnedPolicy with a RulePolicy fallback.
// If the learned policy defers (low confidence), the rule policy decides.
type HybridPolicy struct {
	learned  *LearnedPolicy
	fallback *RulePolicy
}

// NewHybridPolicy creates a hybrid policy.
// Both learned and fallback must be non-nil.
func NewHybridPolicy(learned *LearnedPolicy, fallback *RulePolicy) *HybridPolicy {
	return &HybridPolicy{
		learned:  learned,
		fallback: fallback,
	}
}

// Decide tries the learned policy first. Falls back to rules if learned defers.
func (p *HybridPolicy) Decide(ctx PolicyContext) PolicyDecision {
	decision := p.learned.Decide(ctx)

	if decision.Action == "defer" {
		// Learned policy is not confident — use rules.
		fallbackDecision := p.fallback.Decide(ctx)
		fallbackDecision.IsFallback = true
		fallbackDecision.FallbackReason = decision.Reason
		return fallbackDecision
	}

	return decision
}

// Metadata returns identity information for the hybrid policy.
// Reports as "hybrid" type with the learned policy's version.
func (p *HybridPolicy) Metadata() PolicyMetadata {
	learnedMeta := p.learned.Metadata()
	return PolicyMetadata{
		ID:              learnedMeta.ID + "+rules",
		Version:         learnedMeta.Version,
		Type:            "hybrid",
		Domain:          learnedMeta.Domain,
		DecisionType:    learnedMeta.DecisionType,
		TrainedSamples:  learnedMeta.TrainedSamples,
		CreatedAt:       learnedMeta.CreatedAt,
		ValidationScore: learnedMeta.ValidationScore,
	}
}
