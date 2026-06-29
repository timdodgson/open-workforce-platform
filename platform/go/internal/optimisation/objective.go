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
func ObjectiveScore(assignments []assignment.Assignment, capacities []ResourceInput) int {
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
// Balance is measured by comparing remaining capacity (minutes) across resources.
// A plan where resources have similar remaining time scores higher than one
// where some resources are heavily loaded and others are idle.
//
// Formula: availableCount - (maxRemaining - minRemaining) / scale
// The bonus is deliberately small relative to assignment so it never
// sacrifices assignment count.
func balanceObjective(assignments []assignment.Assignment, capacities []ResourceInput) int {
	if len(assignments) == 0 {
		return 0
	}

	availableCount := 0
	for _, rc := range capacities {
		if rc.Available {
			availableCount++
		}
	}

	if availableCount <= 1 {
		return availableCount
	}

	// Count assigned items per available resource (simple count for bonus).
	loadOf := make(map[string]int)
	for _, rc := range capacities {
		if rc.Available {
			loadOf[rc.ResourceID] = 0
		}
	}
	for _, a := range assignments {
		if _, ok := loadOf[a.ResourceID()]; ok {
			loadOf[a.ResourceID()]++
		}
	}

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
func ObjectiveBreakdown(assignments []assignment.Assignment, capacities []ResourceInput) []ObjectiveContribution {
	return []ObjectiveContribution{
		{Name: "Assignment", Score: assignmentObjective(assignments, len(capacities))},
		{Name: "Workload Balance", Score: balanceObjective(assignments, capacities)},
	}
}
