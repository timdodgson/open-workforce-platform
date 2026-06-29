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

func makeEvent(id, eventType string, priority int, requiredSkill string) event.BusinessEvent {
	var details string
	if requiredSkill != "" {
		details = fmt.Sprintf(`{"priority":%d,"requiredSkill":"%s"}`, priority, requiredSkill)
	} else {
		details = fmt.Sprintf(`{"priority":%d}`, priority)
	}
	e, _ := event.New(id, eventType, time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC), json.RawMessage(details))
	return e
}

func makeResource(id string, capacity int, available bool, skills []string) resource.Resource {
	skillsJSON, _ := json.Marshal(skills)
	details := fmt.Sprintf(`{"capacity":%d,"available":%t,"skills":%s}`, capacity, available, skillsJSON)
	r, _ := resource.New(id, "person", json.RawMessage(details))
	return r
}

func TestOptimise_SkillMatchAssigns(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "clinical")}
	resources := []resource.Resource{makeResource("RES-001", 2, true, []string{"clinical", "assessment"})}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 1 {
		t.Errorf("expected 1 assignment, got %d", result.Size())
	}
}

func TestOptimise_SkillMismatchUnassigned(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "clinical")}
	resources := []resource.Resource{makeResource("RES-001", 5, true, []string{"electrical"})}

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

func TestOptimise_NoRequiredSkillAssignsToAny(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "")}
	resources := []resource.Resource{makeResource("RES-001", 2, true, []string{"clinical"})}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 1 {
		t.Errorf("expected 1 assignment, got %d", result.Size())
	}
}

func TestOptimise_SkillMatchSkipsToCorrectResource(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "clinical")}
	resources := []resource.Resource{
		makeResource("RES-WRONG", 5, true, []string{"electrical"}),
		makeResource("RES-RIGHT", 5, true, []string{"clinical"}),
	}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Assignments()[0].ResourceID() != "RES-RIGHT" {
		t.Errorf("expected RES-RIGHT, got %s", result.Assignments()[0].ResourceID())
	}
}

func TestOptimise_PriorityStillRespectedWithSkills(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-LOW", "task", 10, "clinical"),
		makeEvent("EVT-HIGH", "task", 100, "clinical"),
	}
	resources := []resource.Resource{makeResource("RES-001", 1, true, []string{"clinical"})}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Assignments()[0].WorkItemID() != "WI-EVT-HIGH" {
		t.Errorf("expected WI-EVT-HIGH first, got %s", result.Assignments()[0].WorkItemID())
	}
}

func TestOptimise_AvailabilityStillRespectedWithSkills(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "clinical")}
	resources := []resource.Resource{
		makeResource("RES-UNAVAIL", 5, false, []string{"clinical"}),
		makeResource("RES-AVAIL", 5, true, []string{"clinical"}),
	}

	result, err := application.Optimise(events, resources)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Assignments()[0].ResourceID() != "RES-AVAIL" {
		t.Errorf("expected RES-AVAIL, got %s", result.Assignments()[0].ResourceID())
	}
}

func TestOptimise_CapacityStillRespectedWithSkills(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-001", "task", 50, "clinical"),
		makeEvent("EVT-002", "task", 50, "clinical"),
		makeEvent("EVT-003", "task", 50, "clinical"),
	}
	resources := []resource.Resource{makeResource("RES-001", 2, true, []string{"clinical"})}

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
	resources := []resource.Resource{makeResource("RES-001", 2, true, []string{"clinical"})}
	_, err := application.Optimise([]event.BusinessEvent{}, resources)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}

func TestOptimise_EmptyResources(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "clinical")}
	_, err := application.Optimise(events, []resource.Resource{})
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}
