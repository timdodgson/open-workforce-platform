package inrc2

import (
	"encoding/json"
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/plan"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// SolveWeek converts an INRC-II problem into OWP format, runs the specified
// algorithm, and returns an official INRC-II solution.
func SolveWeek(sc Scenario, wd WeekData, hist History, algorithm string, profile optimisation.AlgorithmProfile) (Solution, plan.OptimisedPlan, error) {
	// Build work items and resources directly as optimisation inputs.
	workItems := buildWorkItemInputs(wd)
	resources := buildResourceInputs(sc)
	domainItems := buildDomainWorkItems(workItems)

	// Build optimisation context with INRC-II constraints.
	ctx := optimisation.NewContextWithTravel(domainItems, resources, workItems, nil)
	ctx = ctx.WithProfile(profile)
	ctx = applyINRC2Context(ctx, sc, wd, hist)

	// Run algorithm.
	alg, err := optimisation.Get(algorithm)
	if err != nil {
		return Solution{}, plan.OptimisedPlan{}, fmt.Errorf("algorithm: %w", err)
	}

	result, err := alg.Solve(ctx)
	if err != nil {
		return Solution{}, plan.OptimisedPlan{}, fmt.Errorf("solve: %w", err)
	}

	// Convert OWP result back to INRC-II solution format.
	sol := convertToINRC2Solution(sc, wd, result, hist.Week)

	return sol, result, nil
}

// buildWorkItemInputs creates WorkItemInput for each coverage demand unit.
func buildWorkItemInputs(wd WeekData) []optimisation.WorkItemInput {
	var result []optimisation.WorkItemInput
	idx := 0

	for _, req := range wd.Requirements {
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			dayReq := req.RequirementForDay(dayIdx)
			total := dayReq.Optimal
			if total < dayReq.Minimum {
				total = dayReq.Minimum
			}
			for i := 0; i < total; i++ {
				idx++
				mandatory := i < dayReq.Minimum
				priority := 100
				if !mandatory {
					priority = 50
				}
				result = append(result, optimisation.WorkItemInput{
					WorkItemID:    fmt.Sprintf("WI-EVT-%04d", idx),
					Priority:      priority,
					RequiredSkill: req.Skill,
					Duration:      1,
					Day:           dayIdx + 1, // 1-indexed for same-day constraint
					ShiftType:     req.ShiftType,
					Mandatory:     mandatory,
				})
			}
		}
	}

	return result
}

// buildResourceInputs creates ResourceInput from scenario nurses.
func buildResourceInputs(sc Scenario) []optimisation.ResourceInput {
	var result []optimisation.ResourceInput
	for _, nurse := range sc.Nurses {
		result = append(result, optimisation.ResourceInput{
			ResourceID: nurse.ID,
			Capacity:   1440,
			Available:  true,
			Skills:     nurse.Skills,
			ShiftStart: 0,
			ShiftEnd:   1440,
			ContractID: nurse.Contract,
		})
	}
	return result
}

// buildDomainWorkItems creates minimal domain work item objects for the context.
func buildDomainWorkItems(inputs []optimisation.WorkItemInput) []workitem.WorkItem {
	items := make([]workitem.WorkItem, 0, len(inputs))
	for _, wi := range inputs {
		item, _ := workitem.New(wi.WorkItemID, "shift.demand", json.RawMessage(`{}`))
		items = append(items, item)
	}
	return items
}

