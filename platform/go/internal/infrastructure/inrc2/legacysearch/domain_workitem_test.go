package legacysearch

import (
	"encoding/json"
	"testing"

)

func validDetails() json.RawMessage {
	return json.RawMessage(`{"location":"Site-A","duration":60}`)
}

// --- Construction: valid ---

func TestWorkItem_New_ValidWorkItem(t *testing.T) {
	w, err := NewWorkItem("WI-001", "maintenance.visit", validDetails())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if w.ID() != "WI-001" {
		t.Errorf("expected id WI-001, got %s", w.ID())
	}
	if w.Type() != "maintenance.visit" {
		t.Errorf("expected type maintenance.visit, got %s", w.Type())
	}
}

func TestWorkItem_New_TrimsWhitespaceFromID(t *testing.T) {
	w, err := NewWorkItem("  WI-002  ", "delivery.drop", validDetails())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if w.ID() != "WI-002" {
		t.Errorf("expected trimmed id WI-002, got %q", w.ID())
	}
}

func TestWorkItem_New_TrimsWhitespaceFromType(t *testing.T) {
	w, err := NewWorkItem("WI-003", "  delivery.drop  ", validDetails())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if w.Type() != "delivery.drop" {
		t.Errorf("expected trimmed type delivery.drop, got %q", w.Type())
	}
}

// --- Construction: invalid ---

func TestWorkItem_New_EmptyID(t *testing.T) {
	_, err := NewWorkItem("", "maintenance.visit", validDetails())
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestWorkItem_New_WhitespaceOnlyID(t *testing.T) {
	_, err := NewWorkItem("   ", "maintenance.visit", validDetails())
	if err == nil {
		t.Fatal("expected error for whitespace-only id")
	}
}

func TestWorkItem_New_EmptyType(t *testing.T) {
	_, err := NewWorkItem("WI-001", "", validDetails())
	if err == nil {
		t.Fatal("expected error for empty type")
	}
}

func TestWorkItem_New_WhitespaceOnlyType(t *testing.T) {
	_, err := NewWorkItem("WI-001", "   ", validDetails())
	if err == nil {
		t.Fatal("expected error for whitespace-only type")
	}
}

func TestWorkItem_New_EmptyDetails(t *testing.T) {
	_, err := NewWorkItem("WI-001", "maintenance.visit", json.RawMessage{})
	if err == nil {
		t.Fatal("expected error for empty details")
	}
}

func TestWorkItem_New_NilDetails(t *testing.T) {
	_, err := NewWorkItem("WI-001", "maintenance.visit", nil)
	if err == nil {
		t.Fatal("expected error for nil details")
	}
}

func TestWorkItem_New_InvalidJSONDetails(t *testing.T) {
	_, err := NewWorkItem("WI-001", "maintenance.visit", json.RawMessage(`{not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON details")
	}
}

// --- Immutability ---

func TestWorkItem_Details_ReturnsCopy(t *testing.T) {
	w, _ := NewWorkItem("WI-001", "maintenance.visit", validDetails())

	details := w.Details()
	details[0] = 'X'

	original := w.Details()
	if original[0] == 'X' {
		t.Fatal("Details() should return a defensive copy; internal state was mutated")
	}
}

func TestWorkItem_New_DefensiveCopyOfDetails(t *testing.T) {
	details := json.RawMessage(`{"location":"Site-B"}`)
	w, _ := NewWorkItem("WI-001", "maintenance.visit", details)

	details[2] = 'X'

	got := w.Details()
	if got[2] == 'X' {
		t.Fatal("constructor should take a defensive copy; mutating original input changed the work item")
	}
}

// --- Identity ---

func TestWorkItem_Equal_SameID(t *testing.T) {
	w1, _ := NewWorkItem("WI-001", "maintenance.visit", validDetails())
	w2, _ := NewWorkItem("WI-001", "delivery.drop", json.RawMessage(`{"parcel":"P-100"}`))

	if !w1.Equal(w2) {
		t.Error("work items with the same ID should be equal")
	}
}

func TestWorkItem_Equal_DifferentID(t *testing.T) {
	w1, _ := NewWorkItem("WI-001", "maintenance.visit", validDetails())
	w2, _ := NewWorkItem("WI-002", "maintenance.visit", validDetails())

	if w1.Equal(w2) {
		t.Error("work items with different IDs should not be equal")
	}
}

// --- Serialisation ---

func TestWorkItem_MarshalJSON(t *testing.T) {
	w, _ := NewWorkItem("WI-001", "maintenance.visit", validDetails())

	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("expected no error marshalling, got %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if string(raw["id"]) != `"WI-001"` {
		t.Errorf("expected id WI-001 in JSON, got %s", raw["id"])
	}
	if string(raw["type"]) != `"maintenance.visit"` {
		t.Errorf("expected type maintenance.visit in JSON, got %s", raw["type"])
	}
}

func TestWorkItem_UnmarshalJSON_Valid(t *testing.T) {
	input := `{"id":"WI-010","type":"shift.cover","details":{"shiftId":"S-300","urgency":"high"}}`

	var w WorkItem
	if err := json.Unmarshal([]byte(input), &w); err != nil {
		t.Fatalf("expected no error unmarshalling, got %v", err)
	}

	if w.ID() != "WI-010" {
		t.Errorf("expected id WI-010, got %s", w.ID())
	}
	if w.Type() != "shift.cover" {
		t.Errorf("expected type shift.cover, got %s", w.Type())
	}
}

func TestWorkItem_UnmarshalJSON_InvalidData(t *testing.T) {
	input := `{"id":"","type":"shift.cover","details":{"shiftId":"S-300"}}`

	var w WorkItem
	if err := json.Unmarshal([]byte(input), &w); err == nil {
		t.Fatal("expected validation error for empty id during unmarshal")
	}
}

func TestWorkItem_RoundTrip(t *testing.T) {
	original, _ := NewWorkItem("WI-099", "patient.visit", json.RawMessage(`{"patientId":"P-500","skills":["wound care"]}`))

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored WorkItem
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !original.Equal(restored) {
		t.Error("round-tripped work item should be equal to original")
	}
	if original.Type() != restored.Type() {
		t.Errorf("type mismatch: %s vs %s", original.Type(), restored.Type())
	}
}
