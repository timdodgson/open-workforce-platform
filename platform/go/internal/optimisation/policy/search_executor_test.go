package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckpointEngine_RulesMode(t *testing.T) {
	e := NewCheckpointEngine(PolicySearchConfig{PolicyMode: "rules"})
	snap := SearchHookSnapshot{MinBudgetFraction: 0.2}

	action := e.Evaluate(CheckpointInput{
		Algorithm: "sa", Candidates: 50000,
		CurrentPenalty: 800, BestPenalty: 784, InitialPenalty: 1000,
		Temperature: 0.5, BaseAction: CheckpointContinue,
	}, snap)

	if action != CheckpointContinue {
		t.Errorf("Action = %q, want continue", action)
	}
	decisions := e.Decisions()
	if len(decisions) == 0 {
		t.Fatal("should record decisions")
	}
	if decisions[0].PolicyUsed != "rule" {
		t.Errorf("PolicyUsed = %q, want rule", decisions[0].PolicyUsed)
	}
}

func TestCheckpointEngine_LearnedMode_NoModel(t *testing.T) {
	e := NewCheckpointEngine(PolicySearchConfig{PolicyMode: "learned", PolicyDir: "/nonexistent"})
	snap := SearchHookSnapshot{}

	e.Evaluate(CheckpointInput{
		Algorithm: "sa", Candidates: 80000,
		CurrentPenalty: 800, BestPenalty: 784, InitialPenalty: 1000,
		Temperature: 0.001, BaseAction: CheckpointEarlyStop,
	}, snap)

	decisions := e.Decisions()
	if len(decisions) == 0 {
		t.Fatal("should record decisions")
	}
	if decisions[0].PolicyUsed != "rule" {
		t.Errorf("PolicyUsed = %q, want rule (no model loaded)", decisions[0].PolicyUsed)
	}
}

func TestCheckpointEngine_RecordsDecisions(t *testing.T) {
	e := NewCheckpointEngine(PolicySearchConfig{PolicyMode: "rules"})
	snap := SearchHookSnapshot{}

	for i := 1; i <= 5; i++ {
		snap.CheckpointNum = i
		e.Evaluate(CheckpointInput{
			Algorithm: "sa", Candidates: i * 10000,
			CurrentPenalty: 800 - i*10, BestPenalty: 750, InitialPenalty: 1000,
			Temperature: 50.0 / float64(i), BaseAction: CheckpointContinue,
		}, snap)
	}

	if len(e.Decisions()) != 5 {
		t.Errorf("expected 5 decisions, got %d", len(e.Decisions()))
	}
}

func TestCheckpointEngine_BuildFeaturesDomain(t *testing.T) {
	e := NewCheckpointEngine(PolicySearchConfig{
		PolicyMode: "hybrid",
		Domain:     "cvrp",
		Instance:   "A-n32-k5",
	})
	fv := e.BuildFeatures(CheckpointInput{
		Algorithm: "sa", Candidates: 50000,
		CurrentPenalty: 800, BestPenalty: 784, InitialPenalty: 1000,
		Temperature: 0.5,
	}, SearchHookSnapshot{IterationsTotal: 100000})

	if fv.Problem != "cvrp" {
		t.Errorf("Problem = %q, want cvrp", fv.Problem)
	}
	if fv.Instance != "A-n32-k5" {
		t.Errorf("Instance = %q, want A-n32-k5", fv.Instance)
	}
}

func TestCheckpointEngine_LearnedStagnationWithDomain(t *testing.T) {
	policyDir := filepath.Join("..", "..", "..", "ml", "policies")
	if _, err := os.Stat(filepath.Join(policyDir, StagnationPolicyFile)); err != nil {
		t.Skip("policy dir not available")
	}

	e := NewCheckpointEngine(PolicySearchConfig{
		PolicyMode: "hybrid",
		PolicyDir:  policyDir,
		Domain:     "cvrp",
		Instance:   "A-n32-k5",
	})
	if e.StagnationDetector() == nil {
		t.Fatal("expected stagnation model to load")
	}

	fv := e.BuildFeatures(CheckpointInput{
		Algorithm: "sa", Candidates: 180000,
		CurrentPenalty: 814, BestPenalty: 814, InitialPenalty: 1145,
		Temperature: 0.001,
	}, SearchHookSnapshot{IterationsTotal: 500000})

	assessment := e.StagnationDetector().Assess(fv)
	if assessment.PolicyConfidence <= 0 {
		t.Fatalf("expected learned model confidence, got %v reason=%s", assessment.PolicyConfidence, assessment.Reason)
	}
	if assessment.Reason == "no_curve_for_config" {
		t.Fatal("stagnation model did not match cvrp/sa domain")
	}
}

func TestCheckpointEngine_ShadowNeverStops(t *testing.T) {
	e := NewCheckpointEngine(PolicySearchConfig{PolicyMode: "learned", PolicyDir: "/nonexistent"})
	in := CheckpointInput{
		Algorithm: "sa", Candidates: 80000,
		CurrentPenalty: 800, BestPenalty: 784, InitialPenalty: 1000,
		Temperature: 0.001, BaseAction: CheckpointEarlyStop,
	}

	shadowAction := e.Evaluate(in, SearchHookSnapshot{ShadowMode: true})
	assistAction := e.Evaluate(CheckpointInput{
		Algorithm: in.Algorithm, Candidates: in.Candidates,
		CurrentPenalty: in.CurrentPenalty, BestPenalty: in.BestPenalty,
		InitialPenalty: in.InitialPenalty, Temperature: in.Temperature,
		BaseAction: in.BaseAction,
	}, SearchHookSnapshot{ShadowMode: false})

	if shadowAction != CheckpointContinue {
		t.Errorf("shadow action = %q, want continue", shadowAction)
	}
	if assistAction != CheckpointEarlyStop {
		t.Errorf("assist action = %q, want early_stop", assistAction)
	}
}
