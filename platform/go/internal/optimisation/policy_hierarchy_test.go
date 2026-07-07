package optimisation

import (
	"testing"
)

func globalRulePolicy() *RulePolicy {
	return NewRulePolicy("global-rules", "1.0.0", "*", "search", []Rule{
		{Name: "default", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "continue", Reason: "global_default"}
			}},
	})
}

func domainPolicy(domain string) *RulePolicy {
	return NewRulePolicy(domain+"-rules", "1.0.0", domain, "search", []Rule{
		{Name: "domain_default", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "early_stop", Reason: "domain:" + domain}
			}},
	})
}

func instancePolicy(domain, instance string) *RulePolicy {
	return NewRulePolicy(domain+"-"+instance, "1.0.0", domain, "search", []Rule{
		{Name: "instance_specific", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "extend", Reason: "instance:" + instance}
			}},
	})
}

func TestHierarchy_ResolvesInstance(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", globalRulePolicy())
	h.RegisterDomain("cvrp", "search", domainPolicy("cvrp"))
	h.RegisterInstance("cvrp", "A-n32-k5", "search", instancePolicy("cvrp", "A-n32-k5"))

	p, level := h.Resolve("cvrp", "A-n32-k5", "search")
	if level != LevelInstance {
		t.Errorf("level = %q, want instance", level)
	}
	m := p.Metadata()
	if m.ID != "cvrp-A-n32-k5" {
		t.Errorf("policy ID = %q, want cvrp-A-n32-k5", m.ID)
	}
}

func TestHierarchy_FallsToDomain(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", globalRulePolicy())
	h.RegisterDomain("cvrp", "search", domainPolicy("cvrp"))

	p, level := h.Resolve("cvrp", "A-n80-k10", "search")
	if level != LevelDomain {
		t.Errorf("level = %q, want domain", level)
	}
	m := p.Metadata()
	if m.ID != "cvrp-rules" {
		t.Errorf("policy ID = %q, want cvrp-rules", m.ID)
	}
}

func TestHierarchy_FallsToGlobal(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", globalRulePolicy())

	_, level := h.Resolve("jss", "la01", "search")
	if level != LevelGlobal {
		t.Errorf("level = %q, want global", level)
	}
}

func TestHierarchy_FallsToNone(t *testing.T) {
	h := NewPolicyHierarchy()

	_, level := h.Resolve("vrptw", "C101", "portfolio")
	if level != LevelNone {
		t.Errorf("level = %q, want none", level)
	}
}

func TestHierarchy_DecideWithCascade(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", globalRulePolicy())
	h.RegisterDomain("cvrp", "search", domainPolicy("cvrp"))
	h.RegisterInstance("cvrp", "A-n32-k5", "search", instancePolicy("cvrp", "A-n32-k5"))

	ctx := PolicyContext{
		DecisionType: "search",
		Domain:       "cvrp",
		Instance:     "A-n32-k5",
	}

	d := h.DecideWithHierarchy(ctx)

	if d.Action != "extend" {
		t.Errorf("Action = %q, want extend (instance level)", d.Action)
	}
	if d.ResolvedLevel != LevelInstance {
		t.Errorf("ResolvedLevel = %q, want instance", d.ResolvedLevel)
	}
	if d.Cascaded {
		t.Error("should not cascade when instance answers")
	}
}

func TestHierarchy_DecideCascadesOnDefer(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", globalRulePolicy())
	h.RegisterDomain("cvrp", "search", domainPolicy("cvrp"))

	// Instance policy that always defers.
	deferPolicy := NewRulePolicy("defers", "1.0.0", "cvrp", "search", []Rule{
		{Name: "low_conf", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "defer", Reason: "low_confidence"}
			}},
	})
	h.RegisterInstance("cvrp", "A-n32-k5", "search", deferPolicy)

	ctx := PolicyContext{
		DecisionType: "search",
		Domain:       "cvrp",
		Instance:     "A-n32-k5",
	}

	d := h.DecideWithHierarchy(ctx)

	// Should cascade past instance to domain.
	if d.Action != "early_stop" {
		t.Errorf("Action = %q, want early_stop (domain level)", d.Action)
	}
	if d.ResolvedLevel != LevelDomain {
		t.Errorf("ResolvedLevel = %q, want domain", d.ResolvedLevel)
	}
	if !d.Cascaded {
		t.Error("should be marked as cascaded")
	}
	if len(d.LevelsAttempted) != 2 {
		t.Errorf("LevelsAttempted = %d, want 2", len(d.LevelsAttempted))
	}
}

