package optimisation_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func makeTestAssignment(resourceID, workItemID string) assignment.Assignment {
	a, _ := assignment.New(resourceID, workItemID)
	return a
}

func TestObjectiveScore_MoreAssignmentsScoresHigher(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-001", 5, true, nil),
	}

	one := []assignment.Assignment{makeTestAssignment("RES-001", "WI-001")}
	two := []assignment.Assignment{makeTestAssignment("RES-001", "WI-001"), makeTestAssignment("RES-001", "WI-002")}

	scoreOne := optimisation.ObjectiveScore(one, optimisation.NewContext(nil, capacities, nil))
	scoreTwo := optimisation.ObjectiveScore(two, optimisation.NewContext(nil, capacities, nil))

	if scoreTwo <= scoreOne {
		t.Errorf("expected more assignments to score higher: %d vs %d", scoreTwo, scoreOne)
	}
}

func TestObjectiveScore_AssignmentDominatesBalance(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-001", 5, true, nil),
		makeCapacity("RES-002", 5, true, nil),
	}

	// 3 items, imbalanced (3+0): higher assignment count
	threeImbalanced := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-001", "WI-002"),
		makeTestAssignment("RES-001", "WI-003"),
	}

	// 2 items, balanced (1+1): better balance but fewer assignments
	twoBalanced := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
	}

	scoreThree := optimisation.ObjectiveScore(threeImbalanced, optimisation.NewContext(nil, capacities, nil))
	scoreTwo := optimisation.ObjectiveScore(twoBalanced, optimisation.NewContext(nil, capacities, nil))

	if scoreTwo >= scoreThree {
		t.Errorf("assignment should dominate balance: 3 imbalanced=%d, 2 balanced=%d", scoreThree, scoreTwo)
	}
}

func TestObjectiveScore_BalancedBetterThanImbalanced(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-001", 5, true, nil),
		makeCapacity("RES-002", 5, true, nil),
	}

	// Same number of assignments, different distribution.
	balanced := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
	}
	imbalanced := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-001", "WI-002"),
	}

	scoreBalanced := optimisation.ObjectiveScore(balanced, optimisation.NewContext(nil, capacities, nil))
	scoreImbalanced := optimisation.ObjectiveScore(imbalanced, optimisation.NewContext(nil, capacities, nil))

	if scoreBalanced <= scoreImbalanced {
		t.Errorf("balanced should score higher: balanced=%d, imbalanced=%d", scoreBalanced, scoreImbalanced)
	}
}

func TestObjectiveScore_Deterministic(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-001", 3, true, nil),
		makeCapacity("RES-002", 3, true, nil),
	}
	assignments := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
	}

	score1 := optimisation.ObjectiveScore(assignments, optimisation.NewContext(nil, capacities, nil))
	score2 := optimisation.ObjectiveScore(assignments, optimisation.NewContext(nil, capacities, nil))

	if score1 != score2 {
		t.Errorf("scoring should be deterministic: %d vs %d", score1, score2)
	}
}

func TestObjectiveScore_EmptyAssignments(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-001", 3, true, nil),
	}

	score := optimisation.ObjectiveScore([]assignment.Assignment{}, optimisation.NewContext(nil, capacities, nil))
	if score != 0 {
		t.Errorf("expected 0 for empty assignments, got %d", score)
	}
}

func TestObjectiveBreakdown_SumEqualsTotal(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-001", 3, true, nil),
		makeCapacity("RES-002", 3, true, nil),
	}
	assignments := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
		makeTestAssignment("RES-001", "WI-003"),
	}

	total := optimisation.ObjectiveScore(assignments, optimisation.NewContext(nil, capacities, nil))
	breakdown := optimisation.ObjectiveBreakdown(assignments, optimisation.NewContext(nil, capacities, nil))

	sum := 0
	for _, entry := range breakdown {
		sum += entry.Score
	}

	if sum != total {
		t.Errorf("breakdown sum %d does not equal total %d", sum, total)
	}
}

func TestObjectiveBreakdown_ContainsExpectedObjectives(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-001", 3, true, nil),
	}
	assignments := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
	}

	breakdown := optimisation.ObjectiveBreakdown(assignments, optimisation.NewContext(nil, capacities, nil))

	if len(breakdown) != 5 {
		t.Fatalf("expected 5 objectives, got %d", len(breakdown))
	}
	if breakdown[0].Name != "Assignment" {
		t.Errorf("expected first objective 'Assignment', got %q", breakdown[0].Name)
	}
	if breakdown[1].Name != "Workload Balance" {
		t.Errorf("expected second objective 'Workload Balance', got %q", breakdown[1].Name)
	}
	if breakdown[2].Name != "Travel Time" {
		t.Errorf("expected third objective 'Travel Time', got %q", breakdown[2].Name)
	}
	if breakdown[3].Name != "Preferred Resource" {
		t.Errorf("expected fourth objective 'Preferred Resource', got %q", breakdown[3].Name)
	}
	if breakdown[4].Name != "Plan Stability" {
		t.Errorf("expected fifth objective 'Plan Stability', got %q", breakdown[4].Name)
	}
}

func TestObjectiveBreakdown_AssignmentContribution(t *testing.T) {
	capacities := []optimisation.ResourceInput{
		makeCapacity("RES-001", 5, true, nil),
	}
	assignments := []assignment.Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-001", "WI-002"),
	}

	breakdown := optimisation.ObjectiveBreakdown(assignments, optimisation.NewContext(nil, capacities, nil))

	// 2 items × 1000 = 2000
	if breakdown[0].Score != 2000 {
		t.Errorf("expected assignment contribution 2000, got %d", breakdown[0].Score)
	}
}
