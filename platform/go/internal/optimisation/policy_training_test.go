package optimisation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempPipeline(t *testing.T) (*TrainingPipeline, string) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "policy_registry.json")
	config := DefaultTrainingPipelineConfig()
	config.RegistryPath = regPath
	config.OutputDir = dir

	p, err := NewTrainingPipeline(config)
	if err != nil {
		t.Fatalf("NewTrainingPipeline failed: %v", err)
	}
	return p, dir
}

func TestPipeline_ValidateTrainingData_Pass(t *testing.T) {
	p, _ := tempPipeline(t)
	dataset := TrainingDataset{
		Samples:  make([]TrainingSample, 100),
		Features: []string{"plateau_length", "improvement_rate"},
	}
	gate := p.ValidateTrainingData(dataset)
	if !gate.Passed {
		t.Errorf("expected pass, got: %s", gate.Reason)
	}
}

func TestPipeline_ValidateTrainingData_InsufficientSamples(t *testing.T) {
	p, _ := tempPipeline(t)
	dataset := TrainingDataset{
		Samples:  make([]TrainingSample, 10),
		Features: []string{"plateau_length"},
	}
	gate := p.ValidateTrainingData(dataset)
	if gate.Passed {
		t.Error("should fail with insufficient samples")
	}
	if gate.Gate != "min_samples" {
		t.Errorf("Gate = %q, want min_samples", gate.Gate)
	}
}

func TestPipeline_ValidateOfflineAccuracy_Pass(t *testing.T) {
	p, _ := tempPipeline(t)
	result := TrainingResult{OfflineAccuracy: 0.75}
	gate := p.ValidateOfflineAccuracy(result)
	if !gate.Passed {
		t.Errorf("expected pass, got: %s", gate.Reason)
	}
}

func TestPipeline_ValidateOfflineAccuracy_Fail(t *testing.T) {
	p, _ := tempPipeline(t)
	result := TrainingResult{OfflineAccuracy: 0.50}
	gate := p.ValidateOfflineAccuracy(result)
	if gate.Passed {
		t.Error("should fail below threshold")
	}
}

func TestPipeline_ValidateShadow_Pass(t *testing.T) {
	p, _ := tempPipeline(t)
	gate := p.ValidateShadowPerformance(0.72, 30)
	if !gate.Passed {
		t.Errorf("expected pass, got: %s", gate.Reason)
	}
}

func TestPipeline_ValidateShadow_InsufficientRuns(t *testing.T) {
	p, _ := tempPipeline(t)
	gate := p.ValidateShadowPerformance(0.80, 5)
	if gate.Passed {
		t.Error("should fail with insufficient runs")
	}
}

func TestPipeline_FullLifecycle(t *testing.T) {
	p, dir := tempPipeline(t)

	// 1. Register candidate.
	result := TrainingResult{
		PolicyID: "cvrp-budget", Version: "3.0.0", Domain: "cvrp",
		DecisionType: "portfolio", Algorithm: "*",
		TrainingSamples: 200, Features: []string{"win_rate", "plateau_length"},
		TrainedAt: time.Now(), OfflineAccuracy: 0.78,
		ValidationSamples: 50, ModelPath: filepath.Join(dir, "model.json"),
		PassedValidation: true,
	}
	if err := p.RegisterCandidate(result); err != nil {
		t.Fatalf("RegisterCandidate failed: %v", err)
	}

	// Verify registered.
	v := p.Registry().FindVersion("cvrp-budget", "3.0.0")
	if v == nil {
		t.Fatal("version not found after register")
	}
	if v.Status != PolicyStatusTraining {
		t.Errorf("status = %q, want training", v.Status)
	}

	// 2. Deploy to shadow.
	if err := p.DeployToShadow("cvrp-budget", "3.0.0"); err != nil {
		t.Fatalf("DeployToShadow failed: %v", err)
	}
	v = p.Registry().FindVersion("cvrp-budget", "3.0.0")
	if v.Status != PolicyStatusShadow {
		t.Errorf("status = %q, want shadow", v.Status)
	}

	// 3. Promote to production.
	gate, err := p.PromoteToProduction("cvrp-budget", "3.0.0", 0.74, 25)
	if err != nil {
		t.Fatalf("PromoteToProduction failed: %v", err)
	}
	if !gate.Passed {
		t.Errorf("promotion gate should pass, got: %s", gate.Reason)
	}
	v = p.Registry().FindVersion("cvrp-budget", "3.0.0")
	if v.Status != PolicyStatusActive {
		t.Errorf("status = %q, want active", v.Status)
	}

	// 4. Verify persistence.
	regPath := filepath.Join(dir, "policy_registry.json")
	loaded, err := LoadPolicyRegistry(regPath)
	if err != nil {
		t.Fatalf("LoadPolicyRegistry failed: %v", err)
	}
	lv := loaded.FindVersion("cvrp-budget", "3.0.0")
	if lv == nil || lv.Status != PolicyStatusActive {
		t.Error("persisted registry should show active status")
	}
}

