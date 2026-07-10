package policy

import (
	"path/filepath"
	"testing"
	"time"
)

func testRegistry() *PolicyLifecycleRegistry {
	now := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	return &PolicyLifecycleRegistry{
		Versions: []PolicyVersionRecord{
			{
				ID: "cvrp-budget", Version: "1.0.0", Domain: "cvrp", DecisionType: "portfolio",
				Algorithm: "*", Status: PolicyStatusActive, CreatedAt: now,
				TrainingSamples: 100, TrainingDate: now, Features: []string{"win_rate", "improvement_rate"},
				OfflineAccuracy: 0.75, ShadowAccuracy: 0.72, ProductionAccuracy: 0.70,
				ProductionRuns: 50, ModelPath: "policies/cvrp/portfolio_v1.json",
			},
			{
				ID: "cvrp-budget", Version: "2.0.0", Domain: "cvrp", DecisionType: "portfolio",
				Algorithm: "*", Status: PolicyStatusShadow, CreatedAt: now.Add(24 * time.Hour),
				TrainingSamples: 240, TrainingDate: now.Add(24 * time.Hour),
				Features:        []string{"win_rate", "improvement_rate", "plateau_length", "acceptance_rate"},
				OfflineAccuracy: 0.82, ShadowAccuracy: 0.79, ProductionAccuracy: -1,
				ModelPath: "policies/cvrp/portfolio_v2.json",
			},
			{
				ID: "jss-stagnation", Version: "1.0.0", Domain: "jss", DecisionType: "search",
				Algorithm: "tabu", Status: PolicyStatusTraining, CreatedAt: now,
				TrainingSamples: 30, TrainingDate: now, Features: []string{"plateau_length", "budget_consumed"},
				OfflineAccuracy: 0.68, ShadowAccuracy: -1, ProductionAccuracy: -1,
			},
		},
	}
}

func TestRegistry_ActiveVersion(t *testing.T) {
	r := testRegistry()

	active := r.ActiveVersion("cvrp", "portfolio")
	if active == nil {
		t.Fatal("expected active version for cvrp/portfolio")
	}
	if active.Version != "1.0.0" {
		t.Errorf("active version = %q, want 1.0.0", active.Version)
	}

	// No active for jss/search.
	if r.ActiveVersion("jss", "search") != nil {
		t.Error("expected no active version for jss/search")
	}
}

func TestRegistry_VersionHistory(t *testing.T) {
	r := testRegistry()
	history := r.VersionHistory("cvrp", "portfolio")

	if len(history) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(history))
	}
	if history[0].Version != "1.0.0" {
		t.Errorf("first version = %q, want 1.0.0", history[0].Version)
	}
	if history[1].Version != "2.0.0" {
		t.Errorf("second version = %q, want 2.0.0", history[1].Version)
	}
}

func TestRegistry_PromoteToShadow(t *testing.T) {
	r := testRegistry()
	err := r.PromoteToShadow("jss-stagnation", "1.0.0")
	if err != nil {
		t.Fatalf("PromoteToShadow failed: %v", err)
	}

	v := r.FindVersion("jss-stagnation", "1.0.0")
	if v.Status != PolicyStatusShadow {
		t.Errorf("status = %q, want shadow", v.Status)
	}
}

func TestRegistry_PromoteToShadow_WrongStatus(t *testing.T) {
	r := testRegistry()
	// v2 is already shadow, can't promote to shadow again.
	err := r.PromoteToShadow("cvrp-budget", "2.0.0")
	if err == nil {
		t.Error("expected error promoting shadow → shadow")
	}
}

func TestRegistry_PromoteToActive(t *testing.T) {
	r := testRegistry()
	err := r.PromoteToActive("cvrp-budget", "2.0.0")
	if err != nil {
		t.Fatalf("PromoteToActive failed: %v", err)
	}

	// v2 should be active.
	v2 := r.FindVersion("cvrp-budget", "2.0.0")
	if v2.Status != PolicyStatusActive {
		t.Errorf("v2 status = %q, want active", v2.Status)
	}
	if v2.PromotedAt == nil {
		t.Error("v2 PromotedAt should be set")
	}

	// v1 should be retired.
	v1 := r.FindVersion("cvrp-budget", "1.0.0")
	if v1.Status != PolicyStatusRetired {
		t.Errorf("v1 status = %q, want retired", v1.Status)
	}
	if v1.RetiredAt == nil {
		t.Error("v1 RetiredAt should be set")
	}
}

