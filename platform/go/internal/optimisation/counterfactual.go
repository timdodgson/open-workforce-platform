// counterfactual.go — Counterfactual Learning for Search Intelligence.
//
// Every policy decision records what was done, what alternatives existed,
// and what happened. This data is the primary training signal for future
// policy improvement.
//
// Output: counterfactual_learning.csv (one row per decision point).
package optimisation

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// ───────────────────────────────────────────────────────────────
// CounterfactualRecord
// ───────────────────────────────────────────────────────────────

// CounterfactualRecord captures a single decision with its alternatives and outcome.
type CounterfactualRecord struct {
	// Identity
	Timestamp    time.Time
	RunID        string
	DecisionType string // worker, search, portfolio

	// Context
	Domain   string
	Instance string
	Algorithm string

	// Actual decision
	ActualAction   string
	ActualConfidence float64
	PolicyID       string
	PolicyVersion  string
	PolicyType     string // rule, learned, hybrid

	// Counterfactual alternatives
	CounterfactualActions []CounterfactualAction

	// Expected value at decision time
	ExpectedValue float64

	// Outcome (filled after execution)
	ActualOutcome float64 // observed result (objective, runtime savings, etc.)
	OutcomeMetric string  // what was measured: "objective", "compute_saved", "runtime"

	// Regret
	Regret         float64 // actual - best_counterfactual (positive = bad decision)
	BestAlternative string // which counterfactual would have been better
}

// CounterfactualAction represents one alternative that was not taken.
type CounterfactualAction struct {
	Action         string  // what would have been done
	Source         string  // who suggested it: "rule", "learned_v1", "random", "opposite"
	ExpectedValue  float64 // estimated outcome if this action were taken
	EstimatedRegret float64 // estimated regret vs actual (filled post-outcome)
}

// ───────────────────────────────────────────────────────────────
// CounterfactualRecorder
// ───────────────────────────────────────────────────────────────

// CounterfactualRecorder writes counterfactual records to CSV.
// Thread-safe for concurrent writes from parallel workers.
type CounterfactualRecorder struct {
	mu       sync.Mutex
	dir      string
	file     *os.File
	writer   *csv.Writer
	count    int
	disabled bool
}

// NewCounterfactualRecorder creates a recorder writing to the given directory.
// If dir is empty, recording is disabled (no-op).
func NewCounterfactualRecorder(dir string) *CounterfactualRecorder {
	if dir == "" {
		return &CounterfactualRecorder{disabled: true}
	}
	return &CounterfactualRecorder{dir: dir}
}

// Record writes a counterfactual record. Thread-safe.
func (cr *CounterfactualRecorder) Record(rec CounterfactualRecord) error {
	if cr.disabled {
		return nil
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.writer == nil {
		if err := cr.open(); err != nil {
			return err
		}
	}

	// Flatten counterfactual actions to top-3 for CSV.
	cf1Action, cf1Source, cf1Expected := "", "", 0.0
	cf2Action, cf2Source, cf2Expected := "", "", 0.0
	cf3Action, cf3Source, cf3Expected := "", "", 0.0
	if len(rec.CounterfactualActions) > 0 {
		cf1Action = rec.CounterfactualActions[0].Action
		cf1Source = rec.CounterfactualActions[0].Source
		cf1Expected = rec.CounterfactualActions[0].ExpectedValue
	}
	if len(rec.CounterfactualActions) > 1 {
		cf2Action = rec.CounterfactualActions[1].Action
		cf2Source = rec.CounterfactualActions[1].Source
		cf2Expected = rec.CounterfactualActions[1].ExpectedValue
	}
	if len(rec.CounterfactualActions) > 2 {
		cf3Action = rec.CounterfactualActions[2].Action
		cf3Source = rec.CounterfactualActions[2].Source
		cf3Expected = rec.CounterfactualActions[2].ExpectedValue
	}

	row := []string{
		rec.Timestamp.Format(time.RFC3339),
		rec.RunID,
		rec.DecisionType,
		rec.Domain,
		rec.Instance,
		rec.Algorithm,
		rec.ActualAction,
		formatFloat(rec.ActualConfidence),
		rec.PolicyID,
		rec.PolicyVersion,
		rec.PolicyType,
		formatFloat(rec.ExpectedValue),
		cf1Action, cf1Source, formatFloat(cf1Expected),
		cf2Action, cf2Source, formatFloat(cf2Expected),
		cf3Action, cf3Source, formatFloat(cf3Expected),
		formatFloat(rec.ActualOutcome),
		rec.OutcomeMetric,
		formatFloat(rec.Regret),
		rec.BestAlternative,
	}

	if err := cr.writer.Write(row); err != nil {
		return fmt.Errorf("counterfactual: write error: %w", err)
	}
	cr.writer.Flush()
	cr.count++
	return nil
}

// Count returns the number of records written.
func (cr *CounterfactualRecorder) Count() int {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	return cr.count
}

// Close flushes and closes the underlying file.
func (cr *CounterfactualRecorder) Close() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	if cr.writer != nil {
		cr.writer.Flush()
	}
	if cr.file != nil {
		return cr.file.Close()
	}
	return nil
}

func (cr *CounterfactualRecorder) open() error {
	if err := os.MkdirAll(cr.dir, 0o755); err != nil {
		return fmt.Errorf("counterfactual: mkdir error: %w", err)
	}
	filename := filepath.Join(cr.dir, "counterfactual_learning.csv")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("counterfactual: open error: %w", err)
	}

	cr.file = f
	cr.writer = csv.NewWriter(f)

	// Write header if file is new (size 0).
	info, _ := f.Stat()
	if info != nil && info.Size() == 0 {
		header := []string{
			"timestamp", "run_id", "decision_type",
			"domain", "instance", "algorithm",
			"actual_action", "actual_confidence", "policy_id", "policy_version", "policy_type",
			"expected_value",
			"cf1_action", "cf1_source", "cf1_expected",
			"cf2_action", "cf2_source", "cf2_expected",
			"cf3_action", "cf3_source", "cf3_expected",
			"actual_outcome", "outcome_metric",
			"regret", "best_alternative",
		}
		if err := cr.writer.Write(header); err != nil {
			return fmt.Errorf("counterfactual: header write error: %w", err)
		}
		cr.writer.Flush()
	}

	return nil
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 4, 64)
}
