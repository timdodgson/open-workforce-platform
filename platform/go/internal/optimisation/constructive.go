package optimisation

import (
	"errors"
	"math"
	"sort"
	"time"

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
	start := time.Now()
	items := ctx.Items()
	capacities := ctx.Resources()
	priorities := ctx.WorkItems()

	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	sorted := orderByPriority(items, priorities)
	assignments, unassigned, _ := assignItems(sorted, capacities, priorities, ctx)

	stats := plan.Statistics{
		Algorithm:            "constructive",
		DurationMs:           time.Since(start).Milliseconds(),
		Iterations:           1,
		CandidatesEvaluated:  0,
		ImprovementsAccepted: 0,
		FinalObjectiveScore:  ObjectiveScore(assignments, ctx),
	}

	return buildResult(assignments, unassigned, len(items), capacities, ctx, stats)
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
// Returns assignments, unassigned IDs, and detailed explanations for unassigned items.
func assignItems(sorted []workitem.WorkItem, capacities []ResourceInput, priorities []WorkItemInput, ctx OptimisationContext) ([]assignment.Assignment, []string, []plan.UnassignedItem) {
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
	var unassignedDetails []plan.UnassignedItem

	for _, item := range sorted {
		required := requiredSkillOf[item.ID()]
		duration := durationOf[item.ID()]
		earliest := earliestOf[item.ID()]
		latest := latestOf[item.ID()]
		itemLocation := locationOf[item.ID()]

		placed := false
		reasons := make(map[string]bool)

		for i, rc := range capacities {
			if !rc.Available {
				reasons["NoAvailableResource"] = true
				continue
			}
			if required != "" && !hasSkill(rc.Skills, required) {
				reasons["SkillMismatch"] = true
				continue
			}

			travel := lookupTravel(travelLookup, currentLocation[i], itemLocation)
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
				if travel > 0 && nextAvailable[i]+duration <= shiftEnd {
					reasons["TravelTimeExceeded"] = true
				} else {
					reasons["ShiftEnded"] = true
				}
				continue
			}
			if finish > latest {
				if travel > 0 && nextAvailable[i]+duration <= latest {
					reasons["TravelTimeExceeded"] = true
				} else {
					reasons["TimeWindowExceeded"] = true
				}
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
			reasonList := make([]string, 0, len(reasons))
			for r := range reasons {
				reasonList = append(reasonList, r)
			}
			// Sort for determinism.
			sort.Strings(reasonList)
			unassignedDetails = append(unassignedDetails, plan.UnassignedItem{
				WorkItemID: item.ID(),
				Reasons:    reasonList,
			})
		}
	}

	return assignments, unassigned, unassignedDetails
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
func buildResult(assignments []assignment.Assignment, unassigned []string, totalItems int, capacities []ResourceInput, ctx OptimisationContext, stats plan.Statistics) (plan.OptimisedPlan, error) {
	totalCapacity := availableCapacity(capacities)
	score := calculateScore(len(assignments), totalItems)
	utilisation := calculateUtilisation(len(assignments), totalCapacity)
	objScore := ObjectiveScore(assignments, ctx)
	breakdown := ObjectiveBreakdown(assignments, ctx)

	entries := make([]plan.ObjectiveEntry, len(breakdown))
	for i, b := range breakdown {
		entries[i] = plan.ObjectiveEntry{Name: b.Name, Score: b.Score}
	}

	// Generate constraint explanations for unassigned items from the final plan state.
	details := explainUnassigned(unassigned, assignments, capacities, ctx)

	return plan.New(plan.Result{
		Assignments:        assignments,
		Unassigned:         unassigned,
		UnassignedDetails:  details,
		TotalCapacity:      totalCapacity,
		Utilisation:        utilisation,
		Score:              score,
		ObjectiveScore:     objScore,
		ObjectiveBreakdown: entries,
		Statistics:         stats,
	})
}

// explainUnassigned generates constraint explanation codes for each unassigned work item.
// It evaluates each item against all resources in the final plan state.
func explainUnassigned(unassigned []string, assignments []assignment.Assignment, capacities []ResourceInput, ctx OptimisationContext) []plan.UnassignedItem {
	if len(unassigned) == 0 {
		return nil
	}

	priorities := ctx.WorkItems()
	travelLookup := buildTravelLookup(ctx.TravelMatrix())

	// Build lookups.
	durationOf := make(map[string]int, len(priorities))
	earliestOf := make(map[string]int, len(priorities))
	latestOf := make(map[string]int, len(priorities))
	locationOf := make(map[string]string, len(priorities))
	requiredSkillOf := make(map[string]string, len(priorities))
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
		requiredSkillOf[p.WorkItemID] = p.RequiredSkill
	}

	// Replay schedule to get final state per resource.
	type resourceState struct {
		nextAvailable   int
		currentLocation string
	}
	states := make([]resourceState, len(capacities))
	for i, rc := range capacities {
		states[i] = resourceState{nextAvailable: rc.ShiftStart, currentLocation: rc.Location}
	}

	// Process assigned items in order to build final schedule state.
	// Group by resource to replay.
	byResource := make(map[string][]string)
	for _, a := range assignments {
		byResource[a.ResourceID()] = append(byResource[a.ResourceID()], a.WorkItemID())
	}

	for i, rc := range capacities {
		itemIDs := byResource[rc.ResourceID]
		cursor := rc.ShiftStart
		loc := rc.Location
		for _, itemID := range itemIDs {
			travel := lookupTravel(travelLookup, loc, locationOf[itemID])
			arrival := cursor + travel
			start := arrival
			if start < earliestOf[itemID] {
				start = earliestOf[itemID]
			}
			cursor = start + durationOf[itemID]
			if locationOf[itemID] != "" {
				loc = locationOf[itemID]
			}
		}
		states[i] = resourceState{nextAvailable: cursor, currentLocation: loc}
	}

	// Now check each unassigned item against all resources.
	var details []plan.UnassignedItem
	for _, itemID := range unassigned {
		required := requiredSkillOf[itemID]
		duration := durationOf[itemID]
		earliest := earliestOf[itemID]
		latest := latestOf[itemID]
		itemLoc := locationOf[itemID]

		reasons := make(map[string]bool)
		for i, rc := range capacities {
			if !rc.Available {
				reasons["NoAvailableResource"] = true
				continue
			}
			if required != "" && !hasSkill(rc.Skills, required) {
				reasons["SkillMismatch"] = true
				continue
			}

			travel := lookupTravel(travelLookup, states[i].currentLocation, itemLoc)
			arrival := states[i].nextAvailable + travel
			start := arrival
			if start < earliest {
				start = earliest
			}
			finish := start + duration

			shiftEnd := rc.ShiftEnd
			if shiftEnd <= 0 {
				shiftEnd = rc.ShiftStart + rc.Capacity
			}

			if finish > shiftEnd {
				if travel > 0 && states[i].nextAvailable+duration <= shiftEnd {
					reasons["TravelTimeExceeded"] = true
				} else {
					reasons["ShiftEnded"] = true
				}
				continue
			}
			if finish > latest {
				if travel > 0 && states[i].nextAvailable+duration <= latest {
					reasons["TravelTimeExceeded"] = true
				} else {
					reasons["TimeWindowExceeded"] = true
				}
				continue
			}
		}

		reasonList := make([]string, 0, len(reasons))
		for r := range reasons {
			reasonList = append(reasonList, r)
		}
		sort.Strings(reasonList)
		details = append(details, plan.UnassignedItem{WorkItemID: itemID, Reasons: reasonList})
	}

	return details
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

		// Sort items by earliest start to determine feasible execution order.
		sort.Slice(itemIDs, func(i, j int) bool {
			return earliestOf[itemIDs[i]] < earliestOf[itemIDs[j]]
		})

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
