// Package optimisation provides optimisation capabilities for the platform.
//
// The optimiser assigns work items to resources while respecting availability,
// capacity, skills, and priority.
package optimisation

import (
	"errors"
	"math"
	"sort"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// ResourceCapacity provides resource constraint information to the optimiser.
//
// This is not a domain object. It is structured input that the application
// layer prepares by interpreting business knowledge from resource details.
type ResourceCapacity struct {
	ResourceID string
	Capacity   int
	Available  bool
	Skills     []string
}

// WorkItemPriority provides work item optimisation input to the optimiser.
//
// This is not a domain object. It is structured input that the application
// layer prepares by interpreting business knowledge from work item details.
type WorkItemPriority struct {
	WorkItemID    string
	Priority      int
	RequiredSkill string
}

// Solve accepts work items, resource capacities, and priorities, and produces
// an OptimisedPlan.
//
// It sorts work items by priority (highest first), then assigns each to the
// first available resource with remaining capacity and matching skills.
func Solve(items []workitem.WorkItem, capacities []ResourceCapacity, priorities []WorkItemPriority) (plan.OptimisedPlan, error) {
	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	sorted := orderByPriority(items, priorities)
	assignments, unassigned := assignItems(sorted, capacities, priorities)

	return buildResult(assignments, unassigned, len(items), capacities)
}

// validate checks that the optimiser has been given valid input.
func validate(items []workitem.WorkItem, capacities []ResourceCapacity) error {
	if len(items) == 0 {
		return errors.New("optimiser requires at least one work item")
	}
	if len(capacities) == 0 {
		return errors.New("optimiser requires at least one resource")
	}
	return nil
}

// orderByPriority returns work items sorted by priority (highest first).
//
// A stable sort preserves original order for equal priorities (determinism).
func orderByPriority(items []workitem.WorkItem, priorities []WorkItemPriority) []workitem.WorkItem {
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
// suitable resource. Returns the assignments and IDs of unassigned work items.
func assignItems(sorted []workitem.WorkItem, capacities []ResourceCapacity, priorities []WorkItemPriority) ([]assignment.Assignment, []string) {
	requiredSkillOf := make(map[string]string, len(priorities))
	for _, p := range priorities {
		requiredSkillOf[p.WorkItemID] = p.RequiredSkill
	}

	remaining := make([]int, len(capacities))
	for i, rc := range capacities {
		remaining[i] = rc.Capacity
	}

	var assignments []assignment.Assignment
	var unassigned []string

	for _, item := range sorted {
		required := requiredSkillOf[item.ID()]

		if idx, ok := findResource(capacities, remaining, required); ok {
			a, err := assignment.New(capacities[idx].ResourceID, item.ID())
			if err != nil {
				// This should not happen with valid IDs, but fail safe.
				unassigned = append(unassigned, item.ID())
				continue
			}
			assignments = append(assignments, a)
			remaining[idx]--
		} else {
			unassigned = append(unassigned, item.ID())
		}
	}

	return assignments, unassigned
}

// findResource returns the index of the first resource that can accept a work
// item with the given required skill. Returns false if no resource qualifies.
func findResource(capacities []ResourceCapacity, remaining []int, requiredSkill string) (int, bool) {
	for i, rc := range capacities {
		if canAccept(rc, remaining[i], requiredSkill) {
			return i, true
		}
	}
	return 0, false
}

// canAccept returns true if a resource is eligible to receive a work item.
//
// A resource can accept a work item when it is:
//   - available
//   - has remaining capacity
//   - has the required skill (or no skill is required)
func canAccept(rc ResourceCapacity, remaining int, requiredSkill string) bool {
	if !rc.Available {
		return false
	}
	if remaining <= 0 {
		return false
	}
	if requiredSkill != "" && !hasSkill(rc.Skills, requiredSkill) {
		return false
	}
	return true
}

// hasSkill returns true if the skills slice contains the required skill.
// Matching is exact and case-sensitive.
func hasSkill(skills []string, required string) bool {
	for _, s := range skills {
		if s == required {
			return true
		}
	}
	return false
}

// buildResult calculates scoring and constructs the OptimisedPlan.
func buildResult(assignments []assignment.Assignment, unassigned []string, totalItems int, capacities []ResourceCapacity) (plan.OptimisedPlan, error) {
	totalCapacity := availableCapacity(capacities)
	score := calculateScore(len(assignments), totalItems)
	utilisation := calculateUtilisation(len(assignments), totalCapacity)

	return plan.New(plan.Result{
		Assignments:   assignments,
		Unassigned:    unassigned,
		TotalCapacity: totalCapacity,
		Utilisation:   utilisation,
		Score:         score,
	})
}

// availableCapacity returns the total capacity of available resources.
func availableCapacity(capacities []ResourceCapacity) int {
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
