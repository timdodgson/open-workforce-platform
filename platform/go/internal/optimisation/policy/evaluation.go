// policy_evaluation.go — Online Policy Evaluation.
//
// Every decision records expected vs actual outcome. This data is used to
// compute policy quality metrics in real time:
//   - Accuracy: fraction of decisions where outcome matched prediction
//   - Calibration: does 80% confidence mean 80% correct?
//   - Regret: cumulative loss vs rule-based baseline
//   - Drift: change in accuracy over time (detects degradation)
//
// The evaluator runs alongside the policy, requiring zero additional compute.
// Results are written to policy_evaluation.csv per run.
package policy

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// ───────────────────────────────────────────────────────────────
// Evaluation Record
// ───────────────────────────────────────────────────────────────

// PolicyEvaluationRecord captures one decision with prediction vs outcome.
type PolicyEvaluationRecord struct {
	// Identity
	Timestamp    time.Time
	RunID        string
	DecisionType string
	Domain       string
	Instance     string
	Algorithm    string

	// Prediction (at decision time)
	PolicyID            string
	PolicyVersion       string
	PolicyType          string // rule, learned, hybrid
	Action              string
	Confidence          float64
	ExpectedImprovement float64

	// Outcome (after execution)
	ActualImprovement float64
	PredictionError   float64 // expected - actual
	Correct           bool    // did the action produce expected outcome?

	// Evaluation metrics
	AbsoluteError float64
	SquaredError  float64
	Regret        float64 // loss vs best alternative
}

// ───────────────────────────────────────────────────────────────
// Policy Evaluator
// ───────────────────────────────────────────────────────────────

// PolicyEvaluator tracks policy quality metrics online.
// Thread-safe for concurrent decision recording.
type PolicyEvaluator struct {
	mu      sync.Mutex
	records []PolicyEvaluationRecord

	// Running metrics.
	totalDecisions int
	correctCount   int
	totalRegret    float64
	totalAbsError  float64
	totalSqError   float64

	// Calibration buckets (10 buckets for confidence 0.0–1.0).
	calibrationTotal   [10]int
	calibrationCorrect [10]int

	// Drift detection (sliding window).
	recentWindow  int
	recentCorrect []bool
}

// NewPolicyEvaluator creates an evaluator.
func NewPolicyEvaluator() *PolicyEvaluator {
	return &PolicyEvaluator{
		recentWindow:  50,
		recentCorrect: make([]bool, 0, 50),
	}
}

// Record adds an evaluation record and updates running metrics.
func (pe *PolicyEvaluator) Record(rec PolicyEvaluationRecord) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	// Compute derived fields.
	rec.PredictionError = rec.ExpectedImprovement - rec.ActualImprovement
	rec.AbsoluteError = math.Abs(rec.PredictionError)
	rec.SquaredError = rec.PredictionError * rec.PredictionError

	pe.records = append(pe.records, rec)
	pe.totalDecisions++
	pe.totalAbsError += rec.AbsoluteError
	pe.totalSqError += rec.SquaredError
	pe.totalRegret += rec.Regret

	if rec.Correct {
		pe.correctCount++
	}

	// Update calibration.
	bucket := int(rec.Confidence * 10)
	if bucket > 9 {
		bucket = 9
	}
	if bucket < 0 {
		bucket = 0
	}
	pe.calibrationTotal[bucket]++
	if rec.Correct {
		pe.calibrationCorrect[bucket]++
	}

	// Update drift window.
	pe.recentCorrect = append(pe.recentCorrect, rec.Correct)
	if len(pe.recentCorrect) > pe.recentWindow {
		pe.recentCorrect = pe.recentCorrect[1:]
	}
}

// ───────────────────────────────────────────────────────────────
// Metrics
// ───────────────────────────────────────────────────────────────

// PolicyMetrics holds computed quality metrics.
type PolicyMetrics struct {
	TotalDecisions int
	Accuracy       float64 // correct / total
	MeanAbsError   float64
	RMSE           float64 // root mean squared error
	TotalRegret    float64
	MeanRegret     float64

	// Calibration: per-bucket actual accuracy vs predicted confidence.
	Calibration [10]CalibrationBucket

	// Drift: recent accuracy vs overall accuracy.
	RecentAccuracy  float64
	OverallAccuracy float64
	DriftMagnitude  float64 // |recent - overall|, high = policy degrading
}

