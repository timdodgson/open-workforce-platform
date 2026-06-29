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
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/loader"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// Optimise takes validated BusinessEvents, Resources, and travel data, converts events
// into WorkItems, extracts constraints, runs the selected optimiser, and returns
// an OptimisedPlan with assignments.
func Optimise(events []event.BusinessEvent, resources []resource.Resource, travel []optimisation.TravelEntry, algorithm string, profile ...optimisation.AlgorithmProfile) (plan.OptimisedPlan, error) {
	return OptimiseWithNRP(events, resources, travel, nil, algorithm, profile...)
}

// OptimiseWithNRP extends Optimise with optional NRP context data.
func OptimiseWithNRP(events []event.BusinessEvent, resources []resource.Resource, travel []optimisation.TravelEntry, nrpCtx *loader.NRPContext, algorithm string, profile ...optimisation.AlgorithmProfile) (plan.OptimisedPlan, error) {
	items, err := convertToWorkItems(events)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("conversion failed: %w", err)
	}

	capacities, err := extractCapacities(resources)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("capacity extraction failed: %w", err)
	}

	priorities := extractPriorities(items)

	alg, err := optimisation.Get(algorithm)
	if err != nil {
		if algorithm == "" {
			alg, _ = optimisation.Get("constructive")
		} else {
			return plan.OptimisedPlan{}, fmt.Errorf("algorithm selection failed: %w", err)
		}
	}

	ctx := optimisation.NewContextWithTravel(items, capacities, priorities, travel)
	if len(profile) > 0 {
		ctx = ctx.WithProfile(profile[0])
	}

	// Apply NRP context if present.
	if nrpCtx != nil {
		ctx = applyNRPContext(ctx, nrpCtx)
	}

	result, err := alg.Solve(ctx)
	if err != nil {
		return plan.OptimisedPlan{}, fmt.Errorf("optimisation failed: %w", err)
	}

	return result, nil
}

// applyNRPContext maps loader NRP context data to optimisation context types.
func applyNRPContext(ctx optimisation.OptimisationContext, nrpCtx *loader.NRPContext) optimisation.OptimisationContext {
	if len(nrpCtx.Contracts) > 0 {
		contracts := make([]optimisation.Contract, len(nrpCtx.Contracts))
		for i, c := range nrpCtx.Contracts {
			contracts[i] = optimisation.Contract{
				ID:                        c.ID,
				MinAssignments:            c.MinAssignments,
				MaxAssignments:            c.MaxAssignments,
				MinConsecutiveWorkingDays: c.MinConsecutiveWorkingDays,
				MaxConsecutiveWorkingDays: c.MaxConsecutiveWorkingDays,
				MinConsecutiveDaysOff:     c.MinConsecutiveDaysOff,
				MaxConsecutiveDaysOff:     c.MaxConsecutiveDaysOff,
				MaxWorkingWeekends:        c.MaxWorkingWeekends,
				CompleteWeekend:           c.CompleteWeekend,
			}
		}
		ctx = ctx.WithContracts(contracts)
	}

	if len(nrpCtx.ShiftTypes) > 0 {
		shiftTypes := make([]optimisation.ShiftTypeInfo, len(nrpCtx.ShiftTypes))
		for i, s := range nrpCtx.ShiftTypes {
			shiftTypes[i] = optimisation.ShiftTypeInfo{
				ID:                        s.ID,
				StartMinute:               s.StartMinute,
				EndMinute:                 s.EndMinute,
				MinConsecutiveAssignments: s.MinConsecutiveAssignments,
				MaxConsecutiveAssignments: s.MaxConsecutiveAssignments,
			}
		}
		ctx = ctx.WithShiftTypes(shiftTypes)
	}

	if len(nrpCtx.ForbiddenSuccessions) > 0 {
		successions := make([]optimisation.ForbiddenSuccession, len(nrpCtx.ForbiddenSuccessions))
		for i, f := range nrpCtx.ForbiddenSuccessions {
			successions[i] = optimisation.ForbiddenSuccession{
				PrecedingShift: f.PrecedingShift,
				SuccessorShift: f.SuccessorShift,
			}
		}
		ctx = ctx.WithForbiddenSuccessions(successions)
	}

	if len(nrpCtx.Requests) > 0 {
		requests := make([]optimisation.Request, len(nrpCtx.Requests))
		for i, r := range nrpCtx.Requests {
			requests[i] = optimisation.Request{
				ResourceID: r.NurseID,
				Day:        r.Day,
				ShiftType:  r.ShiftType,
				Type:       r.Type,
				Weight:     r.Weight,
			}
		}
		ctx = ctx.WithRequests(requests)
	}

	if len(nrpCtx.CoverageRequirements) > 0 {
		reqs := make([]optimisation.CoverageRequirement, len(nrpCtx.CoverageRequirements))
		for i, cr := range nrpCtx.CoverageRequirements {
			reqs[i] = optimisation.CoverageRequirement{
				Day:       cr.Day,
				ShiftType: cr.ShiftType,
				Skill:     cr.Skill,
				Minimum:   cr.Minimum,
				Optimal:   cr.Optimal,
			}
		}
		ctx = ctx.WithCoverageRequirements(reqs)
	}

	// If NRP context is present, use NRP weights by default.
	ctx = ctx.WithWeights(optimisation.NRPWeights())

	return ctx
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

// extractCapacities reads capacity, availability, skills, shift times, location and contractId from each resource's details JSON.
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
			ContractID string   `json:"contractId"`
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
			ContractID: details.ContractID,
		})
	}

	return capacities, nil
}

// extractPriorities reads priority, required skill, duration, time windows, location, shift type, mandatory flag, demand group, and preferred resource from each work item's details JSON.
func extractPriorities(items []workitem.WorkItem) []optimisation.WorkItemInput {
	priorities := make([]optimisation.WorkItemInput, 0, len(items))

	for _, item := range items {
		var details struct {
			Priority          int    `json:"priority"`
			RequiredSkill     string `json:"requiredSkill"`
			Duration          int    `json:"duration"`
			EarliestStart     int    `json:"earliestStart"`
			LatestFinish      int    `json:"latestFinish"`
			Location          string `json:"location"`
			PreferredResource string `json:"preferredResource"`
			Day               int    `json:"day"`
			ShiftType         string `json:"shiftType"`
			Mandatory         bool   `json:"mandatory"`
			DemandGroup       string `json:"demandGroup"`
		}

		json.Unmarshal(item.Details(), &details)

		priorities = append(priorities, optimisation.WorkItemInput{
			WorkItemID:        item.ID(),
			Priority:          details.Priority,
			RequiredSkill:     details.RequiredSkill,
			Duration:          details.Duration,
			EarliestStart:     details.EarliestStart,
			LatestFinish:      details.LatestFinish,
			Location:          details.Location,
			PreferredResource: details.PreferredResource,
			Day:               details.Day,
			ShiftType:         details.ShiftType,
			Mandatory:         details.Mandatory,
			DemandGroup:       details.DemandGroup,
		})
	}

	return priorities
}
