package plan_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
)

func makeAssignment(resourceID, workItemID string) assignment.Assignment {
	a, _ := assignment.New(resourceID, workItemID)
	return a
}

func TestNew_Valid(t *testing.T) {
	p, err := plan.New([]assignment.Assignment{makeAssignment("RES-001", "WI-001")})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Size() != 1 {
		t.Errorf("expected size 1, got %d", p.Size())
	}
}

func TestNew_MultipleAssignments(t *testing.T) {
	assignments := []assignment.Assignment{
		makeAssignment("RES-001", "WI-001"),
		makeAssignment("RES-001", "WI-002"),
		makeAssignment("RES-002", "WI-003"),
	}
	p, err := plan.New(assignments)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Size() != 3 {
		t.Errorf("expected size 3, got %d", p.Size())
	}
}

func TestNew_EmptySlice(t *testing.T) {
	_, err := plan.New([]assignment.Assignment{})
	if err == nil {
		t.Fatal("expected error for empty assignments")
	}
}

func TestNew_NilSlice(t *testing.T) {
	_, err := plan.New(nil)
	if err == nil {
		t.Fatal("expected error for nil assignments")
	}
}

func TestAssignments_ReturnsDefensiveCopy(t *testing.T) {
	assignments := []assignment.Assignment{
		makeAssignment("RES-001", "WI-001"),
		makeAssignment("RES-002", "WI-002"),
	}
	p, _ := plan.New(assignments)

	got := p.Assignments()
	got[0] = makeAssignment("RES-999", "WI-999")

	original := p.Assignments()
	if original[0].ResourceID() == "RES-999" {
		t.Fatal("Assignments() should return a defensive copy; internal state was mutated")
	}
}

func TestNew_DefensiveCopyOfInput(t *testing.T) {
	assignments := []assignment.Assignment{
		makeAssignment("RES-001", "WI-001"),
		makeAssignment("RES-002", "WI-002"),
	}
	p, _ := plan.New(assignments)

	assignments[0] = makeAssignment("RES-999", "WI-999")

	got := p.Assignments()
	if got[0].ResourceID() == "RES-999" {
		t.Fatal("constructor should take a defensive copy; mutating original input changed the plan")
	}
}
