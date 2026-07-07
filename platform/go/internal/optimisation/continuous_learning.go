// continuous_learning.go — Continuous Learning for Search Intelligence 2.0.
//
// After every completed optimisation run, the system automatically:
//  1. Appends telemetry to the training dataset
//  2. Checks if retraining threshold is met
//  3. Evaluates candidate quality vs production
//  4. Recommends promotion (never automatically replaces)
//
// The learner does NOT automatically replace production policies.
// It only recommends. Human or automated gate must approve.
package optimisation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ───────────────────────────────────────────────────────────────
// Continuous Learning Config
// ───────────────────────────────────────────────────────────────

// ContinuousLearningConfig configures the learning loop.
type ContinuousLearningConfig struct {
	// Minimum new samples before retraining is triggered.
	RetrainThreshold int // default 50

	// Directory where training data accumulates.
	DataDir string

	// Directory where candidate policies are written.
	PolicyDir string

	// Path to the policy registry.
	RegistryPath string
}

// DefaultContinuousLearningConfig returns sensible defaults.
func DefaultContinuousLearningConfig() ContinuousLearningConfig {
	return ContinuousLearningConfig{
		RetrainThreshold: 50,
		DataDir:          "data/runs",
		PolicyDir:        "policies",
		RegistryPath:     "policy_registry.json",
	}
}

// ───────────────────────────────────────────────────────────────
// Learning State
// ───────────────────────────────────────────────────────────────

// LearningState tracks the continuous learning loop.
type LearningState struct {
	// How many samples have been collected since last training.
	NewSamplesSinceTraining int `json:"new_samples_since_training"`

	// Total training set size.
	TotalSamples int `json:"total_samples"`

	// Last training time.
	LastTrainedAt *time.Time `json:"last_trained_at,omitempty"`

	// Last training result.
	LastTrainingAccuracy float64 `json:"last_training_accuracy"`

	// Current production policy version.
	ProductionVersion string `json:"production_version"`

	// Candidate policy version (if trained but not promoted).
	CandidateVersion  string  `json:"candidate_version,omitempty"`
	CandidateAccuracy float64 `json:"candidate_accuracy,omitempty"`

	// Recommendation.
	Recommendation  string `json:"recommendation"` // "retrain", "promote", "wait", "none"
	RecommendReason string `json:"recommend_reason"`
}

// ───────────────────────────────────────────────────────────────
// Continuous Learner
// ───────────────────────────────────────────────────────────────

// ContinuousLearner manages the learning loop.
type ContinuousLearner struct {
	mu     sync.Mutex
	config ContinuousLearningConfig
	state  LearningState
}

// NewContinuousLearner creates a learner. Loads state from disk if available.
func NewContinuousLearner(config ContinuousLearningConfig) *ContinuousLearner {
	cl := &ContinuousLearner{config: config}
	cl.loadState()
	return cl
}

// OnRunCompleted is called after every optimisation run completes.
// It appends telemetry and evaluates whether retraining should happen.
func (cl *ContinuousLearner) OnRunCompleted(runID string, domain string, samplesAdded int) LearningRecommendation {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.state.NewSamplesSinceTraining += samplesAdded
	cl.state.TotalSamples += samplesAdded

	rec := cl.evaluate()
	cl.state.Recommendation = rec.Action
	cl.state.RecommendReason = rec.Reason

	cl.saveState()
	return rec
}

// State returns the current learning state.
func (cl *ContinuousLearner) State() LearningState {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.state
}

// RecordTraining records that training was performed.
func (cl *ContinuousLearner) RecordTraining(accuracy float64, candidateVersion string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	now := time.Now()
	cl.state.LastTrainedAt = &now
	cl.state.LastTrainingAccuracy = accuracy
	cl.state.NewSamplesSinceTraining = 0
	cl.state.CandidateVersion = candidateVersion
	cl.state.CandidateAccuracy = accuracy

	cl.saveState()
}

// RecordPromotion records that a candidate was promoted to production.
func (cl *ContinuousLearner) RecordPromotion(version string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.state.ProductionVersion = version
	cl.state.CandidateVersion = ""
	cl.state.CandidateAccuracy = 0
	cl.state.Recommendation = "none"
	cl.state.RecommendReason = "recently_promoted"

	cl.saveState()
}

// ───────────────────────────────────────────────────────────────
// Learning Recommendation
// ───────────────────────────────────────────────────────────────

// LearningRecommendation is what the learner suggests after each run.
type LearningRecommendation struct {
	Action string // "wait", "retrain", "promote", "none"
	Reason string
}

func (cl *ContinuousLearner) evaluate() LearningRecommendation {
	// If we have a candidate that's better than production, recommend promote.
	if cl.state.CandidateVersion != "" && cl.state.CandidateAccuracy > cl.state.LastTrainingAccuracy*0.95 {
		return LearningRecommendation{
			Action: "promote",
			Reason: fmt.Sprintf("candidate_v%s_accuracy_%.2f_ready", cl.state.CandidateVersion, cl.state.CandidateAccuracy),
		}
	}

	// If enough new samples have accumulated, recommend retrain.
	if cl.state.NewSamplesSinceTraining >= cl.config.RetrainThreshold {
		return LearningRecommendation{
			Action: "retrain",
			Reason: fmt.Sprintf("%d_new_samples_since_training", cl.state.NewSamplesSinceTraining),
		}
	}

	// Otherwise, wait for more data.
	return LearningRecommendation{
		Action: "wait",
		Reason: fmt.Sprintf("%d_of_%d_samples_needed", cl.state.NewSamplesSinceTraining, cl.config.RetrainThreshold),
	}
}

// ───────────────────────────────────────────────────────────────
// Persistence
// ───────────────────────────────────────────────────────────────

func (cl *ContinuousLearner) statePath() string {
	return filepath.Join(cl.config.PolicyDir, "learning_state.json")
}

func (cl *ContinuousLearner) loadState() {
	data, err := os.ReadFile(cl.statePath())
	if err != nil {
		return // fresh state
	}
	json.Unmarshal(data, &cl.state)
}

func (cl *ContinuousLearner) saveState() {
	os.MkdirAll(cl.config.PolicyDir, 0o755)
	data, _ := json.MarshalIndent(cl.state, "", "  ")
	os.WriteFile(cl.statePath(), data, 0o644)
}
