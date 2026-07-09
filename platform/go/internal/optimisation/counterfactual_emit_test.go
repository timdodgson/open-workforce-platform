package optimisation

import "testing"

func TestBuildCounterfactualRecords(t *testing.T) {
	records := BuildCounterfactualRecords(CounterfactualEmitInput{
		RunID:          "run-1",
		Domain:         "cvrp",
		Instance:       "A-n32-k5",
		Algorithm:      "sa",
		InitialPenalty: 1000,
		BestPenalty:    800,
		Decisions: []PolicySearchDecision{
			{Checkpoint: 1, PolicyUsed: "hybrid_learned", Action: "early_stop", Confidence: 0.9},
		},
	})
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Domain != "cvrp" {
		t.Fatalf("domain = %q", records[0].Domain)
	}
	if len(records[0].CounterfactualActions) == 0 {
		t.Fatal("expected counterfactual alternatives")
	}
}

func TestWriteCounterfactualFromPolicyDecisions(t *testing.T) {
	dir := t.TempDir()
	err := WriteCounterfactualFromPolicyDecisions(dir, CounterfactualEmitInput{
		RunID: "run-2", Domain: "jss", Instance: "ft06", Algorithm: "lahc",
		InitialPenalty: 500, BestPenalty: 400,
		Decisions: []PolicySearchDecision{
			{PolicyUsed: "learned", Action: "continue", Confidence: 0.7},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}
