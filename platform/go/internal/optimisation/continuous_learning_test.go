package optimisation

import (
	"testing"
)

func testLearnerConfig(t *testing.T) ContinuousLearningConfig {
	return ContinuousLearningConfig{
		RetrainThreshold: 50,
		PolicyDir:        t.TempDir(),
	}
}

func TestContinuousLearner_InitialState(t *testing.T) {
	cl := NewContinuousLearner(testLearnerConfig(t))
	state := cl.State()

	if state.TotalSamples != 0 {
		t.Errorf("TotalSamples = %d, want 0", state.TotalSamples)
	}
	if state.Recommendation != "" {
		t.Errorf("Recommendation = %q, want empty", state.Recommendation)
	}
}

func TestContinuousLearner_AccumulatesSamples(t *testing.T) {
	cl := NewContinuousLearner(testLearnerConfig(t))

	cl.OnRunCompleted("run-1", "cvrp", 10)
	cl.OnRunCompleted("run-2", "cvrp", 15)

	state := cl.State()
	if state.TotalSamples != 25 {
		t.Errorf("TotalSamples = %d, want 25", state.TotalSamples)
	}
	if state.NewSamplesSinceTraining != 25 {
		t.Errorf("NewSamplesSinceTraining = %d, want 25", state.NewSamplesSinceTraining)
	}
}

func TestContinuousLearner_RecommendsWait(t *testing.T) {
	cl := NewContinuousLearner(testLearnerConfig(t))

	rec := cl.OnRunCompleted("run-1", "cvrp", 10)

	if rec.Action != "wait" {
		t.Errorf("Action = %q, want wait (only 10 samples)", rec.Action)
	}
}

func TestContinuousLearner_RecommendsRetrain(t *testing.T) {
	cl := NewContinuousLearner(testLearnerConfig(t))

	// Add enough samples to trigger retrain.
	for i := 0; i < 5; i++ {
		cl.OnRunCompleted("run", "cvrp", 10)
	}

	rec := cl.OnRunCompleted("run-final", "cvrp", 5)

	if rec.Action != "retrain" {
		t.Errorf("Action = %q, want retrain (55 samples >= 50 threshold)", rec.Action)
	}
}

func TestContinuousLearner_RecordTrainingResetsSamples(t *testing.T) {
	cl := NewContinuousLearner(testLearnerConfig(t))

	cl.OnRunCompleted("run-1", "cvrp", 60)
	cl.RecordTraining(0.78, "2.0.0")

	state := cl.State()
	if state.NewSamplesSinceTraining != 0 {
		t.Errorf("NewSamplesSinceTraining = %d, want 0 after training", state.NewSamplesSinceTraining)
	}
	if state.CandidateVersion != "2.0.0" {
		t.Errorf("CandidateVersion = %q, want 2.0.0", state.CandidateVersion)
	}
	if state.LastTrainingAccuracy != 0.78 {
		t.Errorf("LastTrainingAccuracy = %f, want 0.78", state.LastTrainingAccuracy)
	}
}

func TestContinuousLearner_RecommendsPromote(t *testing.T) {
	cl := NewContinuousLearner(testLearnerConfig(t))

	cl.RecordTraining(0.80, "2.0.0")

	// After training, next run should recommend promote (candidate exists).
	rec := cl.OnRunCompleted("run", "cvrp", 5)

	if rec.Action != "promote" {
		t.Errorf("Action = %q, want promote (candidate exists with good accuracy)", rec.Action)
	}
}

func TestContinuousLearner_RecordPromotionClears(t *testing.T) {
	cl := NewContinuousLearner(testLearnerConfig(t))

	cl.RecordTraining(0.80, "2.0.0")
	cl.RecordPromotion("2.0.0")

	state := cl.State()
	if state.ProductionVersion != "2.0.0" {
		t.Errorf("ProductionVersion = %q, want 2.0.0", state.ProductionVersion)
	}
	if state.CandidateVersion != "" {
		t.Errorf("CandidateVersion = %q, want empty after promotion", state.CandidateVersion)
	}
}

func TestContinuousLearner_PersistsState(t *testing.T) {
	config := testLearnerConfig(t)
	cl := NewContinuousLearner(config)

	cl.OnRunCompleted("run-1", "jss", 30)
	cl.RecordTraining(0.75, "1.0.0")

	// Create new learner from same dir — should load persisted state.
	cl2 := NewContinuousLearner(config)
	state := cl2.State()

	if state.TotalSamples != 30 {
		t.Errorf("persisted TotalSamples = %d, want 30", state.TotalSamples)
	}
	if state.CandidateVersion != "1.0.0" {
		t.Errorf("persisted CandidateVersion = %q, want 1.0.0", state.CandidateVersion)
	}
}

func TestContinuousLearner_NeverAutoPromotes(t *testing.T) {
	cl := NewContinuousLearner(testLearnerConfig(t))

	cl.RecordTraining(0.90, "3.0.0")
	rec := cl.OnRunCompleted("run", "cvrp", 5)

	// It recommends but never actually promotes.
	if rec.Action != "promote" {
		t.Errorf("should recommend promote")
	}

	// Production version should NOT be changed.
	state := cl.State()
	if state.ProductionVersion == "3.0.0" {
		t.Error("should NOT auto-promote — only recommend")
	}
}
