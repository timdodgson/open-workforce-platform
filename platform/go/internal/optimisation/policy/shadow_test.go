package policy

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
)

func TestShadowRunner_Disabled(t *testing.T) {
	r := NewPolicyShadowRunner("")
	rule := NewRulePolicy("rules", "1.0.0", "*", "search", []Rule{
		{Name: "default", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue", Confidence: 1.0}
			}},
	})
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "test", Version: "1.0.0", Domain: "cvrp", DecisionType: "search",
		Model: &mockModel{action: "early_stop", confidence: 0.85}, Threshold: 0.01,
	})

	ctx := PolicyContext{DecisionType: "search", Domain: "cvrp", Features: FeatureVector{}}
	d := r.Evaluate(ctx, rule, learned)

	if d.Action != "continue" {
		t.Errorf("Action = %q, want continue (rule always wins)", d.Action)
	}
}

func TestShadowRunner_RuleAlwaysControls(t *testing.T) {
	r := NewPolicyShadowRunner("")
	rule := NewRulePolicy("rules", "1.0.0", "*", "search", []Rule{
		{Name: "rule_stop", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "early_stop", Confidence: 0.70}
			}},
	})
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "test", Version: "2.0.0", Domain: "cvrp", DecisionType: "search",
		Model: &mockModel{action: "continue", confidence: 0.95}, Threshold: 0.01,
	})

	ctx := PolicyContext{DecisionType: "search", Domain: "cvrp", Features: FeatureVector{}}
	d := r.Evaluate(ctx, rule, learned)

	// Rule wins — even though learned is more confident.
	if d.Action != "early_stop" {
		t.Errorf("Action = %q, want early_stop (rule controls)", d.Action)
	}
}

func TestShadowRunner_RecordsAgreement(t *testing.T) {
	r := NewPolicyShadowRunner("")
	rule := NewRulePolicy("rules", "1.0.0", "*", "search", []Rule{
		{Name: "default", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue", Confidence: 0.80}
			}},
	})
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "test", Version: "1.0.0", Domain: "cvrp", DecisionType: "search",
		Model: &mockModel{action: "continue", confidence: 0.85}, Threshold: 0.01,
	})

	ctx := PolicyContext{DecisionType: "search", Domain: "cvrp", Features: FeatureVector{}}
	r.Evaluate(ctx, rule, learned)

	recs := r.Records()
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}
	if !recs[0].Agreement {
		t.Error("should agree (both say continue)")
	}
}

func TestShadowRunner_RecordsDisagreement(t *testing.T) {
	r := NewPolicyShadowRunner("")
	rule := NewRulePolicy("rules", "1.0.0", "*", "search", []Rule{
		{Name: "default", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue", Confidence: 0.50}
			}},
	})
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "test", Version: "2.0.0", Domain: "cvrp", DecisionType: "search",
		Model: &mockModel{action: "early_stop", confidence: 0.85}, Threshold: 0.01,
	})

	ctx := PolicyContext{DecisionType: "search", Domain: "cvrp", Features: FeatureVector{}}
	r.Evaluate(ctx, rule, learned)

	recs := r.Records()
	if recs[0].Agreement {
		t.Error("should disagree (continue vs early_stop)")
	}
	// Expected regret should be positive (learned more confident).
	if recs[0].ExpectedRegret <= 0 {
		t.Errorf("ExpectedRegret = %f, expected positive", recs[0].ExpectedRegret)
	}
}

func TestShadowRunner_Metrics(t *testing.T) {
	r := NewPolicyShadowRunner("")
	rule := NewRulePolicy("rules", "1.0.0", "*", "search", []Rule{
		{Name: "default", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue", Confidence: 0.60}
			}},
	})

	// 3 agreements.
	agreeModel := &mockModel{action: "continue", confidence: 0.80}
	learnedAgree := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "a", Version: "1.0.0", Model: agreeModel, Threshold: 0.01,
	})

	// 2 disagreements.
	disagreeModel := &mockModel{action: "early_stop", confidence: 0.90}
	learnedDisagree := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "d", Version: "1.0.0", Model: disagreeModel, Threshold: 0.01,
	})

	ctx := PolicyContext{DecisionType: "search", Domain: "cvrp", Features: FeatureVector{}}

	for i := 0; i < 3; i++ {
		r.Evaluate(ctx, rule, learnedAgree)
	}
	for i := 0; i < 2; i++ {
		r.Evaluate(ctx, rule, learnedDisagree)
	}

	m := r.Metrics()

	if m.TotalDecisions != 5 {
		t.Errorf("TotalDecisions = %d, want 5", m.TotalDecisions)
	}
	if m.Agreements != 3 {
		t.Errorf("Agreements = %d, want 3", m.Agreements)
	}
	if m.Disagreements != 2 {
		t.Errorf("Disagreements = %d, want 2", m.Disagreements)
	}
	if m.AgreementRate != 0.6 {
		t.Errorf("AgreementRate = %f, want 0.6", m.AgreementRate)
	}
}

func TestShadowRunner_WriteCSV(t *testing.T) {
	dir := t.TempDir()
	r := NewPolicyShadowRunner(dir)

	rule := NewRulePolicy("rules", "1.0.0", "*", "search", []Rule{
		{Name: "default", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue", Confidence: 0.70, Reason: "rule:default"}
			}},
	})
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "cvrp-v2", Version: "2.0.0", Domain: "cvrp", DecisionType: "search",
		Model: &mockModel{action: "early_stop", confidence: 0.82}, Threshold: 0.01,
	})

	ctx := PolicyContext{
		DecisionType: "search", Domain: "cvrp", Instance: "A-n32-k5",
		Features: FeatureVector{RunID: "test-shadow", Algorithm: "sa"},
	}
	r.Evaluate(ctx, rule, learned)
	r.RecordOutcome(784.0)

	if err := r.WriteCSV(); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	f, err := os.Open(filepath.Join(dir, "policy_shadow.csv"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (header + data), got %d", len(rows))
	}
	if rows[0][0] != "timestamp" {
		t.Errorf("header[0] = %q, want timestamp", rows[0][0])
	}
	if rows[1][6] != "continue" {
		t.Errorf("rule_action = %q, want continue", rows[1][6])
	}
	if rows[1][9] != "early_stop" {
		t.Errorf("learned_action = %q, want early_stop", rows[1][9])
	}
	if rows[1][14] != "0" {
		t.Errorf("agreement = %q, want 0 (disagree)", rows[1][14])
	}
}

func TestShadowRunner_Concurrent(t *testing.T) {
	r := NewPolicyShadowRunner("")
	rule := NewRulePolicy("rules", "1.0.0", "*", "search", []Rule{
		{Name: "d", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue"}
			}},
	})
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "t", Version: "1.0.0", Model: &mockModel{action: "continue", confidence: 0.7}, Threshold: 0.01,
	})

	done := make(chan struct{}, 50)
	for i := 0; i < 50; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			ctx := PolicyContext{DecisionType: "search", Features: FeatureVector{}}
			r.Evaluate(ctx, rule, learned)
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}

	m := r.Metrics()
	if m.TotalDecisions != 50 {
		t.Errorf("TotalDecisions = %d, want 50", m.TotalDecisions)
	}
}
