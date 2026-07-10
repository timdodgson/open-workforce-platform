package policy

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/assist"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

// PolicySearchHookRunner extends the SI v1 hook runner with policy-based decisions.
type PolicySearchHookRunner struct {
	*assist.SearchHookRunner
	engine *CheckpointEngine
}

// NewPolicySearchHookRunner creates a policy-aware hook runner.
// Returns nil if assist mode is "off".
func NewPolicySearchHookRunner(assistMode string, assistConfig searchdef.SearchAssistConfig, iterations int, policyConfig PolicySearchConfig) *PolicySearchHookRunner {
	base := assist.NewSearchHookRunner(assistMode, assistConfig, iterations)
	if base == nil {
		return nil
	}
	return &PolicySearchHookRunner{
		SearchHookRunner: base,
		engine:           NewCheckpointEngine(policyConfig),
	}
}

// RunCheckpoint evaluates search state using the configured policy engine.
func (r *PolicySearchHookRunner) RunCheckpoint(algorithm string, candidates int, currentPenalty int, bestPenalty int, initialPenalty int, temperature float64) searchdef.SearchAction {
	if r == nil {
		return searchdef.SearchContinue
	}

	baseAction := r.SearchHookRunner.RunCheckpoint(algorithm, candidates, currentPenalty, bestPenalty, initialPenalty, temperature)
	if r.engine == nil {
		return baseAction
	}

	snap := r.SearchHookRunner.Snapshot()
	final := r.engine.Evaluate(CheckpointInput{
		Algorithm:      algorithm,
		Candidates:     candidates,
		CurrentPenalty: currentPenalty,
		BestPenalty:    bestPenalty,
		InitialPenalty: initialPenalty,
		Temperature:    temperature,
		BaseAction:     string(baseAction),
	}, SearchHookSnapshot{
		ShadowMode:        snap.ShadowMode,
		CheckpointNum:     snap.CheckpointNum,
		IterationsTotal:   snap.IterationsTotal,
		LastImproveAt:     snap.LastImproveAt,
		MinBudgetFraction: snap.MinBudgetFraction,
	})

	return searchdef.SearchAction(final)
}

// PolicyDecisions returns all policy decisions made during this search.
func (r *PolicySearchHookRunner) PolicyDecisions() []PolicySearchDecision {
	if r == nil || r.engine == nil {
		return nil
	}
	return r.engine.Decisions()
}
