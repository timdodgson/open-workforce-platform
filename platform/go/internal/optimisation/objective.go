package optimisation

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
)

// ObjectiveScore evaluates the quality of a candidate plan.
//
// It returns an additive score combining all objectives weighted by context weights.
// Algorithms compare these scores to decide whether a move is an improvement.
// Higher is better.
func ObjectiveScore(assignments []assignment.Assignment, ctx OptimisationContext) int {
	w := ctx.Weights()
	capacities := ctx.Resources()
	return rawAssignment(assignments)*w.Assignment +
		rawBalance(assignments, capacities)*w.Balance +
		rawTravel(assignments, ctx)*w.Travel +
		rawPreferredResource(assignments, ctx)*w.PreferredResource +
		rawStability(assignments, ctx)*w.PlanStability
}

// assignmentObjective is kept for backward compatibility in tests.
func assignmentObjective(assignments []assignment.Assignment) int {
	return len(assignments) * DefaultWeights().Assignment
}

// ObjectiveWeights defines the multiplier for each objective.
type ObjectiveWeights struct {
	Assignment        int
	Balance           int
	Travel            int
	PreferredResource int
	PlanStability     int
}

// DefaultWeights returns the default objective weights that reproduce original behaviour.
func DefaultWeights() ObjectiveWeights {
	return ObjectiveWeights{
		Assignment:        1000,
		Balance:           1,
		Travel:            -1,
		PreferredResource: 25,
		PlanStability:     10,
	}
}

// GetWeightProfile returns a named weight profile. Only "default" is currently supported.
func GetWeightProfile(name string) (ObjectiveWeights, bool) {
	switch name {
	case "default", "":
		return DefaultWeights(), true
	default:
		return ObjectiveWeights{}, false
	}
}

// --- Raw objective values (unweighted) ---

func rawAssignment(assignments []assignment.Assignment) int {
	return len(assignments)
}

func rawBalance(assignments []assignment.Assignment, capacities []ResourceInput) int {
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

func rawTravel(assignments []assignment.Assignment, ctx OptimisationContext) int {
	travel := ctx.TravelMatrix()
	if len(travel) == 0 {
		return 0
	}
	travelLookup := make(map[string]int, len(travel))
	for _, t := range travel {
		travelLookup[t.From+"|"+t.To] = t.Minutes
	}
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
	byResource := make(map[string][]string)
	for _, a := range assignments {
		byResource[a.ResourceID()] = append(byResource[a.ResourceID()], a.WorkItemID())
	}
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
	return totalTravel // positive; weight is negative
}

func rawPreferredResource(assignments []assignment.Assignment, ctx OptimisationContext) int {
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
	count := 0
	for _, a := range assignments {
		if pref, ok := preferredOf[a.WorkItemID()]; ok && pref == a.ResourceID() {
			count++
		}
	}
	return count
}

func rawStability(assignments []assignment.Assignment, ctx OptimisationContext) int {
	existing := ctx.ExistingAssignments()
	if len(existing) == 0 {
		return 0
	}
	previousResource := make(map[string]string, len(existing))
	for _, a := range existing {
		previousResource[a.WorkItemID()] = a.ResourceID()
	}
	count := 0
	for _, a := range assignments {
		if prev, ok := previousResource[a.WorkItemID()]; ok && prev == a.ResourceID() {
			count++
		}
	}
	return count
}


// ObjectiveContribution represents a named objective's contribution to the total score.
type ObjectiveContribution struct {
	Name  string
	Score int
}

// ObjectiveBreakdown returns the individual objective contributions.
func ObjectiveBreakdown(assignments []assignment.Assignment, ctx OptimisationContext) []ObjectiveContribution {
	w := ctx.Weights()
	capacities := ctx.Resources()
	return []ObjectiveContribution{
		{Name: "Assignment", Score: rawAssignment(assignments) * w.Assignment},
		{Name: "Workload Balance", Score: rawBalance(assignments, capacities) * w.Balance},
		{Name: "Travel Time", Score: rawTravel(assignments, ctx) * w.Travel},
		{Name: "Preferred Resource", Score: rawPreferredResource(assignments, ctx) * w.PreferredResource},
		{Name: "Plan Stability", Score: rawStability(assignments, ctx) * w.PlanStability},
	}
}
