// policy_rule.go — RulePolicy implementation.
//
// Wraps deterministic heuristic rules in the Policy interface.
// This preserves all v1 behaviour while conforming to the policy architecture.
// Rule policies always have confidence 1.0 (deterministic) and never fallback.
package policy

import "time"

// Rule is a single condition → action mapping.
type Rule struct {
	// Name identifies this rule for logging.
	Name string

	// Matches returns true if this rule applies to the given context.
	Matches func(ctx PolicyContext) bool

	// Decide returns the action for this rule.
	Decide func(ctx PolicyContext) PolicyDecision
}

// RulePolicy evaluates rules in order, returning the first match.
// If no rule matches, returns a default "continue" action.
type RulePolicy struct {
	id      string
	version string
	domain  string
	decType string
	rules   []Rule
}

// NewRulePolicy creates a rule-based policy.
func NewRulePolicy(id string, version string, domain string, decisionType string, rules []Rule) *RulePolicy {
	return &RulePolicy{
		id:      id,
		version: version,
		domain:  domain,
		decType: decisionType,
		rules:   rules,
	}
}

// Decide evaluates rules in order and returns the first matching decision.
func (p *RulePolicy) Decide(ctx PolicyContext) PolicyDecision {
	for _, rule := range p.rules {
		if rule.Matches(ctx) {
			decision := rule.Decide(ctx)
			decision.PolicyID = p.id
			decision.PolicyVersion = p.version
			decision.IsFallback = false
			decision.FallbackReason = ""
			if decision.Confidence == 0 {
				decision.Confidence = 1.0 // rules are deterministic
			}
			if decision.Reason == "" {
				decision.Reason = "rule:" + rule.Name
			}
			return decision
		}
	}

	// No rule matched — default to continue.
	return PolicyDecision{
		Action:        "continue",
		Confidence:    0.5,
		Reason:        "no_rule_matched",
		PolicyID:      p.id,
		PolicyVersion: p.version,
	}
}

// Metadata returns identity information for this rule policy.
func (p *RulePolicy) Metadata() PolicyMetadata {
	return PolicyMetadata{
		ID:              p.id,
		Version:         p.version,
		Type:            "rule",
		Domain:          p.domain,
		DecisionType:    p.decType,
		TrainedSamples:  0,
		CreatedAt:       time.Time{},
		ValidationScore: -1,
	}
}
