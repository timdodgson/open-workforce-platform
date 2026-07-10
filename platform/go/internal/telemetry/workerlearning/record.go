package workerlearning

import (
	"fmt"
	"os"
	"strings"
)

// Record captures one completed worker's full context for ML training.
// One row per completed worker. Pure observation — does not change optimiser behaviour.
type Record struct {
	// --- Run Metadata ---
	ProblemType string
	Instance    string
	Algorithm   string
	RunSeed     int64

	// --- Spawn State ---
	Week            int
	Phase           string // "search", "refinement"
	Depth           int
	ParentWorkerID  int
	FamilyID        int
	BeamRank        int     // position in beam (0 = best)
	BeamScore       int     // objective of beam path at spawn
	Entropy         float64 // lineage entropy at spawn
	Diversity       float64 // near-duplicate rate at spawn
	BeamHealth      float64 // retained / total ratio
	Temperature     float64 // initial temperature for this worker
	LAHCLength      int     // LAHC buffer length (0 if not LAHC)
	TabuTenure      int     // tabu tenure (0 if not tabu)
	IterationsAlloc int     // iteration budget allocated

	// --- Environment at Spawn ---
	GlobalBest       int     // current global best objective
	ParentObjective  int     // parent worker's best objective
	DistanceFromBest int     // parentObjective - globalBest
	PlateauLength    int     // iterations since last global improvement
	RecentImprovRate float64 // improvements per 10K candidates (recent window)
	WorkerCount      int     // total workers spawned so far
	ActiveFamilies   int     // distinct family lineages still active

	// --- Outcome ---
	Improved           bool
	ProducedGlobalBest bool
	ImprovementAmount  int // startObjective - bestObjective (0 if no improvement)
	FinalObjective     int
	RuntimeMs          int64
	CandidatesEval     int
	Accepted           int
	Rejected           int
	PlateauCount       int
	BranchesSpawned    int

	// --- Derived (computed at emit time) ---
	ROI           float64 // improvementAmount / max(runtimeMs, 1)
	ImprovPerCPU  float64 // improvementAmount / max(runtimeMs, 1) * 1000
	ImprovPer100K float64 // improvementAmount / max(candidatesEval/100000, 1)
}

// ComputeDerived fills the derived fields from the outcome fields.
func (r *Record) ComputeDerived() {
	if r.RuntimeMs > 0 {
		r.ROI = float64(r.ImprovementAmount) / float64(r.RuntimeMs)
		r.ImprovPerCPU = float64(r.ImprovementAmount) / float64(r.RuntimeMs) * 1000
	}
	cands100K := float64(r.CandidatesEval) / 100000.0
	if cands100K > 0 {
		r.ImprovPer100K = float64(r.ImprovementAmount) / cands100K
	}
}

// CSVHeader returns the header row for worker_learning.csv.
func CSVHeader() string {
	cols := []string{
		// Run metadata
		"problem_type", "instance", "algorithm", "run_seed",
		// Spawn state
		"week", "phase", "depth", "parent_worker_id", "family_id",
		"beam_rank", "beam_score", "entropy", "diversity", "beam_health",
		"temperature", "lahc_length", "tabu_tenure", "iterations_alloc",
		// Environment at spawn
		"global_best", "parent_objective", "distance_from_best",
		"plateau_length", "recent_improv_rate", "worker_count", "active_families",
		// Outcome
		"improved", "produced_global_best", "improvement_amount",
		"final_objective", "runtime_ms", "candidates_eval",
		"accepted", "rejected", "plateau_count", "branches_spawned",
		// Derived
		"roi", "improv_per_cpu", "improv_per_100k",
	}
	return strings.Join(cols, ",")
}

// CSVRow formats a record as a CSV row.
func CSVRow(r Record) string {
	improved := 0
	if r.Improved {
		improved = 1
	}
	pgb := 0
	if r.ProducedGlobalBest {
		pgb = 1
	}

	fields := []string{
		// Run metadata
		r.ProblemType, r.Instance, r.Algorithm, fmt.Sprintf("%d", r.RunSeed),
		// Spawn state
		fmt.Sprintf("%d", r.Week), r.Phase, fmt.Sprintf("%d", r.Depth),
		fmt.Sprintf("%d", r.ParentWorkerID), fmt.Sprintf("%d", r.FamilyID),
		fmt.Sprintf("%d", r.BeamRank), fmt.Sprintf("%d", r.BeamScore),
		fmt.Sprintf("%.4f", r.Entropy), fmt.Sprintf("%.4f", r.Diversity),
		fmt.Sprintf("%.4f", r.BeamHealth),
		fmt.Sprintf("%.4f", r.Temperature), fmt.Sprintf("%d", r.LAHCLength),
		fmt.Sprintf("%d", r.TabuTenure), fmt.Sprintf("%d", r.IterationsAlloc),
		// Environment at spawn
		fmt.Sprintf("%d", r.GlobalBest), fmt.Sprintf("%d", r.ParentObjective),
		fmt.Sprintf("%d", r.DistanceFromBest),
		fmt.Sprintf("%d", r.PlateauLength), fmt.Sprintf("%.4f", r.RecentImprovRate),
		fmt.Sprintf("%d", r.WorkerCount), fmt.Sprintf("%d", r.ActiveFamilies),
		// Outcome
		fmt.Sprintf("%d", improved), fmt.Sprintf("%d", pgb),
		fmt.Sprintf("%d", r.ImprovementAmount),
		fmt.Sprintf("%d", r.FinalObjective), fmt.Sprintf("%d", r.RuntimeMs),
		fmt.Sprintf("%d", r.CandidatesEval),
		fmt.Sprintf("%d", r.Accepted), fmt.Sprintf("%d", r.Rejected),
		fmt.Sprintf("%d", r.PlateauCount), fmt.Sprintf("%d", r.BranchesSpawned),
		// Derived
		fmt.Sprintf("%.6f", r.ROI), fmt.Sprintf("%.4f", r.ImprovPerCPU),
		fmt.Sprintf("%.4f", r.ImprovPer100K),
	}
	return strings.Join(fields, ",")
}

// WriteCSV writes a complete worker_learning.csv file.
func WriteCSV(path string, records []Record) error {
	if len(records) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(CSVHeader())
	sb.WriteString("\n")
	for _, r := range records {
		r.ComputeDerived()
		sb.WriteString(CSVRow(r))
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

