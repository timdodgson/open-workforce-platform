package application_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/application"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
)

func makeEvent(id string, eventType string) event.BusinessEvent {
	e, _ := event.New(id, eventType, time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC), json.RawMessage(`{"key":"value"}`))
	return e
}

func makeResourceWithCapacity(id string, capacity int) resource.Resource {
	details := json.RawMessage(fmt.Sprintf(`{"name":"Test","capacity":%d}`, capacity))
	r, _ := resource.New(id, "person", details)
	return r
}

func TestOptimise_AssignsWithinCapacity(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-001", "patient.referred"),
		makeEvent("EVT-002", "maintenance.requested"),
	}
	resources := []resource.Resource{makeResourceWithCapacity("RES-001", 2)}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
	if result.Score() != 100 {
		t.Errorf("expected score 100, got %d", result.Score())
	}
}

func TestOptimise_RespectsCapacityLimit(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-001", "patient.referred"),
		makeEvent("EVT-002", "maintenance.requested"),
		makeEvent("EVT-003", "delivery.scheduled"),
	}
	resources := []resource.Resource{makeResourceWithCapacity("RES-001", 2)}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 1 {
		t.Errorf("expected 1 unassigned, got %d", result.UnassignedCount())
	}
	if result.Score() != 67 {
		t.Errorf("expected score 67, got %d", result.Score())
	}
}

func TestOptimise_SpillsToNextResource(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-001", "patient.referred"),
		makeEvent("EVT-002", "maintenance.requested"),
		makeEvent("EVT-003", "delivery.scheduled"),
	}
	resources := []resource.Resource{
		makeResourceWithCapacity("RES-001", 2),
		makeResourceWithCapacity("RES-002", 2),
	}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 3 {
		t.Errorf("expected 3 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 0 {
		t.Errorf("expected 0 unassigned, got %d", result.UnassignedCount())
	}
	if result.Score() != 100 {
		t.Errorf("expected score 100, got %d", result.Score())
	}

	// Verify first two go to RES-001, third to RES-002.
	assignments := result.Assignments()
	if assignments[0].ResourceID() != "RES-001" {
		t.Errorf("expected first assignment to RES-001, got %s", assignments[0].ResourceID())
	}
	if assignments[2].ResourceID() != "RES-002" {
		t.Errorf("expected third assignment to RES-002, got %s", assignments[2].ResourceID())
	}
}

func TestOptimise_EmptyEvents(t *testing.T) {
	resources := []resource.Resource{makeResourceWithCapacity("RES-001", 2)}
	_, err := application.Optimise([]event.BusinessEvent{}, resources)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}

func TestOptimise_EmptyResources(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "patient.referred")}
	_, err := application.Optimise(events, []resource.Resource{})
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}
