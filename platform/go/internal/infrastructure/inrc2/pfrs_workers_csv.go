package inrc2

import (
	"fmt"
	"os"
	"strings"
)

// --- Worker Lifecycle CSV Export ---
// Exports per-worker lifecycle data for the PFRS Research Lab dashboard.
// Pure observation — does not alter algorithm behaviour.

// WorkerLifecycleRow holds one worker's complete lifecycle data for CSV export.
type WorkerLifecycleRow struct {
	WorkerID          int
	ParentWorkerID    int
	Week              int
	Seed              int64
	Depth             int
	StartTimeMs       int64 // relative to PFRS week start
	FinishTimeMs      int64 // relative to PFRS week start
	FinishCandidate   int
	InitialTemperature float64
	FinalTemperature  float64
	TempAtBest        float64
	BestCandidate     int
	PlateauCount      int
	BranchCount       int
	ProducedGlobalBest bool
	FinalPenalty      int
	BestPenalty       int
	StartPenalty      int
}

// WorkerLifecycleCSVHeader returns the header row.
func WorkerLifecycleCSVHeader() string {
	cols := []string{
		"worker_id", "parent_worker_id", "week", "seed", "depth",
		"start_time_ms", "finish_time_ms", "finish_candidate",
		"initial_temperature", "final_temperature", "temperature_at_best",
		"best_candidate", "plateau_count", "branch_count",
		"produced_global_best", "final_penalty", "best_penalty", "start_penalty",
	}
	return strings.Join(cols, ",")
}

// WorkerLifecycleCSVRow formats a row as CSV.
func WorkerLifecycleCSVRow(r WorkerLifecycleRow) string {
	pgb := 0
	if r.ProducedGlobalBest {
		pgb = 1
	}
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%.6f,%.6f,%.6f,%d,%d,%d,%d,%d,%d,%d",
		r.WorkerID, r.ParentWorkerID, r.Week, r.Seed, r.Depth,
		r.StartTimeMs, r.FinishTimeMs, r.FinishCandidate,
		r.InitialTemperature, r.FinalTemperature, r.TempAtBest,
		r.BestCandidate, r.PlateauCount, r.BranchCount,
		pgb, r.FinalPenalty, r.BestPenalty, r.StartPenalty)
}

// BuildWorkerLifecycleRows converts WorkerAudit data into lifecycle CSV rows.
// week, seed, depth, initialTemperature are context from the calling code.
// branchCounts maps workerID -> number of branches that worker triggered.
func BuildWorkerLifecycleRows(workers []WorkerAudit, week int, seed int64,
	initialTemp float64, branchCounts map[int]int, depthMap map[int]int) []WorkerLifecycleRow {

	var rows []WorkerLifecycleRow
	for _, w := range workers {
		depth := 0
		if depthMap != nil {
			depth = depthMap[w.WorkerID]
		}
		bc := 0
		if branchCounts != nil {
			bc = branchCounts[w.WorkerID]
		}

		// StartTimeMs: we only have DurationMs, not absolute start.
		// Approximate: finishTime = DurationMs from first worker, startTime = finishTime - DurationMs.
		// This is imprecise for concurrent workers but close enough for the Gantt visualisation.
		finishMs := w.DurationMs
		startMs := int64(0) // workers start roughly at 0 relative to week; concurrent overlap

		rows = append(rows, WorkerLifecycleRow{
			WorkerID:           w.WorkerID,
			ParentWorkerID:     w.ParentWorkerID,
			Week:               week,
			Seed:               seed,
			Depth:              depth,
			StartTimeMs:        startMs,
			FinishTimeMs:       finishMs,
			FinishCandidate:    w.CandidatesEval,
			InitialTemperature: initialTemp,
			FinalTemperature:   w.FinalTemperature,
			TempAtBest:         w.TempAtBest,
			BestCandidate:      w.BestIteration,
			PlateauCount:       len(w.Plateaus),
			BranchCount:        bc,
			ProducedGlobalBest: w.ProducedGlobal,
			FinalPenalty:       w.FinalPenalty,
			BestPenalty:        w.BestPenalty,
			StartPenalty:       w.StartPenalty,
		})
	}
	return rows
}

// WriteWorkerLifecycleCSV writes worker lifecycle data to a CSV file.
func WriteWorkerLifecycleCSV(path string, rows []WorkerLifecycleRow) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, WorkerLifecycleCSVHeader())
	for _, r := range rows {
		fmt.Fprintln(f, WorkerLifecycleCSVRow(r))
	}
	return nil
}
