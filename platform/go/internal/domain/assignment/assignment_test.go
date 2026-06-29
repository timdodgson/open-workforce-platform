package assignment_test

import (
	"encoding/json"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
)

// --- Construction: valid ---

func TestNew_Valid(t *testing.T) {
	a, err := assignment.New("RES-001", "WI-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.ResourceID() != "RES-001" {
		t.Errorf("expected resource id RES-001, got %s", a.ResourceID())
	}
	if a.WorkItemID() != "WI-001" {
		t.Errorf("expected work item id WI-001, got %s", a.WorkItemID())
	}
}

func TestNew_TrimsWhitespace(t *testing.T) {
	a, err := assignment.New("  RES-002  ", "  WI-002  ")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if a.ResourceID() != "RES-002" {
		t.Errorf("expected trimmed resource id RES-002, got %q", a.ResourceID())
	}
	if a.WorkItemID() != "WI-002" {
		t.Errorf("expected trimmed work item id WI-002, got %q", a.WorkItemID())
	}
}

// --- Construction: invalid ---

func TestNew_EmptyResourceID(t *testing.T) {
	_, err := assignment.New("", "WI-001")
	if err == nil {
		t.Fatal("expected error for empty resource id")
	}
}

func TestNew_WhitespaceOnlyResourceID(t *testing.T) {
	_, err := assignment.New("   ", "WI-001")
	if err == nil {
		t.Fatal("expected error for whitespace-only resource id")
	}
}

func TestNew_EmptyWorkItemID(t *testing.T) {
	_, err := assignment.New("RES-001", "")
	if err == nil {
		t.Fatal("expected error for empty work item id")
	}
}

func TestNew_WhitespaceOnlyWorkItemID(t *testing.T) {
	_, err := assignment.New("RES-001", "   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only work item id")
	}
}

// --- Serialisation ---

func TestMarshalJSON(t *testing.T) {
	a, _ := assignment.New("RES-001", "WI-001")

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("expected no error marshalling, got %v", err)
	}

	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("expected valid JSON, got %v", err)
	}

	if raw["resourceId"] != "RES-001" {
		t.Errorf("expected resourceId RES-001, got %s", raw["resourceId"])
	}
	if raw["workItemId"] != "WI-001" {
		t.Errorf("expected workItemId WI-001, got %s", raw["workItemId"])
	}
}

func TestUnmarshalJSON_Valid(t *testing.T) {
	input := `{"resourceId":"RES-010","workItemId":"WI-010"}`

	var a assignment.Assignment
	if err := json.Unmarshal([]byte(input), &a); err != nil {
		t.Fatalf("expected no error unmarshalling, got %v", err)
	}

	if a.ResourceID() != "RES-010" {
		t.Errorf("expected resource id RES-010, got %s", a.ResourceID())
	}
	if a.WorkItemID() != "WI-010" {
		t.Errorf("expected work item id WI-010, got %s", a.WorkItemID())
	}
}

func TestUnmarshalJSON_Invalid(t *testing.T) {
	input := `{"resourceId":"","workItemId":"WI-010"}`

	var a assignment.Assignment
	if err := json.Unmarshal([]byte(input), &a); err == nil {
		t.Fatal("expected validation error for empty resource id during unmarshal")
	}
}

func TestRoundTrip(t *testing.T) {
	original, _ := assignment.New("RES-099", "WI-099")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored assignment.Assignment
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if original.ResourceID() != restored.ResourceID() {
		t.Errorf("resource id mismatch: %s vs %s", original.ResourceID(), restored.ResourceID())
	}
	if original.WorkItemID() != restored.WorkItemID() {
		t.Errorf("work item id mismatch: %s vs %s", original.WorkItemID(), restored.WorkItemID())
	}
}
