package application

import (
	"sort"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
)

// OptimisationResponse is the complete, JSON-serialisable result of an optimisation run.
// This model contains all information needed by a UI or REST API without recalculating
// business logic. The CLI is just another consumer of this response.
type OptimisationResponse struct {
	Algorithm          string                    `json:"algorithm"`
	AssignmentScore    int                       `json:"assignmentScore"`
	ObjectiveScore     int                       `json:"objectiveScore"`
	TotalCapacity      int                       `json:"totalCapacity"`
	Utilisation        int                       `json:"utilisation"`
	ObjectiveBreakdown []ObjectiveEntry          `json:"objectiveBreakdown"`
	Constraints        ConstraintReport          `json:"constraints"`
	Resources          []ResourceUtilisation     `json:"resources"`
	Unassigned         []UnassignedWorkItem      `json:"unassigned"`
	Travel             []ResourceTravel          `json:"travel"`
	Statistics         OptimisationStatistics    `json:"statistics"`
}

// ObjectiveEntry represents a named objective contribution.
type ObjectiveEntry struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// ConstraintReport contains all constraint violation data.
type ConstraintReport struct {
	HardCount    int                 `json:"hardCount"`
	SoftCount    int                 `json:"softCount"`
	TotalPenalty int                 `json:"totalPenalty"`
	Summary      []ConstraintSummary `json:"summary"`
	Matches      []ConstraintMatch   `json:"matches"`
}

// ConstraintSummary is a grouped count and penalty by constraint name.
type ConstraintSummary struct {
	Constraint string `json:"constraint"`
	Count      int    `json:"count"`
	Penalty    int    `json:"penalty"`
}

// ConstraintMatch represents a single constraint violation.
type ConstraintMatch struct {
	Constraint  string `json:"constraint"`
	Severity    string `json:"severity"`
	ResourceID  string `json:"resourceId,omitempty"`
	WorkItemID  string `json:"workItemId,omitempty"`
	Day         int    `json:"day"`
	Penalty     int    `json:"penalty"`
	Description string `json:"description"`
}

// ResourceUtilisation shows assignments and used capacity per resource.
type ResourceUtilisation struct {
	ResourceID string   `json:"resourceId"`
	UsedMins   int      `json:"usedMins"`
	CapacityMins int   `json:"capacityMins"`
	WorkItems  []string `json:"workItems"`
}

// UnassignedWorkItem represents a work item that could not be assigned.
type UnassignedWorkItem struct {
	WorkItemID string   `json:"workItemId"`
	Reasons    []string `json:"reasons"`
}

// ResourceTravel shows travel information per resource.
type ResourceTravel struct {
	ResourceID string      `json:"resourceId"`
	TotalMins  int         `json:"totalMins"`
	Legs       []TravelLeg `json:"legs"`
}

// TravelLeg represents one travel segment.
type TravelLeg struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Minutes int    `json:"minutes"`
}

// OptimisationStatistics captures algorithm execution metrics.
type OptimisationStatistics struct {
	Algorithm            string `json:"algorithm"`
	DurationMs           int64  `json:"durationMs"`
	Iterations           int    `json:"iterations"`
	CandidatesEvaluated  int    `json:"candidatesEvaluated"`
	ImprovementsAccepted int    `json:"improvementsAccepted"`
	FinalObjectiveScore  int    `json:"finalObjectiveScore"`
}