func TestRegistry_Rollback(t *testing.T) {
	r := testRegistry()

	// First promote v2 to active.
	r.PromoteToActive("cvrp-budget", "2.0.0")

	// Now rollback to v1.
	err := r.Rollback("cvrp", "portfolio", "1.0.0", "drift_detected")
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// v1 should be active again.
	v1 := r.FindVersion("cvrp-budget", "1.0.0")
	if v1.Status != PolicyStatusActive {
		t.Errorf("v1 status = %q, want active after rollback", v1.Status)
	}
	if v1.RolledBackFrom != "2.0.0" {
		t.Errorf("RolledBackFrom = %q, want 2.0.0", v1.RolledBackFrom)
	}
	if v1.RollbackReason != "drift_detected" {
		t.Errorf("RollbackReason = %q, want drift_detected", v1.RollbackReason)
	}

	// v2 should be retired.
	v2 := r.FindVersion("cvrp-budget", "2.0.0")
	if v2.Status != PolicyStatusRetired {
		t.Errorf("v2 status = %q, want retired after rollback", v2.Status)
	}
}

func TestRegistry_Compare(t *testing.T) {
	r := testRegistry()

	comp, err := r.Compare("cvrp-budget", "1.0.0", "cvrp-budget", "2.0.0")
	if err != nil {
		t.Fatalf("Compare failed: %v", err)
	}

	// B (v2) has shadow accuracy 0.79, A (v1) has production 0.70.
	// AccuracyDelta = 0.79 - 0.70 = 0.09.
	if comp.AccuracyDelta < 0.08 || comp.AccuracyDelta > 0.10 {
		t.Errorf("AccuracyDelta = %f, want ~0.09", comp.AccuracyDelta)
	}
	if comp.TrainingSampleDelta != 140 {
		t.Errorf("TrainingSampleDelta = %d, want 140", comp.TrainingSampleDelta)
	}
	// v2 has no production runs yet → needs_more_data.
	if comp.Recommendation != "needs_more_data" {
		t.Errorf("Recommendation = %q, want needs_more_data", comp.Recommendation)
	}
}

func TestRegistry_SaveAndLoad(t *testing.T) {
	r := testRegistry()
	dir := t.TempDir()
	path := filepath.Join(dir, "policy_registry.json")

	if err := r.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadPolicyRegistry(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(loaded.Versions) != 3 {
		t.Errorf("loaded %d versions, want 3", len(loaded.Versions))
	}
	if loaded.Versions[0].ID != "cvrp-budget" {
		t.Errorf("first version ID = %q, want cvrp-budget", loaded.Versions[0].ID)
	}
}

func TestRegistry_LoadNonexistent(t *testing.T) {
	r, err := LoadPolicyRegistry("/nonexistent/path.json")
	if err != nil {
		t.Fatalf("should not error on missing file: %v", err)
	}
	if len(r.Versions) != 0 {
		t.Errorf("expected empty registry, got %d versions", len(r.Versions))
	}
}

func TestRegistry_Register(t *testing.T) {
	r := &PolicyLifecycleRegistry{}
	r.Register(PolicyVersionRecord{
		ID: "new-policy", Version: "1.0.0", Domain: "vrptw", DecisionType: "restart",
		TrainingSamples: 50, OfflineAccuracy: 0.72,
	})

	if len(r.Versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(r.Versions))
	}
	if r.Versions[0].Status != PolicyStatusTraining {
		t.Errorf("status = %q, want training", r.Versions[0].Status)
	}
	if r.Versions[0].ShadowAccuracy != -1 {
		t.Errorf("ShadowAccuracy = %f, want -1", r.Versions[0].ShadowAccuracy)
	}
}

func TestRegistry_AllActive(t *testing.T) {
	r := testRegistry()
	active := r.AllActive()
	if len(active) != 1 {
		t.Errorf("expected 1 active, got %d", len(active))
	}
}
