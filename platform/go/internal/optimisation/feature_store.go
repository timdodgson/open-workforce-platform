// feature_store.go — Reusable Search Intelligence Feature Store.
//
// Every policy consumes FeatureVectors rather than raw telemetry.
// The Feature Store provides:
//   - FeatureVector: a versioned, domain-agnostic feature representation
//   - FeatureExtractor: builds FeatureVectors from context-specific inputs
//   - FeatureStore: persists vectors for retraining and analysis
//
// Feature schema is versioned. Changes to the schema increment the version.
// Historical vectors retain their original schema version for compatibility.
package optimisation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FeatureSchemaVersion tracks the current feature vector schema.
// Increment when adding, removing, or changing feature semantics.
const FeatureSchemaVersion = "1.0.0"

// ───────────────────────────────────────────────────────────────
// FeatureVector
// ───────────────────────────────────────────────────────────────

// FeatureVector is the universal input to all Search Intelligence policies.
// It captures the full decision context in a structured, versioned format.
type FeatureVector struct {
	// Schema metadata
	SchemaVersion string    `json:"schemaVersion"`
	Timestamp     time.Time `json:"timestamp"`
	RunID         string    `json:"runId"`

	// Problem context
	Problem   string `json:"problem"`   // nrp, cvrp, jss, vrptw
	Instance  string `json:"instance"`  // e.g. A-n32-k5, la01, C101, n012w8
	Algorithm string `json:"algorithm"` // sa, lahc, tabu, portfolio, adaptive

	// Budget and progress
	IterationBudget    int     `json:"iterationBudget"`
	IterationsComplete int     `json:"iterationsComplete"`
	BudgetConsumed     float64 `json:"budgetConsumed"` // 0.0–1.0

	// Temperature (SA-specific, 0 for others)
	Temperature float64 `json:"temperature"`

	// Objective features
	CurrentObjective int     `json:"currentObjective"`
	BestObjective    int     `json:"bestObjective"`
	ParentObjective  int     `json:"parentObjective"`
	DistanceFromBest int     `json:"distanceFromBest"`
	GapToReference   float64 `json:"gapToReference"` // gap to best-known, -1 if unknown

	// Search dynamics
	PlateauLength   int     `json:"plateauLength"`   // iterations since last improvement
	ImprovementRate float64 `json:"improvementRate"` // improvements per 10K iterations
	AcceptanceRate  float64 `json:"acceptanceRate"`  // accepted / evaluated

	// Beam search features (NRP-specific, zero for others)
	Diversity   float64 `json:"diversity"`   // near-duplicate ratio (0 = all unique)
	Entropy     float64 `json:"entropy"`     // lineage entropy (Shannon)
	WorkerCount int     `json:"workerCount"` // active/total workers
	BranchDepth int     `json:"branchDepth"` // depth in search tree
	Week        int     `json:"week"`        // planning week (1–8)
	BeamHealth  float64 `json:"beamHealth"`  // composite 0–100

	// Timing
	ElapsedMs int64   `json:"elapsedMs"`
	TimeRatio float64 `json:"timeRatio"` // elapsed / total expected

	// Decision metadata (filled by caller)
	DecisionType string `json:"decisionType"` // worker, portfolio, search
}

// ───────────────────────────────────────────────────────────────
// FeatureExtractor
// ───────────────────────────────────────────────────────────────

// FeatureExtractor builds FeatureVectors from context-specific inputs.
// Each integration style (worker, portfolio, search) has a dedicated method.
type FeatureExtractor struct{}

// NewFeatureExtractor creates a FeatureExtractor.
func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{}
}

// FromWorkerContext builds a FeatureVector from a WorkerContext.
func (fe *FeatureExtractor) FromWorkerContext(ctx WorkerContext, runID string, instance string) FeatureVector {
	budgetConsumed := 0.0
	if ctx.AllocatedIters > 0 {
		budgetConsumed = 0.0 // at spawn time, worker hasn't started
	}

	return FeatureVector{
		SchemaVersion:      FeatureSchemaVersion,
		Timestamp:          time.Now(),
		RunID:              runID,
		Problem:            "nrp",
		Instance:           instance,
		Algorithm:          ctx.Algorithm,
		IterationBudget:    ctx.AllocatedIters,
		IterationsComplete: 0,
		BudgetConsumed:     budgetConsumed,
		Temperature:        0, // unknown at spawn
		CurrentObjective:   ctx.ParentObjective,
		BestObjective:      ctx.GlobalBest,
		ParentObjective:    ctx.ParentObjective,
		DistanceFromBest:   ctx.DistanceFromBest,
		GapToReference:     -1,
		PlateauLength:      0,
		ImprovementRate:    ctx.RecentImprovRate,
		AcceptanceRate:     0,
		Diversity:          0,
		Entropy:            ctx.Entropy,
		WorkerCount:        ctx.WorkerCount,
		BranchDepth:        ctx.Depth,
		Week:               ctx.Week,
		BeamHealth:         ctx.BeamHealth,
		ElapsedMs:          0,
		TimeRatio:          0,
		DecisionType:       "worker",
	}
}

