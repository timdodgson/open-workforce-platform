package optimisation

import (
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
)

type lnsAlgorithm struct{}

func init() {
	register(&lnsAlgorithm{})
}

func (l *lnsAlgorithm) Name() string {
	return "large-neighbourhood-search"
}

func (l *lnsAlgorithm) Solve(ctx OptimisationContext) (plan.OptimisedPlan, error) {
	startTime := time.Now()
	items := ctx.Items()
	capacities := ctx.Resources()
	priorities := ctx.WorkItems()

	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	sorted := orderByPriority(items, priorities)
	assignments, unassigned, _ := assignItems(sorted, capacities, priorities, ctx)

	// Warm start.
	if existing := ctx.ExistingAssignments(); len(existing) > 0 {
		if scheduleFeasible(existing, capacities, priorities, ctx) {
			assignedSet := make(map[string]bool, len(existing))
			for _, a := range existing {
				assignedSet[a.WorkItemID()] = true
			}
			assignments = existing
			unassigned = nil
			for _, item := range items {
				if !assignedSet[item.ID()] {
					unassigned = append(unassigned, item.ID())
				}
			}
		}
	}

	totalItems := len(items)
	bestAssignments := copyAssignments(assignments)
	bestUnassigned := copyStrings(unassigned)
	bestScore := ObjectiveScore(bestAssignments, ctx)

	candidatesEvaluated := 0
	improvementsAccepted := 0
	iterationsRun := 0
	maxIter := ctx.Profile().LNSIterations
	dSize := ctx.Profile().DestroySize

	for iteration := 0; iteration < maxIter; iteration++ {
		iterationsRun++

		if len(assignments) < dSize {
			break
		}

		// Destroy: remove assignments deterministically based on iteration.
		destroyed, remaining := destroy(assignments, iteration, dSize)

		// Build new unassigned pool: existing unassigned + destroyed items.
		newUnassigned := make([]string, len(unassigned), len(unassigned)+len(destroyed))
		copy(newUnassigned, unassigned)
		for _, d := range destroyed {
			newUnassigned = append(newUnassigned, d.WorkItemID())
		}

		// Repair: reconstruct using existing assignment logic.
		// Keep remaining assignments fixed, try to place all unassigned items.
		repaired := repair(remaining, newUnassigned, capacities, priorities, ctx)
		candidatesEvaluated++

		if !scheduleFeasible(repaired, capacities, priorities, ctx) {
			continue
		}

		score := ObjectiveScore(repaired, ctx)
		if score > bestScore {
			// Derive new unassigned from items not in repaired plan.
			assignedSet := make(map[string]bool, len(repaired))
			for _, a := range repaired {
				assignedSet[a.WorkItemID()] = true
			}
			var newUn []string
			for _, item := range items {
				if !assignedSet[item.ID()] {
					newUn = append(newUn, item.ID())
				}
			}

			assignments = repaired
			unassigned = newUn
			bestAssignments = copyAssignments(assignments)
			bestUnassigned = copyStrings(unassigned)
			bestScore = score
			improvementsAccepted++
		}
	}

	stats := plan.Statistics{
		Algorithm:            "large-neighbourhood-search",
		DurationMs:           time.Since(startTime).Milliseconds(),
		Iterations:           iterationsRun,
		CandidatesEvaluated:  candidatesEvaluated,
		ImprovementsAccepted: improvementsAccepted,
		FinalObjectiveScore:  bestScore,
	}

	return buildResult(bestAssignments, bestUnassigned, totalItems, capacities, ctx, stats)
}

// destroy removes a deterministic subset of assignments based on iteration index.
func destroy(assignments []assignment.Assignment, iteration int, dSize int) ([]assignment.Assignment, []assignment.Assignment) {
	n := len(assignments)
	var destroyed []assignment.Assignment
	removeIndices := make(map[int]bool)

	for i := 0; i < dSize && i < n; i++ {
		idx := (iteration*dSize + i) % n
		removeIndices[idx] = true
	}

	var remaining []assignment.Assignment
	for i, a := range assignments {
		if removeIndices[i] {
			destroyed = append(destroyed, a)
		} else {
			remaining = append(remaining, a)
		}
	}

	return destroyed, remaining
}

