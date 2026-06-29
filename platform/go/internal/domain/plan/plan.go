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
	Assignments        []assignment.Assignment
	Unassigned         []string
	UnassignedDetails  []UnassignedItem
	TotalCapacity      int
	Utilisation        int
	Score              int
	ObjectiveScore     int
	ObjectiveBreakdown []ObjectiveEntry
}

// UnassignedItem represents a work item that could not be assigned,
// along with explanation codes describing why.
type UnassignedItem struct {
	WorkItemID string
	Reasons    []string
}

// ObjectiveEntry represents a named objective's contribution to the total score.
type ObjectiveEntry struct {
	Name  string
	Score int
}

// OptimisedPlan represents the result of an optimisation run.
//
// It is immutable once created.
type OptimisedPlan struct {
	assignments        []assignment.Assignment
	unassigned         []string
	unassignedDetails  []UnassignedItem
	totalCapacity      int
	utilisation        int
	score              int
	objectiveScore     int
	objectiveBreakdown []ObjectiveEntry
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

	// Defensive copy of breakdown.
	breakdownCopy := make([]ObjectiveEntry, len(r.ObjectiveBreakdown))
	copy(breakdownCopy, r.ObjectiveBreakdown)

	// Defensive copy of unassigned details.
	detailsCopy := make([]UnassignedItem, len(r.UnassignedDetails))
	copy(detailsCopy, r.UnassignedDetails)

	return OptimisedPlan{
		assignments:        assignmentsCopy,
		unassigned:         unassignedCopy,
		unassignedDetails:  detailsCopy,
		totalCapacity:      r.TotalCapacity,
		utilisation:        r.Utilisation,
		score:              r.Score,
		objectiveScore:     r.ObjectiveScore,
		objectiveBreakdown: breakdownCopy,
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

// Score returns the assignment score percentage (0-100).
func (p OptimisedPlan) Score() int {
	return p.score
}

// ObjectiveScore returns the objective score used by the optimiser.
//
// This is an additive score combining all objectives (assignment, balance, etc.).
// Higher is better.
func (p OptimisedPlan) ObjectiveScore() int {
	return p.objectiveScore
}

// ObjectiveBreakdown returns the individual objective contributions.
//
// The returned slice is a defensive copy.
func (p OptimisedPlan) ObjectiveBreakdown() []ObjectiveEntry {
	cp := make([]ObjectiveEntry, len(p.objectiveBreakdown))
	copy(cp, p.objectiveBreakdown)
	return cp
}

// UnassignedDetails returns the unassigned work items with explanation codes.
//
// The returned slice is a defensive copy.
func (p OptimisedPlan) UnassignedDetails() []UnassignedItem {
	cp := make([]UnassignedItem, len(p.unassignedDetails))
	copy(cp, p.unassignedDetails)
	return cp
}
