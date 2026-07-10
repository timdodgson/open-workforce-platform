package siadapter

import (
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// NRPPolicyEmitInput configures NRP policy CSV emission after a PFRS run.
type NRPPolicyEmitInput struct {
	PolicyMode       string
	Instance         string
	WorkerMode       string
	BestPenalty      int
	DecisionRecorder *inrc2.ShadowRecorder
	AssistRecorder   *inrc2.AssistRecorder
}

// AdaptWorkerDecisionsToPolicyDecisions maps NRP worker shadow rows to policy_decisions.csv.
func AdaptWorkerDecisionsToPolicyDecisions(records []inrc2.WorkerDecisionRecord, policyMode string) []optimisation.PolicySearchDecision {
	out := make([]optimisation.PolicySearchDecision, 0, len(records))
	for i, r := range records {
		out = append(out, optimisation.PolicySearchDecision{
			Checkpoint:     i,
			Candidates:     workerBudget(r.AllocatedIters, r.SuggestedBudget, 0),
			PolicyMode:     policyMode,
			PolicyUsed:     "rule",
			Action:         workerRecToPolicyAction(r.Recommendation),
			Confidence:     r.Confidence,
			FallbackReason: "nrp_worker_shadow",
			SafetyOverride: false,
		})
	}
	return out
}

// AdaptWorkerAssistToPolicyDecisions maps NRP worker assist rows to policy_decisions.csv.
func AdaptWorkerAssistToPolicyDecisions(records []inrc2.AssistRecord, policyMode string) []optimisation.PolicySearchDecision {
	out := make([]optimisation.PolicySearchDecision, 0, len(records))
	for i, r := range records {
		action := workerRecToPolicyAction(r.FinalAction)
		if action == "continue" {
			action = workerRecToPolicyAction(r.Recommendation)
		}
		fallback := ""
		if r.SafetyTriggered {
			fallback = r.SafetyRule
		}
		out = append(out, optimisation.PolicySearchDecision{
			Checkpoint:     i,
			Candidates:     assistBudget(r, 0),
			PolicyMode:     policyMode,
			PolicyUsed:     assistOutcomeToPolicyUsed(r.Outcome, policyMode),
			Action:         action,
			Confidence:     r.Confidence,
			FallbackReason: fallback,
			SafetyOverride: r.SafetyTriggered,
		})
	}
	return out
}

// MergePolicyDecisions appends b into a without duplicate checkpoints.
func MergePolicyDecisions(a, b []optimisation.PolicySearchDecision) []optimisation.PolicySearchDecision {
	if len(b) == 0 {
		return a
	}
	seen := make(map[int]struct{}, len(a))
	for _, d := range a {
		seen[d.Checkpoint] = struct{}{}
	}
	for _, d := range b {
		if _, ok := seen[d.Checkpoint]; ok {
			continue
		}
		a = append(a, d)
		seen[d.Checkpoint] = struct{}{}
	}
	return a
}

// InferNRPPolicyPenalties estimates initial/best objectives for policy evaluation rows.
func InferNRPPolicyPenalties(
	decisions []inrc2.WorkerDecisionRecord,
	assists []inrc2.AssistRecord,
	bestPenalty int,
) (initial int, best int) {
	best = bestPenalty
	initial = bestPenalty
	for _, r := range decisions {
		if r.ParentObjective > initial {
			initial = r.ParentObjective
		}
		if r.FinalObjective > 0 && r.FinalObjective < best {
			best = r.FinalObjective
		}
	}
	for _, r := range assists {
		if r.ParentObjective > initial {
			initial = r.ParentObjective
		}
		if r.FinalObjective > 0 && r.FinalObjective < best {
			best = r.FinalObjective
		}
	}
	if best <= 0 {
		best = bestPenalty
	}
	if initial < best {
		initial = best
	}
	return initial, best
}

// EmitNRPPolicyCSVs writes policy_decisions/evaluation/counterfactual from worker telemetry.
func EmitNRPPolicyCSVs(outputDir string, in NRPPolicyEmitInput, written map[string]bool) {
	if in.PolicyMode == "" {
		return
	}

	var decisions []optimisation.PolicySearchDecision
	if in.DecisionRecorder != nil {
		decisions = MergePolicyDecisions(
			decisions,
			AdaptWorkerDecisionsToPolicyDecisions(in.DecisionRecorder.Records(), in.PolicyMode),
		)
	}
	if in.AssistRecorder != nil {
		decisions = MergePolicyDecisions(
			decisions,
			AdaptWorkerAssistToPolicyDecisions(in.AssistRecorder.Records(), in.PolicyMode),
		)
	}
	if len(decisions) == 0 {
		return
	}

	optimisation.WritePolicyDecisionsCSV(filepath.Join(outputDir, "policy_decisions.csv"), decisions)
	written["policy_decisions.csv"] = true

	initial, best := InferNRPPolicyPenalties(
		recordsOrNil(in.DecisionRecorder),
		assistRecordsOrNil(in.AssistRecorder),
		in.BestPenalty,
	)
	evalInput := optimisation.PolicyEvaluationInput{
		RunID:          filepath.Base(outputDir),
		Domain:         "nrp",
		Instance:       in.Instance,
		Algorithm:      in.WorkerMode,
		InitialPenalty: initial,
		BestPenalty:    best,
		Decisions:      decisions,
	}
	_ = optimisation.WritePolicyEvaluationCSV(outputDir, evalInput)
	written["policy_evaluation.csv"] = true

	if err := optimisation.WriteCounterfactualFromPolicyDecisions(outputDir, optimisation.CounterfactualEmitInput{
		RunID:          evalInput.RunID,
		Domain:         evalInput.Domain,
		Instance:       evalInput.Instance,
		Algorithm:      evalInput.Algorithm,
		InitialPenalty: evalInput.InitialPenalty,
		BestPenalty:    evalInput.BestPenalty,
		Decisions:      evalInput.Decisions,
	}); err == nil {
		written["counterfactual_learning.csv"] = true
	}
}

func workerRecToPolicyAction(rec inrc2.Recommendation) string {
	switch rec {
	case inrc2.RecSkip:
		return "early_stop"
	case inrc2.RecReduceBudget:
		return "reduce_budget"
	case inrc2.RecIncreaseBudget:
		return "extend_budget"
	case inrc2.RecChangeAlgo:
		return "restart"
	default:
		return "continue"
	}
}

func assistOutcomeToPolicyUsed(outcome inrc2.AssistOutcome, policyMode string) string {
	if outcome == inrc2.AssistRejected {
		return "rule"
	}
	if policyMode == "learned" {
		return "learned"
	}
	if policyMode == "hybrid" {
		return "hybrid_learned"
	}
	return "rule"
}

func recordsOrNil(r *inrc2.ShadowRecorder) []inrc2.WorkerDecisionRecord {
	if r == nil {
		return nil
	}
	return r.Records()
}

func assistRecordsOrNil(r *inrc2.AssistRecorder) []inrc2.AssistRecord {
	if r == nil {
		return nil
	}
	return r.Records()
}