// applyINRC2Context adds INRC-II constraints to the optimisation context.
func applyINRC2Context(ctx optimisation.OptimisationContext, sc Scenario, wd WeekData, hist History) optimisation.OptimisationContext {
	// Contracts.
	var contracts []optimisation.Contract
	for _, c := range sc.Contracts {
		contracts = append(contracts, optimisation.Contract{
			ID:                        c.ID,
			MinAssignments:            c.MinimumNumberOfAssignments,
			MaxAssignments:            c.MaximumNumberOfAssignments,
			MinConsecutiveWorkingDays: c.MinimumNumberOfConsecutiveWorkingDays,
			MaxConsecutiveWorkingDays: c.MaximumNumberOfConsecutiveWorkingDays,
			MinConsecutiveDaysOff:     c.MinimumNumberOfConsecutiveDaysOff,
			MaxConsecutiveDaysOff:     c.MaximumNumberOfConsecutiveDaysOff,
			MaxWorkingWeekends:        c.MaximumNumberOfWorkingWeekends,
			CompleteWeekend:           c.CompleteWeekends == 1,
		})
	}
	ctx = ctx.WithContracts(contracts)

	// Shift types.
	var shiftTypes []optimisation.ShiftTypeInfo
	for _, st := range sc.ShiftTypes {
		shiftTypes = append(shiftTypes, optimisation.ShiftTypeInfo{
			ID:                        st.ID,
			MinConsecutiveAssignments: st.MinimumNumberOfConsecutiveAssignments,
			MaxConsecutiveAssignments: st.MaximumNumberOfConsecutiveAssignments,
		})
	}
	ctx = ctx.WithShiftTypes(shiftTypes)

	// Forbidden successions.
	var successions []optimisation.ForbiddenSuccession
	for _, fs := range sc.ForbiddenShiftTypeSuccessions {
		for _, succ := range fs.SucceedingShiftTypes {
			successions = append(successions, optimisation.ForbiddenSuccession{
				PrecedingShift: fs.PrecedingShiftType,
				SuccessorShift: succ,
			})
		}
	}
	ctx = ctx.WithForbiddenSuccessions(successions)

	// Requests.
	var requests []optimisation.Request
	for _, req := range wd.ShiftOffRequests {
		dayIdx := DayIndex(req.Day)
		requests = append(requests, optimisation.Request{
			ResourceID: req.Nurse,
			Day:        dayIdx + 1,
			ShiftType:  req.ShiftType,
			Type:       "shiftOff",
			Weight:     10,
		})
	}
	ctx = ctx.WithRequests(requests)

	// Use NRP weights.
	ctx = ctx.WithWeights(optimisation.NRPWeights())

	return ctx
}

// convertToINRC2Solution converts an OWP optimised plan into INRC-II solution format.
func convertToINRC2Solution(sc Scenario, wd WeekData, result plan.OptimisedPlan, week int) Solution {
	// The OWP plan assigns work items to resources.
	// We need to map back: work item ID -> (day, shiftType, skill), resource -> nurse.
	itemLookup := buildItemLookup(wd)

	var assignments []Assignment
	for _, a := range result.Assignments() {
		info, ok := itemLookup[a.WorkItemID()]
		if !ok {
			continue
		}
		assignments = append(assignments, Assignment{
			Nurse:     a.ResourceID(),
			Day:       DayName(info.dayIdx),
			ShiftType: info.shiftType,
			Skill:     info.skill,
		})
	}

	return Solution{
		Scenario:    sc.ID,
		Week:        week,
		Assignments: assignments,
	}
}

type itemInfo struct {
	dayIdx    int
	shiftType string
	skill     string
}

// buildItemLookup creates a mapping from work item ID to its INRC-II properties.
func buildItemLookup(wd WeekData) map[string]itemInfo {
	lookup := make(map[string]itemInfo)
	idx := 0

	for _, req := range wd.Requirements {
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			dayReq := req.RequirementForDay(dayIdx)
			total := dayReq.Optimal
			if total < dayReq.Minimum {
				total = dayReq.Minimum
			}
			for i := 0; i < total; i++ {
				idx++
				id := fmt.Sprintf("WI-EVT-%04d", idx)
				lookup[id] = itemInfo{
					dayIdx:    dayIdx,
					shiftType: req.ShiftType,
					skill:     req.Skill,
				}
			}
		}
	}

	return lookup
}
