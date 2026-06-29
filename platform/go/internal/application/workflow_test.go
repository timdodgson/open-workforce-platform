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

func makeEventNoPriority(id string, eventType string) event.BusinessEvent {
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
		makeEventWithPriority("EVT-001", "patient.referred", 50),
		makeEventWithPriority("EVT-002", "maintenance.requested", 50),
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

func TestOptimise_HigherPriorityAssignedFirst(t *testing.T) {
	events := []event.BusinessEvent{
		makeEventWithPriority("EVT-LOW", "task.low", 10),
		makeEventWithPriority("EVT-HIGH", "task.high", 100),
	}
	resources := []resource.Resource{makeResourceWithCapacity("RES-001", 1)}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assigned := result.Assignments()[0]
	if assigned.WorkItemID() != "WI-EVT-HIGH" {
		t.Errorf("expected WI-EVT-HIGH assigned first, got %s", assigned.WorkItemID())
	}

	unassigned := result.Unassigned()
	if unassigned[0] != "WI-EVT-LOW" {
		t.Errorf("expected WI-EVT-LOW unassigned, got %s", unassigned[0])
	}
}

func TestOptimise_MissingPriorityDefaultsToZero(t *testing.T) {
	events := []event.BusinessEvent{
		makeEventNoPriority("EVT-NO-PRIO", "task.basic"),
		makeEventWithPriority("EVT-WITH-PRIO", "task.important", 50),
	}
	resources := []resource.Resource{makeResourceWithCapacity("RES-001", 1)}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assigned := result.Assignments()[0]
	if assigned.WorkItemID() != "WI-EVT-WITH-PRIO" {
		t.Errorf("expected WI-EVT-WITH-PRIO assigned (priority 50 > 0), got %s", assigned.WorkItemID())
	}
}

func TestOptimise_SpillsToNextResource(t *testing.T) {
	events := []event.BusinessEvent{
		makeEventWithPriority("EVT-001", "task.a", 50),
		makeEventWithPriority("EVT-002", "task.b", 50),
		makeEventWithPriority("EVT-003", "task.c", 50),
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
	if result.Score() != 100 {
		t.Errorf("expected score 100, got %d", result.Score())
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
	events := []event.BusinessEvent{makeEventWithPriority("EVT-001", "task.a", 50)}
	_, err := application.Optimise(events, []resource.Resource{})
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}
