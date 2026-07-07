// policy.go — Search Intelligence 2.0 Policy Architecture.
//
// Structured policy framework for SI 2.0. All policy types implement Policy:
//   - RulePolicy: deterministic rules (preserves v1 heuristic behaviour)
//   - LearnedPolicy: model-based decisions with confidence thresholds
//   - HybridPolicy: learned when confident, RulePolicy as fallback
//
// SI v1 production paths (SearchHookRunner, RuleBasedPortfolioAdvisor,
// inrc2.WorkerDecisionEngine) still run in parallel until PolicySearchHookRunner
// and HybridExecutor are wired into solver loops.
package optimisation

import "time"

// ───────────────────────────────────────────────────────────────
// Core Interfaces
// ───────────────────────────────────────────────────────────────

// Policy is the universal interface for all SI decision makers.
type Policy interface {
	// Decide produces a PolicyDecision for the given context.
	Decide(ctx PolicyContext) PolicyDecision

	// Metadata returns identity and version information.
	Metadata() PolicyMetadata
}

// ───────────────────────────────────────────────────────────────
// PolicyContext
// ───────────────────────────────────────────────────────────────

// PolicyContext is the input to every policy decision.
// It wraps a FeatureVector with decision-type routing information.
type PolicyContext struct {
	// What kind of decision is being made.
	DecisionType string // "worker", "search", "portfolio"

	// The extracted features for this decision point.
	Features FeatureVector

	// Domain and instance (convenience, also in Features).
	Domain   string
	Instance string

	// How many historical samples exist for this domain+decision.
	HistoricalSamples int
}

// ───────────────────────────────────────────────────────────────
// PolicyDecision
// ───────────────────────────────────────────────────────────────

// PolicyDecision is the output of every policy evaluation.
// It records everything needed for analysis and retraining.
type PolicyDecision struct {
	// The recommended action (e.g. "run", "skip", "early_stop", "allocate").
	Action string

	// Confidence in this recommendation (0.0–1.0).
	Confidence float64

	// Why this decision was made (human-readable).
	Reason string

	// Which policy produced this decision.
	PolicyID string

	// Which version of the policy was used.
	PolicyVersion string

	// Whether this decision is a fallback from a higher-priority policy.
	IsFallback bool

	// Why the primary policy deferred (empty if not a fallback).
	FallbackReason string

	// Action-specific parameters.
	Parameters map[string]any
}

// ───────────────────────────────────────────────────────────────
// PolicyMetadata
// ───────────────────────────────────────────────────────────────

// PolicyMetadata identifies a policy and its training provenance.
type PolicyMetadata struct {
	// Unique identifier for this policy.
	ID string

	// Semantic version (e.g. "1.0.0", "2.3.1").
	Version string

	// Policy type: "rule", "learned", "hybrid".
	Type string

	// Which domain this policy is trained for. "*" means universal.
	Domain string

	// Which decision type this policy handles.
	DecisionType string

	// Number of training samples (0 for rule policies).
	TrainedSamples int

	// When this policy version was created.
	CreatedAt time.Time

	// Offline validation accuracy (0.0–1.0, -1 if not evaluated).
	ValidationScore float64
}
