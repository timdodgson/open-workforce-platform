// Package plan provides the OptimisedPlan domain object.
//
// An OptimisedPlan represents the output of the optimisation process.
// It contains the WorkItems selected by the optimiser.
package plan

import (
	"errors"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// OptimisedPlan represents the result of an optimisation run.
//
// It is immutable once created and contains the work items
// that the optimiser has selected for scheduling.
type OptimisedPlan struct {
	items []workitem.WorkItem
}

// New creates a validated OptimisedPlan.
//
// An OptimisedPlan must contain at least one work item.
func New(items []workitem.WorkItem) (OptimisedPlan, error) {
	if len(items) == 0 {
		return OptimisedPlan{}, errors.New("optimised plan must contain at least one work item")
	}

	// Defensive copy so callers cannot mutate internal state.
	cp := make([]workitem.WorkItem, len(items))
	copy(cp, items)

	return OptimisedPlan{items: cp}, nil
}

// Items returns the work items in the plan.
//
// The returned slice is a defensive copy.
func (p OptimisedPlan) Items() []workitem.WorkItem {
	cp := make([]workitem.WorkItem, len(p.items))
	copy(cp, p.items)
	return cp
}

// Size returns the number of work items in the plan.
func (p OptimisedPlan) Size() int {
	return len(p.items)
}
