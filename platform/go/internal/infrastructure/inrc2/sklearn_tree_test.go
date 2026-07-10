package inrc2

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func TestSklearnTree_PositiveClassProbability(t *testing.T) {
	tree := &optimisation.SklearnTree{
		FeatureNames:  []string{"distance_from_best", "depth"},
		ChildrenLeft:  []int{1, -1, -1},
		ChildrenRight: []int{2, -1, -1},
		Feature:       []int{0, -2, -2},
		Threshold:     []float64{10, -2, -2},
		Value:         [][]float64{{}, {1, 3}, {4, 1}},
	}

	low := tree.PositiveClassProbability([]float64{5, 1})
	if low < 0.7 {
		t.Fatalf("expected high positive prob for low distance, got %.3f", low)
	}

	high := tree.PositiveClassProbability([]float64{50, 1})
	if high > 0.3 {
		t.Fatalf("expected low positive prob for high distance, got %.3f", high)
	}
}

func TestHybridWorkerDecisionEngine_FallsBackWithoutTree(t *testing.T) {
	engine := NewHybridWorkerDecisionEngine("hybrid", "/nonexistent/policy/dir")
	decision := engine.Evaluate(WorkerDecisionInput{
		GlobalBest:       100,
		ParentObjective:  200,
		DistanceFromBest: 100,
		AllocatedIters:   1000,
		WorkerCount:      100,
		Depth:            5,
	})
	if decision.Recommendation == "" {
		t.Fatal("expected rule-based fallback decision")
	}
}
