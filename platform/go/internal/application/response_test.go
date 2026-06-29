package application

import (
	"encoding/json"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
)

func TestBuildResponse_JSONSerialisable(t *testing.T) {
	a1, _ := assignment.New("R1", "WI-1")
	a2, _ := assignment.New("R1", "WI-2")
	a3, _ := assignment.New("R2", "WI-3")

	p, _ := plan.New(plan.Result{
		Assignments: []assignment.Assignment{a1, a2, a3},
		Unassigned:  []string{"WI-4"},
		UnassignedDetails: []plan.UnassignedItem{
			{WorkItemID: "WI-4", Reasons: []string{"ShiftEnded"}},
		},
		HardViolations: []plan.HardViolation{
			{Code: "UnderStaffed", Message: "1 mandatory item unassigned"},
		},
		ConstraintMatches: []plan.ConstraintMatch{
			{Constraint: "UnderStaffed", Severity: "hard", Day: -1, Description: "1 mandatory item unassigned"},
			{Constraint: "Coverage", Severity: "soft", Day: 1, Penalty: 30, Description: "coverage gap"},
		},
		TotalCapacity:  960,
		Utilisation:    50,
		Score:          75,
		ObjectiveScore: 3000,
		ObjectiveBreakdown: []plan.ObjectiveEntry{
			{Name: "Assignment", Score: 3000},
		},
		Statistics: plan.Statistics{Algorithm: "constructive", DurationMs: 1, Iterations: 1},
	})

	resp := BuildResponse(
		p,
		"constructive",
		map[string]int{"R1": 480, "R2": 480},
		map[string]int{"WI-1": 60, "WI-2": 60, "WI-3": 60},
		map[string]string{},
		map[string]string{},
		map[string]int{},
	)

	// Verify it serialises to valid JSON.
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	// Verify it deserialises back.
	var decoded OptimisationResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Algorithm != "constructive" {
		t.Errorf("expected algorithm constructive, got %s", decoded.Algorithm)
	}
	if decoded.ObjectiveScore != 3000 {
		t.Errorf("expected objective 3000, got %d", decoded.ObjectiveScore)
	}
	if decoded.Constraints.HardCount != 1 {
		t.Errorf("expected 1 hard, got %d", decoded.Constraints.HardCount)
	}
	if decoded.Constraints.SoftCount != 1 {
		t.Errorf("expected 1 soft, got %d", decoded.Constraints.SoftCount)
	}
	if decoded.Constraints.TotalPenalty != 30 {
		t.Errorf("expected penalty 30, got %d", decoded.Constraints.TotalPenalty)
	}
	if len(decoded.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(decoded.Resources))
	}
	if len(decoded.Unassigned) != 1 {
		t.Errorf("expected 1 unassigned, got %d", len(decoded.Unassigned))
	}
}

func TestBuildResponse_DeterministicOrdering(t *testing.T) {
	a1, _ := assignment.New("R1", "WI-1")
	a2, _ := assignment.New("R2", "WI-2")

	p, _ := plan.New(plan.Result{
		Assignments: []assignment.Assignment{a1, a2},
		Unassigned:  []string{},
		ConstraintMatches: []plan.ConstraintMatch{
			{Constraint: "Zebra", Severity: "soft", Penalty: 10},
			{Constraint: "Alpha", Severity: "soft", Penalty: 20},
		},
		TotalCapacity:      960,
		Score:              100,
		ObjectiveScore:     2000,
		ObjectiveBreakdown: []plan.ObjectiveEntry{{Name: "A", Score: 2000}},
		Statistics:         plan.Statistics{Algorithm: "test"},
	})

	resp := BuildResponse(p, "test",
		map[string]int{"R1": 480, "R2": 480},
		map[string]int{"WI-1": 60, "WI-2": 60},
		map[string]string{}, map[string]string{}, map[string]int{},
	)

	// Summary should be sorted alphabetically by constraint name.
	if len(resp.Constraints.Summary) != 2 {
		t.Fatalf("expected 2 summary entries, got %d", len(resp.Constraints.Summary))
	}
	if resp.Constraints.Summary[0].Constraint != "Alpha" {
		t.Errorf("expected first summary Alpha, got %s", resp.Constraints.Summary[0].Constraint)
	}
	if resp.Constraints.Summary[1].Constraint != "Zebra" {
		t.Errorf("expected second summary Zebra, got %s", resp.Constraints.Summary[1].Constraint)
	}

	// Resource order follows assignment order.
	if resp.Resources[0].ResourceID != "R1" {
		t.Errorf("expected first resource R1, got %s", resp.Resources[0].ResourceID)
	}
}

func TestBuildResponse_ResourceUtilisation(t *testing.T) {
	a1, _ := assignment.New("R1", "WI-1")
	a2, _ := assignment.New("R1", "WI-2")

	p, _ := plan.New(plan.Result{
		Assignments:        []assignment.Assignment{a1, a2},
		Unassigned:         []string{},
		TotalCapacity:      480,
		Score:              100,
		ObjectiveScore:     2000,
		ObjectiveBreakdown: []plan.ObjectiveEntry{},
		Statistics:         plan.Statistics{Algorithm: "test"},
	})

	resp := BuildResponse(p, "test",
		map[string]int{"R1": 480},
		map[string]int{"WI-1": 120, "WI-2": 180},
		map[string]string{}, map[string]string{}, map[string]int{},
	)

	if len(resp.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resp.Resources))
	}
	if resp.Resources[0].UsedMins != 300 {
		t.Errorf("expected 300 used mins, got %d", resp.Resources[0].UsedMins)
	}
	if resp.Resources[0].CapacityMins != 480 {
		t.Errorf("expected 480 capacity, got %d", resp.Resources[0].CapacityMins)
	}
}
