// policy_hierarchy.go — Hierarchical Policy Resolution.
//
// Policies are resolved through a three-level hierarchy:
//  1. Instance policy (most specific, e.g. CVRP A-n32-k5)
//  2. Domain policy (e.g. CVRP)
//  3. Global policy (universal fallback)
//
// Resolution: use the most specific policy that exists AND has sufficient confidence.
// If a specific policy defers (low confidence), fall up to the next level.
//
// Transfer learning:
//   - Supported via TransferPolicy which adapts a source domain policy to a target domain.
//   - Never assumed automatic — transfer requires explicit confidence check.
//   - Transfer confidence is penalised relative to native training.
//
// Each domain (NRP, CVRP, JSS, VRPTW) can have its own policies.
// Instance-level overrides allow fine-tuning for specific problem instances.
package policy

import (
	"fmt"
)

// ───────────────────────────────────────────────────────────────
// Policy Hierarchy
// ───────────────────────────────────────────────────────────────

// PolicyHierarchy resolves policies through instance → domain → global levels.
type PolicyHierarchy struct {
	// Instance-level policies: domain/instance/decisionType → Policy
	instancePolicies map[string]Policy

	// Domain-level policies: domain/decisionType → Policy
	domainPolicies map[string]Policy

	// Global policies: decisionType → Policy
	globalPolicies map[string]Policy
}

// NewPolicyHierarchy creates an empty hierarchy.
func NewPolicyHierarchy() *PolicyHierarchy {
	return &PolicyHierarchy{
		instancePolicies: make(map[string]Policy),
		domainPolicies:   make(map[string]Policy),
		globalPolicies:   make(map[string]Policy),
	}
}

// RegisterInstance registers a policy for a specific domain/instance/decision type.
func (h *PolicyHierarchy) RegisterInstance(domain string, instance string, decisionType string, policy Policy) {
	key := fmt.Sprintf("%s/%s/%s", domain, instance, decisionType)
	h.instancePolicies[key] = policy
}

// RegisterDomain registers a policy for a domain/decision type.
func (h *PolicyHierarchy) RegisterDomain(domain string, decisionType string, policy Policy) {
	key := fmt.Sprintf("%s/%s", domain, decisionType)
	h.domainPolicies[key] = policy
}

// RegisterGlobal registers a global fallback policy for a decision type.
func (h *PolicyHierarchy) RegisterGlobal(decisionType string, policy Policy) {
	h.globalPolicies[decisionType] = policy
}

// Resolve returns the best policy for the given context, traversing the hierarchy.
// Returns: the policy and which level it came from.
func (h *PolicyHierarchy) Resolve(domain string, instance string, decisionType string) (Policy, PolicyLevel) {
	// Level 1: Instance-specific.
	instanceKey := fmt.Sprintf("%s/%s/%s", domain, instance, decisionType)
	if p, ok := h.instancePolicies[instanceKey]; ok {
		return p, LevelInstance
	}

	// Level 2: Domain-wide.
	domainKey := fmt.Sprintf("%s/%s", domain, decisionType)
	if p, ok := h.domainPolicies[domainKey]; ok {
		return p, LevelDomain
	}

	// Level 3: Global.
	if p, ok := h.globalPolicies[decisionType]; ok {
		return p, LevelGlobal
	}

	// No policy at any level.
	return &defaultPolicy{}, LevelNone
}

// PolicyLevel indicates where in the hierarchy a policy was resolved from.
type PolicyLevel string

const (
	LevelInstance PolicyLevel = "instance"
	LevelDomain   PolicyLevel = "domain"
	LevelGlobal   PolicyLevel = "global"
	LevelNone     PolicyLevel = "none"
)

// ───────────────────────────────────────────────────────────────
// Hierarchical Decision
// ───────────────────────────────────────────────────────────────

// HierarchicalDecision extends PolicyDecision with resolution metadata.
type HierarchicalDecision struct {
	PolicyDecision

	// Which level provided the decision.
	ResolvedLevel PolicyLevel

	// Whether higher-level policies were consulted (cascade happened).
	Cascaded bool

	// Levels attempted before finding a confident answer.
	LevelsAttempted []PolicyLevel
}

// DecideWithHierarchy resolves and executes the best policy, cascading up
// if a specific policy defers.
func (h *PolicyHierarchy) DecideWithHierarchy(ctx PolicyContext) HierarchicalDecision {
	levels := []struct {
		level  PolicyLevel
		policy Policy
	}{
		{LevelInstance, h.getInstancePolicy(ctx)},
		{LevelDomain, h.getDomainPolicy(ctx)},
		{LevelGlobal, h.getGlobalPolicy(ctx)},
	}

	attempted := []PolicyLevel{}

	for _, l := range levels {
		if l.policy == nil {
			continue
		}
		attempted = append(attempted, l.level)

		decision := l.policy.Decide(ctx)

		// If policy defers, cascade to next level.
		if decision.Action == "defer" {
			continue
		}

		return HierarchicalDecision{
			PolicyDecision:  decision,
			ResolvedLevel:   l.level,
			Cascaded:        len(attempted) > 1,
			LevelsAttempted: attempted,
		}
	}

	// All levels deferred or missing — return safe default.
	return HierarchicalDecision{
		PolicyDecision: PolicyDecision{
			Action:        "continue",
			Confidence:    0.0,
			Reason:        "all_levels_deferred",
			PolicyID:      "hierarchy-default",
			PolicyVersion: "0.0.0",
		},
		ResolvedLevel:   LevelNone,
		Cascaded:        true,
		LevelsAttempted: attempted,
	}
}

