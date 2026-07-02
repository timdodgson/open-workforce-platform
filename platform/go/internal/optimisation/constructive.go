package optimisation

import (
	"errors"
	"fmt"
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
	dayOf := make(map[string]int, len(priorities))
	shiftTypeOf := make(map[string]string, len(priorities))
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
		dayOf[p.WorkItemID] = p.Day
		shiftTypeOf[p.WorkItemID] = p.ShiftType
	}

	travelLookup := buildTravelLookup(ctx.TravelMatrix())
	forbiddenSet := buildForbiddenSet(ctx.ForbiddenSuccessions())

	// Track next available time and current location per resource.
	nextAvailable := make([]int, len(capacities))
	currentLocation := make([]string, len(capacities))
	for i, rc := range capacities {
		nextAvailable[i] = rc.ShiftStart
		currentLocation[i] = rc.Location
	}

	// Track which days each resource already has assigned (for same-day constraint).
	resourceDays := make([]map[int]bool, len(capacities))
	for i := range resourceDays {
		resourceDays[i] = make(map[int]bool)
	}

	// Track shift type assigned per day per resource (for forbidden succession check).
	resourceDayShift := make([]map[int]string, len(capacities))
	for i := range resourceDayShift {
		resourceDayShift[i] = make(map[int]string)
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

			// Same-day constraint: a resource can only have one item per day.
			itemDay := dayOf[item.ID()]
			if itemDay > 0 && resourceDays[i][itemDay] {
				reasons["SameDayAssignment"] = true
				continue
			}

			// Forbidden shift succession check.
			itemShiftType := shiftTypeOf[item.ID()]
			if itemDay > 0 && itemShiftType != "" && len(forbiddenSet) > 0 {
				if !isSuccessionLegal(resourceDayShift[i], itemDay, itemShiftType, forbiddenSet) {
					reasons["IllegalShiftSuccession"] = true
					continue
				}
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
			if itemDay > 0 {
				resourceDays[i][itemDay] = true
				if itemShiftType != "" {
					resourceDayShift[i][itemDay] = itemShiftType
				}
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

	// Evaluate hard violations.
	violations := evaluateHardViolations(unassigned, ctx)

	// Generate constraint matches from hard violations and soft penalties.
	matches := generateConstraintMatches(violations, breakdown)

	return plan.New(plan.Result{
		Assignments:        assignments,
		Unassigned:         unassigned,
		UnassignedDetails:  details,
		HardViolations:     violations,
		ConstraintMatches:  matches,
		TotalCapacity:      totalCapacity,
		Utilisation:        utilisation,
		Score:              score,
		ObjectiveScore:     objScore,
		ObjectiveBreakdown: entries,
		Statistics:         stats,
	})
}

// generateConstraintMatches produces generic constraint match entries from
// hard violations and soft objective penalties.
func generateConstraintMatches(violations []plan.HardViolation, breakdown []ObjectiveContribution) []plan.ConstraintMatch {
	var matches []plan.ConstraintMatch

	// Convert hard violations to matches.
	for _, v := range violations {
		matches = append(matches, plan.ConstraintMatch{
			Constraint:  v.Code,
			Severity:    "hard",
			Day:         -1,
			Penalty:     0,
			Description: v.Message,
		})
	}

	// Convert soft objective penalties to matches.
	// Each non-zero penalty entry becomes a soft constraint match.
	for _, b := range breakdown {
		if b.Score < 0 {
			// Negative score = penalty (soft violation).
			matches = append(matches, plan.ConstraintMatch{
				Constraint:  b.Name,
				Severity:    "soft",
				Day:         -1,
				Penalty:     -b.Score, // Convert to positive penalty
				Description: b.Name + " penalty",
			})
		}
	}

	return matches
}

// evaluateHardViolations checks for hard constraint violations in the final plan.
func evaluateHardViolations(unassigned []string, ctx OptimisationContext) []plan.HardViolation {
	var violations []plan.HardViolation

	// H2: Under-staffing — unassigned mandatory demand items.
	priorities := ctx.WorkItems()
	mandatoryOf := make(map[string]bool, len(priorities))
	for _, p := range priorities {
		if p.Mandatory {
			mandatoryOf[p.WorkItemID] = true
		}
	}

	mandatoryUnassigned := 0
	for _, id := range unassigned {
		if mandatoryOf[id] {
			mandatoryUnassigned++
		}
	}

	if mandatoryUnassigned > 0 {
		violations = append(violations, plan.HardViolation{
			Code:    "UnderStaffed",
			Message: fmt.Sprintf("%d mandatory demand item(s) unassigned", mandatoryUnassigned),
		})
	}

	return violations
}

// explainUnassigned generates constraint explanation codes for each unassigned work item.
// It evaluates each item against all resources in the final plan state.
func explainUnassigned(unassigned []string, assignments []assignment.Assignment, capacities []ResourceInput, ctx OptimisationContext) []plan.UnassignedItem {
	if len(unassigned) == 0 {
		return nil
	}

	priorities := ctx.WorkItems()
	travelLookup := buildTravelLookup(ctx.TravelMatrix())
	forbiddenSet := buildForbiddenSet(ctx.ForbiddenSuccessions())

	// Build lookups.
	durationOf := make(map[string]int, len(priorities))
	earliestOf := make(map[string]int, len(priorities))
	latestOf := make(map[string]int, len(priorities))
	locationOf := make(map[string]string, len(priorities))
	requiredSkillOf := make(map[string]string, len(priorities))
	dayOf := make(map[string]int, len(priorities))
	shiftTypeOf := make(map[string]string, len(priorities))
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
		dayOf[p.WorkItemID] = p.Day
		shiftTypeOf[p.WorkItemID] = p.ShiftType
	}

	// Group by resource to replay schedule. Track days and shift types used.
	byResource := make(map[string][]string)
	for _, a := range assignments {
		byResource[a.ResourceID()] = append(byResource[a.ResourceID()], a.WorkItemID())
	}

	resourceDaysUsed := make([]map[int]bool, len(capacities))
	resourceDayShift := make([]map[int]string, len(capacities))
	for i, rc := range capacities {
		resourceDaysUsed[i] = make(map[int]bool)
		resourceDayShift[i] = make(map[int]string)
		for _, itemID := range byResource[rc.ResourceID] {
			if d := dayOf[itemID]; d > 0 {
				resourceDaysUsed[i][d] = true
				if st := shiftTypeOf[itemID]; st != "" {
					resourceDayShift[i][d] = st
				}
			}
		}
	}

	// Replay schedule to get final time state per resource (for time-based explanations).
	type resourceState struct {
		nextAvailable   int
		currentLocation string
	}
	states := make([]resourceState, len(capacities))
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
		itemDay := dayOf[itemID]
		itemShiftType := shiftTypeOf[itemID]

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

			// Same-day constraint.
			if itemDay > 0 && resourceDaysUsed[i][itemDay] {
				reasons["SameDayAssignment"] = true
				continue
			}

			// Forbidden shift succession.
			if itemDay > 0 && itemShiftType != "" && len(forbiddenSet) > 0 {
				if !isSuccessionLegal(resourceDayShift[i], itemDay, itemShiftType, forbiddenSet) {
					reasons["IllegalShiftSuccession"] = true
					continue
				}
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
// scheduled within time windows including travel time, same-day constraints,
// and forbidden shift successions.
func scheduleFeasible(assignments []assignment.Assignment, capacities []ResourceInput, priorities []WorkItemInput, ctx OptimisationContext) bool {
	// Build lookups.
	durationOf := make(map[string]int, len(priorities))
	earliestOf := make(map[string]int, len(priorities))
	latestOf := make(map[string]int, len(priorities))
	locationOf := make(map[string]string, len(priorities))
	dayOf := make(map[string]int, len(priorities))
	shiftTypeOf := make(map[string]string, len(priorities))
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
		dayOf[p.WorkItemID] = p.Day
		shiftTypeOf[p.WorkItemID] = p.ShiftType
	}

	travelLookup := buildTravelLookup(ctx.TravelMatrix())
	forbiddenSet := buildForbiddenSet(ctx.ForbiddenSuccessions())

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

		// Same-day constraint: check no two items share the same non-zero day.
		daysUsed := make(map[int]bool)
		for _, itemID := range itemIDs {
			d := dayOf[itemID]
			if d > 0 {
				if daysUsed[d] {
					return false
				}
				daysUsed[d] = true
			}
		}

		// Forbidden shift succession check (across consecutive days).
		if len(forbiddenSet) > 0 {
			if !checkShiftSuccessions(itemIDs, dayOf, shiftTypeOf, forbiddenSet) {
				return false
			}
		}

		// Sort items by day then by earliest start for scheduling validation.
		sort.Slice(itemIDs, func(i, j int) bool {
			di, dj := dayOf[itemIDs[i]], dayOf[itemIDs[j]]
			if di != dj {
				return di < dj
			}
			return earliestOf[itemIDs[i]] < earliestOf[itemIDs[j]]
		})

		// Validate scheduling per day: items on different days don't compete for time.
		dayGroups := groupByDay(itemIDs, dayOf)
		for _, dayItems := range dayGroups {
			shiftEnd := rc.ShiftEnd
			if shiftEnd <= 0 {
				shiftEnd = rc.ShiftStart + rc.Capacity
			}

			cursor := rc.ShiftStart
			currentLoc := rc.Location
			for _, itemID := range dayItems {
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
	}

	return true
}

// buildForbiddenSet creates a set for O(1) forbidden succession lookups.
// Key format: "preceding|successor"
func buildForbiddenSet(successions []ForbiddenSuccession) map[string]bool {
	if len(successions) == 0 {
		return nil
	}
	set := make(map[string]bool, len(successions))
	for _, s := range successions {
		set[s.PrecedingShift+"|"+s.SuccessorShift] = true
	}
	return set
}

// checkShiftSuccessions validates that no forbidden shift type transitions exist
// across consecutive days for a resource's assignments.
func checkShiftSuccessions(itemIDs []string, dayOf map[string]int, shiftTypeOf map[string]string, forbiddenSet map[string]bool) bool {
	// Build day -> shift type mapping for this resource.
	dayShift := make(map[int]string)
	for _, itemID := range itemIDs {
		d := dayOf[itemID]
		st := shiftTypeOf[itemID]
		if d > 0 && st != "" {
			dayShift[d] = st
		}
	}

	// Check consecutive days.
	for d, shiftType := range dayShift {
		if nextShift, ok := dayShift[d+1]; ok {
			key := shiftType + "|" + nextShift
			if forbiddenSet[key] {
				return false
			}
		}
	}
	return true
}

// isSuccessionLegal checks whether placing a shift type on a given day for a resource
// would violate any forbidden succession rules.
func isSuccessionLegal(dayShiftMap map[int]string, day int, shiftType string, forbiddenSet map[string]bool) bool {
	// Check preceding day -> this day.
	if prevShift, ok := dayShiftMap[day-1]; ok {
		if forbiddenSet[prevShift+"|"+shiftType] {
			return false
		}
	}
	// Check this day -> following day.
	if nextShift, ok := dayShiftMap[day+1]; ok {
		if forbiddenSet[shiftType+"|"+nextShift] {
			return false
		}
	}
	return true
}

// groupByDay groups item IDs by their day value. Items with day=0 form their own group.
func groupByDay(itemIDs []string, dayOf map[string]int) [][]string {
	dayMap := make(map[int][]string)
	var dayKeys []int
	for _, itemID := range itemIDs {
		d := dayOf[itemID]
		if _, exists := dayMap[d]; !exists {
			dayKeys = append(dayKeys, d)
		}
		dayMap[d] = append(dayMap[d], itemID)
	}
	sort.Ints(dayKeys)
	groups := make([][]string, 0, len(dayKeys))
	for _, d := range dayKeys {
		groups = append(groups, dayMap[d])
	}
	return groups
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
