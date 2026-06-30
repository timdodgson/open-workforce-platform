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
	EventType          string // "LOCAL_BEST" or "GLOBAL_BEST"
	BranchDepth        int
	SeedUsed           int64
	AcceptedWorseCount int
	HardRejectCount    int
	SoftRejectCount    int

	// Derived metrics (computed during row building).
	DiscoveryNumber        int
	CandsSincePrevious     int
	TimeSincePreviousMs    int64
	ImprovementPer10K      float64
	ImprovementPerSecond   float64
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
	}
	return strings.Join(cols, ",")
}

// DiscoveriesCSVRow formats a DiscoveryRow as a CSV line.
func DiscoveriesCSVRow(r DiscoveryRow) string {
	return fmt.Sprintf("%s,%s,%d,%d,%d,%.1f,%s,%s,%d,%d,%d,%d,%d,%.6f,%d,%d,%d,%d,%.4f,%s,%d,%d,%d,%d,%d,%d,%d,%d,%.4f,%.4f",
		r.RunID, r.Instance, r.Seed, r.BeamWidth, r.Iterations,
		r.Temperature, r.CoolingMode, r.Timestamp,
		r.Week, r.WorkerID, r.BeamPath, r.Candidate, r.ElapsedMs,
		r.TemperatureAtEvent, r.CurrentPenalty, r.PreviousBest, r.NewBest,
		r.Improvement, r.ImprovementPercent, r.EventType,
		r.BranchDepth, r.SeedUsed, r.AcceptedWorseCount,
		r.HardRejectCount, r.SoftRejectCount,
		r.DiscoveryNumber, r.CandsSincePrevious, r.TimeSincePreviousMs,
		r.ImprovementPer10K, r.ImprovementPerSecond)
}

// BuildDiscoveryRows converts DiscoveryEvents from a PFRS audit into CSV rows
// with derived metrics (candidates between discoveries, yield, etc).
func BuildDiscoveryRows(ctx RunContext, week int, beamPath int, seed int64,
	events []DiscoveryEvent, depthMap map[int]int) []DiscoveryRow {

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
	return rows
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
