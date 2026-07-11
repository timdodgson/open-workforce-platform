package optimisation

import "testing"

func TestPortfolioBanditMultiplier(t *testing.T) {
	bandit := &BanditPolicy{
		Status: "trained",
		Entries: []BanditEntry{
			{
				Domain: "cvrp", Instance: "a32k5", Strategy: "sa",
				BestArm: "1.25", BestValue: 1.25, Confidence: 0.8, Samples: 20,
			},
		},
	}
	mult, ok, reason := PortfolioBanditMultiplier(bandit, "cvrp", "a32k5", "sa")
	if !ok {
		t.Fatalf("expected bandit hit, got %s", reason)
	}
	if mult != 1.25 {
		t.Fatalf("expected 1.25, got %v", mult)
	}
}

func TestWorkerBanditDecision(t *testing.T) {
	bandit := &BanditPolicy{
		Status: "trained",
		Entries: []BanditEntry{
			{
				Context: "week=1|depth=2|dist=near",
				BestArm: "boost", BestValue: 1.25, Confidence: 0.75, Samples: 30,
			},
		},
	}
	ctx := WorkerContextKey(1, 2, 10)
	arm, mult, ok := WorkerBanditDecision(bandit, ctx)
	if !ok || arm != "boost" || mult != 1.25 {
		t.Fatalf("unexpected bandit decision: %s %v %v", arm, mult, ok)
	}
}

func TestPortfolioBanditInstanceFallback(t *testing.T) {
	bandit := &BanditPolicy{
		Status: "trained",
		Entries: []BanditEntry{
			{
				Domain: "cvrp", Strategy: "sa",
				BestArm: "1.0", BestValue: 1.0, Confidence: 0.7, Samples: 15,
			},
		},
	}
	mult, ok, _ := PortfolioBanditMultiplier(bandit, "cvrp", "unknown", "sa")
	if !ok || mult != 1.0 {
		t.Fatalf("expected domain fallback, got %v %v", mult, ok)
	}
}