func TestPipeline_PromotionFailsWithLowAccuracy(t *testing.T) {
	p, _ := tempPipeline(t)

	result := TrainingResult{
		PolicyID: "test", Version: "1.0.0", Domain: "jss",
		DecisionType: "search", TrainingSamples: 100,
		Features: []string{"plateau_length"}, TrainedAt: time.Now(),
		OfflineAccuracy: 0.70, PassedValidation: true,
	}
	p.RegisterCandidate(result)
	p.DeployToShadow("test", "1.0.0")

	// Shadow accuracy too low.
	gate, err := p.PromoteToProduction("test", "1.0.0", 0.50, 25)
	if err == nil {
		t.Error("expected promotion to fail with low shadow accuracy")
	}
	if gate.Passed {
		t.Error("gate should not pass")
	}
}

func TestPipeline_Rollback(t *testing.T) {
	p, _ := tempPipeline(t)

	// Register and promote v1.
	p.RegisterCandidate(TrainingResult{
		PolicyID: "x", Version: "1.0.0", Domain: "vrptw", DecisionType: "restart",
		TrainingSamples: 60, Features: []string{"a"}, TrainedAt: time.Now(), OfflineAccuracy: 0.70,
	})
	p.DeployToShadow("x", "1.0.0")
	p.PromoteToProduction("x", "1.0.0", 0.68, 25)

	// Register and promote v2.
	p.RegisterCandidate(TrainingResult{
		PolicyID: "x", Version: "2.0.0", Domain: "vrptw", DecisionType: "restart",
		TrainingSamples: 120, Features: []string{"a", "b"}, TrainedAt: time.Now(), OfflineAccuracy: 0.80,
	})
	p.DeployToShadow("x", "2.0.0")
	p.PromoteToProduction("x", "2.0.0", 0.75, 30)

	// Rollback to v1.
	if err := p.Rollback("vrptw", "restart", "1.0.0", "regression_detected"); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	v1 := p.Registry().FindVersion("x", "1.0.0")
	if v1.Status != PolicyStatusActive {
		t.Errorf("v1 status = %q, want active after rollback", v1.Status)
	}
}

func TestPipeline_GenerateReport(t *testing.T) {
	p, dir := tempPipeline(t)

	result := TrainingResult{
		PolicyID: "report-test", Version: "1.0.0", Domain: "cvrp",
		DecisionType: "portfolio", TrainingSamples: 150, ValidationSamples: 40,
		Features: []string{"win_rate", "plateau_length"}, TrainedAt: time.Now(),
		OfflineAccuracy: 0.76, PassedValidation: true,
	}

	report := p.GenerateReport(result, 0.72, 25)

	if report.FinalStatus != "promoted" {
		t.Errorf("FinalStatus = %q, want promoted", report.FinalStatus)
	}
	if report.OfflineAccuracy != 0.76 {
		t.Errorf("OfflineAccuracy = %f, want 0.76", report.OfflineAccuracy)
	}
	if len(report.Gates) != 3 {
		t.Errorf("expected 3 gates, got %d", len(report.Gates))
	}

	// Save and verify.
	if err := SaveReport(report, dir); err != nil {
		t.Fatalf("SaveReport failed: %v", err)
	}
	files, _ := os.ReadDir(dir)
	found := false
	for _, f := range files {
		if f.Name() == "policy_report_report-test_1.0.0.json" {
			found = true
			break
		}
	}
	if !found {
		t.Error("report file not written")
	}

	// Verify JSON is valid.
	data, _ := os.ReadFile(filepath.Join(dir, "policy_report_report-test_1.0.0.json"))
	var loaded PolicyReport
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("report JSON invalid: %v", err)
	}
	if loaded.PolicyID != "report-test" {
		t.Errorf("loaded PolicyID = %q", loaded.PolicyID)
	}
}

func TestPipeline_ReportRejected(t *testing.T) {
	p, _ := tempPipeline(t)

	result := TrainingResult{
		PolicyID: "bad", Version: "1.0.0", Domain: "nrp",
		DecisionType: "worker", TrainingSamples: 20, // below threshold
		Features: []string{"entropy"}, TrainedAt: time.Now(),
		OfflineAccuracy: 0.50,
	}

	report := p.GenerateReport(result, 0, 0)

	if report.FinalStatus != "rejected" {
		t.Errorf("FinalStatus = %q, want rejected", report.FinalStatus)
	}
}
