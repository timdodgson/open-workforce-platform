package policy_test

import (
	"os"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/policy"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

func TestPolicySearchHookRunner_NilWhenOff(t *testing.T) {
	pConfig := policy.PolicySearchConfig{PolicyMode: "learned"}
	runner := policy.NewPolicySearchHookRunner("off", searchdef.DefaultSearchAssistConfig(), 100000, pConfig)
	if runner != nil {
		t.Error("runner should be nil when assist mode is off")
	}
}

func TestPolicySearchHookRunner_ShadowAssistNeverStops(t *testing.T) {
	config := searchdef.DefaultSearchAssistConfig()
	config.StagnationWindow = 10000
	config.MinBudgetFraction = 0.20
	pConfig := policy.PolicySearchConfig{PolicyMode: "learned", PolicyDir: "/nonexistent"}

	shadow := policy.NewPolicySearchHookRunner("shadow", config, 100000, pConfig)
	assistRunner := policy.NewPolicySearchHookRunner("assist", config, 100000, pConfig)

	shadowAction := shadow.RunCheckpoint("sa", 80000, 800, 784, 1000, 0.001)
	assistAction := assistRunner.RunCheckpoint("sa", 80000, 800, 784, 1000, 0.001)

	if shadowAction != searchdef.SearchContinue {
		t.Errorf("shadow action = %q, want continue", shadowAction)
	}
	if assistAction != searchdef.SearchEarlyStop {
		t.Errorf("assist action = %q, want early_stop", assistAction)
	}
}

func TestWritePolicyDecisionsCSV(t *testing.T) {
	decisions := []policy.PolicySearchDecision{
		{Checkpoint: 0, Candidates: 10000, PolicyMode: "hybrid", PolicyUsed: "hybrid_learned", Action: "continue", Confidence: 0.75},
		{Checkpoint: 1, Candidates: 20000, PolicyMode: "hybrid", PolicyUsed: "hybrid_rule", Action: "early_stop", Confidence: 0.45, FallbackReason: "learned_low_confidence"},
	}

	path := t.TempDir() + "/policy_decisions.csv"
	err := policy.WritePolicyDecisionsCSV(path, decisions)
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