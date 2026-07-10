package optimisation

// search_hooks_bridge.go wires SI v1/v2 hook runners into search.go without import cycles.

type searchHookRunner interface {
	OnImprovement(candidates int)
	OnRestart(candidates int)
	ShouldCheckpoint(candidates int) bool
	RunCheckpoint(algorithm string, candidates int, currentPenalty int, bestPenalty int, initialPenalty int, temperature float64) SearchAction
	GetIterationsTotal() int
	Finalise(bestPenalty int, totalCandidates int) *SearchAssistRecorder
	PolicyDecisions() []PolicySearchDecision
}

// siV1HookRunner wraps the assist hook runner with a nil PolicyDecisions implementation.
type siV1HookRunner struct {
	*SearchHookRunner
}

func (siV1HookRunner) PolicyDecisions() []PolicySearchDecision { return nil }

// newSearchHooks creates SI v1 or SI 2.0 hook runners from SearchConfig.
// When PolicyMode is set without AssistMode, mirrors CLI behaviour (shadow assist).
func newSearchHooks(config SearchConfig) searchHookRunner {
	assistMode := config.AssistMode
	if assistMode == "" && config.PolicyMode != "" {
		assistMode = "shadow"
	}
	assistConfig := config.AssistConfig
	if assistMode != "" && assistMode != "off" && assistConfig == (SearchAssistConfig{}) {
		assistConfig = DefaultSearchAssistConfig()
	}

	if config.PolicyMode != "" {
		if p := NewPolicySearchHookRunner(assistMode, assistConfig, config.Iterations, PolicySearchConfig{
			PolicyMode: config.PolicyMode,
			PolicyDir:  config.PolicyDir,
			Domain:     config.PolicyDomain,
			Instance:   config.PolicyInstance,
		}); p != nil {
			return p
		}
	}
	return siV1HookRunner{NewSearchHookRunner(assistMode, assistConfig, config.Iterations)}
}

// finalizeSearchHooks collects assist and policy telemetry at end of a search run.
func finalizeSearchHooks(hooks searchHookRunner, bestPenalty, totalCandidates int) ([]SearchAssistRecord, []PolicySearchDecision) {
	if hooks == nil {
		return nil, nil
	}
	var assistRecords []SearchAssistRecord
	if recorder := hooks.Finalise(bestPenalty, totalCandidates); recorder != nil {
		assistRecords = recorder.Records()
	}
	return assistRecords, hooks.PolicyDecisions()
}
