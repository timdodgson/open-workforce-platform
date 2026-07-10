package optimisation

import (
	"os"
	"testing"
)

func TestPolicySearchHookRunner_NilWhenOff(t *testing.T) {
	pConfig := PolicySearchConfig{PolicyMode: "learned"}
	runner := NewPolicySearchHookRunner("off", DefaultSearchAssistConfig(), 100000, pConfig)
	if runner != nil {
		t.Error("runner should be nil when assist mode is off")
	}
}

func TestPolicySearchHookRunner_ShadowAssistNeverStops(t *testing.T) {
	config := DefaultSearchAssistConfig()
	config.StagnationWindow = 10000
	config.MinBudgetFraction = 0.20
	pConfig := PolicySearchConfig{PolicyMode: "learned", PolicyDir: "/nonexistent"}

	shadow := NewPolicySearchHookRunner("shadow", config, 100000, pConfig)
	assist := NewPolicySearchHookRunner("assist", config, 100000, pConfig)

	shadowAction := shadow.RunCheckpoint("sa", 80000, 800, 784, 1000, 0.001)
	assistAction := assist.RunCheckpoint("sa", 80000, 800, 784, 1000, 0.001)

	if shadowAction != SearchContinue {
		t.Errorf("shadow action = %q, want continue", shadowAction)
	}
	if assistAction != SearchEarlyStop {
		t.Errorf("assist action = %q, want early_stop", assistAction)
	}
}

func TestNewSearchHooks_PolicyModeDefaultsToShadowAssist(t *testing.T) {
	hooks := newSearchHooks(SearchConfig{
		Iterations: 100000,
		PolicyMode: "rules",
	})
	if hooks == nil {
		t.Fatal("expected non-nil hooks when PolicyMode is set")
	}
	_, policyDecisions := finalizeSearchHooks(hooks, 100, 10000)
	_ = policyDecisions
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

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(string(data)) < 50 {
		t.Errorf("CSV too short: %d bytes", len(data))
	}
}
