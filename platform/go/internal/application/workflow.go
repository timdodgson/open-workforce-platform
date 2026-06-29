// Package application provides application-level orchestration.
//
// It coordinates the workflow between domain objects and the optimisation layer.
// The application layer is responsible for interpreting business knowledge
// from domain objects and providing structured input to the optimiser.
package application

import (
	"encoding/json"
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/resource"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// Optimise takes validated BusinessEvents and Resources, converts events into
// WorkItems, extracts capacity from resources, runs the optimiser, and returns
// an OptimisedPlan with assignments.
func Optimise(events []event.BusinessEvent, resources []resource.Resource) (plan.OptimisedPlan, error) {
	items, err := convertToWorkItems(events)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("conversion failed: %w", err)
	}

	capacities, err := extractCapacities(resources)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("capacity extraction failed: %w", err)
	}

	result, err := optimisation.Solve(items, capacities)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("optimisation failed: %w", err)
	}

	return result, nil
}

// convertToWorkItems creates a WorkItem for each BusinessEvent.
//
// For the walking skeleton, this is a simple 1:1 mapping.
func convertToWorkItems(events []event.BusinessEvent) ([]workitem.WorkItem, error) {
	items := make([]workitem.WorkItem, 0, len(events))

	for _, evt := range events {
		id := "WI-" + evt.ID()
		details := evt.Details()

		item, err := workitem.New(id, evt.Type(), json.RawMessage(details))
		if err != nil {
			return nil, fmt.Errorf("failed to create work item from event %s: %w", evt.ID(), err)
		}

		items = append(items, item)
	}

	return items, nil
}

// extractCapacities reads capacity from each resource's details JSON.
//
// If a resource has no capacity field, it defaults to 0 (fail safe).
// This is application-layer responsibility — the resource domain object
// remains generic per the architecture.
func extractCapacities(resources []resource.Resource) ([]optimisation.ResourceCapacity, error) {
	capacities := make([]optimisation.ResourceCapacity, 0, len(resources))

	for _, res := range resources {
		var details struct {
			Capacity int `json:"capacity"`
		}

		if err := json.Unmarshal(res.Details(), &details); err != nil {
			return nil, fmt.Errorf("failed to read capacity from resource %s: %w", res.ID(), err)
		}

		capacities = append(capacities, optimisation.ResourceCapacity{
			ResourceID: res.ID(),
			Capacity:   details.Capacity,
		})
	}

	return capacities, nil
}
