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

// Optimise takes validated BusinessEvents, Resources, and travel data, converts events
// into WorkItems, extracts constraints, runs the selected optimiser, and returns
// an OptimisedPlan with assignments.
func Optimise(events []event.BusinessEvent, resources []resource.Resource, travel []optimisation.TravelEntry, algorithm string) (plan.OptimisedPlan, error) {
	items, err := convertToWorkItems(events)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("conversion failed: %w", err)
	}

	capacities, err := extractCapacities(resources)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("capacity extraction failed: %w", err)
	}

	priorities := extractPriorities(items)

	var result plan.OptimisedPlan

	alg, err := optimisation.Get(algorithm)
	if err != nil {
		// Default to constructive for empty string.
		if algorithm == "" {
			alg, _ = optimisation.Get("constructive")
		} else {
			return plan.OptimisedPlan{}, fmt.Errorf("algorithm selection failed: %w", err)
		}
	}

	ctx := optimisation.NewContextWithTravel(items, capacities, priorities, travel)
	result, err = alg.Solve(ctx)
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

// extractCapacities reads capacity, availability, skills, shift times, and location from each resource's details JSON.
func extractCapacities(resources []resource.Resource) ([]optimisation.ResourceInput, error) {
	capacities := make([]optimisation.ResourceInput, 0, len(resources))

	for _, res := range resources {
		var details struct {
			Capacity   int      `json:"capacity"`
			Available  bool     `json:"available"`
			Skills     []string `json:"skills"`
			ShiftStart int      `json:"shiftStart"`
			ShiftEnd   int      `json:"shiftEnd"`
			Location   string   `json:"location"`
		}

		if err := json.Unmarshal(res.Details(), &details); err != nil {
			return nil, fmt.Errorf("failed to read resource details from %s: %w", res.ID(), err)
		}

		capacities = append(capacities, optimisation.ResourceInput{
			ResourceID: res.ID(),
			Capacity:   details.Capacity,
			Available:  details.Available,
			Skills:     details.Skills,
			ShiftStart: details.ShiftStart,
			ShiftEnd:   details.ShiftEnd,
			Location:   details.Location,
		})
	}

	return capacities, nil
}

// extractPriorities reads priority, required skill, duration, time windows, and location from each work item's details JSON.
func extractPriorities(items []workitem.WorkItem) []optimisation.WorkItemInput {
	priorities := make([]optimisation.WorkItemInput, 0, len(items))

	for _, item := range items {
		var details struct {
			Priority      int    `json:"priority"`
			RequiredSkill string `json:"requiredSkill"`
			Duration      int    `json:"duration"`
			EarliestStart int    `json:"earliestStart"`
			LatestFinish  int    `json:"latestFinish"`
			Location      string `json:"location"`
		}

		json.Unmarshal(item.Details(), &details)

		priorities = append(priorities, optimisation.WorkItemInput{
			WorkItemID:    item.ID(),
			Priority:      details.Priority,
			RequiredSkill: details.RequiredSkill,
			Duration:      details.Duration,
			EarliestStart: details.EarliestStart,
			LatestFinish:  details.LatestFinish,
			Location:      details.Location,
		})
	}

	return priorities
}