// CalibrationBucket holds one calibration bin.
type CalibrationBucket struct {
	ConfidenceRange  string // e.g. "0.50–0.60"
	Total            int
	Correct          int
	ActualAccuracy   float64 // correct / total
	ExpectedMidpoint float64 // midpoint of confidence range
	CalibrationError float64 // |actual - expected|
}

// Metrics computes current policy quality metrics.
func (pe *PolicyEvaluator) Metrics() PolicyMetrics {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	m := PolicyMetrics{
		TotalDecisions: pe.totalDecisions,
	}

	if pe.totalDecisions == 0 {
		return m
	}

	m.Accuracy = float64(pe.correctCount) / float64(pe.totalDecisions)
	m.OverallAccuracy = m.Accuracy
	m.MeanAbsError = pe.totalAbsError / float64(pe.totalDecisions)
	m.RMSE = math.Sqrt(pe.totalSqError / float64(pe.totalDecisions))
	m.TotalRegret = pe.totalRegret
	m.MeanRegret = pe.totalRegret / float64(pe.totalDecisions)

	// Calibration.
	for i := 0; i < 10; i++ {
		b := CalibrationBucket{
			ConfidenceRange:  fmt.Sprintf("%.1f–%.1f", float64(i)*0.1, float64(i+1)*0.1),
			Total:            pe.calibrationTotal[i],
			Correct:          pe.calibrationCorrect[i],
			ExpectedMidpoint: float64(i)*0.1 + 0.05,
		}
		if b.Total > 0 {
			b.ActualAccuracy = float64(b.Correct) / float64(b.Total)
			b.CalibrationError = math.Abs(b.ActualAccuracy - b.ExpectedMidpoint)
		}
		m.Calibration[i] = b
	}

	// Drift.
	if len(pe.recentCorrect) > 0 {
		recentCorrectCount := 0
		for _, c := range pe.recentCorrect {
			if c {
				recentCorrectCount++
			}
		}
		m.RecentAccuracy = float64(recentCorrectCount) / float64(len(pe.recentCorrect))
		m.DriftMagnitude = math.Abs(m.RecentAccuracy - m.OverallAccuracy)
	}

	return m
}

// Records returns all evaluation records.
func (pe *PolicyEvaluator) Records() []PolicyEvaluationRecord {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	return append([]PolicyEvaluationRecord{}, pe.records...)
}

// ───────────────────────────────────────────────────────────────
// CSV Output
// ───────────────────────────────────────────────────────────────

// WriteCSV writes all evaluation records to policy_evaluation.csv.
func (pe *PolicyEvaluator) WriteCSV(dir string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if len(pe.records) == 0 {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("policy_evaluation: mkdir error: %w", err)
	}

	path := filepath.Join(dir, "policy_evaluation.csv")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("policy_evaluation: create error: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{
		"timestamp", "run_id", "decision_type", "domain", "instance", "algorithm",
		"policy_id", "policy_version", "policy_type", "action", "confidence",
		"expected_improvement", "actual_improvement", "prediction_error",
		"correct", "absolute_error", "squared_error", "regret",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, r := range pe.records {
		correct := "0"
		if r.Correct {
			correct = "1"
		}
		row := []string{
			r.Timestamp.Format(time.RFC3339),
			r.RunID, r.DecisionType, r.Domain, r.Instance, r.Algorithm,
			r.PolicyID, r.PolicyVersion, r.PolicyType, r.Action,
			strconv.FormatFloat(r.Confidence, 'f', 4, 64),
			strconv.FormatFloat(r.ExpectedImprovement, 'f', 2, 64),
			strconv.FormatFloat(r.ActualImprovement, 'f', 2, 64),
			strconv.FormatFloat(r.PredictionError, 'f', 2, 64),
			correct,
			strconv.FormatFloat(r.AbsoluteError, 'f', 2, 64),
			strconv.FormatFloat(r.SquaredError, 'f', 2, 64),
			strconv.FormatFloat(r.Regret, 'f', 2, 64),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}
