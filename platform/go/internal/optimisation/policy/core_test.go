package policy

import (
	"testing"
	"time"
)

// ─── Test helpers ───

type mockModel struct {
	action     string
	confidence float64
}

func (m *mockModel) Predict(_ FeatureVector) ModelPrediction {
	return ModelPrediction{
		Action:     m.action,
		Confidence: m.confidence,
		Reason:     "mock_prediction",
	}
}

func testRulePolicy() *RulePolicy {
	rules := []Rule{
		{
			Name:    "high_distance",
			Matches: func(ctx PolicyContext) bool { return ctx.Features.DistanceFromBest > 1000 },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "skip", Reason: "rule:high_distance"}
			},
		},
		{
			Name:    "default_run",
			Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "run", Reason: "rule:default_run"}
			},
		},
	}
	return NewRulePolicy("worker-rules", "1.0.0", "nrp", "worker", rules)
}

func testContext(distFromBest int) PolicyContext {
	return PolicyContext{
		DecisionType: "worker",
		Domain:       "nrp",
		Instance:     "n012w8",
		Features: FeatureVector{
			SchemaVersion:    FeatureSchemaVersion,
			Problem:          "nrp",
			Instance:         "n012w8",
			Algorithm:        "sa",
			DistanceFromBest: distFromBest,
		},
	}
}

// ─── RulePolicy tests ───

func TestRulePolicy_MatchesFirstRule(t *testing.T) {
	p := testRulePolicy()
	ctx := testContext(2000) // distance > 1000

	d := p.Decide(ctx)

	if d.Action != "skip" {
		t.Errorf("Action = %q, want skip", d.Action)
	}
	if d.PolicyID != "worker-rules" {
		t.Errorf("PolicyID = %q, want worker-rules", d.PolicyID)
	}
	if d.PolicyVersion != "1.0.0" {
		t.Errorf("PolicyVersion = %q, want 1.0.0", d.PolicyVersion)
	}
	if d.IsFallback {
		t.Error("IsFallback should be false for direct rule match")
	}
}

func TestRulePolicy_FallsToDefault(t *testing.T) {
	p := testRulePolicy()
	ctx := testContext(500) // distance < 1000, matches default_run

	d := p.Decide(ctx)

	if d.Action != "run" {
		t.Errorf("Action = %q, want run", d.Action)
	}
	if d.Confidence != 1.0 {
		t.Errorf("Confidence = %f, want 1.0", d.Confidence)
	}
}

func TestRulePolicy_NoRulesMatch(t *testing.T) {
	p := NewRulePolicy("empty", "1.0.0", "*", "worker", nil)
	ctx := testContext(100)

	d := p.Decide(ctx)

	if d.Action != "continue" {
		t.Errorf("Action = %q, want continue", d.Action)
	}
	if d.Reason != "no_rule_matched" {
		t.Errorf("Reason = %q, want no_rule_matched", d.Reason)
	}
}

func TestRulePolicy_Metadata(t *testing.T) {
	p := testRulePolicy()
	m := p.Metadata()

	if m.Type != "rule" {
		t.Errorf("Type = %q, want rule", m.Type)
	}
	if m.Domain != "nrp" {
		t.Errorf("Domain = %q, want nrp", m.Domain)
	}
	if m.TrainedSamples != 0 {
		t.Errorf("TrainedSamples = %d, want 0", m.TrainedSamples)
	}
}

// ─── LearnedPolicy tests ───

func TestLearnedPolicy_HighConfidence(t *testing.T) {
	model := &mockModel{action: "early_stop", confidence: 0.85}
	p := NewLearnedPolicy(LearnedPolicyConfig{
		ID:             "search-learned",
		Version:        "2.1.0",
		Domain:         "cvrp",
		DecisionType:   "search",
		Model:          model,
		Threshold:      0.60,
		TrainedSamples: 240,
		CreatedAt:      time.Now(),
	})

	d := p.Decide(testContext(100))

	if d.Action != "early_stop" {
		t.Errorf("Action = %q, want early_stop", d.Action)
	}
	if d.Confidence != 0.85 {
		t.Errorf("Confidence = %f, want 0.85", d.Confidence)
	}
	if d.PolicyVersion != "2.1.0" {
		t.Errorf("PolicyVersion = %q, want 2.1.0", d.PolicyVersion)
	}
}

func TestLearnedPolicy_LowConfidence_Defers(t *testing.T) {
	model := &mockModel{action: "early_stop", confidence: 0.45}
	p := NewLearnedPolicy(LearnedPolicyConfig{
		ID:        "search-learned",
		Version:   "2.1.0",
		Domain:    "cvrp",
		Model:     model,
		Threshold: 0.60,
	})

	d := p.Decide(testContext(100))

	if d.Action != "defer" {
		t.Errorf("Action = %q, want defer", d.Action)
	}
	if d.Reason != "learned_low_confidence" {
		t.Errorf("Reason = %q, want learned_low_confidence", d.Reason)
	}
}

