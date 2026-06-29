package optimisation_test

import (
	"encoding/json"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func makeItem(id string) workitem.WorkItem {
	w, _ := workitem.New(id, "test.type", json.RawMessage(`{"key":"value"}`))
	return w
}

func makeResource(id string) resource.Resource {
	r, _ := resource.New(id, "person", json.RawMessage(`{"name":"Test"}`))
	return r
}

func TestSolve_AssignsAllItemsToFirstResource(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	resources := []resource.Resource{makeResource("RES-001"), makeResource("RES-002")}

	result, err := optimisation.Solve(items, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 3 {
		t.Errorf("expected 3 assignments, got %d", result.Size())
	}

	for _, a := range result.Assignments() {
		if a.ResourceID() != "RES-001" {
			t.Errorf("expected all assignments to RES-001, got %s", a.ResourceID())
		}
	}
}

func TestSolve_AssignmentContainsCorrectWorkItemIDs(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-A"), makeItem("WI-B")}
	resources := []resource.Resource{makeResource("RES-001")}

	result, err := optimisation.Solve(items, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assignments := result.Assignments()
	if assignments[0].WorkItemID() != "WI-A" {
		t.Errorf("expected first assignment work item WI-A, got %s", assignments[0].WorkItemID())
	}
	if assignments[1].WorkItemID() != "WI-B" {
		t.Errorf("expected second assignment work item WI-B, got %s", assignments[1].WorkItemID())
	}
}

func TestSolve_EmptyItems(t *testing.T) {
	resources := []resource.Resource{makeResource("RES-001")}
	_, err := optimisation.Solve([]workitem.WorkItem{}, resources)
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestSolve_EmptyResources(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001")}
	_, err := optimisation.Solve(items, []resource.Resource{})
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}
