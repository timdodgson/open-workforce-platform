package inrc2

import (
	"fmt"
	"os"
	"strings"
)

// --- Branch Event CSV Export ---
// Exports best-update events (which trigger branches) for visualisation.
// Pure observation — does not alter algorithm behaviour.

// BranchCSVHeader returns the CSV header for branch event output.
func BranchCSVHeader() string {
	cols := []string{
		"week", "worker_id", "candidate", "old_penalty", "new_penalty",
		"improvement", "timestamp_ms",
	}
	return strings.Join(cols, ",")
}

// BranchCSVRow formats a BestUpdateEvent as a CSV row.
func BranchCSVRow(week int, e BestUpdateEvent) string {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d",
		week, e.WorkerID, e.Iteration,
		e.OldPenalty, e.NewPenalty,
		e.OldPenalty-e.NewPenalty, e.TimestampMs)
}

// WriteBranchCSV writes branch events from all weeks to a CSV file.
func WriteBranchCSV(path string, weekEvents map[int][]BestUpdateEvent) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, BranchCSVHeader())
	for week := 1; week <= len(weekEvents)+1; week++ {
		events, ok := weekEvents[week]
		if !ok {
			continue
		}
		for _, e := range events {
			fmt.Fprintln(f, BranchCSVRow(week, e))
		}
	}
	return nil
}
