package optimisation_test

import (
	"encoding/json"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func makeItem(id string) workitem.WorkItem {
	w, _ := workitem.New(id, "test.type", json.RawMessage(`{"key":"value"}`))
	return w
}

func TestSolve_ReturnsAllItems(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}

	result, err := optimisation.Solve(items)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 3 {
		t.Errorf("expected 3 items in plan, got %d", result.Size())
	}
}

func TestSolve_EmptyItems(t *testing.T) {
	_, err := optimisation.Solve([]workitem.WorkItem{})
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestSolve_PreservesOrder(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-A"), makeItem("WI-B"), makeItem("WI-C")}

	result, err := optimisation.Solve(items)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := result.Items()
	for i, expected := range []string{"WI-A", "WI-B", "WI-C"} {
		if got[i].ID() != expected {
			t.Errorf("expected item %d to be %s, got %s", i, expected, got[i].ID())
		}
	}
}
