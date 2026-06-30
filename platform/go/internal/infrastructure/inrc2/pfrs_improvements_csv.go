package inrc2

import (
	"fmt"
	"math"
	"os"
	"strings"
)

// --- Improvements CSV Export ---
// Records every global best update across the entire run.
// Pure observation — does not alter algorithm behaviour.

// ImprovementRow holds one global-best improvement event for CSV export.
type ImprovementRow struct {
	// Run context (for future DynamoDB compatibility).
	RunID     string
	Instance  string
	Seed      int64
	BeamWidth int
	Iterations int
	Temperature float64
	CoolingMode string
	Timestamp  string // ISO 8601

	// Event data.
	Week          int
	WorkerID      int
	Candidate     int
	Temperature_  float64 // at the moment of improvement
	OldGlobalBest int
	NewGlobalBest int
	Improvement   int
	ElapsedMs     int64
}

// ImprovementsCSVHeader returns the CSV header.
func ImprovementsCSVHeader() string {
	cols := []string{
		"run_id", "instance", "seed", "beam_width", "iterations",
		"temperature", "cooling_mode", "timestamp",
		"week", "worker_id", "candidate",
		"temperature_at_event", "old_global_best", "new_global_best",
		"improvement", "elapsed_ms",
	}
	return strings.Join(cols, ",")
}

// ImprovementsCSVRow formats a row.
func ImprovementsCSVRow(r ImprovementRow) string {
	return fmt.Sprintf("%s,%s,%d,%d,%d,%.1f,%s,%s,%d,%d,%d,%.6f,%d,%d,%d,%d",
		r.RunID, r.Instance, r.Seed, r.BeamWidth, r.Iterations,
		r.Temperature, r.CoolingMode, r.Timestamp,
		r.Week, r.WorkerID, r.Candidate,
		r.Temperature_, r.OldGlobalBest, r.NewGlobalBest,
		r.Improvement, r.ElapsedMs)
}

// RunContext holds the run-level metadata to include in every CSV row.
type RunContext struct {
	RunID       string
	Instance    string
	Seed        int64
	BeamWidth   int
	Iterations  int
	Temperature float64
	CoolingMode string
	Timestamp   string
}

// BuildImprovementRows converts BestUpdateEvents into CSV rows.
func BuildImprovementRows(ctx RunContext, week int, events []BestUpdateEvent, effectiveRate float64) []ImprovementRow {
	var rows []ImprovementRow
	for _, e := range events {
		// Calculate theoretical temperature at this candidate.
		temp := ctx.Temperature * math.Pow(1-effectiveRate, float64(e.Iteration))

		rows = append(rows, ImprovementRow{
			RunID:         ctx.RunID,
			Instance:      ctx.Instance,
			Seed:          ctx.Seed,
			BeamWidth:     ctx.BeamWidth,
			Iterations:    ctx.Iterations,
			Temperature:   ctx.Temperature,
			CoolingMode:   ctx.CoolingMode,
			Timestamp:     ctx.Timestamp,
			Week:          week,
			WorkerID:      e.WorkerID,
			Candidate:     e.Iteration,
			Temperature_:  temp,
			OldGlobalBest: e.OldPenalty,
			NewGlobalBest: e.NewPenalty,
			Improvement:   e.OldPenalty - e.NewPenalty,
			ElapsedMs:     e.TimestampMs,
		})
	}
	return rows
}

// WriteImprovementsCSV writes improvement events to a CSV file.
func WriteImprovementsCSV(path string, rows []ImprovementRow) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, ImprovementsCSVHeader())
	for _, r := range rows {
		fmt.Fprintln(f, ImprovementsCSVRow(r))
	}
	return nil
}
