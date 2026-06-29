package plan

// ResultView is a JSON-serialisable representation of an optimisation result.
// This is the model a future REST API would return directly.
// It contains all information required by a UI without recalculating business rules.
type ResultView struct {
	// Core metrics.
	AssignmentScore int `json:"assignmentScore"`
	ObjectiveScore  int `json:"objectiveScore"`
	TotalCapacity   int `json:"totalCapacity"`
	Utilisation     int `json:"utilisation"`

	// Assignments.
	Assignments []AssignmentView `json:"assignments"`
	Unassigned  []UnassignedView `json:"unassigned"`

	// Objective breakdown.
	ObjectiveBreakdown []ObjectiveEntryView `json:"objectiveBreakdown"`

	// Constraint reporting.
	Constraints ConstraintReportView `json:"constraints"`

	// Optimisation statistics.
	Statistics StatisticsView `json:"statistics"`
}

// AssignmentView represents a single assignment in API output.
type AssignmentView struct {
	ResourceID string `json:"resourceId"`
	WorkItemID string `json:"workItemId"`
}

// UnassignedView represents an unassigned work item with reasons.
type UnassignedView struct {
	WorkItemID string   `json:"workItemId"`
	Reasons    []string `json:"reasons"`
}

// ObjectiveEntryView represents a single objective contribution.
type ObjectiveEntryView struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// ConstraintReportView is the JSON-serialisable constraint report.
type ConstraintReportView struct {
	HardCount    int                     `json:"hardCount"`
	SoftCount    int                     `json:"softCount"`
	TotalPenalty int                     `json:"totalPenalty"`
	Summary      []ConstraintSummaryView `json:"summary"`
	Matches      []ConstraintMatchView   `json:"matches"`
}

// ConstraintSummaryView represents a grouped constraint summary entry.
type ConstraintSummaryView struct {
	Constraint string `json:"constraint"`
	Count      int    `json:"count"`
	Penalty    int    `json:"penalty"`
}

// ConstraintMatchView is the JSON-serialisable constraint match.
type ConstraintMatchView struct {
	Constraint  string `json:"constraint"`
	Severity    string `json:"severity"`
	ResourceID  string `json:"resourceId,omitempty"`
	WorkItemID  string `json:"workItemId,omitempty"`
	Day         int    `json:"day"`
	Penalty     int    `json:"penalty"`
	Description string `json:"description"`
}

// StatisticsView is the JSON-serialisable statistics model.
type StatisticsView struct {
	Algorithm            string `json:"algorithm"`
	DurationMs           int64  `json:"durationMs"`
	Iterations           int    `json:"iterations"`
	CandidatesEvaluated  int    `json:"candidatesEvaluated"`
	ImprovementsAccepted int    `json:"improvementsAccepted"`
	FinalObjectiveScore  int    `json:"finalObjectiveScore"`
}

// ToResultView converts an OptimisedPlan to a JSON-serialisable ResultView.
// This is the single transformation point between the domain model and API output.
func (p OptimisedPlan) ToResultView() ResultView {
	// Assignments.
	assignments := make([]AssignmentView, len(p.assignments))
	for i, a := range p.assignments {
		assignments[i] = AssignmentView{
			ResourceID: a.ResourceID(),
			WorkItemID: a.WorkItemID(),
		}
	}

	// Unassigned.
	unassigned := make([]UnassignedView, len(p.unassignedDetails))
	for i, u := range p.unassignedDetails {
		reasons := make([]string, len(u.Reasons))
		copy(reasons, u.Reasons)
		unassigned[i] = UnassignedView{
			WorkItemID: u.WorkItemID,
			Reasons:    reasons,
		}
	}

	// Objective breakdown.
	breakdown := make([]ObjectiveEntryView, len(p.objectiveBreakdown))
	for i, e := range p.objectiveBreakdown {
		breakdown[i] = ObjectiveEntryView{Name: e.Name, Score: e.Score}
	}

	// Constraint report.
	report := p.constraintReport
	summaryMap := report.Summary()
	penaltyMap := report.PenaltyByConstraint()
	summaryViews := make([]ConstraintSummaryView, 0, len(summaryMap))
	for constraint, count := range summaryMap {
		summaryViews = append(summaryViews, ConstraintSummaryView{
			Constraint: constraint,
			Count:      count,
			Penalty:    penaltyMap[constraint],
		})
	}

	matchViews := make([]ConstraintMatchView, len(report.Matches))
	for i, m := range report.Matches {
		matchViews[i] = ConstraintMatchView{
			Constraint:  m.Constraint,
			Severity:    m.Severity,
			ResourceID:  m.ResourceID,
			WorkItemID:  m.WorkItemID,
			Day:         m.Day,
			Penalty:     m.Penalty,
			Description: m.Description,
		}
	}

	constraintView := ConstraintReportView{
		HardCount:    report.HardCount(),
		SoftCount:    report.SoftCount(),
		TotalPenalty: report.TotalPenalty(),
		Summary:      summaryViews,
		Matches:      matchViews,
	}

	// Statistics.
	stats := StatisticsView{
		Algorithm:            p.statistics.Algorithm,
		DurationMs:           p.statistics.DurationMs,
		Iterations:           p.statistics.Iterations,
		CandidatesEvaluated:  p.statistics.CandidatesEvaluated,
		ImprovementsAccepted: p.statistics.ImprovementsAccepted,
		FinalObjectiveScore:  p.statistics.FinalObjectiveScore,
	}

	return ResultView{
		AssignmentScore:    p.score,
		ObjectiveScore:     p.objectiveScore,
		TotalCapacity:      p.totalCapacity,
		Utilisation:        p.utilisation,
		Assignments:        assignments,
		Unassigned:         unassigned,
		ObjectiveBreakdown: breakdown,
		Constraints:        constraintView,
		Statistics:         stats,
	}
}
