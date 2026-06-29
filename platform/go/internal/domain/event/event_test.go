package event_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
)

var validTime = time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC)

func validDetails() json.RawMessage {
	return json.RawMessage(`{"patient":"P-001","priority":"urgent"}`)
}

// --- Construction: valid ---

func TestNew_ValidEvent(t *testing.T) {
	e, err := event.New("EVT-001", "patient.referred", validTime, validDetails())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if e.ID() != "EVT-001" {
		t.Errorf("expected id EVT-001, got %s", e.ID())
	}
	if e.Type() != "patient.referred" {
		t.Errorf("expected type patient.referred, got %s", e.Type())
	}
	if !e.OccurredAt().Equal(validTime) {
		t.Errorf("expected occurredAt %v, got %v", validTime, e.OccurredAt())
	}
}

func TestNew_TrimsWhitespaceFromID(t *testing.T) {
	e, err := event.New("  EVT-002  ", "order.placed", validTime, validDetails())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if e.ID() != "EVT-002" {
		t.Errorf("expected trimmed id EVT-002, got %q", e.ID())
	}
}

func TestNew_TrimsWhitespaceFromType(t *testing.T) {
	e, err := event.New("EVT-003", "  order.placed  ", validTime, validDetails())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if e.Type() != "order.placed" {
		t.Errorf("expected trimmed type order.placed, got %q", e.Type())
	}
}

// --- Construction: invalid ---

func TestNew_EmptyID(t *testing.T) {
	_, err := event.New("", "patient.referred", validTime, validDetails())
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestNew_WhitespaceOnlyID(t *testing.T) {
	_, err := event.New("   ", "patient.referred", validTime, validDetails())
	if err == nil {
		t.Fatal("expected error for whitespace-only id")
	}
}

func TestNew_EmptyType(t *testing.T) {
	_, err := event.New("EVT-001", "", validTime, validDetails())
	if err == nil {
		t.Fatal("expected error for empty type")
	}
}

func TestNew_WhitespaceOnlyType(t *testing.T) {
	_, err := event.New("EVT-001", "   ", validTime, validDetails())
	if err == nil {
		t.Fatal("expected error for whitespace-only type")
	}
}

func TestNew_ZeroTime(t *testing.T) {
	_, err := event.New("EVT-001", "patient.referred", time.Time{}, validDetails())
	if err == nil {
		t.Fatal("expected error for zero time")
	}
}

func TestNew_EmptyDetails(t *testing.T) {
	_, err := event.New("EVT-001", "patient.referred", validTime, json.RawMessage{})
	if err == nil {
		t.Fatal("expected error for empty details")
	}
}

func TestNew_NilDetails(t *testing.T) {
	_, err := event.New("EVT-001", "patient.referred", validTime, nil)
	if err == nil {
		t.Fatal("expected error for nil details")
	}
}

func TestNew_InvalidJSONDetails(t *testing.T) {
	_, err := event.New("EVT-001", "patient.referred", validTime, json.RawMessage(`{not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON details")
	}
}

// --- Immutability ---

func TestDetails_ReturnsCopy(t *testing.T) {
	e, _ := event.New("EVT-001", "patient.referred", validTime, validDetails())

	details := e.Details()
	details[0] = 'X' // mutate the returned slice

	original := e.Details()
	if original[0] == 'X' {
		t.Fatal("Details() should return a defensive copy; internal state was mutated")
	}
}

func TestNew_DefensiveCopyOfDetails(t *testing.T) {
	details := json.RawMessage(`{"patient":"P-001"}`)
	e, _ := event.New("EVT-001", "patient.referred", validTime, details)

	// Mutate the original slice passed to the constructor.
	details[2] = 'X'

	got := e.Details()
	if got[2] == 'X' {
		t.Fatal("constructor should take a defensive copy; mutating original input changed the event")
	}
}

// --- Identity ---

func TestEqual_SameID(t *testing.T) {
	e1, _ := event.New("EVT-001", "patient.referred", validTime, validDetails())
	e2, _ := event.New("EVT-001", "order.placed", validTime, json.RawMessage(`{"order":"O-100"}`))

	if !e1.Equal(e2) {
		t.Error("events with the same ID should be equal")
	}
}

func TestEqual_DifferentID(t *testing.T) {
	e1, _ := event.New("EVT-001", "patient.referred", validTime, validDetails())
	e2, _ := event.New("EVT-002", "patient.referred", validTime, validDetails())

	if e1.Equal(e2) {
		t.Error("events with different IDs should not be equal")
	}
}

// --- Serialisation ---

func TestMarshalJSON(t *testing.T) {
	e, _ := event.New("EVT-001", "patient.referred", validTime, validDetails())

	data, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("expected no error marshalling, got %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("expected valid JSON output, got %v", err)
	}

	if string(raw["id"]) != `"EVT-001"` {
		t.Errorf("expected id EVT-001 in JSON, got %s", raw["id"])
	}
	if string(raw["type"]) != `"patient.referred"` {
		t.Errorf("expected type patient.referred in JSON, got %s", raw["type"])
	}
}

func TestUnmarshalJSON_Valid(t *testing.T) {
	input := `{"id":"EVT-010","type":"maintenance.scheduled","occurredAt":"2026-06-15T10:30:00Z","details":{"asset":"A-500"}}`

	var e event.BusinessEvent
	if err := json.Unmarshal([]byte(input), &e); err != nil {
		t.Fatalf("expected no error unmarshalling, got %v", err)
	}

	if e.ID() != "EVT-010" {
		t.Errorf("expected id EVT-010, got %s", e.ID())
	}
	if e.Type() != "maintenance.scheduled" {
		t.Errorf("expected type maintenance.scheduled, got %s", e.Type())
	}
	if !e.OccurredAt().Equal(validTime) {
		t.Errorf("expected occurredAt %v, got %v", validTime, e.OccurredAt())
	}
}

func TestUnmarshalJSON_InvalidData(t *testing.T) {
	input := `{"id":"","type":"maintenance.scheduled","occurredAt":"2026-06-15T10:30:00Z","details":{"asset":"A-500"}}`

	var e event.BusinessEvent
	if err := json.Unmarshal([]byte(input), &e); err == nil {
		t.Fatal("expected validation error for empty id during unmarshal")
	}
}

func TestRoundTrip(t *testing.T) {
	original, _ := event.New("EVT-099", "shift.cancelled", validTime, json.RawMessage(`{"shiftId":"S-200","reason":"staff shortage"}`))

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var restored event.BusinessEvent
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !original.Equal(restored) {
		t.Error("round-tripped event should be equal to original")
	}
	if original.Type() != restored.Type() {
		t.Errorf("type mismatch: %s vs %s", original.Type(), restored.Type())
	}
	if !original.OccurredAt().Equal(restored.OccurredAt()) {
		t.Errorf("occurredAt mismatch: %v vs %v", original.OccurredAt(), restored.OccurredAt())
	}
}
