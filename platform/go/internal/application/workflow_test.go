package application_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
)

func makeEvent(id string, eventType string) event.BusinessEvent {
	e, _ := event.New(id, eventType, time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC), json.RawMessage(`{"key":"value"}`))
	return e
}

func TestOptimise_ProducesPlan(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-001", "patient.referred"),
		makeEvent("EVT-002", "maintenance.requested"),
	}

	result, err := application.Optimise(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 items in plan, got %d", result.Size())
	}
}

func TestOptimise_WorkItemIDsDerivedFromEvents(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-001", "patient.referred"),
	}

	result, err := application.Optimise(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	items := result.Items()
	if items[0].ID() != "WI-EVT-001" {
		t.Errorf("expected work item id WI-EVT-001, got %s", items[0].ID())
	}
}

func TestOptimise_WorkItemTypesMatchEventTypes(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-001", "delivery.scheduled"),
	}

	result, err := application.Optimise(events)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	items := result.Items()
	if items[0].Type() != "delivery.scheduled" {
		t.Errorf("expected work item type delivery.scheduled, got %s", items[0].Type())
	}
}

func TestOptimise_EmptyEvents(t *testing.T) {
	_, err := application.Optimise([]event.BusinessEvent{})
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}
