// Package optimisation provides optimisation capabilities for the platform.
//
// The trivial optimiser is the first implementation. It assigns all work items
// to the first available resource. It exists to prove the assignment flow
// works end-to-end.
package optimisation

import (
	"errors"
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// Solve accepts work items and resources, and produces an OptimisedPlan.
//
// The trivial implementation assigns all work items to the first resource.
// Future implementations will consider constraints, objectives and scoring.
func Solve(items []workitem.WorkItem, resources []resource.Resource) (plan.OptimisedPlan, error) {
	if len(items) == 0 {
		return plan.OptimisedPlan{}, errors.New("optimiser requires at least one work item")
	}

	if len(resources) == 0 {
		return plan.OptimisedPlan{}, errors.New("optimiser requires at least one resource")
	}

	// Trivial strategy: assign all work items to the first resource.
	target := resources[0]
	assignments := make([]assignment.Assignment, 0, len(items))

	for _, item := range items {
		a, err := assignment.New(target.ID(), item.ID())
		if err != nil {
			return plan.OptimisedPlan{}, fmt.Errorf("failed to create assignment: %w", err)
		}
		assignments = append(assignments, a)
	}

	return plan.New(assignments)
}
