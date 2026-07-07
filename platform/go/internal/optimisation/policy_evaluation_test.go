package optimisation

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPolicyEvaluator_EmptyMetrics(t *testing.T) {
	pe := NewPolicyEvaluator()
	m := pe.Metrics()

	if m.TotalDecisions != 0 {
		t.Errorf("TotalDecisions = %d, want 0", m.TotalDecisions)
	}
	if m.Accuracy != 0 {
		t.Errorf("Accuracy = %f, want 0", m.Accuracy)
	}
}

func TestPolicyEvaluator_SingleCorrectDecision(t *testing.T) {
	pe := NewPolicyEvaluator()
	pe.Record(PolicyEvaluationRecord{
		Timestamp:           time.Now(),
		RunID:               "test-1",
		DecisionType:        "search",
		Domain:              "cvrp",
		PolicyID:            "cvrp-search-v2",
		PolicyVersion:       "2.0.0",
		PolicyType:          "learned",
		Action:              "early_stop",
		Confidence:          0.85,
		ExpectedImprovement: 0.0,
		ActualImprovement:   0.0,
		Correct:             true,
		Regret:              0.0,
	})

	m := pe.Metrics()

	if m.TotalDecisions != 1 {
		t.Errorf("TotalDecisions = %d, want 1", m.TotalDecisions)
	}
	if m.Accuracy != 1.0 {
		t.Errorf("Accuracy = %f, want 1.0", m.Accuracy)
	}
}

func TestPolicyEvaluator_MixedDecisions(t *testing.T) {
	pe := NewPolicyEvaluator()

	// 7 correct, 3 incorrect.
	for i := 0; i < 7; i++ {
		pe.Record(PolicyEvaluationRecord{
			Confidence:          0.75,
			ExpectedImprovement: 10.0,
			ActualImprovement:   9.0,
			Correct:             true,
			Regret:              0.0,
		})
	}
	for i := 0; i < 3; i++ {
		pe.Record(PolicyEvaluationRecord{
			Confidence:          0.75,
			ExpectedImprovement: 10.0,
			ActualImprovement:   0.0,
			Correct:             false,
			Regret:              5.0,
		})
	}

	m := pe.Metrics()

	if m.TotalDecisions != 10 {
		t.Errorf("TotalDecisions = %d, want 10", m.TotalDecisions)
	}
	if m.Accuracy != 0.7 {
		t.Errorf("Accuracy = %f, want 0.7", m.Accuracy)
	}
	if m.TotalRegret != 15.0 {
		t.Errorf("TotalRegret = %f, want 15.0", m.TotalRegret)
	}
	if m.MeanRegret != 1.5 {
		t.Errorf("MeanRegret = %f, want 1.5", m.MeanRegret)
	}
}

func TestPolicyEvaluator_PredictionError(t *testing.T) {
	pe := NewPolicyEvaluator()
	pe.Record(PolicyEvaluationRecord{
		ExpectedImprovement: 50.0,
		ActualImprovement:   30.0,
		Correct:             true,
	})

	recs := pe.Records()
	if len(recs) != 1 {
		t.Fatalf("expected 1 record, got %d", len(recs))
	}

	// Prediction error = expected - actual = 50 - 30 = 20.
	if recs[0].PredictionError != 20.0 {
		t.Errorf("PredictionError = %f, want 20.0", recs[0].PredictionError)
	}
	if recs[0].AbsoluteError != 20.0 {
		t.Errorf("AbsoluteError = %f, want 20.0", recs[0].AbsoluteError)
	}
	if recs[0].SquaredError != 400.0 {
		t.Errorf("SquaredError = %f, want 400.0", recs[0].SquaredError)
	}
}

func TestPolicyEvaluator_RMSE(t *testing.T) {
	pe := NewPolicyEvaluator()
	// Two decisions: errors of 3 and 4.
	pe.Record(PolicyEvaluationRecord{ExpectedImprovement: 10, ActualImprovement: 7, Correct: true})
	pe.Record(PolicyEvaluationRecord{ExpectedImprovement: 10, ActualImprovement: 6, Correct: true})

	m := pe.Metrics()
	// RMSE = sqrt((9 + 16) / 2) = sqrt(12.5) ≈ 3.536
	if m.RMSE < 3.5 || m.RMSE > 3.6 {
		t.Errorf("RMSE = %f, want ~3.536", m.RMSE)
	}
}

