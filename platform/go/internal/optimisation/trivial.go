// Package optimisation provides optimisation capabilities for the platform.
//
// The capacity-aware optimiser assigns work items to resources while
// respecting each resource's maximum capacity. It processes work items
// in order and assigns each to the first resource with available capacity.
package optimisation

import (
	"errors"
	"fmt"
	"math"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// ResourceCapacity provides capacity information to the optimiser.
//
// This is not a domain object. It is structured input that the application
// layer prepares by interpreting business knowledge from resource details.
type ResourceCapacity struct {
	ResourceID string
	Capacity   int
}

// Solve accepts work items and resource capacities, and produces an OptimisedPlan.
//
// It assigns each work item to the first resource with available capacity.
// When all resources are at capacity, remaining work items are left unassigned.
func Solve(items []workitem.WorkItem, capacities []ResourceCapacity) (plan.OptimisedPlan, error) {
	if len(items) == 0 {
		return plan.OptimisedPlan{}, errors.New("optimiser requires at least one work item")
	}

	if len(capacities) == 0 {
		return plan.OptimisedPlan{}, errors.New("optimiser requires at least one resource")
	}

	// Track remaining capacity per resource.
	remaining := make([]int, len(capacities))
	for i, rc := range capacities {
		remaining[i] = rc.Capacity
	}

	var assignments []assignment.Assignment
	var unassigned []string

	for _, item := range items {
		assigned := false
		for i, rc := range capacities {
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

	// Calculate total capacity.
	totalCapacity := 0
	for _, rc := range capacities {
		totalCapacity += rc.Capacity
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
