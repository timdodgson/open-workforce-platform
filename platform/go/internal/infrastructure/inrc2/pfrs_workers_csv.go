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
	// Run context.
	RunContext

	// Identity.
	WorkerID          int
	ParentWorkerID    int
	Week              int
	Seed              int64
	Depth             int

	// Timing.
	StartTimeMs       int64
	FinishTimeMs      int64
	FinishCandidate   int

	// Temperature.
	InitialTemperature float64
	FinalTemperature  float64
	TempAtBest        float64

	// Outcomes.
	BestCandidate     int
	PlateauCount      int
	BranchCount       int
	ProducedGlobalBest bool
	FinalPenalty      int
	BestPenalty       int
	StartPenalty      int

	// Efficiency metrics.
	ImprovementPer10K float64
	BranchesPer10K    float64
	AcceptedWorsePct  float64
	HardRejectPct     float64
	FirstPlateauCand  int
	LastImproveCand   int

	// Lifecycle events.
	LocalBestImprovements int
	AcceptedWorse         int
	HardRejects           int
}

// WorkerLifecycleCSVHeader returns the header row.
func WorkerLifecycleCSVHeader() string {
	cols := []string{
		"run_id", "instance", "seed", "beam_width", "iterations",
		"temperature", "cooling_mode", "timestamp",
		"worker_id", "parent_worker_id", "week", "worker_seed", "depth",
		"start_time_ms", "finish_time_ms", "finish_candidate",
		"initial_temperature", "final_temperature", "temperature_at_best",
		"best_candidate", "plateau_count", "branch_count",
		"produced_global_best", "final_penalty", "best_penalty", "start_penalty",
		"improvement_per_10k", "branches_per_10k",
		"accepted_worse_pct", "hard_reject_pct",
		"first_plateau_candidate", "last_improvement_candidate",
		"local_best_improvements", "accepted_worse", "hard_rejects",
	}
	return strings.Join(cols, ",")
}

// WorkerLifecycleCSVRow formats a row as CSV.
func WorkerLifecycleCSVRow(r WorkerLifecycleRow) string {
	pgb := 0
	if r.ProducedGlobalBest {
		pgb = 1
	}
	return fmt.Sprintf("%s,%s,%d,%d,%d,%.1f,%s,%s,%d,%d,%d,%d,%d,%d,%d,%d,%.6f,%.6f,%.6f,%d,%d,%d,%d,%d,%d,%d,%.4f,%.4f,%.2f,%.2f,%d,%d,%d,%d,%d",
		r.RunID, r.Instance, r.RunContext.Seed, r.BeamWidth, r.Iterations,
		r.RunContext.Temperature, r.CoolingMode, r.Timestamp,
		r.WorkerID, r.ParentWorkerID, r.Week, r.Seed, r.Depth,
		r.StartTimeMs, r.FinishTimeMs, r.FinishCandidate,
		r.InitialTemperature, r.FinalTemperature, r.TempAtBest,
		r.BestCandidate, r.PlateauCount, r.BranchCount,
		pgb, r.FinalPenalty, r.BestPenalty, r.StartPenalty,
		r.ImprovementPer10K, r.BranchesPer10K,
		r.AcceptedWorsePct, r.HardRejectPct,
		r.FirstPlateauCand, r.LastImproveCand,
		r.LocalBestImprovements, r.AcceptedWorse, r.HardRejects)
}

// BuildWorkerLifecycleRows converts WorkerAudit data into lifecycle CSV rows.
func BuildWorkerLifecycleRows(ctx RunContext, workers []WorkerAudit, week int, seed int64,
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

		finishMs := w.DurationMs
		startMs := int64(0)

		// Efficiency metrics.
		cands := w.CandidatesEval
		improvPer10K := 0.0
		branchesPer10K := 0.0
		if cands > 0 {
			improvement := w.StartPenalty - w.BestPenalty
			improvPer10K = float64(improvement) / float64(cands) * 10000
			branchesPer10K = float64(bc) / float64(cands) * 10000
		}
		acceptedWorsePct := 0.0
		if cands > 0 {
			acceptedWorsePct = float64(w.AcceptedWorse) / float64(cands) * 100
		}
		hardRejectPct := 0.0
		if w.Attempts > 0 {
			hardRejectPct = float64(w.Rejected) / float64(w.Attempts) * 100
		}

		// First plateau candidate.
		firstPlateauCand := 0
		if len(w.Plateaus) > 0 {
			firstPlateauCand = w.Plateaus[0].Candidate
		}

		// Local best improvement count = number of times penalty decreased.
		// Approximated as: accepted_better that actually improved local best.
		// We use BestIteration as last improvement candidate.
		lastImproveCand := w.BestIteration

		rows = append(rows, WorkerLifecycleRow{
			RunContext:         ctx,
			WorkerID:           w.WorkerID,
			ParentWorkerID:     w.ParentWorkerID,
			Week:               week,
			Seed:               seed,
			Depth:              depth,
			StartTimeMs:        startMs,
			FinishTimeMs:       finishMs,
			FinishCandidate:    cands,
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
			ImprovementPer10K:  improvPer10K,
			BranchesPer10K:     branchesPer10K,
			AcceptedWorsePct:   acceptedWorsePct,
			HardRejectPct:      hardRejectPct,
			FirstPlateauCand:   firstPlateauCand,
			LastImproveCand:    lastImproveCand,
			LocalBestImprovements: w.AcceptedBetter,
			AcceptedWorse:      w.AcceptedWorse,
			HardRejects:        w.Rejected,
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
