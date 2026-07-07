package optimisation

import (
	"os"
	"testing"
)

func TestPolicySearchHookRunner_RulesMode(t *testing.T) {
	config := DefaultSearchAssistConfig()
	config.StagnationWindow = 10000
	pConfig := PolicySearchConfig{PolicyMode: "rules"}

	runner := NewPolicySearchHookRunner("shadow", config, 100000, pConfig)
	if runner == nil {
		t.Fatal("runner should not be nil for shadow mode")
	}

	action := runner.RunPolicyCheckpoint("sa", 50000, 800, 784, 1000, 0.5)
	if action != SearchContinue {
		t.Errorf("Action = %q, want continue (no stagnation yet)", action)
	}

	decisions := runner.Decisions()
	if len(decisions) == 0 {
		t.Fatal("should record decisions")
	}
	if decisions[0].PolicyUsed != "rule" {
		t.Errorf("PolicyUsed = %q, want rule", decisions[0].PolicyUsed)
	}
}

func TestPolicySearchHookRunner_LearnedMode_NoModel(t *testing.T) {
	config := DefaultSearchAssistConfig()
	pConfig := PolicySearchConfig{PolicyMode: "learned", PolicyDir: "/nonexistent"}

	runner := NewPolicySearchHookRunner("assist", config, 100000, pConfig)
	if runner == nil {
		t.Fatal("runner should not be nil")
	}

	// Without models loaded, falls back to rule behaviour.
	action := runner.RunPolicyCheckpoint("sa", 80000, 800, 784, 1000, 0.001)
	// Stagnation window default is 50000. Plateau = 80000 (no improvements).
	// Rule would stop. Learned not loaded → uses rule.
	decisions := runner.Decisions()
	if len(decisions) == 0 {
		t.Fatal("should record decisions")
	}
	// With no learned model, should use rule.
	if decisions[0].PolicyUsed != "rule" {
		t.Errorf("PolicyUsed = %q, want rule (no model loaded)", decisions[0].PolicyUsed)
	}
	_ = action
}

func TestPolicySearchHookRunner_NilWhenOff(t *testing.T) {
	pConfig := PolicySearchConfig{PolicyMode: "learned"}
	runner := NewPolicySearchHookRunner("off", DefaultSearchAssistConfig(), 100000, pConfig)
	if runner != nil {
		t.Error("runner should be nil when assist mode is off")
	}
}

func TestPolicySearchHookRunner_RecordsDecisions(t *testing.T) {
	config := DefaultSearchAssistConfig()
	config.CheckpointInterval = 10000
	pConfig := PolicySearchConfig{PolicyMode: "rules"}

	runner := NewPolicySearchHookRunner("shadow", config, 100000, pConfig)

	// Simulate 5 checkpoints.
	for i := 1; i <= 5; i++ {
		runner.RunPolicyCheckpoint("sa", i*10000, 800-i*10, 750, 1000, 50.0/float64(i))
	}

	decisions := runner.Decisions()
	if len(decisions) != 5 {
		t.Errorf("expected 5 decisions, got %d", len(decisions))
	}
	for _, d := range decisions {
		if d.PolicyMode != "rules" {
			t.Errorf("PolicyMode = %q, want rules", d.PolicyMode)
		}
	}
}

func TestWritePolicyDecisionsCSV(t *testing.T) {
	decisions := []PolicySearchDecision{
		{Checkpoint: 0, Candidates: 10000, PolicyMode: "hybrid", PolicyUsed: "hybrid_learned", Action: "continue", Confidence: 0.75},
		{Checkpoint: 1, Candidates: 20000, PolicyMode: "hybrid", PolicyUsed: "hybrid_rule", Action: "early_stop", Confidence: 0.45, FallbackReason: "learned_low_confidence"},
	}

	path := t.TempDir() + "/policy_decisions.csv"
	err := WritePolicyDecisionsCSV(path, decisions)
	if err != nil {
		t.Fatalf("WritePolicyDecisionsCSV failed: %v", err)
	}

	// Verify file exists and has content.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	if len(content) < 50 {
		t.Errorf("CSV too short: %d bytes", len(content))
	}
}
