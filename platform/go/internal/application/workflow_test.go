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

// --- Default algorithm (constructive) ---

func TestOptimise_SkillMatchAssigns(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "clinical")}
	resources := []resource.Resource{makeResource("RES-001", 2, true, []string{"clinical", "assessment"})}

	result, err := application.Optimise(events, resources, "constructive")
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

	result, err := application.Optimise(events, resources, "constructive")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 0 {
		t.Errorf("expected 0 assignments, got %d", result.Size())
	}
}

func TestOptimise_NoRequiredSkillAssignsToAny(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "")}
	resources := []resource.Resource{makeResource("RES-001", 2, true, []string{"clinical"})}

	result, err := application.Optimise(events, resources, "constructive")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 1 {
		t.Errorf("expected 1 assignment, got %d", result.Size())
	}
}

func TestOptimise_PriorityStillRespected(t *testing.T) {
	events := []event.BusinessEvent{
		makeEvent("EVT-LOW", "task", 10, "clinical"),
		makeEvent("EVT-HIGH", "task", 100, "clinical"),
	}
	resources := []resource.Resource{makeResource("RES-001", 1, true, []string{"clinical"})}

	result, err := application.Optimise(events, resources, "constructive")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Assignments()[0].WorkItemID() != "WI-EVT-HIGH" {
		t.Errorf("expected WI-EVT-HIGH first, got %s", result.Assignments()[0].WorkItemID())
	}
}

func TestOptimise_EmptyEvents(t *testing.T) {
	resources := []resource.Resource{makeResource("RES-001", 2, true, []string{"clinical"})}
	_, err := application.Optimise([]event.BusinessEvent{}, resources, "constructive")
	if err == nil {
		t.Fatal("expected error for empty events")
	}
}

func TestOptimise_EmptyResources(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "clinical")}
	_, err := application.Optimise(events, []resource.Resource{}, "constructive")
	if err == nil {
		t.Fatal("expected error for empty resources")
	}
}

// --- Algorithm selection ---

func TestOptimise_DefaultAlgorithmIsConstructive(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "")}
	resources := []resource.Resource{makeResource("RES-001", 2, true, nil)}

	// Empty string defaults to constructive.
	result, err := application.Optimise(events, resources, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 1 {
		t.Errorf("expected 1 assignment, got %d", result.Size())
	}
}

func TestOptimise_HillClimbingAlgorithm(t *testing.T) {
	events := []event.BusinessEvent{makeEvent("EVT-001", "task", 50, "")}
	resources := []resource.Resource{makeResource("RES-001", 2, true, nil)}

	result, err := application.Optimise(events, resources, "hill-climbing")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Size() != 1 {
		t.Errorf("expected 1 assignment, got %d", result.Size())
	}
}
