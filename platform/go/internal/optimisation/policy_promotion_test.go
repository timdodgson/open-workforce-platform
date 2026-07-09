package optimisation

import (
	"testing"
	"time"
)

func promotionRegistry() *PolicyLifecycleRegistry {
	now := time.Now()
	return &PolicyLifecycleRegistry{
		Versions: []PolicyVersionRecord{
			{
				ID: "cvrp-stag", Version: "1.0.0", Domain: "cvrp", DecisionType: "stagnation",
				Status: PolicyStatusTraining, CreatedAt: now,
				OfflineAccuracy: 0.85, RegretVsRules: -0.1, ShadowAccuracy: -1, ProductionAccuracy: -1,
			},
			{
				ID: "cvrp-stag", Version: "2.0.0", Domain: "cvrp", DecisionType: "search",
				Status: PolicyStatusShadow, CreatedAt: now,
				OfflineAccuracy: 0.82, ShadowAccuracy: 0.78, ProductionRuns: 30,
				RegretVsRules: -0.5,
			},
			{
				ID: "jss-budget", Version: "1.0.0", Domain: "jss", DecisionType: "budget",
				Status: PolicyStatusTraining, CreatedAt: now,
				OfflineAccuracy: 0.50, RegretVsRules: 0.2, // below accuracy + positive regret
			},
			{
				ID: "vrptw-restart", Version: "1.0.0", Domain: "vrptw", DecisionType: "restart",
				Status: PolicyStatusShadow, CreatedAt: now,
				OfflineAccuracy: 0.70, ShadowAccuracy: 0.55, ProductionRuns: 25,
				RegretVsRules: 0.0, // below accuracy threshold
			},
		},
	}
}

func TestPromoter_CandidateToShadow_Pass(t *testing.T) {
	reg := promotionRegistry()
	p := NewPolicyPromoter(DefaultPromotionRules(), reg)

	result := p.EvaluateOne("cvrp-stag", "1.0.0")

	if !result.Promoted {
		t.Errorf("should promote (accuracy 0.85 >= %.2f), blocked: %s", MinLearnedPolicyAgreement, result.BlockedReason)
	}
	if result.ToStatus != PolicyStatusShadow {
		t.Errorf("ToStatus = %q, want shadow", result.ToStatus)
	}
	v := reg.FindVersion("cvrp-stag", "1.0.0")
	if v.Status != PolicyStatusShadow {
		t.Errorf("registry status = %q, want shadow", v.Status)
	}
}

func TestPromoter_CandidateToShadow_Blocked(t *testing.T) {
	reg := promotionRegistry()
	p := NewPolicyPromoter(DefaultPromotionRules(), reg)

	result := p.EvaluateOne("jss-budget", "1.0.0")

	if result.Promoted {
		t.Error("should NOT promote (accuracy 0.50 < 0.80)")
	}
	if result.BlockedReason == "" {
		t.Error("should have a blocked reason")
	}
}

func TestPromoter_ShadowToActive_Pass(t *testing.T) {
	reg := promotionRegistry()
	p := NewPolicyPromoter(DefaultPromotionRules(), reg)

	result := p.EvaluateOne("cvrp-stag", "2.0.0")

	if !result.Promoted {
		t.Errorf("should promote (shadow accuracy 0.78, 30 runs, regret -0.5), blocked: %s", result.BlockedReason)
	}
	if result.ToStatus != PolicyStatusActive {
		t.Errorf("ToStatus = %q, want active", result.ToStatus)
	}
}

func TestPromoter_ShadowToActive_BlockedAccuracy(t *testing.T) {
	reg := promotionRegistry()
	p := NewPolicyPromoter(DefaultPromotionRules(), reg)

	result := p.EvaluateOne("vrptw-restart", "1.0.0")

	if result.Promoted {
		t.Error("should NOT promote (shadow accuracy 0.55 < 0.60)")
	}
}

func TestPromoter_BlocksOnDrift(t *testing.T) {
	reg := &PolicyLifecycleRegistry{
		Versions: []PolicyVersionRecord{
			{
				ID: "drifty", Version: "1.0.0", Domain: "cvrp", DecisionType: "search",
				Status: PolicyStatusTraining, OfflineAccuracy: 0.80,
				DriftDetected: true,
			},
		},
	}
	p := NewPolicyPromoter(DefaultPromotionRules(), reg)

	result := p.EvaluateOne("drifty", "1.0.0")

	if result.Promoted {
		t.Error("should NOT promote when drift detected")
	}
}

func TestPromoter_EvaluateAll(t *testing.T) {
	reg := promotionRegistry()
	p := NewPolicyPromoter(DefaultPromotionRules(), reg)

	results := p.EvaluateAll()

	// 4 versions: 2 training, 2 shadow.
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}

	promoted := 0
	for _, r := range results {
		if r.Promoted {
			promoted++
		}
	}
	// cvrp-stag 1.0.0 (training→shadow): passes (0.85 >= 0.80)
	// cvrp-stag 2.0.0 (shadow→active): passes (0.78 >= 0.60, 30 >= 20, -0.5 <= 0)
	// jss-budget 1.0.0 (training): fails (0.50 < 0.80)
	// vrptw-restart 1.0.0 (shadow): fails (0.55 < 0.60)
	if promoted != 2 {
		t.Errorf("expected 2 promotions, got %d", promoted)
	}
}

func TestPromoter_Rollback(t *testing.T) {
	reg := &PolicyLifecycleRegistry{
		Versions: []PolicyVersionRecord{
			{ID: "x", Version: "1.0.0", Domain: "cvrp", DecisionType: "search", Status: PolicyStatusRetired},
			{ID: "x", Version: "2.0.0", Domain: "cvrp", DecisionType: "search", Status: PolicyStatusActive},
		},
	}
	p := NewPolicyPromoter(DefaultPromotionRules(), reg)

	result, err := p.Rollback("cvrp", "search", "1.0.0", "regression")
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	if !result.Promoted {
		t.Error("rollback should succeed")
	}

	v1 := reg.FindVersion("x", "1.0.0")
	if v1.Status != PolicyStatusActive {
		t.Errorf("v1 status = %q, want active", v1.Status)
	}
	v2 := reg.FindVersion("x", "2.0.0")
	if v2.Status != PolicyStatusRetired {
		t.Errorf("v2 status = %q, want retired", v2.Status)
	}
}

func TestPromoter_History(t *testing.T) {
	reg := promotionRegistry()
	p := NewPolicyPromoter(DefaultPromotionRules(), reg)

	p.EvaluateAll()

	history := p.History()
	if len(history) != 4 {
		t.Errorf("history length = %d, want 4", len(history))
	}
}

func TestPromoter_NeverPromotesFailed(t *testing.T) {
	reg := &PolicyLifecycleRegistry{
		Versions: []PolicyVersionRecord{
			{
				ID: "bad", Version: "1.0.0", Domain: "nrp", DecisionType: "worker",
				Status: PolicyStatusTraining, OfflineAccuracy: 0.30, // terrible
			},
		},
	}
	rules := DefaultPromotionRules()
	p := NewPolicyPromoter(rules, reg)

	result := p.EvaluateOne("bad", "1.0.0")
	if result.Promoted {
		t.Error("should NEVER promote a policy with 0.30 accuracy")
	}
}
