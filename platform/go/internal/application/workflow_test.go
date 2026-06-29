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

func makeEventWithPriority(id string, eventType string, priority int) event.BusinessEvent {
	details := json.RawMessage(fmt.Sprintf(`{"key":"value","priority":%d}`, priority))
	e, _ := event.New(id, eventType, time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC), details)
	return e
}

func makeResourceAvailable(id string, capacity int) resource.Resource {
	details := json.RawMessage(fmt.Sprintf(`{"name":"Test","capacity":%d,"available":true}`, capacity))
	r, _ := resource.New(id, "person", details)
	return r
}

func makeResourceUnavailable(id string, capacity int) resource.Resource {
	details := json.RawMessage(fmt.Sprintf(`{"name":"Test","capacity":%d,"available":false}`, capacity))
	r, _ := resource.New(id, "person", details)
	return r
}

func makeResourceNoAvailability(id string, capacity int) resource.Resource {
	details := json.RawMessage(fmt.Sprintf(`{"name":"Test","capacity":%d}`, capacity))
	r, _ := resource.New(id, "person", details)
	return r
}

func TestOptimise_AssignsToAvailableResource(t *testing.T) {
	events := []event.BusinessEvent{
		makeEventWithPriority("EVT-001", "task.a", 50),
		makeEventWithPriority("EVT-002", "task.b", 50),
	}
	resources := []resource.Resource{makeResourceAvailable("RES-001", 2)}

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

func TestOptimise_SkipsUnavailableResource(t *testing.T) {
	events := []event.BusinessEvent{
		makeEventWithPriority("EVT-001", "task.a", 50),
	}
	resources := []resource.Resource{
		makeResourceUnavailable("RES-UNAVAIL", 5),
		makeResourceAvailable("RES-AVAIL", 2),
	}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assigned := result.Assignments()[0]
	if assigned.ResourceID() != "RES-AVAIL" {
		t.Errorf("expected assignment to RES-AVAIL, got %s", assigned.ResourceID())
	}
}

func TestOptimise_MissingAvailabilityDefaultsToUnavailable(t *testing.T) {
	events := []event.BusinessEvent{
		makeEventWithPriority("EVT-001", "task.a", 50),
	}
	resources := []resource.Resource{
		makeResourceNoAvailability("RES-NO-FIELD", 5),
		makeResourceAvailable("RES-AVAIL", 2),
	}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assigned := result.Assignments()[0]
	if assigned.ResourceID() != "RES-AVAIL" {
		t.Errorf("expected assignment to RES-AVAIL (missing available defaults to false), got %s", assigned.ResourceID())
	}
}

func TestOptimise_AllUnavailableResultsInUnassigned(t *testing.T) {
	events := []event.BusinessEvent{
		makeEventWithPriority("EVT-001", "task.a", 50),
	}
	resources := []resource.Resource{
		makeResourceUnavailable("RES-001", 5),
	}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments, got %d", result.Size())
	}
	if result.UnassignedCount() != 1 {
		t.Errorf("expected 1 unassigned, got %d", result.UnassignedCount())
	}
}

func TestOptimise_HigherPriorityStillAssignedFirst(t *testing.T) {
	events := []event.BusinessEvent{
		makeEventWithPriority("EVT-LOW", "task.low", 10),
		makeEventWithPriority("EVT-HIGH", "task.high", 100),
	}
	resources := []resource.Resource{makeResourceAvailable("RES-001", 1)}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assigned := result.Assignments()[0]
	if assigned.WorkItemID() != "WI-EVT-HIGH" {
		t.Errorf("expected WI-EVT-HIGH assigned first, got %s", assigned.WorkItemID())
	}
}

func TestOptimise_CapacityStillRespected(t *testing.T) {
	events := []event.BusinessEvent{
		makeEventWithPriority("EVT-001", "task.a", 50),
		makeEventWithPriority("EVT-002", "task.b", 50),
		makeEventWithPriority("EVT-003", "task.c", 50),
	}
	resources := []resource.Resource{makeResourceAvailable("RES-001", 2)}

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
}

func TestOptimise_EmptyEvents(t *testing.T) {
	resources := []resource.Resource{makeResourceAvailable("RES-001", 2)}
	_, err := application.Optimise([]event.BusinessEvent{}, resources)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}

func TestOptimise_EmptyResources(t *testing.T) {
	events := []event.BusinessEvent{makeEventWithPriority("EVT-001", "task.a", 50)}
	_, err := application.Optimise(events, []resource.Resource{})
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}