// repair attempts to place unassigned items onto the existing partial plan.
// It uses the same sequential scheduling logic as the constructive algorithm.
func repair(existing []assignment.Assignment, unassignedIDs []string, capacities []ResourceInput, priorities []WorkItemInput, ctx OptimisationContext) []assignment.Assignment {
	// Build work item lookup for the items we need to place.
	itemInputs := make(map[string]WorkItemInput, len(priorities))
	for _, p := range priorities {
		itemInputs[p.WorkItemID] = p
	}

	// Sort unassigned by priority (highest first) for greedy repair.
	type itemPriority struct {
		id       string
		priority int
	}
	toPlace := make([]itemPriority, 0, len(unassignedIDs))
	for _, id := range unassignedIDs {
		p := 0
		if inp, ok := itemInputs[id]; ok {
			p = inp.Priority
		}
		toPlace = append(toPlace, itemPriority{id: id, priority: p})
	}
	// Simple insertion sort for determinism (small N).
	for i := 1; i < len(toPlace); i++ {
		for j := i; j > 0 && toPlace[j].priority > toPlace[j-1].priority; j-- {
			toPlace[j], toPlace[j-1] = toPlace[j-1], toPlace[j]
		}
	}

	// Replay schedule for existing assignments to get resource states.
	travelLookup := buildTravelLookup(ctx.TravelMatrix())

	type resourceState struct {
		nextAvailable   int
		currentLocation string
	}
	states := make([]resourceState, len(capacities))
	for i, rc := range capacities {
		states[i] = resourceState{nextAvailable: rc.ShiftStart, currentLocation: rc.Location}
	}

	// Replay existing assignments in order per resource.
	byResource := make(map[string][]string)
	for _, a := range existing {
		byResource[a.ResourceID()] = append(byResource[a.ResourceID()], a.WorkItemID())
	}
	for i, rc := range capacities {
		for _, itemID := range byResource[rc.ResourceID] {
			inp := itemInputs[itemID]
			dur := inp.Duration
			if dur <= 0 {
				dur = 1
			}
			travel := lookupTravel(travelLookup, states[i].currentLocation, inp.Location)
			arrival := states[i].nextAvailable + travel
			start := arrival
			if start < inp.EarliestStart {
				start = inp.EarliestStart
			}
			states[i].nextAvailable = start + dur
			if inp.Location != "" {
				states[i].currentLocation = inp.Location
			}
		}
	}

	// Try to place each unassigned item.
	result := make([]assignment.Assignment, len(existing))
	copy(result, existing)

	for _, ip := range toPlace {
		inp := itemInputs[ip.id]
		dur := inp.Duration
		if dur <= 0 {
			dur = 1
		}
		latest := inp.LatestFinish
		if latest <= 0 {
			latest = 1440
		}

		for i, rc := range capacities {
			if !rc.Available {
				continue
			}
			if inp.RequiredSkill != "" && !hasSkill(rc.Skills, inp.RequiredSkill) {
				continue
			}

			travel := lookupTravel(travelLookup, states[i].currentLocation, inp.Location)
			arrival := states[i].nextAvailable + travel
			start := arrival
			if start < inp.EarliestStart {
				start = inp.EarliestStart
			}
			finish := start + dur

			shiftEnd := rc.ShiftEnd
			if shiftEnd <= 0 {
				shiftEnd = rc.ShiftStart + rc.Capacity
			}

			if finish > shiftEnd || finish > latest {
				continue
			}

			a, err := assignment.New(rc.ResourceID, ip.id)
			if err != nil {
				continue
			}
			result = append(result, a)
			states[i].nextAvailable = finish
			if inp.Location != "" {
				states[i].currentLocation = inp.Location
			}
			break
		}
	}

	return result
}
