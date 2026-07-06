package inrc2

import (
	"os"
	"strings"
	"testing"
)

func TestRuleBasedEngine_DefaultRun(t *testing.T) {
	engine := NewRuleBasedEngine()
	decision := engine.Evaluate(WorkerDecisionInput{
		Algorithm:        "sa",
		Week:             3,
		ParentObjective:  3500,
		GlobalBest:       3400,
		DistanceFromBest: 100,
		AllocatedIters:   60000,
	})

	if decision.Recommendation != RecRun && decision.Recommendation != RecIncreaseBudget {
		// Small gap — should run or increase budget.
		t.Errorf("Expected run/increase_budget, got %s", decision.Recommendation)
	}
}

func TestRuleBasedEngine_SkipFarFromBest(t *testing.T) {
	engine := NewRuleBasedEngine()
	// Single parent_gap signal alone should NOT skip — it should reduce_budget.
	// SKIP requires multiple independent negative signals.
	decision := engine.Evaluate(WorkerDecisionInput{
		Algorithm:        "sa",
		ParentObjective:  6000,
		GlobalBest:       3000,
		DistanceFromBest: 3000,
		AllocatedIters:   60000,
	})

	if decision.Recommendation != RecReduceBudget {
		t.Errorf("Expected reduce_budget (single signal is advisory only), got %s", decision.Recommendation)
	}
}

func TestRuleBasedEngine_SkipRequiresMultipleSignals(t *testing.T) {
	engine := NewRuleBasedEngine()
	// Large gap + stale lineage + crowded = 3+ signals → SKIP.
	decision := engine.Evaluate(WorkerDecisionInput{
		Algorithm:                  "sa",
		ParentObjective:            6000,
		GlobalBest:                 3000,
		DistanceFromBest:           3000,
		AllocatedIters:             60000,
		WorkerCount:                60,
		GenerationsSinceGlobalBest: 15,
	})

	if decision.Recommendation != RecSkip {
		t.Errorf("Expected skip (3+ negative signals), got %s", decision.Recommendation)
	}
	if decision.Confidence < 0.6 {
		t.Errorf("Expected confidence >= 0.6, got %f", decision.Confidence)
	}
}

func TestRuleBasedEngine_NeverSkipGlobalBestLineage(t *testing.T) {
	engine := NewRuleBasedEngine()
	// Even with large gap, if in global best lineage → never skip.
	decision := engine.Evaluate(WorkerDecisionInput{
		Algorithm:                  "sa",
		ParentObjective:            6000,
		GlobalBest:                 3000,
		DistanceFromBest:           3000,
		AllocatedIters:             60000,
		WorkerCount:                60,
		GenerationsSinceGlobalBest: 15,
		IsGlobalBestLineage:        true,
	})

	if decision.Recommendation == RecSkip {
		t.Errorf("Should never skip a global best lineage worker, got %s", decision.Recommendation)
	}
	if decision.Recommendation != RecIncreaseBudget {
		t.Errorf("Expected increase_budget for lineage protection, got %s", decision.Recommendation)
	}
}

func TestRuleBasedEngine_IncreaseBudgetFromGlobalBest(t *testing.T) {
	engine := NewRuleBasedEngine()
	decision := engine.Evaluate(WorkerDecisionInput{
		Algorithm:        "sa",
		ParentObjective:  3000,
		GlobalBest:       3000,
		DistanceFromBest: 0,
		AllocatedIters:   60000,
	})

	if decision.Recommendation != RecIncreaseBudget {
		t.Errorf("Expected increase_budget (spawned from global best), got %s", decision.Recommendation)
	}
	if decision.SuggestedBudget != 120000 {
		t.Errorf("Expected 120000, got %d", decision.SuggestedBudget)
	}
}

func TestRuleBasedEngine_LowEntropy(t *testing.T) {
	engine := NewRuleBasedEngine()
	decision := engine.Evaluate(WorkerDecisionInput{
		Algorithm:        "sa",
		ParentObjective:  3500,
		GlobalBest:       3400,
		DistanceFromBest: 100,
		Entropy:          0.5,
		AllocatedIters:   60000,
	})

	// With DistanceFromBest=100 and GlobalBest=3400, gapPct ~2.9% < 25%,
	// so no negative signals. Low entropy should trigger change_algorithm.
	if decision.Recommendation != RecChangeAlgo {
		t.Errorf("Expected change_algorithm (low entropy), got %s", decision.Recommendation)
	}
	if decision.SuggestedAlgorithm != "lahc" {
		t.Errorf("Expected lahc, got %s", decision.SuggestedAlgorithm)
	}
}

func TestShadowRecorder(t *testing.T) {
	recorder := NewShadowRecorder()
	engine := NewRuleBasedEngine()

	input := WorkerDecisionInput{
		Algorithm: "sa", Week: 1, ParentObjective: 3800, GlobalBest: 3800,
		DistanceFromBest: 0, AllocatedIters: 60000,
	}
	decision := engine.Evaluate(input)
	idx := recorder.RecordDecision(0, input, decision)

	// Simulate worker completion.
	recorder.RecordOutcome(idx, true, true, 200, 3600, 50)

	records := recorder.Records()
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}
	if !records[0].Improved {
		t.Error("Expected improved=true")
	}
	if records[0].ImprovementAmount != 200 {
		t.Errorf("Expected improvement 200, got %d", records[0].ImprovementAmount)
	}
	if records[0].ROI != 4000.0 { // 200 / 50 * 1000
		t.Errorf("Expected ROI 4000, got %f", records[0].ROI)
	}
}

func TestWriteWorkerDecisionsCSV(t *testing.T) {
	records := []WorkerDecisionRecord{
		{
			WorkerID: 1, Week: 1, Algorithm: "sa",
			ParentObjective: 3800, GlobalBest: 3800,
			Recommendation: RecIncreaseBudget, Confidence: 0.7,
			ReasonCodes: "spawned_from_global_best", SuggestedBudget: 120000,
			Improved: true, ImprovementAmount: 200, FinalObjective: 3600,
			RuntimeMs: 50, ROI: 4000,
		},
		{
			WorkerID: 2, Week: 1, Algorithm: "sa",
			ParentObjective: 5000, GlobalBest: 3000,
			Recommendation: RecSkip, Confidence: 0.7,
			ReasonCodes: "parent_gap_67_pct",
			Improved: true, ImprovementAmount: 50, FinalObjective: 4950,
			RuntimeMs: 50, ROI: 1000,
		},
	}

	path := t.TempDir() + "/worker_decisions.csv"
	err := WriteWorkerDecisionsCSV(path, records)
	if err != nil {
		t.Fatalf("WriteWorkerDecisionsCSV: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "increase_budget") {
		t.Error("Expected increase_budget in row 1")
	}
	if !strings.Contains(lines[2], "skip") {
		t.Error("Expected skip in row 2")
	}
}
