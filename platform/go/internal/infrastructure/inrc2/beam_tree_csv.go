package inrc2

import (
	"fmt"
	"os"
	"strings"
)

// --- Beam Tree CSV Export ---
// Writes beam search path data for visualisation in the PFRS Research Lab.
// Pure observation — does not alter algorithm behaviour.

// BeamTreeCSVHeader returns the CSV header for beam tree output.
func BeamTreeCSVHeader() string {
	cols := []string{
		"path_id", "parent_id", "week", "seed",
		"week_penalty", "cumulative_penalty",
		"workers_started", "candidates", "accepted", "rejected",
		"sa_accepted_better", "sa_accepted_worse", "sa_rejected_by_prob",
		"hard_reject_rate", "duration_ms",
		"retained", "retained_rank", "winning",
	}
	return strings.Join(cols, ",")
}

// BeamTreeRow holds the data for one beam path CSV row.
type BeamTreeRow struct {
	PathID            int
	ParentID          int
	Week              int
	Seed              int64
	WeekPenalty       int
	CumulativePenalty int
	WorkersStarted    int
	Candidates        int
	Accepted          int
	Rejected          int
	SAAcceptedBetter  int
	SAAcceptedWorse   int
	SARejectedByProb  int
	HardRejectRate    float64
	DurationMs        int64
	Retained          bool
	RetainedRank      int // 0 if not retained
	Winning           bool
}

// BeamTreeCSVRow formats a BeamTreeRow as a CSV line.
func BeamTreeCSVRow(r BeamTreeRow) string {
	retained := 0
	if r.Retained {
		retained = 1
	}
	winning := 0
	if r.Winning {
		winning = 1
	}
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.4f,%d,%d,%d,%d",
		r.PathID, r.ParentID, r.Week, r.Seed,
		r.WeekPenalty, r.CumulativePenalty,
		r.WorkersStarted, r.Candidates, r.Accepted, r.Rejected,
		r.SAAcceptedBetter, r.SAAcceptedWorse, r.SARejectedByProb,
		r.HardRejectRate, r.DurationMs,
		retained, r.RetainedRank, winning)
}

// BuildBeamTreeRows converts a BeamResult into CSV rows for export.
func BuildBeamTreeRows(result BeamResult) []BeamTreeRow {
	// Build winning path ID set for marking.
	winningIDs := make(map[int]bool, len(result.WinningPath))
	for _, wp := range result.WinningPath {
		winningIDs[wp.ID] = true
	}

	// Build retained set per week with ranks.
	type retainedInfo struct {
		rank int
	}
	retainedByWeek := make(map[int]map[int]retainedInfo) // week -> pathID -> info
	for _, ws := range result.WeekSummaries {
		retainedByWeek[ws.Week] = make(map[int]retainedInfo)
	}

	// WeekSummaries tell us how many were retained per week.
	// AllPaths is ordered: candidates per week are consecutive and sorted by cumulative penalty.
	// We need to mark the first N as retained per week.
	pathsByWeek := make(map[int][]BeamPath)
	for _, p := range result.AllPaths {
		pathsByWeek[p.Week] = append(pathsByWeek[p.Week], p)
	}
	for _, ws := range result.WeekSummaries {
		paths := pathsByWeek[ws.Week]
		for rank, p := range paths {
			if rank < ws.Retained {
				if retainedByWeek[ws.Week] == nil {
					retainedByWeek[ws.Week] = make(map[int]retainedInfo)
				}
				retainedByWeek[ws.Week][p.ID] = retainedInfo{rank: rank + 1}
			}
		}
	}

	var rows []BeamTreeRow
	for _, p := range result.AllPaths {
		hardRejectRate := 0.0
		if p.Stats.CandidatesEvaluated+p.Stats.InvalidMovesRejected > 0 {
			hardRejectRate = float64(p.Stats.InvalidMovesRejected) /
				float64(p.Stats.CandidatesEvaluated+p.Stats.InvalidMovesRejected) * 100
		}

		// Aggregate SA metrics from audit workers.
		saBetter, saWorse, saRejProb := 0, 0, 0
		for _, w := range p.Audit.Workers {
			saBetter += w.AcceptedBetter
			saWorse += w.AcceptedWorse
			saRejProb += w.RejectedByProb
		}

		ri, isRetained := retainedByWeek[p.Week][p.ID]
		rank := 0
		if isRetained {
			rank = ri.rank
		}

		rows = append(rows, BeamTreeRow{
			PathID:            p.ID,
			ParentID:          p.ParentID,
			Week:              p.Week,
			Seed:              p.Seed,
			WeekPenalty:       p.WeekPenalty,
			CumulativePenalty: p.CumulativePenalty,
			WorkersStarted:    p.Stats.WorkersStarted,
			Candidates:        p.Stats.CandidatesEvaluated,
			Accepted:          p.Stats.ImprovementsAccepted,
			Rejected:          p.Stats.InvalidMovesRejected,
			SAAcceptedBetter:  saBetter,
			SAAcceptedWorse:   saWorse,
			SARejectedByProb:  saRejProb,
			HardRejectRate:    hardRejectRate,
			DurationMs:        p.Stats.DurationMs,
			Retained:          isRetained,
			RetainedRank:      rank,
			Winning:           winningIDs[p.ID],
		})
	}
	return rows
}

// WriteBeamTreeCSV writes the beam tree CSV file.
func WriteBeamTreeCSV(path string, result BeamResult) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, BeamTreeCSVHeader())
	for _, row := range BuildBeamTreeRows(result) {
		fmt.Fprintln(f, BeamTreeCSVRow(row))
	}
	return nil
}