func (h *PolicyHierarchy) getInstancePolicy(ctx PolicyContext) Policy {
	key := fmt.Sprintf("%s/%s/%s", ctx.Domain, ctx.Instance, ctx.DecisionType)
	return h.instancePolicies[key]
}

func (h *PolicyHierarchy) getDomainPolicy(ctx PolicyContext) Policy {
	key := fmt.Sprintf("%s/%s", ctx.Domain, ctx.DecisionType)
	return h.domainPolicies[key]
}

func (h *PolicyHierarchy) getGlobalPolicy(ctx PolicyContext) Policy {
	return h.globalPolicies[ctx.DecisionType]
}

// ───────────────────────────────────────────────────────────────
// Transfer Policy
// ───────────────────────────────────────────────────────────────

// TransferPolicy adapts a policy trained on one domain to another domain.
// Transfer confidence is penalised by a configurable factor.
// Never assumes behaviour transfers automatically.
type TransferPolicy struct {
	source            Policy
	sourceDomain      string
	targetDomain      string
	confidencePenalty float64 // multiplier on confidence (e.g. 0.7 = 30% penalty)
}

// TransferPolicyConfig configures a transfer policy.
type TransferPolicyConfig struct {
	Source            Policy
	SourceDomain      string
	TargetDomain      string
	ConfidencePenalty float64 // 0.0–1.0, lower = more penalty. Default: 0.70.
}

// NewTransferPolicy creates a policy that adapts source domain knowledge to a target.
func NewTransferPolicy(cfg TransferPolicyConfig) *TransferPolicy {
	penalty := cfg.ConfidencePenalty
	if penalty <= 0 || penalty > 1.0 {
		penalty = 0.70
	}
	return &TransferPolicy{
		source:            cfg.Source,
		sourceDomain:      cfg.SourceDomain,
		targetDomain:      cfg.TargetDomain,
		confidencePenalty: penalty,
	}
}

// Decide delegates to the source policy but penalises confidence.
func (p *TransferPolicy) Decide(ctx PolicyContext) PolicyDecision {
	decision := p.source.Decide(ctx)

	// Penalise confidence for cross-domain transfer.
	decision.Confidence *= p.confidencePenalty
	decision.Reason = fmt.Sprintf("transfer(%s→%s): %s", p.sourceDomain, p.targetDomain, decision.Reason)
	decision.PolicyID = fmt.Sprintf("transfer:%s→%s:%s", p.sourceDomain, p.targetDomain, decision.PolicyID)

	return decision
}

// Metadata returns metadata identifying this as a transfer policy.
func (p *TransferPolicy) Metadata() PolicyMetadata {
	sourceMeta := p.source.Metadata()
	return PolicyMetadata{
		ID:              fmt.Sprintf("transfer:%s→%s:%s", p.sourceDomain, p.targetDomain, sourceMeta.ID),
		Version:         sourceMeta.Version,
		Type:            "transfer",
		Domain:          p.targetDomain,
		DecisionType:    sourceMeta.DecisionType,
		TrainedSamples:  sourceMeta.TrainedSamples,
		CreatedAt:       sourceMeta.CreatedAt,
		ValidationScore: sourceMeta.ValidationScore * p.confidencePenalty,
	}
}

// ───────────────────────────────────────────────────────────────
// Hierarchy Summary (for dashboard)
// ───────────────────────────────────────────────────────────────

// HierarchyEntry describes one policy in the hierarchy.
type HierarchyEntry struct {
	Level        PolicyLevel
	Domain       string
	Instance     string
	DecisionType string
	PolicyID     string
	Version      string
	Type         string // rule, learned, hybrid, transfer
}

// Summary returns a flat list of all registered policies with their hierarchy level.
func (h *PolicyHierarchy) Summary() []HierarchyEntry {
	var entries []HierarchyEntry

	for key, p := range h.globalPolicies {
		m := p.Metadata()
		entries = append(entries, HierarchyEntry{
			Level: LevelGlobal, Domain: "*", Instance: "*",
			DecisionType: key, PolicyID: m.ID, Version: m.Version, Type: m.Type,
		})
	}

	for key, p := range h.domainPolicies {
		m := p.Metadata()
		// Parse domain/decisionType from key.
		domain, decType := parseHierarchyKey(key)
		entries = append(entries, HierarchyEntry{
			Level: LevelDomain, Domain: domain, Instance: "*",
			DecisionType: decType, PolicyID: m.ID, Version: m.Version, Type: m.Type,
		})
	}

	for key, p := range h.instancePolicies {
		m := p.Metadata()
		domain, instance, decType := parseInstanceKey(key)
		entries = append(entries, HierarchyEntry{
			Level: LevelInstance, Domain: domain, Instance: instance,
			DecisionType: decType, PolicyID: m.ID, Version: m.Version, Type: m.Type,
		})
	}

	return entries
}

func parseHierarchyKey(key string) (string, string) {
	for i, c := range key {
		if c == '/' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}

func parseInstanceKey(key string) (string, string, string) {
	parts := splitKey(key)
	if len(parts) >= 3 {
		return parts[0], parts[1], parts[2]
	}
	if len(parts) == 2 {
		return parts[0], parts[1], ""
	}
	return key, "", ""
}

func splitKey(key string) []string {
	var parts []string
	start := 0
	for i, c := range key {
		if c == '/' {
			parts = append(parts, key[start:i])
			start = i + 1
		}
	}
	parts = append(parts, key[start:])
	return parts
}
