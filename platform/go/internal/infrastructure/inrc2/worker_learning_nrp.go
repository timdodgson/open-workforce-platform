package inrc2

import (
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/telemetry/workerlearning"
)

// EmitNRPWorkerLearning converts per-week WorkerAudit data into WorkerLearningRecords
// and writes worker_learning.csv. One row per completed NRP worker.
//
// This is pure telemetry — no optimiser decisions change.
func EmitNRPWorkerLearning(outputDir string, cfg NRPLearningConfig, weekAudits []WeekAuditBundle) error {
	var records []workerlearning.Record

	for _, wa := range weekAudits {
		for _, worker := range wa.Workers {
			improved := worker.BestPenalty < worker.StartPenalty
			improvementAmount := 0
			if improved {
				improvementAmount = worker.StartPenalty - worker.BestPenalty
			}

			record := workerlearning.Record{
				// Run metadata.
				ProblemType: "nrp",
				Instance:    cfg.Instance,
				Algorithm:   worker.Algorithm,
				RunSeed:     cfg.RunSeed,

				// Spawn state.
				Week:            wa.Week,
				Phase:           "search",
				Depth:           wa.Depth,
				ParentWorkerID:  worker.ParentWorkerID,
				FamilyID:        wa.FamilyID,
				BeamRank:        wa.BeamRank,
				BeamScore:       wa.BeamScore,
				Entropy:         wa.Entropy,
				Diversity:       wa.Diversity,
				BeamHealth:      wa.BeamHealth,
				Temperature:     cfg.Temperature,
				LAHCLength:      cfg.LAHCLength,
				TabuTenure:      cfg.TabuTenure,
				IterationsAlloc: cfg.IterationsPerWorker,

				// Environment at spawn.
				GlobalBest:       wa.GlobalBestAtSpawn,
				ParentObjective:  worker.StartPenalty,
				DistanceFromBest: worker.StartPenalty - wa.GlobalBestAtSpawn,
				PlateauLength:    wa.PlateauLength,
				RecentImprovRate: wa.RecentImprovRate,
				WorkerCount:      wa.TotalWorkersStarted,
				ActiveFamilies:   wa.ActiveFamilies,

				// Outcome.
				Improved:           improved,
				ProducedGlobalBest: worker.ProducedGlobal,
				ImprovementAmount:  improvementAmount,
				FinalObjective:     worker.BestPenalty,
				RuntimeMs:          worker.DurationMs,
				CandidatesEval:     worker.CandidatesEval,
				Accepted:           worker.Accepted,
				Rejected:           worker.Rejected,
				PlateauCount:       len(worker.Plateaus),
				BranchesSpawned:    0, // not tracked per-worker
			}

			record.ComputeDerived()
			records = append(records, record)
		}
	}

	if len(records) == 0 {
		return nil
	}

	path := filepath.Join(outputDir, "worker_learning.csv")
	return workerlearning.WriteCSV(path, records)
}

// NRPLearningConfig holds run-level parameters for NRP learning records.
type NRPLearningConfig struct {
	Instance           string
	RunSeed            int64
	Temperature        float64
	LAHCLength         int
	TabuTenure         int
	IterationsPerWorker int
}

// WeekAuditBundle holds per-week context and the worker audits for that week.
// Populated by the CLI from the beam search results.
type WeekAuditBundle struct {
	Week               int
	Depth              int
	FamilyID           int
	BeamRank           int
	BeamScore          int
	Entropy            float64
	Diversity          float64
	BeamHealth         float64
	GlobalBestAtSpawn  int
	PlateauLength      int
	RecentImprovRate   float64
	TotalWorkersStarted int
	ActiveFamilies     int
	Workers            []WorkerAudit
}
