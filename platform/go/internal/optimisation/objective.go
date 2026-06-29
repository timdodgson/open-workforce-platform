package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
)

// ObjectiveScore evaluates the quality of a candidate plan.
//
// It returns an additive score combining all objectives.
// Algorithms compare these scores to decide whether a move is an improvement.
// Higher is better.
//
// Algorithms should not know how individual objectives are calculated.
// They simply compare total scores.
func ObjectiveScore(assignments []assignment.Assignment, capacities []ResourceCapacity) int {
	return assignmentObjective(assignments, len(capacities)) +
		balanceObjective(assignments, capacities)
}

// assignmentObjective rewards assigning work items.
//
// This is the dominant objective. Each assigned item contributes 1000 points.
// This ensures that assigning more work is always preferred over improving balance.
func assignmentObjective(assignments []assignment.Assignment, _ int) int {
	return len(assignments) * 1000
}

// balanceObjective rewards even distribution of work across available resources.
//
// A perfectly balanced plan (all resources have the same load) scores the
// maximum balance bonus. Imbalanced plans score less.
//
// The bonus is deliberately small relative to assignment (max possible is
// the number of available resources) so it never sacrifices assignment count.
//
// Formula: availableCount - (maxLoad - minLoad)
// When perfectly balanced: maxLoad == minLoad, bonus = availableCount
// When maximally imbalanced: bonus approaches 0
func balanceObjective(assignments []assignment.Assignment, capacities []ResourceCapacity) int {
	if len(assignments) == 0 {
		return 0
	}

	// Count assignments per available resource.
	loadOf := make(map[string]int)
	availableCount := 0
	for _, rc := range capacities {
		if rc.Available {
			loadOf[rc.ResourceID] = 0
			availableCount++
		}
	}

	if availableCount == 0 {
		return 0
	}

	for _, a := range assignments {
		if _, ok := loadOf[a.ResourceID()]; ok {
			loadOf[a.ResourceID()]++
		}
	}

	// Find min and max load across available resources.
	minLoad := -1
	maxLoad := 0
	for _, load := range loadOf {
		if minLoad == -1 || load < minLoad {
			minLoad = load
		}
		if load > maxLoad {
			maxLoad = load
		}
	}

	if minLoad == -1 {
		minLoad = 0
	}

	spread := maxLoad - minLoad
	bonus := availableCount - spread
	if bonus < 0 {
		bonus = 0
	}

	return bonus
}

// ObjectiveContribution represents a named objective's contribution to the total score.
type ObjectiveContribution struct {
	Name  string
	Score int
}

// ObjectiveBreakdown returns the individual objective contributions.
//
// The total of all contributions equals the ObjectiveScore.
// The order is deterministic.
func ObjectiveBreakdown(assignments []assignment.Assignment, capacities []ResourceCapacity) []ObjectiveContribution {
	return []ObjectiveContribution{
		{Name: "Assignment", Score: assignmentObjective(assignments, len(capacities))},
		{Name: "Workload Balance", Score: balanceObjective(assignments, capacities)},
	}
}
