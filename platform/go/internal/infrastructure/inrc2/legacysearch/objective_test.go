package legacysearch

import (
	"testing"

)

func makeTestAssignment(resourceID, workItemID string) Assignment {
	a, _ := NewAssignment(resourceID, workItemID)
	return a
}

func TestObjectiveScore_MoreAssignmentsScoresHigher(t *testing.T) {
	capacities := []ResourceInput{
		makeCapacity("RES-001", 5, true, nil),
	}

	one := []Assignment{makeTestAssignment("RES-001", "WI-001")}
	two := []Assignment{makeTestAssignment("RES-001", "WI-001"), makeTestAssignment("RES-001", "WI-002")}

	scoreOne := ObjectiveScore(one, NewContext(nil, capacities, nil))
	scoreTwo := ObjectiveScore(two, NewContext(nil, capacities, nil))

	if scoreTwo <= scoreOne {
		t.Errorf("expected more assignments to score higher: %d vs %d", scoreTwo, scoreOne)
	}
}

func TestObjectiveScore_AssignmentDominatesBalance(t *testing.T) {
	capacities := []ResourceInput{
		makeCapacity("RES-001", 5, true, nil),
		makeCapacity("RES-002", 5, true, nil),
	}

	// 3 items, imbalanced (3+0): higher assignment count
	threeImbalanced := []Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-001", "WI-002"),
		makeTestAssignment("RES-001", "WI-003"),
	}

	// 2 items, balanced (1+1): better balance but fewer assignments
	twoBalanced := []Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
	}

	scoreThree := ObjectiveScore(threeImbalanced, NewContext(nil, capacities, nil))
	scoreTwo := ObjectiveScore(twoBalanced, NewContext(nil, capacities, nil))

	if scoreTwo >= scoreThree {
		t.Errorf("assignment should dominate balance: 3 imbalanced=%d, 2 balanced=%d", scoreThree, scoreTwo)
	}
}

func TestObjectiveScore_BalancedBetterThanImbalanced(t *testing.T) {
	capacities := []ResourceInput{
		makeCapacity("RES-001", 5, true, nil),
		makeCapacity("RES-002", 5, true, nil),
	}

	// Same number of assignments, different distribution.
	balanced := []Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
	}
	imbalanced := []Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-001", "WI-002"),
	}

	scoreBalanced := ObjectiveScore(balanced, NewContext(nil, capacities, nil))
	scoreImbalanced := ObjectiveScore(imbalanced, NewContext(nil, capacities, nil))

	if scoreBalanced <= scoreImbalanced {
		t.Errorf("balanced should score higher: balanced=%d, imbalanced=%d", scoreBalanced, scoreImbalanced)
	}
}

func TestObjectiveScore_Deterministic(t *testing.T) {
	capacities := []ResourceInput{
		makeCapacity("RES-001", 3, true, nil),
		makeCapacity("RES-002", 3, true, nil),
	}
	assignments := []Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
	}

	score1 := ObjectiveScore(assignments, NewContext(nil, capacities, nil))
	score2 := ObjectiveScore(assignments, NewContext(nil, capacities, nil))

	if score1 != score2 {
		t.Errorf("scoring should be deterministic: %d vs %d", score1, score2)
	}
}

func TestObjectiveScore_EmptyAssignments(t *testing.T) {
	capacities := []ResourceInput{
		makeCapacity("RES-001", 3, true, nil),
	}

	score := ObjectiveScore([]Assignment{}, NewContext(nil, capacities, nil))
	if score != 0 {
		t.Errorf("expected 0 for empty assignments, got %d", score)
	}
}

func TestObjectiveBreakdown_SumEqualsTotal(t *testing.T) {
	capacities := []ResourceInput{
		makeCapacity("RES-001", 3, true, nil),
		makeCapacity("RES-002", 3, true, nil),
	}
	assignments := []Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-002", "WI-002"),
		makeTestAssignment("RES-001", "WI-003"),
	}

	total := ObjectiveScore(assignments, NewContext(nil, capacities, nil))
	breakdown := ObjectiveBreakdown(assignments, NewContext(nil, capacities, nil))

	sum := 0
	for _, entry := range breakdown {
		sum += entry.Score
	}

	if sum != total {
		t.Errorf("breakdown sum %d does not equal total %d", sum, total)
	}
}

func TestObjectiveBreakdown_ContainsExpectedObjectives(t *testing.T) {
	capacities := []ResourceInput{
		makeCapacity("RES-001", 3, true, nil),
	}
	assignments := []Assignment{
		makeTestAssignment("RES-001", "WI-001"),
	}

	breakdown := ObjectiveBreakdown(assignments, NewContext(nil, capacities, nil))

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
	capacities := []ResourceInput{
		makeCapacity("RES-001", 5, true, nil),
	}
	assignments := []Assignment{
		makeTestAssignment("RES-001", "WI-001"),
		makeTestAssignment("RES-001", "WI-002"),
	}

	breakdown := ObjectiveBreakdown(assignments, NewContext(nil, capacities, nil))

	// 2 items × 1000 = 2000
	if breakdown[0].Score != 2000 {
		t.Errorf("expected assignment contribution 2000, got %d", breakdown[0].Score)
	}
}
