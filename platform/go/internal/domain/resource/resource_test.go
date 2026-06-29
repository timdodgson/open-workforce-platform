package resource_test

import (
	"encoding/json"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
)

func validDetails() json.RawMessage {
	return json.RawMessage(`{"name":"Alice","skills":["wound care","palliative"]}`)
}

// --- Construction: valid ---

func TestNew_ValidResource(t *testing.T) {
	r, err := resource.New("RES-001", "person", validDetails())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.ID() != "RES-001" {
		t.Errorf("expected id RES-001, got %s", r.ID())
	}
	if r.Type() != "person" {
		t.Errorf("expected type person, got %s", r.Type())
	}
}

func TestNew_TrimsWhitespaceFromID(t *testing.T) {
	r, err := resource.New("  RES-002  ", "vehicle", validDetails())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.ID() != "RES-002" {
		t.Errorf("expected trimmed id RES-002, got %q", r.ID())
	}
}

func TestNew_TrimsWhitespaceFromType(t *testing.T) {
	r, err := resource.New("RES-003", "  vehicle  ", validDetails())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if r.Type() != "vehicle" {
		t.Errorf("expected trimmed type vehicle, got %q", r.Type())
	}
}

// --- Construction: invalid ---

func TestNew_EmptyID(t *testing.T) {
	_, err := resource.New("", "person", validDetails())
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestNew_WhitespaceOnlyID(t *testing.T) {
	_, err := resource.New("   ", "person", validDetails())
	if err == nil {
		t.Fatal("expected error for whitespace-only id")
	}
}

func TestNew_EmptyType(t *testing.T) {
	_, err := resource.New("RES-001", "", validDetails())
	if err == nil {
		t.Fatal("expected error for empty type")
	}
}

func TestNew_WhitespaceOnlyType(t *testing.T) {
	_, err := resource.New("RES-001", "   ", validDetails())
	if err == nil {
		t.Fatal("expected error for whitespace-only type")
	}
}

func TestNew_EmptyDetails(t *testing.T) {
	_, err := resource.New("RES-001", "person", json.RawMessage{})
	if err == nil {
		t.Fatal("expected error for empty details")
	}
}

func TestNew_NilDetails(t *testing.T) {
	_, err := resource.New("RES-001", "person", nil)
	if err == nil {
		t.Fatal("expected error for nil details")
	}
}

func TestNew_InvalidJSONDetails(t *testing.T) {
	_, err := resource.New("RES-001", "person", json.RawMessage(`{not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON details")
	}
}

// --- Immutability ---

func TestDetails_ReturnsCopy(t *testing.T) {
	r, _ := resource.New("RES-001", "person", validDetails())

	details := r.Details()
	details[0] = 'X'

	original := r.Details()
	if original[0] == 'X' {
		t.Fatal("Details() should return a defensive copy; internal state was mutated")
	}
}

func TestNew_DefensiveCopyOfDetails(t *testing.T) {
	details := json.RawMessage(`{"name":"Bob"}`)
	r, _ := resource.New("RES-001", "person", details)

	details[2] = 'X'

	got := r.Details()
	if got[2] == 'X' {
		t.Fatal("constructor should take a defensive copy; mutating original input changed the resource")
	}
}

// --- Identity ---

func TestEqual_SameID(t *testing.T) {
	r1, _ := resource.New("RES-001", "person", validDetails())
	r2, _ := resource.New("RES-001", "vehicle", json.RawMessage(`{"plate":"AB12 CDE"}`))

	if !r1.Equal(r2) {
		t.Error("resources with the same ID should be equal")
	}
}

func TestEqual_DifferentID(t *testing.T) {
	r1, _ := resource.New("RES-001", "person", validDetails())
	r2, _ := resource.New("RES-002", "person", validDetails())

	if r1.Equal(r2) {
		t.Error("resources with different IDs should not be equal")
	}
}

// --- Serialisation ---

func TestMarshalJSON(t *testing.T) {
	r, _ := resource.New("RES-001", "person", validDetails())

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("expected no error marshalling, got %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if string(raw["id"]) != `"RES-001"` {
		t.Errorf("expected id RES-001 in JSON, got %s", raw["id"])
	}
	if string(raw["type"]) != `"person"` {
		t.Errorf("expected type person in JSON, got %s", raw["type"])
	}
}

func TestUnmarshalJSON_Valid(t *testing.T) {
	input := `{"id":"RES-010","type":"vehicle","details":{"plate":"XY99 ZZZ","capacity":500}}`

	var r resource.Resource
	if err := json.Unmarshal([]byte(input), &r); err != nil {
		t.Fatalf("expected no error unmarshalling, got %v", err)
	}

	if r.ID() != "RES-010" {
		t.Errorf("expected id RES-010, got %s", r.ID())
	}
	if r.Type() != "vehicle" {
		t.Errorf("expected type vehicle, got %s", r.Type())
	}
}

func TestUnmarshalJSON_InvalidData(t *testing.T) {
	input := `{"id":"","type":"person","details":{"name":"Nobody"}}`

	var r resource.Resource
	if err := json.Unmarshal([]byte(input), &r); err == nil {
		t.Fatal("expected validation error for empty id during unmarshal")
	}
}

func TestRoundTrip(t *testing.T) {
	original, _ := resource.New("RES-099", "team", json.RawMessage(`{"members":["Alice","Bob"],"shift":"morning"}`))

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored resource.Resource
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !original.Equal(restored) {
		t.Error("round-tripped resource should be equal to original")
	}
	if original.Type() != restored.Type() {
		t.Errorf("type mismatch: %s vs %s", original.Type(), restored.Type())
	}
}