func TestHierarchy_AllDefer(t *testing.T) {
	h := NewPolicyHierarchy()
	deferPolicy := NewRulePolicy("defers", "1.0.0", "*", "search", []Rule{
		{Name: "defer", Matches: func(_ PolicyContext) bool { return true },
			Decide: func(_ PolicyContext) PolicyDecision {
				return PolicyDecision{Action: "defer"}
			}},
	})
	h.RegisterGlobal("search", deferPolicy)

	ctx := PolicyContext{DecisionType: "search", Domain: "nrp", Instance: "n012w8"}
	d := h.DecideWithHierarchy(ctx)

	if d.Action != "continue" {
		t.Errorf("Action = %q, want continue (safe default)", d.Action)
	}
	if d.ResolvedLevel != LevelNone {
		t.Errorf("ResolvedLevel = %q, want none", d.ResolvedLevel)
	}
}

func TestTransferPolicy_PenalisesConfidence(t *testing.T) {
	source := domainPolicy("cvrp")
	transfer := NewTransferPolicy(TransferPolicyConfig{
		Source:            source,
		SourceDomain:      "cvrp",
		TargetDomain:      "vrptw",
		ConfidencePenalty: 0.70,
	})

	ctx := PolicyContext{DecisionType: "search", Domain: "vrptw"}
	d := transfer.Decide(ctx)

	// Source returns confidence 1.0 (rule). Transfer penalises to 0.70.
	if d.Confidence != 0.70 {
		t.Errorf("Confidence = %f, want 0.70 (penalised)", d.Confidence)
	}
	if d.Action != "early_stop" {
		t.Errorf("Action = %q, want early_stop (from source)", d.Action)
	}
}

func TestTransferPolicy_Metadata(t *testing.T) {
	source := domainPolicy("cvrp")
	transfer := NewTransferPolicy(TransferPolicyConfig{
		Source: source, SourceDomain: "cvrp", TargetDomain: "jss",
		ConfidencePenalty: 0.65,
	})

	m := transfer.Metadata()
	if m.Type != "transfer" {
		t.Errorf("Type = %q, want transfer", m.Type)
	}
	if m.Domain != "jss" {
		t.Errorf("Domain = %q, want jss (target)", m.Domain)
	}
}

func TestHierarchy_Summary(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterGlobal("search", globalRulePolicy())
	h.RegisterDomain("cvrp", "search", domainPolicy("cvrp"))
	h.RegisterDomain("jss", "search", domainPolicy("jss"))
	h.RegisterInstance("cvrp", "A-n32-k5", "search", instancePolicy("cvrp", "A-n32-k5"))

	entries := h.Summary()
	if len(entries) != 4 {
		t.Errorf("Summary returned %d entries, want 4", len(entries))
	}

	// Verify levels are correct.
	levels := map[PolicyLevel]int{}
	for _, e := range entries {
		levels[e.Level]++
	}
	if levels[LevelGlobal] != 1 {
		t.Errorf("global count = %d, want 1", levels[LevelGlobal])
	}
	if levels[LevelDomain] != 2 {
		t.Errorf("domain count = %d, want 2", levels[LevelDomain])
	}
	if levels[LevelInstance] != 1 {
		t.Errorf("instance count = %d, want 1", levels[LevelInstance])
	}
}

func TestHierarchy_AllDomains(t *testing.T) {
	h := NewPolicyHierarchy()
	h.RegisterDomain("nrp", "worker", domainPolicy("nrp"))
	h.RegisterDomain("cvrp", "search", domainPolicy("cvrp"))
	h.RegisterDomain("jss", "search", domainPolicy("jss"))
	h.RegisterDomain("vrptw", "search", domainPolicy("vrptw"))

	// Each domain should resolve its own policy.
	domains := []string{"nrp", "cvrp", "jss", "vrptw"}
	for _, d := range domains {
		decType := "search"
		if d == "nrp" {
			decType = "worker"
		}
		_, level := h.Resolve(d, "", decType)
		if level != LevelDomain {
			t.Errorf("%s resolved at %q, want domain", d, level)
		}
	}
}