func TestLearnedPolicy_Metadata(t *testing.T) {
	p := NewLearnedPolicy(LearnedPolicyConfig{
		ID:              "cvrp-budget",
		Version:         "3.0.0",
		Domain:          "cvrp",
		DecisionType:    "portfolio",
		TrainedSamples:  500,
		ValidationScore: 0.82,
	})

	m := p.Metadata()
	if m.Type != "learned" {
		t.Errorf("Type = %q, want learned", m.Type)
	}
	if m.TrainedSamples != 500 {
		t.Errorf("TrainedSamples = %d, want 500", m.TrainedSamples)
	}
}

// ─── HybridPolicy tests ───

func TestHybridPolicy_UsesLearnedWhenConfident(t *testing.T) {
	model := &mockModel{action: "extend", confidence: 0.90}
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "hybrid-learned", Version: "1.0.0", Domain: "vrptw",
		Model: model, Threshold: 0.60,
	})
	fallback := testRulePolicy()
	hybrid := NewHybridPolicy(learned, fallback)

	d := hybrid.Decide(testContext(100))

	if d.Action != "extend" {
		t.Errorf("Action = %q, want extend", d.Action)
	}
	if d.IsFallback {
		t.Error("should not be fallback when learned is confident")
	}
}

func TestHybridPolicy_FallsBackWhenLowConfidence(t *testing.T) {
	model := &mockModel{action: "extend", confidence: 0.40}
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "hybrid-learned", Version: "1.0.0", Domain: "nrp",
		Model: model, Threshold: 0.60,
	})
	fallback := testRulePolicy()
	hybrid := NewHybridPolicy(learned, fallback)

	ctx := testContext(500) // matches "default_run" rule
	d := hybrid.Decide(ctx)

	if d.Action != "run" {
		t.Errorf("Action = %q, want run (from rules)", d.Action)
	}
	if !d.IsFallback {
		t.Error("should be marked as fallback")
	}
	if d.FallbackReason != "learned_low_confidence" {
		t.Errorf("FallbackReason = %q, want learned_low_confidence", d.FallbackReason)
	}
}

func TestHybridPolicy_Metadata(t *testing.T) {
	model := &mockModel{action: "run", confidence: 0.9}
	learned := NewLearnedPolicy(LearnedPolicyConfig{
		ID: "test", Version: "2.0.0", Domain: "jss", DecisionType: "search",
		Model: model, Threshold: 0.60,
	})
	hybrid := NewHybridPolicy(learned, testRulePolicy())

	m := hybrid.Metadata()
	if m.Type != "hybrid" {
		t.Errorf("Type = %q, want hybrid", m.Type)
	}
	if m.ID != "test+rules" {
		t.Errorf("ID = %q, want test+rules", m.ID)
	}
}

// ─── PolicyProvider tests ───

func TestPolicyProvider_ExactDomainMatch(t *testing.T) {
	pp := NewPolicyProvider()
	pp.Register("nrp", "worker", testRulePolicy())

	p := pp.GetPolicy("nrp", "worker")
	m := p.Metadata()

	if m.ID != "worker-rules" {
		t.Errorf("got policy %q, want worker-rules", m.ID)
	}
}

func TestPolicyProvider_UniversalFallback(t *testing.T) {
	pp := NewPolicyProvider()
	universal := NewRulePolicy("universal", "1.0.0", "*", "worker", nil)
	pp.Register("*", "worker", universal)

	p := pp.GetPolicy("cvrp", "worker") // no cvrp-specific policy
	m := p.Metadata()

	if m.ID != "universal" {
		t.Errorf("got policy %q, want universal", m.ID)
	}
}

func TestPolicyProvider_DefaultWhenNoneRegistered(t *testing.T) {
	pp := NewPolicyProvider()

	p := pp.GetPolicy("jss", "portfolio")
	d := p.Decide(testContext(0))

	if d.Action != "continue" {
		t.Errorf("Action = %q, want continue", d.Action)
	}
	if d.PolicyID != "default" {
		t.Errorf("PolicyID = %q, want default", d.PolicyID)
	}
}

func TestPolicyProvider_ListPolicies(t *testing.T) {
	pp := NewPolicyProvider()
	pp.Register("nrp", "worker", testRulePolicy())
	pp.Register("cvrp", "search", NewRulePolicy("cvrp-search", "1.0.0", "cvrp", "search", nil))

	list := pp.ListPolicies()
	if len(list) != 2 {
		t.Errorf("ListPolicies returned %d, want 2", len(list))
	}
}
