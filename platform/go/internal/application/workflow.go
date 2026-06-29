// Package application provides application-level orchestration.
//
// It coordinates the workflow between domain objects and the optimisation layer.
// Business logic for converting events into work items lives here for the
// walking skeleton. As the platform matures, this may evolve into dedicated
// handlers per event type.
package application

import (
	"encoding/json"
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/event"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// Optimise takes validated BusinessEvents, converts them into WorkItems,
// runs the optimiser, and returns an OptimisedPlan.
func Optimise(events []event.BusinessEvent) (plan.OptimisedPlan, error) {
	items, err := convertToWorkItems(events)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("conversion failed: %w", err)
	}

	result, err := optimisation.Solve(items)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("optimisation failed: %w", err)
	}

	return result, nil
}

// convertToWorkItems creates a WorkItem for each BusinessEvent.
//
// For the walking skeleton, this is a simple 1:1 mapping.
// The work item ID is derived from the event ID, the type matches the event type,
// and the event details are passed through as planning details.
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
