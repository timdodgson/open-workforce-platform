package workerlearning

import (
	"path/filepath"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// EmitSingleWorkerLearning creates a worker_learning.csv for a single-search run
// (CVRP, JSS, VRPTW). These runs have exactly one "worker" — the search itself.
// This records the run as a training example for future ML models.
func EmitSingleWorkerLearning(outputDir string, cfg SingleWorkerConfig, result optimisation.SearchResult) error {
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
		BeamScore:       result.InitialPenalty,
		Entropy:         0,
		Diversity:       0,
		BeamHealth:      1.0,
		Temperature:     cfg.Temperature,
		LAHCLength:      cfg.LAHCLength,
		TabuTenure:      cfg.TabuTenure,
		IterationsAlloc: cfg.Iterations,

		// Environment at spawn (single worker — it IS the global state).
		GlobalBest:       result.InitialPenalty,
		ParentObjective:  result.InitialPenalty,
		DistanceFromBest: 0,
		PlateauLength:    0,
		RecentImprovRate: 0,
		WorkerCount:      1,
		ActiveFamilies:   1,

		// Outcome.
		Improved:           result.BestPenalty < result.InitialPenalty,
		ProducedGlobalBest: result.BestPenalty < result.InitialPenalty,
		ImprovementAmount:  result.InitialPenalty - result.BestPenalty,
		FinalObjective:     result.BestPenalty,
		RuntimeMs:          result.DurationMs,
		CandidatesEval:     result.Candidates,
		Accepted:           result.Accepted,
		Rejected:           result.Rejected,
		PlateauCount:       0, // not tracked at this level
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

