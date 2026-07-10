package legacysearch

import "errors"

// PlanResult holds the inputs required to construct an OptimisedPlan.
type PlanResult struct {
	Assignments        []Assignment
	Unassigned         []string
	UnassignedDetails  []UnassignedItem
	HardViolations     []HardViolation
	ConstraintMatches  []ConstraintMatch
	TotalCapacity      int
	Utilisation        int
	Score              int
	ObjectiveScore     int
	ObjectiveBreakdown []ObjectiveEntry
	Statistics         PlanStatistics
}

// UnassignedItem represents a work item that could not be assigned,
// along with explanation codes describing why.
type UnassignedItem struct {
	WorkItemID string
	Reasons    []string
}

// HardViolation represents a hard constraint violation in the plan.
type HardViolation struct {
	Code    string
	Message string
}

// ConstraintMatch represents a single constraint violation or satisfaction.
type ConstraintMatch struct {
	Constraint  string
	Severity    string
	ResourceID  string
	WorkItemID  string
	Day         int
	Penalty     int
	Description string
}

// ConstraintReport aggregates all constraint matches for a plan.
type ConstraintReport struct {
	Matches []ConstraintMatch
}

func (r ConstraintReport) HardCount() int {
	count := 0
	for _, m := range r.Matches {
		if m.Severity == "hard" {
			count++
		}
	}
	return count
}

func (r ConstraintReport) SoftCount() int {
	count := 0
	for _, m := range r.Matches {
		if m.Severity == "soft" {
			count++
		}
	}
	return count
}

func (r ConstraintReport) TotalPenalty() int {
	total := 0
	for _, m := range r.Matches {
		total += m.Penalty
	}
	return total
}

func (r ConstraintReport) Summary() map[string]int {
	summary := make(map[string]int)
	for _, m := range r.Matches {
		summary[m.Constraint]++
	}
	return summary
}

func (r ConstraintReport) PenaltyByConstraint() map[string]int {
	penalties := make(map[string]int)
	for _, m := range r.Matches {
		penalties[m.Constraint] += m.Penalty
	}
	return penalties
}

func (r ConstraintReport) ByResource(resourceID string) []ConstraintMatch {
	var result []ConstraintMatch
	for _, m := range r.Matches {
		if m.ResourceID == resourceID {
			result = append(result, m)
		}
	}
	return result
}

func (r ConstraintReport) ByConstraint(constraint string) []ConstraintMatch {
	var result []ConstraintMatch
	for _, m := range r.Matches {
		if m.Constraint == constraint {
			result = append(result, m)
		}
	}
	return result
}

// PlanStatistics captures optimisation execution metrics.
type PlanStatistics struct {
	Algorithm            string
	DurationMs           int64
	Iterations           int
	CandidatesEvaluated  int
	ImprovementsAccepted int
	FinalObjectiveScore  int
}

// ObjectiveEntry represents a named objective's contribution to the total score.
type ObjectiveEntry struct {
	Name  string
	Score int
}

// OptimisedPlan represents the result of an optimisation run.
type OptimisedPlan struct {
	assignments        []Assignment
	unassigned         []string
	unassignedDetails  []UnassignedItem
	hardViolations     []HardViolation
	constraintReport   ConstraintReport
	totalCapacity      int
	utilisation        int
	score              int
	objectiveScore     int
	objectiveBreakdown []ObjectiveEntry
	statistics         PlanStatistics
}

// NewOptimisedPlan creates a validated OptimisedPlan from an optimisation result.
func NewOptimisedPlan(r PlanResult) (OptimisedPlan, error) {
	if len(r.Assignments) == 0 && len(r.Unassigned) == 0 {
		return OptimisedPlan{}, errors.New("optimised plan must contain at least one assignment or unassigned item")
	}

	assignmentsCopy := make([]Assignment, len(r.Assignments))
	copy(assignmentsCopy, r.Assignments)

	unassignedCopy := make([]string, len(r.Unassigned))
	copy(unassignedCopy, r.Unassigned)

	breakdownCopy := make([]ObjectiveEntry, len(r.ObjectiveBreakdown))
	copy(breakdownCopy, r.ObjectiveBreakdown)

	detailsCopy := make([]UnassignedItem, len(r.UnassignedDetails))
	copy(detailsCopy, r.UnassignedDetails)

	violationsCopy := make([]HardViolation, len(r.HardViolations))
	copy(violationsCopy, r.HardViolations)

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

func (p OptimisedPlan) Assignments() []Assignment {
	cp := make([]Assignment, len(p.assignments))
	copy(cp, p.assignments)
	return cp
}

func (p OptimisedPlan) Unassigned() []string {
	cp := make([]string, len(p.unassigned))
	copy(cp, p.unassigned)
	return cp
}

func (p OptimisedPlan) Size() int { return len(p.assignments) }

func (p OptimisedPlan) UnassignedCount() int { return len(p.unassigned) }

func (p OptimisedPlan) TotalCapacity() int { return p.totalCapacity }

func (p OptimisedPlan) Utilisation() int { return p.utilisation }

func (p OptimisedPlan) Score() int { return p.score }

func (p OptimisedPlan) ObjectiveScore() int { return p.objectiveScore }

func (p OptimisedPlan) ObjectiveBreakdown() []ObjectiveEntry {
	cp := make([]ObjectiveEntry, len(p.objectiveBreakdown))
	copy(cp, p.objectiveBreakdown)
	return cp
}

func (p OptimisedPlan) UnassignedDetails() []UnassignedItem {
	cp := make([]UnassignedItem, len(p.unassignedDetails))
	copy(cp, p.unassignedDetails)
	return cp
}

func (p OptimisedPlan) Statistics() PlanStatistics { return p.statistics }

func (p OptimisedPlan) HardViolations() []HardViolation {
	cp := make([]HardViolation, len(p.hardViolations))
	copy(cp, p.hardViolations)
	return cp
}

func (p OptimisedPlan) HasHardViolations() bool { return len(p.hardViolations) > 0 }

func (p OptimisedPlan) ConstraintMatches() []ConstraintMatch {
	cp := make([]ConstraintMatch, len(p.constraintReport.Matches))
	copy(cp, p.constraintReport.Matches)
	return cp
}

func (p OptimisedPlan) ConstraintReportView() ConstraintReport {
	cp := make([]ConstraintMatch, len(p.constraintReport.Matches))
	copy(cp, p.constraintReport.Matches)
	return ConstraintReport{Matches: cp}
}

func (p OptimisedPlan) SoftConstraintCount() int { return p.constraintReport.SoftCount() }

func (p OptimisedPlan) HardConstraintCount() int { return p.constraintReport.HardCount() }

func (p OptimisedPlan) TotalPenalty() int { return p.constraintReport.TotalPenalty() }

func (p OptimisedPlan) ConstraintSummary() map[string]int { return p.constraintReport.Summary() }
