package plan_test

import (
	"encoding/json"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

func makeItem(id string) workitem.WorkItem {
	w, _ := workitem.New(id, "test.type", json.RawMessage(`{"key":"value"}`))
	return w
}

func TestNew_Valid(t *testing.T) {
	p, err := plan.New([]workitem.WorkItem{makeItem("WI-001")})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Size() != 1 {
		t.Errorf("expected size 1, got %d", p.Size())
	}
}

func TestNew_MultipleItems(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002"), makeItem("WI-003")}
	p, err := plan.New(items)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.Size() != 3 {
		t.Errorf("expected size 3, got %d", p.Size())
	}
}

func TestNew_EmptySlice(t *testing.T) {
	_, err := plan.New([]workitem.WorkItem{})
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}

func TestNew_NilSlice(t *testing.T) {
	_, err := plan.New(nil)
	if err == nil {
		t.Fatal("expected error for nil items")
	}
}

func TestItems_ReturnsDefensiveCopy(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	p, _ := plan.New(items)

	got := p.Items()
	got[0] = makeItem("WI-999")

	original := p.Items()
	if original[0].ID() == "WI-999" {
		t.Fatal("Items() should return a defensive copy; internal state was mutated")
	}
}

func TestNew_DefensiveCopyOfInput(t *testing.T) {
	items := []workitem.WorkItem{makeItem("WI-001"), makeItem("WI-002")}
	p, _ := plan.New(items)

	items[0] = makeItem("WI-999")

	got := p.Items()
	if got[0].ID() == "WI-999" {
		t.Fatal("constructor should take a defensive copy; mutating original input changed the plan")
	}
}