func TestPolicyEvaluator_Calibration(t *testing.T) {
	pe := NewPolicyEvaluator()

	// 10 decisions at confidence 0.75 (bucket 7), 7 correct.
	for i := 0; i < 10; i++ {
		pe.Record(PolicyEvaluationRecord{
			Confidence: 0.75,
			Correct:    i < 7,
		})
	}

	m := pe.Metrics()

	// Bucket 7 (0.70–0.80) should have 10 total, 7 correct.
	b := m.Calibration[7]
	if b.Total != 10 {
		t.Errorf("bucket 7 total = %d, want 10", b.Total)
	}
	if b.Correct != 7 {
		t.Errorf("bucket 7 correct = %d, want 7", b.Correct)
	}
	if b.ActualAccuracy != 0.7 {
		t.Errorf("bucket 7 actual accuracy = %f, want 0.7", b.ActualAccuracy)
	}
	// Expected midpoint = 0.75. Calibration error = |0.7 - 0.75| = 0.05.
	if b.CalibrationError < 0.04 || b.CalibrationError > 0.06 {
		t.Errorf("bucket 7 calibration error = %f, want ~0.05", b.CalibrationError)
	}
}

func TestPolicyEvaluator_Drift(t *testing.T) {
	pe := NewPolicyEvaluator()

	// First 50 decisions: all correct (accuracy = 1.0).
	for i := 0; i < 50; i++ {
		pe.Record(PolicyEvaluationRecord{Confidence: 0.9, Correct: true})
	}

	// Next 50 decisions: all wrong (recent accuracy = 0.0).
	for i := 0; i < 50; i++ {
		pe.Record(PolicyEvaluationRecord{Confidence: 0.9, Correct: false})
	}

	m := pe.Metrics()

	// Overall: 50/100 = 0.5.
	if m.OverallAccuracy != 0.5 {
		t.Errorf("OverallAccuracy = %f, want 0.5", m.OverallAccuracy)
	}
	// Recent (last 50): 0/50 = 0.0.
	if m.RecentAccuracy != 0.0 {
		t.Errorf("RecentAccuracy = %f, want 0.0", m.RecentAccuracy)
	}
	// Drift = |0.0 - 0.5| = 0.5.
	if m.DriftMagnitude != 0.5 {
		t.Errorf("DriftMagnitude = %f, want 0.5", m.DriftMagnitude)
	}
}

func TestPolicyEvaluator_WriteCSV(t *testing.T) {
	pe := NewPolicyEvaluator()
	pe.Record(PolicyEvaluationRecord{
		Timestamp:           time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC),
		RunID:               "csv-test",
		DecisionType:        "portfolio",
		Domain:              "jss",
		Instance:            "la01",
		Algorithm:           "tabu",
		PolicyID:            "jss-portfolio-v1",
		PolicyVersion:       "1.0.0",
		PolicyType:          "hybrid",
		Action:              "allocate",
		Confidence:          0.72,
		ExpectedImprovement: 20.0,
		ActualImprovement:   15.0,
		Correct:             true,
		Regret:              2.0,
	})

	dir := t.TempDir()
	if err := pe.WriteCSV(dir); err != nil {
		t.Fatalf("WriteCSV failed: %v", err)
	}

	f, err := os.Open(filepath.Join(dir, "policy_evaluation.csv"))
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	// Header + 1 data row.
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "timestamp" {
		t.Errorf("header[0] = %q, want timestamp", rows[0][0])
	}
	if rows[1][3] != "jss" {
		t.Errorf("row domain = %q, want jss", rows[1][3])
	}
	if rows[1][14] != "1" {
		t.Errorf("row correct = %q, want 1", rows[1][14])
	}
}

func TestPolicyEvaluator_Concurrent(t *testing.T) {
	pe := NewPolicyEvaluator()
	done := make(chan struct{}, 100)

	for i := 0; i < 100; i++ {
		go func(n int) {
			defer func() { done <- struct{}{} }()
			pe.Record(PolicyEvaluationRecord{
				Confidence:          float64(n%10) * 0.1,
				ExpectedImprovement: float64(n),
				ActualImprovement:   float64(n) * 0.9,
				Correct:             n%3 != 0,
				Regret:              float64(n % 5),
			})
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	m := pe.Metrics()
	if m.TotalDecisions != 100 {
		t.Errorf("TotalDecisions = %d, want 100", m.TotalDecisions)
	}
}
