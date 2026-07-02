package inrc2

import (
	"fmt"
	"os"
	"strings"
)

// AuditCSVHeader returns the CSV header row.
func AuditCSVHeader() string {
	cols := []string{
		"instance", "seed", "mode", "iterations_per_worker", "max_total_workers",
		"max_concurrent", "initial_temperature", "cooling_rate", "cooling_mode",
		"effective_cooling_rate", "min_temperature", "late_acceptance_length",
		"week", "start_penalty", "final_penalty", "improvement",
		"hard_violations", "soft_violations",
		"candidates", "accepted", "rejected", "acceptance_rate",
		"best_iteration", "best_worker_id",
		"workers_started", "branches_created", "branches_dropped",
		"max_queue_depth", "max_concurrent_seen", "duration_ms",
		"sa_final_temp", "sa_temp_at_best", "sa_accepted_better", "sa_accepted_worse", "sa_rejected_by_prob",
		"lahc_accepted_by_current", "lahc_accepted_by_late", "lahc_rejected_by_late",
		"branches_queued", "branches_started", "branches_completed",
		"winning_branch_depth", "workers_improved", "workers_produced_best",
		"rejected_noop", "rejected_skill", "rejected_succession", "rejected_history",
	}
	return strings.Join(cols, ",")
}

// AuditCSVRow formats a WeekAuditRow as a CSV line.
func AuditCSVRow(r WeekAuditRow) string {
	return fmt.Sprintf("%s,%d,%s,%d,%d,%d,%.6f,%.6f,%s,%.10f,%.6f,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%.4f,%d,%d,%d,%d,%d,%d,%d,%d,%.6f,%.6f,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d",
		r.Instance, r.Seed, r.Mode, r.IterationsPerWorker, r.MaxTotalWorkers,
		r.MaxConcurrent, r.InitialTemperature, r.CoolingRate, r.CoolingMode,
		r.EffectiveCoolingRate, r.MinTemperature, r.LateAcceptanceLen,
		r.Week, r.StartPenalty, r.FinalPenalty, r.Improvement,
		r.HardViolations, r.SoftViolations,
		r.Candidates, r.Accepted, r.Rejected, r.AcceptanceRate,
		r.BestIteration, r.BestWorkerID,
		r.WorkersStarted, r.BranchesCreated, r.BranchesDropped,
		r.MaxQueueDepth, r.MaxConcurrentSeen, r.DurationMs,
		r.SAFinalTemp, r.SATempAtBest, r.SAAcceptedBetter, r.SAAcceptedWorse, r.SARejectedByProb,
		r.LAHCAcceptedByCurrent, r.LAHCAcceptedByLate, r.LAHCRejectedByLate,
		r.BranchesQueued, r.BranchesStarted, r.BranchesCompleted,
		r.WinningBranchDepth, r.WorkersImproved, r.WorkersProducedBest,
		r.RejectedNoop, r.RejectedSkill, r.RejectedSuccession, r.RejectedHistory,
	)
}

// WriteAuditCSV writes a complete CSV file from a slice of audit rows.
func WriteAuditCSV(path string, rows []WeekAuditRow) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, AuditCSVHeader())
	for _, r := range rows {
		fmt.Fprintln(f, AuditCSVRow(r))
	}
	return nil
}
