package siadapter_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/siadapter"
)

func TestAdaptWorkerDecisionsToSearchAssistUsesAllocatedIters(t *testing.T) {
	records := []inrc2.WorkerDecisionRecord{{
		Algorithm:        "sa",
		ParentObjective:  4000,
		GlobalBest:       3500,
		DistanceFromBest: 500,
		AllocatedIters:   100000,
		Recommendation:   inrc2.RecRun,
		Confidence:       0.8,
		FinalObjective:   3600,
	}}
	out := siadapter.AdaptWorkerDecisionsToSearchAssist(records, 200000)
	if len(out) != 1 {
		t.Fatalf("got %d rows", len(out))
	}
	if out[0].Candidates != 100000 || out[0].IterationsTotal != 100000 {
		t.Fatalf("budget = %d/%d, want 100000", out[0].Candidates, out[0].IterationsTotal)
	}
	if out[0].PlateauLength != 500 {
		t.Fatalf("plateau = %d, want 500", out[0].PlateauLength)
	}
}

func TestAdaptWorkerAssistToPolicyDecisions(t *testing.T) {
	records := []inrc2.AssistRecord{{
		Algorithm:      "sa",
		Recommendation: inrc2.RecSkip,
		FinalAction:    inrc2.RecSkip,
		Outcome:        inrc2.AssistAccepted,
		FinalBudget:    120000,
		Confidence:     0.9,
	}}
	out := siadapter.AdaptWorkerAssistToPolicyDecisions(records, "hybrid")
	if len(out) != 1 {
		t.Fatalf("got %d decisions", len(out))
	}
	if out[0].Action != "early_stop" {
		t.Fatalf("action = %q", out[0].Action)
	}
	if out[0].Candidates != 120000 {
		t.Fatalf("candidates = %d", out[0].Candidates)
	}
	if out[0].PolicyUsed != "hybrid_learned" {
		t.Fatalf("policy_used = %q", out[0].PolicyUsed)
	}
}

func TestInferNRPPolicyPenalties(t *testing.T) {
	initial, best := siadapter.InferNRPPolicyPenalties(
		[]inrc2.WorkerDecisionRecord{{ParentObjective: 4200, FinalObjective: 3100}},
		nil,
		3000,
	)
	if initial != 4200 || best != 3000 {
		t.Fatalf("initial=%d best=%d", initial, best)
	}
}
