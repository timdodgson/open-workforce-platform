// Package optimisation provides optimisation capabilities for the platform.
//
// The optimiser assigns work items to resources while respecting availability,
// capacity, and priority.
package optimisation

import (
	"errors"
	"fmt"
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
}

// WorkItemPriority provides priority information to the optimiser.
//
// This is not a domain object. It is structured input that the application
// layer prepares by interpreting business knowledge from work item details.
type WorkItemPriority struct {
	WorkItemID string
	Priority   int
}

// Solve accepts work items, resource capacities, and priorities, and produces
// an OptimisedPlan.
//
// It sorts work items by priority (highest first), then assigns each to the
// first available resource with remaining capacity. Unavailable resources are
// never assigned work. When all available resources are at capacity, remaining
// work items are left unassigned.
func Solve(items []workitem.WorkItem, capacities []ResourceCapacity, priorities []WorkItemPriority) (plan.OptimisedPlan, error) {
	if len(items) == 0 {
		return plan.OptimisedPlan{}, errors.New("optimiser requires at least one work item")
	}

	if len(capacities) == 0 {
		return plan.OptimisedPlan{}, errors.New("optimiser requires at least one resource")
	}

	// Build priority lookup.
	priorityOf := make(map[string]int, len(priorities))
	for _, p := range priorities {
		priorityOf[p.WorkItemID] = p.Priority
	}

	// Sort work items by priority (highest first), stable to preserve
	// original order for equal priorities (determinism).
	sorted := make([]workitem.WorkItem, len(items))
	copy(sorted, items)

	sort.SliceStable(sorted, func(i, j int) bool {
		return priorityOf[sorted[i].ID()] > priorityOf[sorted[j].ID()]
	})

	// Track remaining capacity per resource.
	remaining := make([]int, len(capacities))
	for i, rc := range capacities {
		remaining[i] = rc.Capacity
	}

	var assignments []assignment.Assignment
	var unassigned []string

	for _, item := range sorted {
		assigned := false
		for i, rc := range capacities {
			if !rc.Available {
				continue
			}
			if remaining[i] > 0 {
				a, err := assignment.New(rc.ResourceID, item.ID())
				if err != nil {
					return plan.OptimisedPlan{}, fmt.Errorf("failed to create assignment: %w", err)
				}
				assignments = append(assignments, a)
				remaining[i]--
				assigned = true
				break
			}
		}
		if !assigned {
			unassigned = append(unassigned, item.ID())
		}
	}

	// Calculate total capacity from available resources only.
	totalCapacity := 0
	for _, rc := range capacities {
		if rc.Available {
			totalCapacity += rc.Capacity
		}
	}

	// Calculate score.
	score := calculateScore(len(assignments), len(items))

	// Calculate utilisation.
	utilisation := calculateUtilisation(len(assignments), totalCapacity)

	return plan.New(plan.Result{
		Assignments:   assignments,
		Unassigned:    unassigned,
		TotalCapacity: totalCapacity,
		Utilisation:   utilisation,
		Score:         score,
	})
}

// calculateScore returns the optimisation score.
//
// 100 if all items assigned, otherwise (assigned/total) × 100 rounded.
func calculateScore(assigned, total int) int {
	if total == 0 {
		return 0
	}
	if assigned == total {
		return 100
	}
	return int(math.Round(float64(assigned) / float64(total) * 100))
}

// calculateUtilisation returns the utilisation percentage.
//
// (assigned / total capacity) × 100 rounded.
func calculateUtilisation(assigned, totalCapacity int) int {
	if totalCapacity == 0 {
		return 0
	}
	return int(math.Round(float64(assigned) / float64(totalCapacity) * 100))
}
