package inrc2

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// --- Branch Event CSV Export ---
// Records every global-best improvement that triggered a branch.
// Pure observation — does not alter algorithm behaviour.

// BranchCSVHeader returns the CSV header for branch event output.
func BranchCSVHeader() string {
	cols := []string{
		"run_id", "instance", "seed", "beam_width", "iterations",
		"temperature", "cooling_mode", "timestamp",
		"week", "worker_id", "parent_worker_id", "child_worker_id",
		"depth", "branch_reason", "candidate",
		"temperature_at_event", "old_penalty", "new_penalty",
		"improvement", "elapsed_ms",
	}
	return strings.Join(cols, ",")
}

// BranchRow holds one branch event for CSV export.
type BranchRow struct {
	RunContext
	Week             int
	WorkerID         int
	ParentWorkerID   int
	ChildWorkerID    int
	Depth            int
	BranchReason     string
	Candidate        int
	TemperatureAt    float64
	OldPenalty       int
	NewPenalty       int
	Improvement      int
	ElapsedMs        int64
}

// BranchCSVRow formats a BranchRow as a CSV line.
func BranchCSVRow(r BranchRow) string {
	return fmt.Sprintf("%s,%s,%d,%d,%d,%.1f,%s,%s,%d,%d,%d,%d,%d,%s,%d,%.6f,%d,%d,%d,%d",
		r.RunID, r.Instance, r.Seed, r.BeamWidth, r.Iterations,
		r.Temperature, r.CoolingMode, r.Timestamp,
		r.Week, r.WorkerID, r.ParentWorkerID, r.ChildWorkerID,
		r.Depth, r.BranchReason, r.Candidate,
		r.TemperatureAt, r.OldPenalty, r.NewPenalty,
		r.Improvement, r.ElapsedMs)
}

// WriteBranchCSV writes branch events to a CSV file.
func WriteBranchCSV(path string, rows []BranchRow) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, BranchCSVHeader())
	for _, r := range rows {
		fmt.Fprintln(f, BranchCSVRow(r))
	}
	return nil
}

// BuildBranchRows converts BestUpdateEvents into BranchRows with context.
func BuildBranchRows(ctx RunContext, week int, events []BestUpdateEvent,
	effectiveRate float64, depthMap map[int]int, parentMap map[int]int) []BranchRow {

	var rows []BranchRow
	for i, e := range events {
		tempAtEvent := ctx.Temperature * math.Pow(1-effectiveRate, float64(e.Iteration))
		childID := e.WorkerID + 1000 + i
		depth := 0
		if depthMap != nil {
			depth = depthMap[e.WorkerID]
		}
		parentID := -1
		if parentMap != nil {
			parentID = parentMap[e.WorkerID]
		}

		rows = append(rows, BranchRow{
			RunContext:     ctx,
			Week:           week,
			WorkerID:       e.WorkerID,
			ParentWorkerID: parentID,
			ChildWorkerID:  childID,
			Depth:          depth,
			BranchReason:   "global_best_update",
			Candidate:      e.Iteration,
			TemperatureAt:  tempAtEvent,
			OldPenalty:     e.OldPenalty,
			NewPenalty:     e.NewPenalty,
			Improvement:    e.OldPenalty - e.NewPenalty,
			ElapsedMs:      e.TimestampMs,
		})
	}
	return rows
}