// BuildResponse constructs an OptimisationResponse from a plan and input context.
// This moves all business interpretation out of the CLI into the application layer.
func BuildResponse(
	result plan.OptimisedPlan,
	algorithm string,
	capacityOf map[string]int,
	durationOf map[string]int,
	locationOf map[string]string,
	resourceLocationOf map[string]string,
	travelLookup map[string]int,
) OptimisationResponse {
	// Objective breakdown.
	planBreakdown := result.ObjectiveBreakdown()
	breakdown := make([]ObjectiveEntry, len(planBreakdown))
	for i, e := range planBreakdown {
		breakdown[i] = ObjectiveEntry{Name: e.Name, Score: e.Score}
	}

	// Constraint report.
	planMatches := result.ConstraintMatches()
	matches := make([]ConstraintMatch, len(planMatches))
	for i, m := range planMatches {
		matches[i] = ConstraintMatch{
			Constraint:  m.Constraint,
			Severity:    m.Severity,
			ResourceID:  m.ResourceID,
			WorkItemID:  m.WorkItemID,
			Day:         m.Day,
			Penalty:     m.Penalty,
			Description: m.Description,
		}
	}

	// Build summary (sorted for determinism).
	summaryMap := make(map[string]*ConstraintSummary)
	for _, m := range matches {
		s, ok := summaryMap[m.Constraint]
		if !ok {
			s = &ConstraintSummary{Constraint: m.Constraint}
			summaryMap[m.Constraint] = s
		}
		s.Count++
		s.Penalty += m.Penalty
	}
	summary := make([]ConstraintSummary, 0, len(summaryMap))
	for _, s := range summaryMap {
		summary = append(summary, *s)
	}
	sort.Slice(summary, func(i, j int) bool {
		return summary[i].Constraint < summary[j].Constraint
	})

	constraints := ConstraintReport{
		HardCount:    result.HardConstraintCount(),
		SoftCount:    result.SoftConstraintCount(),
		TotalPenalty: result.TotalPenalty(),
		Summary:      summary,
		Matches:      matches,
	}

	// Resource utilisation (grouped by resource, ordered by first assignment).
	var resources []ResourceUtilisation
	var resourceOrder []string
	resourceItems := make(map[string][]string)

	for _, a := range result.Assignments() {
		if _, exists := resourceItems[a.ResourceID()]; !exists {
			resourceOrder = append(resourceOrder, a.ResourceID())
		}
		resourceItems[a.ResourceID()] = append(resourceItems[a.ResourceID()], a.WorkItemID())
	}

	for _, resID := range resourceOrder {
		items := resourceItems[resID]
		usedMins := 0
		for _, itemID := range items {
			usedMins += durationOf[itemID]
		}
		resources = append(resources, ResourceUtilisation{
			ResourceID:   resID,
			UsedMins:     usedMins,
			CapacityMins: capacityOf[resID],
			WorkItems:    items,
		})
	}

	// Travel breakdown.
	var travel []ResourceTravel
	for _, resID := range resourceOrder {
		items := resourceItems[resID]
		current := resourceLocationOf[resID]
		var legs []TravelLeg
		totalMins := 0

		for _, itemID := range items {
			dest := locationOf[itemID]
			if dest != "" && current != "" && dest != current {
				mins := travelLookup[current+"|"+dest]
				if mins > 0 {
					legs = append(legs, TravelLeg{From: current, To: dest, Minutes: mins})
					totalMins += mins
				}
			}
			if dest != "" {
				current = dest
			}
		}

		travel = append(travel, ResourceTravel{
			ResourceID: resID,
			TotalMins:  totalMins,
			Legs:       legs,
		})
	}

	// Unassigned work items.
	planUnassigned := result.UnassignedDetails()
	unassigned := make([]UnassignedWorkItem, len(planUnassigned))
	for i, u := range planUnassigned {
		reasons := make([]string, len(u.Reasons))
		copy(reasons, u.Reasons)
		unassigned[i] = UnassignedWorkItem{
			WorkItemID: u.WorkItemID,
			Reasons:    reasons,
		}
	}

	// Statistics.
	s := result.Statistics()
	stats := OptimisationStatistics{
		Algorithm:            s.Algorithm,
		DurationMs:           s.DurationMs,
		Iterations:           s.Iterations,
		CandidatesEvaluated:  s.CandidatesEvaluated,
		ImprovementsAccepted: s.ImprovementsAccepted,
		FinalObjectiveScore:  s.FinalObjectiveScore,
	}

	return OptimisationResponse{
		Algorithm:          algorithm,
		AssignmentScore:    result.Score(),
		ObjectiveScore:     result.ObjectiveScore(),
		TotalCapacity:      result.TotalCapacity(),
		Utilisation:        result.Utilisation(),
		ObjectiveBreakdown: breakdown,
		Constraints:        constraints,
		Resources:          resources,
		Unassigned:         unassigned,
		Travel:             travel,
		Statistics:         stats,
	}
}