// FromSearchProgress builds a FeatureVector from SearchProgress.
func (fe *FeatureExtractor) FromSearchProgress(
	p SearchProgress, runID string, problem string, instance string, elapsedMs int64,
) FeatureVector {
	budgetConsumed := 0.0
	if p.IterationsTotal > 0 {
		budgetConsumed = float64(p.IterationsComplete) / float64(p.IterationsTotal)
	}
	acceptanceRate := 0.0
	if p.CandidatesEval > 0 {
		acceptanceRate = float64(p.Accepted) / float64(p.CandidatesEval)
	}

	return FeatureVector{
		SchemaVersion:      FeatureSchemaVersion,
		Timestamp:          time.Now(),
		RunID:              runID,
		Problem:            problem,
		Instance:           instance,
		Algorithm:          p.Algorithm,
		IterationBudget:    p.IterationsTotal,
		IterationsComplete: p.IterationsComplete,
		BudgetConsumed:     budgetConsumed,
		Temperature:        p.Temperature,
		CurrentObjective:   p.CurrentPenalty,
		BestObjective:      p.BestPenalty,
		ParentObjective:    p.InitialPenalty,
		DistanceFromBest:   p.CurrentPenalty - p.BestPenalty,
		GapToReference:     -1,
		PlateauLength:      p.PlateauLength,
		ImprovementRate:    p.ImprovementRate,
		AcceptanceRate:     acceptanceRate,
		Diversity:          0,
		Entropy:            0,
		WorkerCount:        0,
		BranchDepth:        0,
		Week:               0,
		BeamHealth:         0,
		ElapsedMs:          elapsedMs,
		TimeRatio:          0,
		DecisionType:       "search",
	}
}

// FromPortfolioContext builds a FeatureVector for a specific strategy within a portfolio.
func (fe *FeatureExtractor) FromPortfolioContext(
	ctx PortfolioContext, strategy string, runID string,
) FeatureVector {
	// Compute historical win rate for this strategy.
	wins := 0
	total := 0
	for _, entry := range ctx.PreviousResults {
		if entry.Strategy == strategy {
			total++
			if entry.Improved {
				wins++
			}
		}
	}
	winRate := 0.0
	if total > 0 {
		winRate = float64(wins) / float64(total)
	}

	return FeatureVector{
		SchemaVersion:      FeatureSchemaVersion,
		Timestamp:          time.Now(),
		RunID:              runID,
		Problem:            ctx.ProblemType,
		Instance:           ctx.Instance,
		Algorithm:          strategy,
		IterationBudget:    ctx.TotalBudget,
		IterationsComplete: 0,
		BudgetConsumed:     0,
		Temperature:        0,
		CurrentObjective:   0,
		BestObjective:      0,
		ParentObjective:    0,
		DistanceFromBest:   0,
		GapToReference:     winRate, // overload: strategy win rate stored here
		PlateauLength:      0,
		ImprovementRate:    0,
		AcceptanceRate:     0,
		Diversity:          0,
		Entropy:            0,
		WorkerCount:        len(ctx.Strategies),
		BranchDepth:        0,
		Week:               0,
		BeamHealth:         0,
		ElapsedMs:          0,
		TimeRatio:          0,
		DecisionType:       "portfolio",
	}
}

// ───────────────────────────────────────────────────────────────
// FeatureStore
// ───────────────────────────────────────────────────────────────

// FeatureStore persists feature vectors for future retraining and analysis.
// Vectors are written to a JSONL file (one JSON object per line).
// Thread-safe for concurrent writes from parallel workers.
type FeatureStore struct {
	mu       sync.Mutex
	dir      string
	file     *os.File
	count    int
	disabled bool
}

// NewFeatureStore creates a store that writes to the given directory.
// If dir is empty, the store is disabled (no-op).
func NewFeatureStore(dir string) *FeatureStore {
	if dir == "" {
		return &FeatureStore{disabled: true}
	}
	return &FeatureStore{dir: dir}
}

// Record persists a FeatureVector with its associated decision and outcome.
func (fs *FeatureStore) Record(entry FeatureRecord) error {
	if fs.disabled {
		return nil
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.file == nil {
		if err := fs.open(); err != nil {
			return err
		}
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("feature_store: marshal error: %w", err)
	}

	if _, err := fs.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("feature_store: write error: %w", err)
	}

	fs.count++
	return nil
}

// Count returns the number of records written.
func (fs *FeatureStore) Count() int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.count
}

// Close flushes and closes the underlying file.
func (fs *FeatureStore) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.file != nil {
		return fs.file.Close()
	}
	return nil
}

func (fs *FeatureStore) open() error {
	if err := os.MkdirAll(fs.dir, 0o755); err != nil {
		return fmt.Errorf("feature_store: mkdir error: %w", err)
	}
	filename := filepath.Join(fs.dir, "features.jsonl")
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("feature_store: open error: %w", err)
	}
	fs.file = f
	return nil
}

// FeatureRecord is a single persisted entry: the feature vector plus
// the decision made and the observed outcome (filled later if available).
type FeatureRecord struct {
	// The feature vector at decision time.
	Features FeatureVector `json:"features"`

	// Decision made by the policy.
	Action       string  `json:"action"`
	Confidence   float64 `json:"confidence"`
	PolicySource string  `json:"policySource"` // "rule", "learned", "hybrid"

	// Outcome (filled after run completes, zero values if not yet known).
	Outcome FeatureOutcome `json:"outcome"`
}

// FeatureOutcome captures what happened after the decision.
type FeatureOutcome struct {
	Improved           bool    `json:"improved"`
	ImprovementAmount  int     `json:"improvementAmount"`
	FinalObjective     int     `json:"finalObjective"`
	ComputeUsed        int     `json:"computeUsed"`
	RuntimeMs          int64   `json:"runtimeMs"`
	ProducedGlobalBest bool    `json:"producedGlobalBest"`
	Regret             float64 `json:"regret"` // estimated regret vs counterfactual
}
