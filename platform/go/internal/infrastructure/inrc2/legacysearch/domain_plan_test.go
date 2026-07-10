package legacysearch

import (
	"testing"

)

func makeAssignment(resourceID, workItemID string) Assignment {
	a, _ := NewAssignment(resourceID, workItemID)
	return a
}

func TestPlan_New_ValidWithAssignments(t *testing.T) {
	p, err := NewOptimisedPlan(PlanResult{
		Assignments:   []Assignment{makeAssignment("RES-001", "WI-001")},
		TotalCapacity: 2,
		Utilisation:   50,
		Score:         100,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Size() != 1 {
		t.Errorf("expected size 1, got %d", p.Size())
	}
	if p.Score() != 100 {
		t.Errorf("expected score 100, got %d", p.Score())
	}
	if p.Utilisation() != 50 {
		t.Errorf("expected utilisation 50, got %d", p.Utilisation())
	}
	if p.TotalCapacity() != 2 {
		t.Errorf("expected total capacity 2, got %d", p.TotalCapacity())
	}
}

func TestPlan_New_ValidWithUnassignedOnly(t *testing.T) {
	p, err := NewOptimisedPlan(PlanResult{
		Unassigned:    []string{"WI-001", "WI-002"},
		TotalCapacity: 0,
		Score:         0,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.UnassignedCount() != 2 {
		t.Errorf("expected 2 unassigned, got %d", p.UnassignedCount())
	}
	if p.Size() != 0 {
		t.Errorf("expected 0 assignments, got %d", p.Size())
	}
}

func TestPlan_New_EmptyPlan(t *testing.T) {
	_, err := NewOptimisedPlan(PlanResult{})
	if err == nil {
		t.Fatal("expected error for completely empty plan")
	}
}

func TestPlan_Assignments_ReturnsDefensiveCopy(t *testing.T) {
	assignments := []Assignment{
		makeAssignment("RES-001", "WI-001"),
		makeAssignment("RES-002", "WI-002"),
	}
	p, _ := NewOptimisedPlan(PlanResult{Assignments: assignments, TotalCapacity: 4, Score: 100})

	got := p.Assignments()
	got[0] = makeAssignment("RES-999", "WI-999")

	original := p.Assignments()
	if original[0].ResourceID() == "RES-999" {
		t.Fatal("Assignments() should return a defensive copy")
	}
}

func TestPlan_Unassigned_ReturnsDefensiveCopy(t *testing.T) {
	p, _ := NewOptimisedPlan(PlanResult{
		Assignments: []Assignment{makeAssignment("RES-001", "WI-001")},
		Unassigned:  []string{"WI-002"},
		Score:       50,
	})

	got := p.Unassigned()
	got[0] = "MUTATED"

	original := p.Unassigned()
	if original[0] == "MUTATED" {
		t.Fatal("Unassigned() should return a defensive copy")
	}
}

func TestPlan_New_DefensiveCopyOfInput(t *testing.T) {
	assignments := []Assignment{
		makeAssignment("RES-001", "WI-001"),
	}
	p, _ := NewOptimisedPlan(PlanResult{Assignments: assignments, TotalCapacity: 2, Score: 100})

	assignments[0] = makeAssignment("RES-999", "WI-999")

	got := p.Assignments()
	if got[0].ResourceID() == "RES-999" {
		t.Fatal("constructor should take a defensive copy")
	}
}
