package inrc2

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// --- Discoveries CSV Export ---
// Records every meaningful discovery (local best or global best) across a run.
// Pure observation — does not alter algorithm behaviour.

// DiscoveryRow holds one discovery event for CSV export.
type DiscoveryRow struct {
	// Run context.
	RunID       string
	Instance    string
	Seed        int64
	BeamWidth   int
	Iterations  int
	Temperature float64
	CoolingMode string
	Timestamp   string

	// Discovery data.
	Week               int
	WorkerID           int
	BeamPath           int
	Candidate          int
	ElapsedMs          int64
	TemperatureAtEvent float64
	CurrentPenalty     int
	PreviousBest       int
	NewBest            int
	Improvement        int
	ImprovementPercent float64
	EventType          string // "LOCAL_BEST", "GLOBAL_BEST", or "REHEAT"
	BranchDepth        int
	SeedUsed           int64
	AcceptedWorseCount int
	HardRejectCount    int
	SoftRejectCount    int

	// Derived metrics (computed during row building).
	DiscoveryNumber      int
	CandsSincePrevious   int
	TimeSincePreviousMs  int64
	ImprovementPer10K    float64
	ImprovementPerSecond float64

	// Post-reheat metrics (only populated for REHEAT rows, computed post-run).
	PostReheatImproved          bool // did a later LOCAL_BEST beat the pre-reheat local best?
	PostReheatBestDelta         int  // how much better than pre-reheat best (0 if not improved)
	PostReheatCandidatesToImprove int  // candidates from reheat to first post-reheat improvement (0 if none)
	PostReheatSpawnedBranch     bool // did any GLOBAL_BEST occur for this worker after this reheat?
	PostReheatBeatGlobal        bool // same as spawned branch (global best = branch trigger)
	PostReheatOnWinningLineage  bool // was this worker on the winning lineage? (false if unknown)
}

// DiscoveriesCSVHeader returns the CSV header for discoveries output.
func DiscoveriesCSVHeader() string {
	cols := []string{
		"run_id", "instance", "seed", "beam_width", "iterations",
		"temperature", "cooling_mode", "timestamp",
		"week", "worker_id", "beam_path", "candidate", "elapsed_ms",
		"temperature_at_event", "current_penalty", "previous_best", "new_best",
		"improvement", "improvement_percent", "event_type",
		"branch_depth", "seed_used", "accepted_worse_count",
		"hard_reject_count", "soft_reject_count",
		"discovery_number", "cands_since_previous", "time_since_previous_ms",
		"improvement_per_10k", "improvement_per_second",
		"post_reheat_improved", "post_reheat_best_delta",
		"post_reheat_candidates_to_improve", "post_reheat_spawned_branch",
		"post_reheat_beat_global", "post_reheat_on_winning_lineage",
	}
	return strings.Join(cols, ",")
}

// DiscoveriesCSVRow formats a DiscoveryRow as a CSV line.
func DiscoveriesCSVRow(r DiscoveryRow) string {
	boolInt := func(b bool) int {
		if b {
			return 1
		}
		return 0
	}
	return fmt.Sprintf("%s,%s,%d,%d,%d,%.1f,%s,%s,%d,%d,%d,%d,%d,%.6f,%d,%d,%d,%d,%.4f,%s,%d,%d,%d,%d,%d,%d,%d,%d,%.4f,%.4f,%d,%d,%d,%d,%d,%d",
		r.RunID, r.Instance, r.Seed, r.BeamWidth, r.Iterations,
		r.Temperature, r.CoolingMode, r.Timestamp,
		r.Week, r.WorkerID, r.BeamPath, r.Candidate, r.ElapsedMs,
		r.TemperatureAtEvent, r.CurrentPenalty, r.PreviousBest, r.NewBest,
		r.Improvement, r.ImprovementPercent, r.EventType,
		r.BranchDepth, r.SeedUsed, r.AcceptedWorseCount,
		r.HardRejectCount, r.SoftRejectCount,
		r.DiscoveryNumber, r.CandsSincePrevious, r.TimeSincePreviousMs,
		r.ImprovementPer10K, r.ImprovementPerSecond,
		boolInt(r.PostReheatImproved), r.PostReheatBestDelta,
		r.PostReheatCandidatesToImprove, boolInt(r.PostReheatSpawnedBranch),
		boolInt(r.PostReheatBeatGlobal), boolInt(r.PostReheatOnWinningLineage))
}

