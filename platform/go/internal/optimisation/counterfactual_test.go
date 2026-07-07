package optimisation

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCounterfactualRecorder_Disabled(t *testing.T) {
	cr := NewCounterfactualRecorder("")
	err := cr.Record(CounterfactualRecord{})
	if err != nil {
		t.Errorf("disabled recorder should not error, got: %v", err)
	}
	if cr.Count() != 0 {
		t.Errorf("disabled recorder count should be 0, got %d", cr.Count())
	}
}

func TestCounterfactualRecorder_WriteAndRead(t *testing.T) {
	dir := t.TempDir()
	cr := NewCounterfactualRecorder(dir)
	defer cr.Close()

	rec := CounterfactualRecord{
		Timestamp:        time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC),
		RunID:            "stat-cvrp-a32k5-sa-adaptive-s42",
		DecisionType:     "search",
		Domain:           "cvrp",
		Instance:         "A-n32-k5",
		Algorithm:        "sa",
		ActualAction:     "early_stop",
		ActualConfidence: 0.82,
		PolicyID:         "cvrp-search-v2",
		PolicyVersion:    "2.1.0",
		PolicyType:       "learned",
		ExpectedValue:    784.0,
		CounterfactualActions: []CounterfactualAction{
			{Action: "continue", Source: "rule", ExpectedValue: 790.0},
			{Action: "restart", Source: "learned_v1", ExpectedValue: 786.0},
		},
		ActualOutcome:   784.0,
		OutcomeMetric:   "objective",
		Regret:          0.0,
		BestAlternative: "",
	}

	if err := cr.Record(rec); err != nil {
		t.Fatalf("Record failed: %v", err)
	}
	if cr.Count() != 1 {
		t.Errorf("Count = %d, want 1", cr.Count())
	}

	// Write a second record with regret.
	rec2 := CounterfactualRecord{
		Timestamp:        time.Date(2026, 7, 7, 12, 1, 0, 0, time.UTC),
		RunID:            "stat-jss-la01-tabu-adaptive-s42",
		DecisionType:     "search",
		Domain:           "jss",
		Instance:         "la01",
		Algorithm:        "tabu",
		ActualAction:     "continue",
		ActualConfidence: 0.55,
		PolicyID:         "jss-search-rules",
		PolicyVersion:    "1.0.0",
		PolicyType:       "rule",
		ExpectedValue:    666.0,
		CounterfactualActions: []CounterfactualAction{
			{Action: "early_stop", Source: "learned_v2", ExpectedValue: 666.0, EstimatedRegret: -20.0},
		},
		ActualOutcome:   686.0,
		OutcomeMetric:   "objective",
		Regret:          20.0,
		BestAlternative: "early_stop",
	}

	if err := cr.Record(rec2); err != nil {
		t.Fatalf("Record #2 failed: %v", err)
	}
	if cr.Count() != 2 {
		t.Errorf("Count = %d, want 2", cr.Count())
	}

	// Close and read back.
	cr.Close()

	f, err := os.Open(filepath.Join(dir, "counterfactual_learning.csv"))
	if err != nil {
		t.Fatalf("Open CSV failed: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}

	// Header + 2 data rows.
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows (header + 2 data), got %d", len(rows))
	}

	// Verify header.
	if rows[0][0] != "timestamp" {
		t.Errorf("header[0] = %q, want timestamp", rows[0][0])
	}
	if rows[0][23] != "regret" {
		t.Errorf("header[23] = %q, want regret", rows[0][23])
	}

	// Verify first data row.
	if rows[1][6] != "early_stop" {
		t.Errorf("row1 actual_action = %q, want early_stop", rows[1][6])
	}
	if rows[1][3] != "cvrp" {
		t.Errorf("row1 domain = %q, want cvrp", rows[1][3])
	}

	// Verify second data row has regret.
	if rows[2][23] != "20.0000" {
		t.Errorf("row2 regret = %q, want 20.0000", rows[2][23])
	}
	if rows[2][24] != "early_stop" {
		t.Errorf("row2 best_alternative = %q, want early_stop", rows[2][24])
	}
}

func TestCounterfactualRecord_Fields(t *testing.T) {
	rec := CounterfactualRecord{
		ActualAction: "skip",
		CounterfactualActions: []CounterfactualAction{
			{Action: "run", Source: "rule", ExpectedValue: 100.0},
			{Action: "reduce_budget", Source: "learned_v1", ExpectedValue: 95.0},
			{Action: "increase_budget", Source: "random", ExpectedValue: 110.0},
		},
		ActualOutcome: 105.0,
		Regret:        5.0, // could have got 100 with "run"
	}

	if len(rec.CounterfactualActions) != 3 {
		t.Errorf("expected 3 counterfactual actions, got %d", len(rec.CounterfactualActions))
	}
	if rec.Regret != 5.0 {
		t.Errorf("Regret = %f, want 5.0", rec.Regret)
	}
}

func TestCounterfactualRecorder_ConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	cr := NewCounterfactualRecorder(dir)
	defer cr.Close()

	// Write 100 records concurrently.
	done := make(chan struct{}, 100)
	for i := 0; i < 100; i++ {
		go func(n int) {
			defer func() { done <- struct{}{} }()
			rec := CounterfactualRecord{
				Timestamp:    time.Now(),
				RunID:        "concurrent-test",
				DecisionType: "worker",
				Domain:       "nrp",
				ActualAction: "run",
			}
			if err := cr.Record(rec); err != nil {
				t.Errorf("concurrent write %d failed: %v", n, err)
			}
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	if cr.Count() != 100 {
		t.Errorf("Count = %d, want 100", cr.Count())
	}
}
