package optimisation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureSchemaVersion(t *testing.T) {
	if FeatureSchemaVersion == "" {
		t.Fatal("FeatureSchemaVersion must not be empty")
	}
}

func TestFeatureExtractor_FromWorkerContext(t *testing.T) {
	fe := NewFeatureExtractor()
	ctx := WorkerContext{
		Algorithm:        "sa",
		Week:             3,
		Depth:            2,
		ParentObjective:  5000,
		GlobalBest:       4500,
		DistanceFromBest: 500,
		BeamRank:         1,
		Entropy:          0.85,
		BeamHealth:       72.0,
		RecentImprovRate: 2.5,
		AllocatedIters:   500000,
		WorkerCount:      8,
	}

	fv := fe.FromWorkerContext(ctx, "run-001", "n012w8")

	if fv.SchemaVersion != FeatureSchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", fv.SchemaVersion, FeatureSchemaVersion)
	}
	if fv.Problem != "nrp" {
		t.Errorf("Problem = %q, want nrp", fv.Problem)
	}
	if fv.Instance != "n012w8" {
		t.Errorf("Instance = %q, want n012w8", fv.Instance)
	}
	if fv.Algorithm != "sa" {
		t.Errorf("Algorithm = %q, want sa", fv.Algorithm)
	}
	if fv.DecisionType != "worker" {
		t.Errorf("DecisionType = %q, want worker", fv.DecisionType)
	}
	if fv.Week != 3 {
		t.Errorf("Week = %d, want 3", fv.Week)
	}
	if fv.BranchDepth != 2 {
		t.Errorf("BranchDepth = %d, want 2", fv.BranchDepth)
	}
	if fv.DistanceFromBest != 500 {
		t.Errorf("DistanceFromBest = %d, want 500", fv.DistanceFromBest)
	}
	if fv.Entropy != 0.85 {
		t.Errorf("Entropy = %f, want 0.85", fv.Entropy)
	}
	if fv.WorkerCount != 8 {
		t.Errorf("WorkerCount = %d, want 8", fv.WorkerCount)
	}
}

func TestFeatureExtractor_FromSearchProgress(t *testing.T) {
	fe := NewFeatureExtractor()
	p := SearchProgress{
		Algorithm:          "tabu",
		IterationsComplete: 50000,
		IterationsTotal:    100000,
		CurrentPenalty:     700,
		BestPenalty:        666,
		InitialPenalty:     1200,
		ImprovementRate:    0.3,
		Temperature:        0,
		PlateauLength:      12000,
		Accepted:           800,
		Rejected:           49200,
		CandidatesEval:     50000,
	}

	fv := fe.FromSearchProgress(p, "run-002", "jss", "la01", 4500)

	if fv.BudgetConsumed != 0.5 {
		t.Errorf("BudgetConsumed = %f, want 0.5", fv.BudgetConsumed)
	}
	if fv.PlateauLength != 12000 {
		t.Errorf("PlateauLength = %d, want 12000", fv.PlateauLength)
	}
	if fv.DecisionType != "search" {
		t.Errorf("DecisionType = %q, want search", fv.DecisionType)
	}
	if fv.Problem != "jss" {
		t.Errorf("Problem = %q, want jss", fv.Problem)
	}
	// Acceptance rate = 800/50000 = 0.016
	if fv.AcceptanceRate < 0.015 || fv.AcceptanceRate > 0.017 {
		t.Errorf("AcceptanceRate = %f, want ~0.016", fv.AcceptanceRate)
	}
}

func TestFeatureExtractor_FromPortfolioContext(t *testing.T) {
	fe := NewFeatureExtractor()
	ctx := PortfolioContext{
		Strategies:  []string{"sa", "lahc", "tabu"},
		TotalBudget: 300000,
		ProblemType: "cvrp",
		Instance:    "A-n32-k5",
		PreviousResults: []PortfolioHistoryEntry{
			{Strategy: "sa", Improved: true},
			{Strategy: "sa", Improved: true},
			{Strategy: "sa", Improved: false},
			{Strategy: "lahc", Improved: true},
		},
	}

	fv := fe.FromPortfolioContext(ctx, "sa", "run-003")

	if fv.Problem != "cvrp" {
		t.Errorf("Problem = %q, want cvrp", fv.Problem)
	}
	if fv.DecisionType != "portfolio" {
		t.Errorf("DecisionType = %q, want portfolio", fv.DecisionType)
	}
	if fv.WorkerCount != 3 {
		t.Errorf("WorkerCount = %d, want 3 (strategy count)", fv.WorkerCount)
	}
	// SA win rate = 2/3 ≈ 0.667
	if fv.GapToReference < 0.66 || fv.GapToReference > 0.67 {
		t.Errorf("GapToReference (win rate) = %f, want ~0.667", fv.GapToReference)
	}
}

func TestFeatureStore_Disabled(t *testing.T) {
	fs := NewFeatureStore("")
	err := fs.Record(FeatureRecord{})
	if err != nil {
		t.Errorf("disabled store should not error, got: %v", err)
	}
	if fs.Count() != 0 {
		t.Errorf("disabled store count should be 0, got %d", fs.Count())
	}
}

func TestFeatureStore_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	fs := NewFeatureStore(dir)
	defer fs.Close()

	fe := NewFeatureExtractor()
	fv := fe.FromSearchProgress(SearchProgress{
		Algorithm:          "sa",
		IterationsComplete: 100000,
		IterationsTotal:    500000,
		CurrentPenalty:     800,
		BestPenalty:        784,
		InitialPenalty:     1200,
		PlateauLength:      5000,
		CandidatesEval:     100000,
		Accepted:           2000,
	}, "run-test", "cvrp", "A-n32-k5", 78)

	record := FeatureRecord{
		Features:     fv,
		Action:       "early_stop",
		Confidence:   0.82,
		PolicySource: "learned",
		Outcome: FeatureOutcome{
			Improved:       false,
			FinalObjective: 784,
			ComputeUsed:    100000,
			RuntimeMs:      78,
		},
	}

	if err := fs.Record(record); err != nil {
		t.Fatalf("Record failed: %v", err)
	}
	if fs.Count() != 1 {
		t.Errorf("Count = %d, want 1", fs.Count())
	}

	// Read back and verify.
	fs.Close()
	data, err := os.ReadFile(filepath.Join(dir, "features.jsonl"))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	var decoded FeatureRecord
	if err := json.Unmarshal([]byte(lines[0]), &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Features.SchemaVersion != FeatureSchemaVersion {
		t.Errorf("decoded SchemaVersion = %q", decoded.Features.SchemaVersion)
	}
	if decoded.Action != "early_stop" {
		t.Errorf("decoded Action = %q, want early_stop", decoded.Action)
	}
	if decoded.Confidence != 0.82 {
		t.Errorf("decoded Confidence = %f, want 0.82", decoded.Confidence)
	}
	if decoded.Features.Problem != "cvrp" {
		t.Errorf("decoded Problem = %q, want cvrp", decoded.Features.Problem)
	}
}
