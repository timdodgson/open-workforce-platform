package inrc2

import (
	"fmt"
	"os"
	"strings"
)

// --- Beam Diversity CSV Export ---
// Records per-path diversity metrics for each week of a beam search.
// Pure observation — does not alter algorithm behaviour.

// DiversityRow holds one beam path's diversity metrics for CSV export.
type DiversityRow struct {
	// Run context.
	RunID       string
	Instance    string
	Seed        int64
	BeamWidth   int
	Iterations  int
	Temperature float64
	CoolingMode string
	Timestamp   string

	// Diversity data.
	Week              int
	PathID            int
	Fingerprint       string
	HammingToBest     float64 // distance to best-ranked path this week
	HammingToParent   float64 // distance to parent path's roster
	BeamSpread        int     // worst_retained_cumulative - best_retained_cumulative
	NearDuplicate     bool    // HammingToBest < 0.05
	Retained          bool
	RetainedRank      int
	Winning           bool
	CumulativePenalty int
	WeekPenalty       int
}

// DiversityCSVHeader returns the CSV header for diversity output.
func DiversityCSVHeader() string {
	cols := []string{
		"run_id", "instance", "seed", "beam_width", "iterations",
		"temperature", "cooling_mode", "timestamp",
		"week", "path_id", "fingerprint",
		"hamming_to_best", "hamming_to_parent", "beam_spread",
		"near_duplicate", "retained", "retained_rank", "winning",
		"cumulative_penalty", "week_penalty",
	}
	return strings.Join(cols, ",")
}

// DiversityCSVRow formats a DiversityRow as a CSV line.
func DiversityCSVRow(r DiversityRow) string {
	nearDup := 0
	if r.NearDuplicate {
		nearDup = 1
	}
	retained := 0
	if r.Retained {
		retained = 1
	}
	winning := 0
	if r.Winning {
		winning = 1
	}
	return fmt.Sprintf("%s,%s,%d,%d,%d,%.1f,%s,%s,%d,%d,%s,%.6f,%.6f,%d,%d,%d,%d,%d,%d,%d",
		r.RunID, r.Instance, r.Seed, r.BeamWidth, r.Iterations,
		r.Temperature, r.CoolingMode, r.Timestamp,
		r.Week, r.PathID, r.Fingerprint,
		r.HammingToBest, r.HammingToParent, r.BeamSpread,
		nearDup, retained, r.RetainedRank, winning,
		r.CumulativePenalty, r.WeekPenalty)
}

// BuildDiversityRows converts a BeamResult into diversity CSV rows.
func BuildDiversityRows(ctx RunContext, result BeamResult, sc Scenario) []DiversityRow {
	// Build winning path ID set.
	winningIDs := make(map[int]bool, len(result.WinningPath))
	for _, wp := range result.WinningPath {
		winningIDs[wp.ID] = true
	}

	// Build retained set per week with ranks (same logic as beam_tree_csv.go).
	type retainedInfo struct {
		rank int
	}
	retainedByWeek := make(map[int]map[int]retainedInfo)
	pathsByWeek := make(map[int][]BeamPath)
	for _, p := range result.AllPaths {
		pathsByWeek[p.Week] = append(pathsByWeek[p.Week], p)
	}
	for _, ws := range result.WeekSummaries {
		retainedByWeek[ws.Week] = make(map[int]retainedInfo)
		paths := pathsByWeek[ws.Week]
		for rank, p := range paths {
			if rank < ws.Retained {
				retainedByWeek[ws.Week][p.ID] = retainedInfo{rank: rank + 1}
			}
		}
	}

	// Build parent fingerprint lookup: parent path ID -> fingerprint.
	// Parent paths are from the previous week's retained set.
	parentFingerprints := make(map[int]string)
	parentRosters := make(map[int]*Roster)
	for _, p := range result.AllPaths {
		if _, exists := parentFingerprints[p.ID]; !exists {
			roster := SolutionToRoster(p.Solution, sc)
			parentFingerprints[p.ID] = RosterFingerprint(roster)
			parentRosters[p.ID] = roster
		}
	}

	var rows []DiversityRow
	for _, ws := range result.WeekSummaries {
		weekPaths := pathsByWeek[ws.Week]
		if len(weekPaths) == 0 {
			continue
		}

		// Beam spread: worst retained cumulative - best retained cumulative.
		beamSpread := 0
		if ws.Retained > 1 {
			best := weekPaths[0].CumulativePenalty
			worst := weekPaths[0].CumulativePenalty
			for rank, p := range weekPaths {
				if rank >= ws.Retained {
					break
				}
				if p.CumulativePenalty < best {
					best = p.CumulativePenalty
				}
				if p.CumulativePenalty > worst {
					worst = p.CumulativePenalty
				}
			}
			beamSpread = worst - best
		}

		for _, p := range weekPaths {
			// Compute Hamming distance to parent path.
			hammingToParent := 0.0
			if p.ParentID > 0 {
				parentRoster, ok := parentRosters[p.ParentID]
				if ok {
					thisRoster := SolutionToRoster(p.Solution, sc)
					hammingToParent = RosterHammingDistance(thisRoster, parentRoster)
				}
			}

			ri, isRetained := retainedByWeek[ws.Week][p.ID]
			rank := 0
			if isRetained {
				rank = ri.rank
			}

			rows = append(rows, DiversityRow{
				RunID:             ctx.RunID,
				Instance:          ctx.Instance,
				Seed:              ctx.Seed,
				BeamWidth:         ctx.BeamWidth,
				Iterations:        ctx.Iterations,
				Temperature:       ctx.Temperature,
				CoolingMode:       ctx.CoolingMode,
				Timestamp:         ctx.Timestamp,
				Week:              ws.Week,
				PathID:            p.ID,
				Fingerprint:       p.Fingerprint,
				HammingToBest:     p.HammingToBest,
				HammingToParent:   hammingToParent,
				BeamSpread:        beamSpread,
				NearDuplicate:     p.HammingToBest < 0.05,
				Retained:          isRetained,
				RetainedRank:      rank,
				Winning:           winningIDs[p.ID],
				CumulativePenalty: p.CumulativePenalty,
				WeekPenalty:       p.WeekPenalty,
			})
		}
	}
	return rows
}

// WriteDiversityCSV writes the diversity CSV file.
func WriteDiversityCSV(path string, rows []DiversityRow) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, DiversityCSVHeader())
	for _, row := range rows {
		fmt.Fprintln(f, DiversityCSVRow(row))
	}
	return nil
}
