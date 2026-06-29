package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
)

// ObjectiveScore evaluates the quality of a candidate plan.
//
// It returns an additive score combining all objectives.
// Algorithms compare these scores to decide whether a move is an improvement.
// Higher is better.
func ObjectiveScore(assignments []assignment.Assignment, ctx OptimisationContext) int {
	capacities := ctx.Resources()
	return assignmentObjective(assignments) +
		balanceObjective(assignments, capacities) +
		travelObjective(assignments, ctx) +
		preferredResourceObjective(assignments, ctx) +
		stabilityObjective(assignments, ctx)
}

// assignmentObjective rewards assigning work items.
//
// This is the dominant objective. Each assigned item contributes 1000 points.
func assignmentObjective(assignments []assignment.Assignment) int {
	return len(assignments) * 1000
}

// balanceObjective rewards even distribution of work across available resources.
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

// travelObjective penalises travel time.
//
// For each resource, calculates total travel from its starting location through
// each assigned work item in order. Returns a negative value (penalty).
// 1 point penalty per minute of travel.
func travelObjective(assignments []assignment.Assignment, ctx OptimisationContext) int {
	travel := ctx.TravelMatrix()
	if len(travel) == 0 {
		return 0
	}

	// Build travel lookup: "from|to" -> minutes
	travelLookup := make(map[string]int, len(travel))
	for _, t := range travel {
		travelLookup[t.From+"|"+t.To] = t.Minutes
	}

	// Build location lookups.
	resources := ctx.Resources()
	workItems := ctx.WorkItems()

	resourceLocation := make(map[string]string, len(resources))
	for _, r := range resources {
		resourceLocation[r.ResourceID] = r.Location
	}

	itemLocation := make(map[string]string, len(workItems))
	for _, w := range workItems {
		itemLocation[w.WorkItemID] = w.Location
	}

	// Group assignments by resource (in order).
	byResource := make(map[string][]string)
	for _, a := range assignments {
		byResource[a.ResourceID()] = append(byResource[a.ResourceID()], a.WorkItemID())
	}

	// Calculate total travel.
	totalTravel := 0
	for resID, itemIDs := range byResource {
		current := resourceLocation[resID]
		for _, itemID := range itemIDs {
			dest := itemLocation[itemID]
			if dest != "" && current != "" && dest != current {
				key := current + "|" + dest
				if minutes, ok := travelLookup[key]; ok {
					totalTravel += minutes
				}
			}
			if dest != "" {
				current = dest
			}
		}
	}

	return -totalTravel
}

// ObjectiveContribution represents a named objective's contribution to the total score.
type ObjectiveContribution struct {
	Name  string
	Score int
}

// ObjectiveBreakdown returns the individual objective contributions.
//
// The total of all contributions equals the ObjectiveScore.
func ObjectiveBreakdown(assignments []assignment.Assignment, ctx OptimisationContext) []ObjectiveContribution {
	capacities := ctx.Resources()
	return []ObjectiveContribution{
		{Name: "Assignment", Score: assignmentObjective(assignments)},
		{Name: "Workload Balance", Score: balanceObjective(assignments, capacities)},
		{Name: "Travel Time", Score: travelObjective(assignments, ctx)},
		{Name: "Preferred Resource", Score: preferredResourceObjective(assignments, ctx)},
		{Name: "Plan Stability", Score: stabilityObjective(assignments, ctx)},
	}
}

// preferredResourceObjective rewards assignments that match a work item's
// preferred resource.
//
// Each matched preference contributes 25 points.
// This is deliberately smaller than the assignment objective (1000 per item)
// so the optimiser never leaves work unassigned to satisfy a preference.
func preferredResourceObjective(assignments []assignment.Assignment, ctx OptimisationContext) int {
	workItems := ctx.WorkItems()

	preferredOf := make(map[string]string, len(workItems))
	for _, w := range workItems {
		if w.PreferredResource != "" {
			preferredOf[w.WorkItemID] = w.PreferredResource
		}
	}

	if len(preferredOf) == 0 {
		return 0
	}

	bonus := 0
	for _, a := range assignments {
		if pref, ok := preferredOf[a.WorkItemID()]; ok && pref == a.ResourceID() {
			bonus += 25
		}
	}

	return bonus
}

// stabilityObjective rewards preserving existing assignments.
//
// Each assignment that matches the existing plan (same item on same resource)
// contributes 10 points. If no existing plan is provided, returns 0.
// This is deliberately much smaller than assignment (1000) so the optimiser
// never refuses to assign work to preserve stability.
func stabilityObjective(assignments []assignment.Assignment, ctx OptimisationContext) int {
	existing := ctx.ExistingAssignments()
	if len(existing) == 0 {
		return 0
	}

	// Build lookup: workItemID -> resourceID from existing plan.
	previousResource := make(map[string]string, len(existing))
	for _, a := range existing {
		previousResource[a.WorkItemID()] = a.ResourceID()
	}

	bonus := 0
	for _, a := range assignments {
		if prev, ok := previousResource[a.WorkItemID()]; ok && prev == a.ResourceID() {
			bonus += 10
		}
	}

	return bonus
}