// BuildDiscoveryRows converts DiscoveryEvents from a PFRS audit into CSV rows
// with derived metrics (candidates between discoveries, yield, etc).
// winningWorkerID: the worker that produced the final global best (-1 if unknown).
func BuildDiscoveryRows(ctx RunContext, week int, beamPath int, seed int64,
	events []DiscoveryEvent, depthMap map[int]int, winningWorkerID int) []DiscoveryRow {

	if len(events) == 0 {
		return nil
	}

	// Sort events by candidate (chronological within the week's PFRS run).
	sorted := make([]DiscoveryEvent, len(events))
	copy(sorted, events)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Candidate < sorted[j].Candidate
	})

	var rows []DiscoveryRow
	prevCandidate := 0
	var prevTimeMs int64

	for i, e := range sorted {
		depth := 0
		if depthMap != nil {
			depth = depthMap[e.WorkerID]
		}

		improvementPct := 0.0
		if e.PreviousBest > 0 {
			improvementPct = float64(e.Improvement) / float64(e.PreviousBest) * 100
		}

		candsSincePrev := e.Candidate - prevCandidate
		timeSincePrev := e.TimestampMs - prevTimeMs

		improvPer10K := 0.0
		if candsSincePrev > 0 {
			improvPer10K = float64(e.Improvement) / float64(candsSincePrev) * 10000
		}
		improvPerSecond := 0.0
		if timeSincePrev > 0 {
			improvPerSecond = float64(e.Improvement) / (float64(timeSincePrev) / 1000.0)
		}

		rows = append(rows, DiscoveryRow{
			RunID:       ctx.RunID,
			Instance:    ctx.Instance,
			Seed:        ctx.Seed,
			BeamWidth:   ctx.BeamWidth,
			Iterations:  ctx.Iterations,
			Temperature: ctx.Temperature,
			CoolingMode: ctx.CoolingMode,
			Timestamp:   ctx.Timestamp,

			Week:               week,
			WorkerID:           e.WorkerID,
			BeamPath:           beamPath,
			Candidate:          e.Candidate,
			ElapsedMs:          e.TimestampMs,
			TemperatureAtEvent: e.Temperature,
			CurrentPenalty:     e.CurrentPenalty,
			PreviousBest:       e.PreviousBest,
			NewBest:            e.NewBest,
			Improvement:        e.Improvement,
			ImprovementPercent: improvementPct,
			EventType:          e.EventType,
			BranchDepth:        depth,
			SeedUsed:           seed,
			AcceptedWorseCount: e.AcceptedWorseCount,
			HardRejectCount:    e.HardRejectCount,
			SoftRejectCount:    e.SoftRejectCount,

			DiscoveryNumber:      i + 1,
			CandsSincePrevious:   candsSincePrev,
			TimeSincePreviousMs:  timeSincePrev,
			ImprovementPer10K:    improvPer10K,
			ImprovementPerSecond: improvPerSecond,
		})

		prevCandidate = e.Candidate
		prevTimeMs = e.TimestampMs
	}

	// Post-processing: enrich REHEAT rows with outcome data.
	enrichReheatRows(rows, winningWorkerID)

	return rows
}

// enrichReheatRows scans the sorted discovery rows and, for each REHEAT event,
// looks ahead at subsequent events for the same worker to determine outcomes.
func enrichReheatRows(rows []DiscoveryRow, winningWorkerID int) {
	for i := range rows {
		if rows[i].EventType != "REHEAT" {
			continue
		}

		reheatWorker := rows[i].WorkerID
		reheatBest := rows[i].PreviousBest // local best at time of reheat
		reheatCandidate := rows[i].Candidate

		// On winning lineage: only true if this worker is the winning worker.
		// We don't guess lineage through parent chains — just direct match.
		if winningWorkerID >= 0 && reheatWorker == winningWorkerID {
			rows[i].PostReheatOnWinningLineage = true
		}

		// Scan forward for events from the same worker after this reheat.
		for j := i + 1; j < len(rows); j++ {
			if rows[j].WorkerID != reheatWorker {
				continue
			}
			// Only look at events after this reheat's candidate.
			if rows[j].Candidate <= reheatCandidate {
				continue
			}

			switch rows[j].EventType {
			case "LOCAL_BEST", "GLOBAL_BEST":
				// Did this discovery beat the pre-reheat local best?
				if rows[j].NewBest < reheatBest {
					if !rows[i].PostReheatImproved {
						// First improvement after this reheat.
						rows[i].PostReheatImproved = true
						rows[i].PostReheatBestDelta = reheatBest - rows[j].NewBest
						rows[i].PostReheatCandidatesToImprove = rows[j].Candidate - reheatCandidate
					} else {
						// Track best delta across all post-reheat improvements.
						delta := reheatBest - rows[j].NewBest
						if delta > rows[i].PostReheatBestDelta {
							rows[i].PostReheatBestDelta = delta
						}
					}
				}

				if rows[j].EventType == "GLOBAL_BEST" {
					rows[i].PostReheatSpawnedBranch = true
					rows[i].PostReheatBeatGlobal = true
				}

			case "REHEAT":
				// Hit the next reheat for this worker — stop looking.
				break
			}

			// If we hit another REHEAT for this worker, stop scanning.
			if rows[j].EventType == "REHEAT" && rows[j].WorkerID == reheatWorker {
				break
			}
		}
	}
}

// WriteDiscoveriesCSV writes the discoveries CSV file.
func WriteDiscoveriesCSV(path string, rows []DiscoveryRow) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, DiscoveriesCSVHeader())
	for _, row := range rows {
		fmt.Fprintln(f, DiscoveriesCSVRow(row))
	}
	return nil
}
