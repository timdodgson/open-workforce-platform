// features.go — versioned feature vectors and persistence for SI policies.
package policy

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

// FeatureVector is the universal input to all Search Intelligence policies.
type FeatureVector struct {
	SchemaVersion      string    `json:"schemaVersion"`
	Timestamp          time.Time `json:"timestamp"`
	RunID              string    `json:"runId"`
	Problem            string    `json:"problem"`
	Instance           string    `json:"instance"`
	Algorithm          string    `json:"algorithm"`
	IterationBudget    int       `json:"iterationBudget"`
	IterationsComplete int       `json:"iterationsComplete"`
	BudgetConsumed     float64   `json:"budgetConsumed"`
	Temperature        float64   `json:"temperature"`
	CurrentObjective   int       `json:"currentObjective"`
	BestObjective      int       `json:"bestObjective"`
	ParentObjective    int       `json:"parentObjective"`
	DistanceFromBest   int       `json:"distanceFromBest"`
	GapToReference     float64   `json:"gapToReference"`
	PlateauLength      int       `json:"plateauLength"`
	ImprovementRate    float64   `json:"improvementRate"`
	AcceptanceRate     float64   `json:"acceptanceRate"`
	Diversity          float64   `json:"diversity"`
	Entropy            float64   `json:"entropy"`
	WorkerCount        int       `json:"workerCount"`
	BranchDepth        int       `json:"branchDepth"`
	Week               int       `json:"week"`
	BeamHealth         float64   `json:"beamHealth"`
	ElapsedMs          int64     `json:"elapsedMs"`
	TimeRatio          float64   `json:"timeRatio"`
	DecisionType       string    `json:"decisionType"`
}

// FeatureStore persists feature vectors for future retraining and analysis.
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
	Features     FeatureVector   `json:"features"`
	Action       string          `json:"action"`
	Confidence   float64         `json:"confidence"`
	PolicySource string          `json:"policySource"`
	Outcome      FeatureOutcome  `json:"outcome"`
}

// FeatureOutcome captures what happened after the decision.
type FeatureOutcome struct {
	Improved           bool    `json:"improved"`
	ImprovementAmount  int     `json:"improvementAmount"`
	FinalObjective     int     `json:"finalObjective"`
	ComputeUsed        int     `json:"computeUsed"`
	RuntimeMs          int64   `json:"runtimeMs"`
	ProducedGlobalBest bool    `json:"producedGlobalBest"`
	Regret             float64 `json:"regret"`
}
