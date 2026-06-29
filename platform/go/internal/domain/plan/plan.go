// Package plan provides the OptimisedPlan domain object.
//
// An OptimisedPlan represents the output of the optimisation process.
// It contains assignments, unassigned work items, capacity information,
// utilisation and an optimisation score.
package plan

import (
	"errors"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
)

// Result holds the inputs required to construct an OptimisedPlan.
//
// This separates construction parameters from the immutable plan itself.
type Result struct {
	Assignments   []assignment.Assignment
	Unassigned    []string
	TotalCapacity int
	Utilisation   int
	Score         int
}

// OptimisedPlan represents the result of an optimisation run.
//
// It is immutable once created.
type OptimisedPlan struct {
	assignments   []assignment.Assignment
	unassigned    []string
	totalCapacity int
	utilisation   int
	score         int
}

// New creates a validated OptimisedPlan from an optimisation result.
//
// A plan must have either assignments or unassigned items (or both).
// A completely empty plan is invalid.
func New(r Result) (OptimisedPlan, error) {
	if len(r.Assignments) == 0 && len(r.Unassigned) == 0 {
		return OptimisedPlan{}, errors.New("optimised plan must contain at least one assignment or unassigned item")
	}

	// Defensive copies.
	assignmentsCopy := make([]assignment.Assignment, len(r.Assignments))
	copy(assignmentsCopy, r.Assignments)

	unassignedCopy := make([]string, len(r.Unassigned))
	copy(unassignedCopy, r.Unassigned)

	return OptimisedPlan{
		assignments:   assignmentsCopy,
		unassigned:    unassignedCopy,
		totalCapacity: r.TotalCapacity,
		utilisation:   r.Utilisation,
		score:         r.Score,
	}, nil
}

// Assignments returns the assignments in the plan.
//
// The returned slice is a defensive copy.
func (p OptimisedPlan) Assignments() []assignment.Assignment {
	cp := make([]assignment.Assignment, len(p.assignments))
	copy(cp, p.assignments)
	return cp
}

// Unassigned returns the IDs of work items that could not be assigned.
//
// The returned slice is a defensive copy.
func (p OptimisedPlan) Unassigned() []string {
	cp := make([]string, len(p.unassigned))
	copy(cp, p.unassigned)
	return cp
}

// Size returns the number of assignments in the plan.
func (p OptimisedPlan) Size() int {
	return len(p.assignments)
}

// UnassignedCount returns the number of unassigned work items.
func (p OptimisedPlan) UnassignedCount() int {
	return len(p.unassigned)
}

// TotalCapacity returns the total resource capacity available.
func (p OptimisedPlan) TotalCapacity() int {
	return p.totalCapacity
}

// Utilisation returns the utilisation percentage (0-100).
func (p OptimisedPlan) Utilisation() int {
	return p.utilisation
}

// Score returns the optimisation score (0-100).
func (p OptimisedPlan) Score() int {
	return p.score
}
