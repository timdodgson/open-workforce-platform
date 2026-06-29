package application_test

import (
	"encoding/json"
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

func makeResource(id string) resource.Resource {
	r, _ := resource.New(id, "person", json.RawMessage(`{"name":"Test"}`))
	return r
}

func TestOptimise_ProducesPlan(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-001", "patient.referred"),
		makeEvent("EVT-002", "maintenance.requested"),
	}
	resources := []resource.Resource{makeResource("RES-001")}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 2 {
		t.Errorf("expected 2 assignments, got %d", result.Size())
	}
}

func TestOptimise_AssignmentsReferenceCorrectWorkItems(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-001", "patient.referred"),
	}
	resources := []resource.Resource{makeResource("RES-001")}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	assignments := result.Assignments()
	if assignments[0].WorkItemID() != "WI-EVT-001" {
		t.Errorf("expected work item id WI-EVT-001, got %s", assignments[0].WorkItemID())
	}
	if assignments[0].ResourceID() != "RES-001" {
		t.Errorf("expected resource id RES-001, got %s", assignments[0].ResourceID())
	}
}

func TestOptimise_EmptyEvents(t *testing.T) {
	resources := []resource.Resource{makeResource("RES-001")}
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
