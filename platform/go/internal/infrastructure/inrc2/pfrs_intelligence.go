package inrc2

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// WorkerIntelligenceWire holds PFRS beam-search worker decision wiring.
type WorkerIntelligenceWire struct {
	Engine           WorkerDecisionEngine
	DecisionRecorder *ShadowRecorder
	AssistRecorder   *AssistRecorder
	AssistMode       bool
}

// WireWorkerIntelligence configures PFRS worker-level SI from CLI mode flags.
// Returns optional log lines for the caller to print.
func WireWorkerIntelligence(mode, policyMode, policyDir string) (WorkerIntelligenceWire, []string) {
	var out WorkerIntelligenceWire
	var lines []string

	if mode == "shadow" || mode == "assist" || mode == "adaptive" {
		resolvedDir := optimisation.ResolvePolicyDir(policyMode, policyDir)
		if policyMode != "" && policyMode != "rules" {
			out.Engine = NewHybridWorkerDecisionEngine(policyMode, resolvedDir)
			lines = append(lines, "  Worker Policy: "+policyMode+" (dir: "+resolvedDir+")")
		} else {
			out.Engine = NewRuleBasedEngine()
		}
		out.DecisionRecorder = NewShadowRecorder()
		if mode == "shadow" {
			lines = append(lines, "  Decision Mode: shadow (recording predictions, no behaviour change)")
		}
	}
	if mode == "assist" || mode == "adaptive" {
		out.AssistMode = true
		out.AssistRecorder = NewAssistRecorder()
		if mode == "adaptive" {
			lines = append(lines, "  Decision Mode: adaptive (live-updating decisions, safety overrides active)")
		} else {
			lines = append(lines, "  Decision Mode: assist (AI advises optimiser, safety overrides active)")
		}
	}
	return out, lines
}
