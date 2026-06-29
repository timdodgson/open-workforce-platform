package optimisation

import (
	"errors"
	"math"
	"sort"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// constructiveAlgorithm implements Algorithm using a single-pass greedy approach.
type constructiveAlgorithm struct{}

func init() {
	register(&constructiveAlgorithm{})
}

func (c *constructiveAlgorithm) Name() string {
	return "constructive"
}

func (c *constructiveAlgorithm) Solve(ctx OptimisationContext) (plan.OptimisedPlan, error) {
	items := ctx.Items()
	capacities := ctx.Resources()
	priorities := ctx.WorkItems()

	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	sorted := orderByPriority(items, priorities)
	assignments, unassigned := assignItems(sorted, capacities, priorities, ctx)

	return buildResult(assignments, unassigned, len(items), capacities, ctx)
}

// --- Shared helpers used by both algorithms ---

// validate checks that the optimiser has been given valid input.
func validate(items []workitem.WorkItem, capacities []ResourceInput) error {
	if len(items) == 0 {
		return errors.New("optimiser requires at least one work item")
	}
	if len(capacities) == 0 {
		return errors.New("optimiser requires at least one resource")
	}
	return nil
}

// orderByPriority returns work items sorted by priority (highest first).
func orderByPriority(items []workitem.WorkItem, priorities []WorkItemInput) []workitem.WorkItem {
	priorityOf := make(map[string]int, len(priorities))
	for _, p := range priorities {
		priorityOf[p.WorkItemID] = p.Priority
	}

	sorted := make([]workitem.WorkItem, len(items))
	copy(sorted, items)

	sort.SliceStable(sorted, func(i, j int) bool {
		return priorityOf[sorted[i].ID()] > priorityOf[sorted[j].ID()]
	})

	return sorted
}

// assignItems iterates through sorted work items and assigns each to the first
// suitable resource using sequential scheduling with time windows and travel time.
func assignItems(sorted []workitem.WorkItem, capacities []ResourceInput, priorities []WorkItemInput, ctx OptimisationContext) ([]assignment.Assignment, []string) {
	requiredSkillOf := make(map[string]string, len(priorities))
	durationOf := make(map[string]int, len(priorities))
	earliestOf := make(map[string]int, len(priorities))
	latestOf := make(map[string]int, len(priorities))
	locationOf := make(map[string]string, len(priorities))
	for _, p := range priorities {
		requiredSkillOf[p.WorkItemID] = p.RequiredSkill
		dur := p.Duration
		if dur <= 0 {
			dur = 1
		}
		durationOf[p.WorkItemID] = dur
		earliestOf[p.WorkItemID] = p.EarliestStart
		latest := p.LatestFinish
		if latest <= 0 {
			latest = 1440
		}
		latestOf[p.WorkItemID] = latest
		locationOf[p.WorkItemID] = p.Location
	}

	travelLookup := buildTravelLookup(ctx.TravelMatrix())

	// Track next available time and current location per resource.
	nextAvailable := make([]int, len(capacities))
	currentLocation := make([]string, len(capacities))
	for i, rc := range capacities {
		nextAvailable[i] = rc.ShiftStart
		currentLocation[i] = rc.Location
	}

	var assignments []assignment.Assignment
	var unassigned []string

	for _, item := range sorted {
		required := requiredSkillOf[item.ID()]
		duration := durationOf[item.ID()]
		earliest := earliestOf[item.ID()]
		latest := latestOf[item.ID()]
		itemLocation := locationOf[item.ID()]

		placed := false
		for i, rc := range capacities {
			if !rc.Available {
				continue
			}
			if required != "" && !hasSkill(rc.Skills, required) {
				continue
			}

			// Calculate travel time from current location to work item.
			travel := lookupTravel(travelLookup, currentLocation[i], itemLocation)

			// Determine when work would start (after travel + waiting).
			arrivalTime := nextAvailable[i] + travel
			start := arrivalTime
			if start < earliest {
				start = earliest
			}

			finish := start + duration
			shiftEnd := rc.ShiftEnd
			if shiftEnd <= 0 {
				shiftEnd = rc.ShiftStart + rc.Capacity
			}

			if finish > shiftEnd {
				continue
			}
			if finish > latest {
				continue
			}

			a, err := assignment.New(rc.ResourceID, item.ID())
			if err != nil {
				continue
			}
			assignments = append(assignments, a)
			nextAvailable[i] = finish
			if itemLocation != "" {
				currentLocation[i] = itemLocation
			}
			placed = true
			break
		}

		if !placed {
			unassigned = append(unassigned, item.ID())
		}
	}

	return assignments, unassigned
}

// findResource returns the index of the first resource that can accept a work item.
func findResource(capacities []ResourceInput, remaining []int, requiredSkill string, duration int) (int, bool) {
	for i, rc := range capacities {
		if canAccept(rc, remaining[i], requiredSkill, duration) {
			return i, true
		}
	}
	return 0, false
}

// canAccept returns true if a resource is eligible to receive a work item.
func canAccept(rc ResourceInput, remaining int, requiredSkill string, duration int) bool {
	if !rc.Available {
		return false
	}
	if remaining < duration {
		return false
	}
	if requiredSkill != "" && !hasSkill(rc.Skills, requiredSkill) {
		return false
	}
	return true
}

// hasSkill returns true if the skills slice contains the required skill.
func hasSkill(skills []string, required string) bool {
	for _, s := range skills {
		if s == required {
			return true
		}
	}
	return false
}

// buildResult calculates scoring and constructs the OptimisedPlan.
func buildResult(assignments []assignment.Assignment, unassigned []string, totalItems int, capacities []ResourceInput, ctx OptimisationContext) (plan.OptimisedPlan, error) {
	totalCapacity := availableCapacity(capacities)
	score := calculateScore(len(assignments), totalItems)
	utilisation := calculateUtilisation(len(assignments), totalCapacity)
	objScore := ObjectiveScore(assignments, ctx)
	breakdown := ObjectiveBreakdown(assignments, ctx)

	// Convert to plan's ObjectiveEntry type.
	entries := make([]plan.ObjectiveEntry, len(breakdown))
	for i, b := range breakdown {
		entries[i] = plan.ObjectiveEntry{Name: b.Name, Score: b.Score}
	}

	return plan.New(plan.Result{
		Assignments:        assignments,
		Unassigned:         unassigned,
		TotalCapacity:      totalCapacity,
		Utilisation:        utilisation,
		Score:              score,
		ObjectiveScore:     objScore,
		ObjectiveBreakdown: entries,
	})
}

// availableCapacity returns the total capacity of available resources.
func availableCapacity(capacities []ResourceInput) int {
	total := 0
	for _, rc := range capacities {
		if rc.Available {
			total += rc.Capacity
		}
	}
	return total
}

func calculateScore(assigned, total int) int {
	if total == 0 {
		return 0
	}
	if assigned == total {
		return 100
	}
	return int(math.Round(float64(assigned) / float64(total) * 100))
}

func calculateUtilisation(assigned, totalCapacity int) int {
	if totalCapacity == 0 {
		return 0
	}
	return int(math.Round(float64(assigned) / float64(totalCapacity) * 100))
}

// scheduleFeasible checks whether a set of assignments can all be sequentially
// scheduled within time windows including travel time.
func scheduleFeasible(assignments []assignment.Assignment, capacities []ResourceInput, priorities []WorkItemInput, ctx OptimisationContext) bool {
	// Build lookups.
	durationOf := make(map[string]int, len(priorities))
	earliestOf := make(map[string]int, len(priorities))
	latestOf := make(map[string]int, len(priorities))
	locationOf := make(map[string]string, len(priorities))
	for _, p := range priorities {
		dur := p.Duration
		if dur <= 0 {
			dur = 1
		}
		durationOf[p.WorkItemID] = dur
		earliestOf[p.WorkItemID] = p.EarliestStart
		latest := p.LatestFinish
		if latest <= 0 {
			latest = 1440
		}
		latestOf[p.WorkItemID] = latest
		locationOf[p.WorkItemID] = p.Location
	}

	travelLookup := buildTravelLookup(ctx.TravelMatrix())

	// Group assignments by resource.
	byResource := make(map[string][]string)
	for _, a := range assignments {
		byResource[a.ResourceID()] = append(byResource[a.ResourceID()], a.WorkItemID())
	}

	// Build resource lookup.
	resourceOf := make(map[string]ResourceInput, len(capacities))
	for _, rc := range capacities {
		resourceOf[rc.ResourceID] = rc
	}

	// Check each resource's schedule.
	for resID, itemIDs := range byResource {
		rc, ok := resourceOf[resID]
		if !ok {
			return false
		}

		shiftEnd := rc.ShiftEnd
		if shiftEnd <= 0 {
			shiftEnd = rc.ShiftStart + rc.Capacity
		}

		cursor := rc.ShiftStart
		currentLoc := rc.Location
		for _, itemID := range itemIDs {
			duration := durationOf[itemID]
			earliest := earliestOf[itemID]
			latest := latestOf[itemID]
			itemLoc := locationOf[itemID]

			travel := lookupTravel(travelLookup, currentLoc, itemLoc)
			arrival := cursor + travel
			start := arrival
			if start < earliest {
				start = earliest
			}
			finish := start + duration
			if finish > shiftEnd || finish > latest {
				return false
			}
			cursor = finish
			if itemLoc != "" {
				currentLoc = itemLoc
			}
		}
	}

	return true
}

// buildTravelLookup creates a map for O(1) travel time lookups.
func buildTravelLookup(matrix []TravelEntry) map[string]int {
	lookup := make(map[string]int, len(matrix))
	for _, t := range matrix {
		lookup[t.From+"|"+t.To] = t.Minutes
	}
	return lookup
}

// lookupTravel returns travel time between two locations, defaulting to 0.
func lookupTravel(lookup map[string]int, from, to string) int {
	if from == "" || to == "" || from == to {
		return 0
	}
	return lookup[from+"|"+to]
}
