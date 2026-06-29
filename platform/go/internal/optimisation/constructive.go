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
	assignments, unassigned := assignItems(sorted, capacities, priorities)

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
// suitable resource using sequential scheduling with time windows.
func assignItems(sorted []workitem.WorkItem, capacities []ResourceInput, priorities []WorkItemInput) ([]assignment.Assignment, []string) {
	requiredSkillOf := make(map[string]string, len(priorities))
	durationOf := make(map[string]int, len(priorities))
	earliestOf := make(map[string]int, len(priorities))
	latestOf := make(map[string]int, len(priorities))
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
			latest = 1440 // no constraint
		}
		latestOf[p.WorkItemID] = latest
	}

	// Track next available time per resource (sequential scheduling).
	nextAvailable := make([]int, len(capacities))
	for i, rc := range capacities {
		nextAvailable[i] = rc.ShiftStart
	}

	var assignments []assignment.Assignment
	var unassigned []string

	for _, item := range sorted {
		required := requiredSkillOf[item.ID()]
		duration := durationOf[item.ID()]
		earliest := earliestOf[item.ID()]
		latest := latestOf[item.ID()]

		placed := false
		for i, rc := range capacities {
			if !rc.Available {
				continue
			}
			if required != "" && !hasSkill(rc.Skills, required) {
				continue
			}

			// Determine when work would start on this resource.
			start := nextAvailable[i]
			if start < earliest {
				start = earliest
			}

			finish := start + duration
			shiftEnd := rc.ShiftEnd
			if shiftEnd <= 0 {
				shiftEnd = rc.ShiftStart + rc.Capacity // backward compatible
			}

			// Check constraints.
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
// scheduled within time windows. It groups assignments by resource and verifies
// each can fit within the resource's shift and the work item's time window.
//
// This is used by search algorithms to validate moves respect time windows.
func scheduleFeasible(assignments []assignment.Assignment, capacities []ResourceInput, priorities []WorkItemInput) bool {
	// Build lookups.
	durationOf := make(map[string]int, len(priorities))
	earliestOf := make(map[string]int, len(priorities))
	latestOf := make(map[string]int, len(priorities))
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
	}

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
		for _, itemID := range itemIDs {
			duration := durationOf[itemID]
			earliest := earliestOf[itemID]
			latest := latestOf[itemID]

			start := cursor
			if start < earliest {
				start = earliest
			}
			finish := start + duration
			if finish > shiftEnd || finish > latest {
				return false
			}
			cursor = finish
		}
	}

	return true
}
