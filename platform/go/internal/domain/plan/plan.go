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
	HardViolations     []HardViolation
	ConstraintMatches  []ConstraintMatch
	TotalCapacity      int
	Utilisation        int
	Score              int
	ObjectiveScore     int
	ObjectiveBreakdown []ObjectiveEntry
	Statistics         Statistics
}

// UnassignedItem represents a work item that could not be assigned,
// along with explanation codes describing why.
type UnassignedItem struct {
	WorkItemID string
	Reasons    []string
}

// HardViolation represents a hard constraint violation in the plan.
type HardViolation struct {
	Code    string // e.g. "UnderStaffed", "IllegalShiftSuccession"
	Message string // human-readable description
}

// ConstraintMatch represents a single constraint violation or satisfaction.
// This is a generic model usable by any optimisation domain.
type ConstraintMatch struct {
	Constraint  string // constraint identifier (e.g. "Coverage", "Weekend", "Succession")
	Severity    string // "hard" or "soft"
	ResourceID  string // resource involved (empty if not resource-specific)
	WorkItemID  string // work item involved (empty if not item-specific)
	Day         int    // day involved (-1 if not day-specific)
	Penalty     int    // weighted penalty contribution
	Description string // human-readable explanation
}

// ConstraintReport aggregates all constraint matches for a plan.
type ConstraintReport struct {
	Matches []ConstraintMatch
}

// HardCount returns the number of hard constraint violations.
func (r ConstraintReport) HardCount() int {
	count := 0
	for _, m := range r.Matches {
		if m.Severity == "hard" {
			count++
		}
	}
	return count
}

// SoftCount returns the number of soft constraint violations.
func (r ConstraintReport) SoftCount() int {
	count := 0
	for _, m := range r.Matches {
		if m.Severity == "soft" {
			count++
		}
	}
	return count
}

// TotalPenalty returns the sum of all penalties across all matches.
func (r ConstraintReport) TotalPenalty() int {
	total := 0
	for _, m := range r.Matches {
		total += m.Penalty
	}
	return total
}

// Summary returns violation counts grouped by constraint name.
func (r ConstraintReport) Summary() map[string]int {
	summary := make(map[string]int)
	for _, m := range r.Matches {
		summary[m.Constraint]++
	}
	return summary
}

// PenaltyByConstraint returns total penalty grouped by constraint name.
func (r ConstraintReport) PenaltyByConstraint() map[string]int {
	penalties := make(map[string]int)
	for _, m := range r.Matches {
		penalties[m.Constraint] += m.Penalty
	}
	return penalties
}

// ByResource returns matches filtered to a specific resource.
func (r ConstraintReport) ByResource(resourceID string) []ConstraintMatch {
	var result []ConstraintMatch
	for _, m := range r.Matches {
		if m.ResourceID == resourceID {
			result = append(result, m)
		}
	}
	return result
}

// ByConstraint returns matches filtered to a specific constraint.
func (r ConstraintReport) ByConstraint(constraint string) []ConstraintMatch {
	var result []ConstraintMatch
	for _, m := range r.Matches {
		if m.Constraint == constraint {
			result = append(result, m)
		}
	}
	return result
}

// Statistics captures optimisation execution metrics.
type Statistics struct {
	Algorithm          string
	DurationMs         int64
	Iterations         int
	CandidatesEvaluated int
	ImprovementsAccepted int
	FinalObjectiveScore  int
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
	hardViolations     []HardViolation
	constraintReport   ConstraintReport
	totalCapacity      int
	utilisation        int
	score              int
	objectiveScore     int
	objectiveBreakdown []ObjectiveEntry
	statistics         Statistics
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

	// Defensive copy of hard violations.
	violationsCopy := make([]HardViolation, len(r.HardViolations))
	copy(violationsCopy, r.HardViolations)

	// Defensive copy of constraint matches.
	matchesCopy := make([]ConstraintMatch, len(r.ConstraintMatches))
	copy(matchesCopy, r.ConstraintMatches)

	return OptimisedPlan{
		assignments:        assignmentsCopy,
		unassigned:         unassignedCopy,
		unassignedDetails:  detailsCopy,
		hardViolations:     violationsCopy,
		constraintReport:   ConstraintReport{Matches: matchesCopy},
		totalCapacity:      r.TotalCapacity,
		utilisation:        r.Utilisation,
		score:              r.Score,
		objectiveScore:     r.ObjectiveScore,
		objectiveBreakdown: breakdownCopy,
		statistics:         r.Statistics,
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

// Statistics returns the optimisation execution statistics.
func (p OptimisedPlan) Statistics() Statistics {
	return p.statistics
}

// HardViolations returns the hard constraint violations in the plan.
//
// The returned slice is a defensive copy.
func (p OptimisedPlan) HardViolations() []HardViolation {
	cp := make([]HardViolation, len(p.hardViolations))
	copy(cp, p.hardViolations)
	return cp
}

// HasHardViolations returns true if the plan has hard constraint violations.
func (p OptimisedPlan) HasHardViolations() bool {
	return len(p.hardViolations) > 0
}

// ConstraintMatches returns all constraint matches (hard and soft).
func (p OptimisedPlan) ConstraintMatches() []ConstraintMatch {
	cp := make([]ConstraintMatch, len(p.constraintReport.Matches))
	copy(cp, p.constraintReport.Matches)
	return cp
}

// ConstraintReportView returns the full constraint report for grouping and filtering.
func (p OptimisedPlan) ConstraintReportView() ConstraintReport {
	cp := make([]ConstraintMatch, len(p.constraintReport.Matches))
	copy(cp, p.constraintReport.Matches)
	return ConstraintReport{Matches: cp}
}

// SoftConstraintCount returns the number of soft constraint violations.
func (p OptimisedPlan) SoftConstraintCount() int {
	return p.constraintReport.SoftCount()
}

// HardConstraintCount returns the number of hard constraint violations.
func (p OptimisedPlan) HardConstraintCount() int {
	return p.constraintReport.HardCount()
}

// TotalPenalty returns the total weighted penalty from all constraint matches.
func (p OptimisedPlan) TotalPenalty() int {
	return p.constraintReport.TotalPenalty()
}

// ConstraintSummary returns violation counts grouped by constraint name.
func (p OptimisedPlan) ConstraintSummary() map[string]int {
	return p.constraintReport.Summary()
}
