// Package plan provides the OptimisedPlan domain object.
//
// An OptimisedPlan represents the output of the optimisation process.
// It contains the assignments produced by the optimiser.
package plan

import (
	"errors"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
)

// OptimisedPlan represents the result of an optimisation run.
//
// It is immutable once created and contains the assignments
// that the optimiser has produced.
type OptimisedPlan struct {
	assignments []assignment.Assignment
}

// New creates a validated OptimisedPlan.
//
// An OptimisedPlan must contain at least one assignment.
func New(assignments []assignment.Assignment) (OptimisedPlan, error) {
	if len(assignments) == 0 {
		return OptimisedPlan{}, errors.New("optimised plan must contain at least one assignment")
	}

	// Defensive copy so callers cannot mutate internal state.
	cp := make([]assignment.Assignment, len(assignments))
	copy(cp, assignments)

	return OptimisedPlan{assignments: cp}, nil
}

// Assignments returns the assignments in the plan.
//
// The returned slice is a defensive copy.
func (p OptimisedPlan) Assignments() []assignment.Assignment {
	cp := make([]assignment.Assignment, len(p.assignments))
	copy(cp, p.assignments)
	return cp
}

// Size returns the number of assignments in the plan.
func (p OptimisedPlan) Size() int {
	return len(p.assignments)
}
