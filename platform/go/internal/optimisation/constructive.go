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

func (c *constructiveAlgorithm) Solve(items []workitem.WorkItem, capacities []ResourceCapacity, priorities []WorkItemPriority) (plan.OptimisedPlan, error) {
	if err := validate(items, capacities); err != nil {
		return plan.OptimisedPlan{}, err
	}

	sorted := orderByPriority(items, priorities)
	assignments, unassigned := assignItems(sorted, capacities, priorities)

	return buildResult(assignments, unassigned, len(items), capacities)
}

// --- Shared helpers used by both algorithms ---

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
// suitable resource.
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

// findResource returns the index of the first resource that can accept a work item.
func findResource(capacities []ResourceCapacity, remaining []int, requiredSkill string) (int, bool) {
	for i, rc := range capacities {
		if canAccept(rc, remaining[i], requiredSkill) {
			return i, true
		}
	}
	return 0, false
}

// canAccept returns true if a resource is eligible to receive a work item.
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
	objScore := ObjectiveScore(assignments, capacities)
	breakdown := ObjectiveBreakdown(assignments, capacities)

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
