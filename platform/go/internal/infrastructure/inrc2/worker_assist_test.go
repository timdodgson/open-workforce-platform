package inrc2

import (
	"os"
	"strings"
	"testing"
)

func TestEvaluateSafety_AllowsSkipWhenSafe(t *testing.T) {
	input := WorkerDecisionInput{
		Algorithm:        "sa",
		ParentObjective:  6000,
		GlobalBest:       3000,
		DistanceFromBest: 3000,
		AllocatedIters:   60000,
	}
	decision := WorkerDecision{
		Recommendation: RecSkip,
		Confidence:     0.7,
		ReasonCodes:    []string{"parent_gap_100_pct", "stale_lineage", "crowded_and_distant"},
	}

	result := EvaluateSafety(input, decision)
	if !result.Safe {
		t.Errorf("Expected safe=true for distant worker, got safe=false (rule=%s)", result.Rule)
	}
}

func TestEvaluateSafety_BlocksSkipForGlobalBestLineage(t *testing.T) {
	input := WorkerDecisionInput{
		Algorithm:           "sa",
		ParentObjective:     3000,
		GlobalBest:          3000,
		DistanceFromBest:    0,
		AllocatedIters:      60000,
		IsGlobalBestLineage: true,
	}
	decision := WorkerDecision{
		Recommendation: RecSkip,
		Confidence:     0.8,
	}

	result := EvaluateSafety(input, decision)
	if result.Safe {
		t.Error("Expected safe=false for global best lineage, got safe=true")
	}
	if result.Rule != "global_best_lineage" {
		t.Errorf("Expected rule='global_best_lineage', got '%s'", result.Rule)
	}
}

func TestEvaluateSafety_BlocksLowConfidenceSkip(t *testing.T) {
	input := WorkerDecisionInput{
		Algorithm:        "sa",
		ParentObjective:  5000,
		GlobalBest:       3000,
		DistanceFromBest: 2000,
		AllocatedIters:   60000,
	}
	decision := WorkerDecision{
		Recommendation: RecSkip,
		Confidence:     0.55, // below 0.65 threshold
	}

	result := EvaluateSafety(input, decision)
	if result.Safe {
		t.Error("Expected safe=false for low confidence skip, got safe=true")
	}
	if result.Rule != "high_uncertainty" {
		t.Errorf("Expected rule='high_uncertainty', got '%s'", result.Rule)
	}
}

func TestAssistRecorder(t *testing.T) {
	ar := NewAssistRecorder()

	record := AssistRecord{
		WorkerID:        1,
		Algorithm:       "sa",
		Recommendation:  RecSkip,
		Confidence:      0.75,
		Outcome:         AssistAccepted,
		FinalAction:     RecSkip,
		FinalBudget:     0,
		FinalAlgorithm:  "sa",
	}

	idx := ar.RecordDecision(record)
	ar.RecordOutcome(idx, false, false, 0, 3000, 0)

	records := ar.Records()
	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}
	if records[0].Outcome != AssistAccepted {
		t.Errorf("Expected outcome=accepted, got %s", records[0].Outcome)
	}
	if records[0].Improved {
		t.Error("Expected improved=false")
	}
}

func TestWriteWorkerAssistCSV(t *testing.T) {
	records := []AssistRecord{
		{
			WorkerID: 1, Algorithm: "sa", Week: 0, Depth: 5,
			ParentObjective: 5000, GlobalBest: 3000, DistanceFromBest: 2000,
			Recommendation: RecSkip, Confidence: 0.75, ReasonCodes: "parent_gap_67_pct;stale;crowded",
			SafetyTriggered: false,
			Outcome: AssistAccepted, FinalAction: RecSkip, FinalBudget: 0, FinalAlgorithm: "sa",
			Improved: false, ImprovementAmount: 0, FinalObjective: 3000, RuntimeMs: 0,
		},
		{
			WorkerID: 2, Algorithm: "sa", Week: 0, Depth: 6,
			ParentObjective: 3000, GlobalBest: 3000, DistanceFromBest: 0,
			Recommendation: RecSkip, Confidence: 0.8, ReasonCodes: "stale",
			SafetyTriggered: true, SafetyRule: "spawned_from_global_best",
			Outcome: AssistRejected, FinalAction: RecIncreaseBudget, FinalBudget: 120000, FinalAlgorithm: "sa",
			Improved: true, ProducedGlobalBest: true, ImprovementAmount: 500, FinalObjective: 2500, RuntimeMs: 600,
		},
	}

	path := t.TempDir() + "/worker_assist.csv"
	err := WriteWorkerAssistCSV(path, records)
	if err != nil {
		t.Fatalf("WriteWorkerAssistCSV: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "accepted") {
		t.Error("Expected 'accepted' in row 1")
	}
	if !strings.Contains(lines[2], "rejected") {
		t.Error("Expected 'rejected' in row 2")
	}
	if !strings.Contains(lines[2], "spawned_from_global_best") {
		t.Error("Expected safety rule in row 2")
	}
}
