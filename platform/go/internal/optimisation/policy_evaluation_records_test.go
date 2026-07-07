package optimisation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildPolicyEvaluationRecords_SkipsCosmeticRules(t *testing.T) {
	records := BuildPolicyEvaluationRecords(PolicyEvaluationInput{
		RunID: "test-run", Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa",
		InitialPenalty: 1000, BestPenalty: 800,
		Decisions: []PolicySearchDecision{
			{PolicyUsed: "rule", FallbackReason: "no_learned_assessment", Action: "continue"},
			{PolicyUsed: "hybrid_learned", Action: "early_stop", Confidence: 0.85},
		},
	})
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].PolicyType != "hybrid" {
		t.Fatalf("expected hybrid policy type, got %s", records[0].PolicyType)
	}
}

func TestWritePolicyEvaluationCSV(t *testing.T) {
	dir := t.TempDir()
	err := WritePolicyEvaluationCSV(dir, PolicyEvaluationInput{
		RunID: "si2-test", Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa",
		InitialPenalty: 1145, BestPenalty: 814,
		Decisions: []PolicySearchDecision{
			{PolicyUsed: "hybrid_learned", Action: "early_stop", Confidence: 0.85},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "policy_evaluation.csv")); err != nil {
		t.Fatalf("csv not written: %v", err)
	}
}
