package workerlearning

import (
	"path/filepath"
)

// EmitSingleWorkerLearning creates a worker_learning.csv for a single-search run
// (CVRP, JSS, VRPTW). These runs have exactly one "worker" — the search itself.
// This records the run as a training example for future ML models.
func EmitSingleWorkerLearning(outputDir string, cfg SingleWorkerConfig, outcome SearchOutcome) error {
	record := Record{
		// Run metadata.
		ProblemType: cfg.ProblemType,
		Instance:    cfg.Instance,
		Algorithm:   cfg.Algorithm,
		RunSeed:     cfg.Seed,

		// Spawn state (single worker — no beam context).
		Week:            1,
		Phase:           "search",
		Depth:           0,
		ParentWorkerID:  -1,
		FamilyID:        0,
		BeamRank:        0,
		BeamScore:       outcome.InitialPenalty,
		Entropy:         0,
		Diversity:       0,
		BeamHealth:      1.0,
		Temperature:     cfg.Temperature,
		LAHCLength:      cfg.LAHCLength,
		TabuTenure:      cfg.TabuTenure,
		IterationsAlloc: cfg.Iterations,

		// Environment at spawn (single worker — it IS the global state).
		GlobalBest:       outcome.InitialPenalty,
		ParentObjective:  outcome.InitialPenalty,
		DistanceFromBest: 0,
		PlateauLength:    0,
		RecentImprovRate: 0,
		WorkerCount:      1,
		ActiveFamilies:   1,

		// Outcome.
		Improved:           outcome.BestPenalty < outcome.InitialPenalty,
		ProducedGlobalBest: outcome.BestPenalty < outcome.InitialPenalty,
		ImprovementAmount:  outcome.InitialPenalty - outcome.BestPenalty,
		FinalObjective:     outcome.BestPenalty,
		RuntimeMs:          outcome.DurationMs,
		CandidatesEval:     outcome.Candidates,
		Accepted:           outcome.Accepted,
		Rejected:           outcome.Rejected,
		PlateauCount:       0,
		BranchesSpawned:    0,
	}

	path := filepath.Join(outputDir, "worker_learning.csv")
	return WriteCSV(path, []Record{record})
}

// SingleWorkerConfig holds the parameters needed to emit a learning record
// for a single-search run (non-beam-search).
type SingleWorkerConfig struct {
	ProblemType string
	Instance    string
	Algorithm   string
	Seed        int64
	Temperature float64
	LAHCLength  int
	TabuTenure  int
	Iterations  int
}
