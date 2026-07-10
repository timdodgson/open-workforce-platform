// policy_executor.go wires PolicySearchHookRunner to the search assist hook loop.
package optimisation

import (
	pol "github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/policy"
)

// PolicySearchHookRunner extends SearchHookRunner with policy-based decisions.
type PolicySearchHookRunner struct {
	*SearchHookRunner
	engine *pol.CheckpointEngine
}

// NewPolicySearchHookRunner creates a policy-aware hook runner.
// Returns nil if assist mode is "off".
func NewPolicySearchHookRunner(assistMode string, assistConfig SearchAssistConfig, iterations int, policyConfig PolicySearchConfig) *PolicySearchHookRunner {
	base := NewSearchHookRunner(assistMode, assistConfig, iterations)
	if base == nil {
		return nil
	}
	return &PolicySearchHookRunner{
		SearchHookRunner: base,
		engine:           pol.NewCheckpointEngine(policyConfig),
	}
}

// RunCheckpoint implements searchHookRunner.
func (r *PolicySearchHookRunner) RunCheckpoint(algorithm string, candidates int, currentPenalty int, bestPenalty int, initialPenalty int, temperature float64) SearchAction {
	return r.RunPolicyCheckpoint(algorithm, candidates, currentPenalty, bestPenalty, initialPenalty, temperature)
}

// PolicyDecisions implements searchHookRunner.
func (r *PolicySearchHookRunner) PolicyDecisions() []PolicySearchDecision {
	return r.Decisions()
}

// RunPolicyCheckpoint evaluates search state using the configured policy engine.
func (r *PolicySearchHookRunner) RunPolicyCheckpoint(algorithm string, candidates int, currentPenalty int, bestPenalty int, initialPenalty int, temperature float64) SearchAction {
	if r == nil {
		return SearchContinue
	}

	baseAction := r.SearchHookRunner.RunCheckpoint(algorithm, candidates, currentPenalty, bestPenalty, initialPenalty, temperature)
	if r.engine == nil {
		return baseAction
	}

	final := r.engine.Evaluate(pol.CheckpointInput{
		Algorithm:      algorithm,
		Candidates:     candidates,
		CurrentPenalty: currentPenalty,
		BestPenalty:    bestPenalty,
		InitialPenalty: initialPenalty,
		Temperature:    temperature,
		BaseAction:     string(baseAction),
	}, pol.SearchHookSnapshot{
		ShadowMode:        r.mode == "shadow",
		CheckpointNum:     r.checkpointNum,
		IterationsTotal:   r.iterationsTotal,
		LastImproveAt:     r.lastImproveAt,
		MinBudgetFraction: float64(r.assist.config.MinBudgetFraction),
	})

	return SearchAction(final)
}

// Decisions returns all policy decisions made during this search.
func (r *PolicySearchHookRunner) Decisions() []PolicySearchDecision {
	if r == nil || r.engine == nil {
		return nil
	}
	return r.engine.Decisions()
}
