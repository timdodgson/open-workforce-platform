// policy_provider.go — PolicyProvider routes decisions to the correct policy.
//
// The provider maintains a registry of policies keyed by domain and decision type.
// If no domain-specific policy exists, falls back to the universal ("*") policy.
// If no policy exists at all, returns a safe default (continue with low confidence).
package optimisation

// PolicyProvider is the single entry point for obtaining policies.
// Solvers call GetPolicy(domain, decisionType) at each decision point.
type PolicyProvider struct {
	// policies maps domain → decisionType → Policy.
	// The special domain "*" holds universal fallback policies.
	policies map[string]map[string]Policy
}

// NewPolicyProvider creates an empty provider.
func NewPolicyProvider() *PolicyProvider {
	return &PolicyProvider{
		policies: make(map[string]map[string]Policy),
	}
}

// Register adds a policy for a specific domain and decision type.
// Use domain="*" for universal fallback policies.
func (pp *PolicyProvider) Register(domain string, decisionType string, policy Policy) {
	if pp.policies[domain] == nil {
		pp.policies[domain] = make(map[string]Policy)
	}
	pp.policies[domain][decisionType] = policy
}

// GetPolicy returns the best available policy for the given domain and decision type.
// Resolution order:
//  1. Exact domain + decision type match
//  2. Universal ("*") + decision type match
//  3. Safe default (no-op policy)
func (pp *PolicyProvider) GetPolicy(domain string, decisionType string) Policy {
	// Try exact domain match.
	if domainPolicies, ok := pp.policies[domain]; ok {
		if policy, ok := domainPolicies[decisionType]; ok {
			return policy
		}
	}

	// Try universal fallback.
	if universalPolicies, ok := pp.policies["*"]; ok {
		if policy, ok := universalPolicies[decisionType]; ok {
			return policy
		}
	}

	// No policy found — return safe default.
	return &defaultPolicy{}
}

// ListPolicies returns metadata for all registered policies.
func (pp *PolicyProvider) ListPolicies() []PolicyMetadata {
	var result []PolicyMetadata
	for _, domainPolicies := range pp.policies {
		for _, policy := range domainPolicies {
			result = append(result, policy.Metadata())
		}
	}
	return result
}

// defaultPolicy is a no-op policy returned when no policy is registered.
type defaultPolicy struct{}

func (p *defaultPolicy) Decide(_ PolicyContext) PolicyDecision {
	return PolicyDecision{
		Action:        "continue",
		Confidence:    0.0,
		Reason:        "no_policy_registered",
		PolicyID:      "default",
		PolicyVersion: "0.0.0",
	}
}

func (p *defaultPolicy) Metadata() PolicyMetadata {
	return PolicyMetadata{
		ID:              "default",
		Version:         "0.0.0",
		Type:            "none",
		Domain:          "*",
		DecisionType:    "*",
		ValidationScore: -1,
	}
}
