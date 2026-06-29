package optimisation

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/workitem"
)

// --- Hard Constraint Tests ---

func TestForbiddenSuccession_Rejected(t *testing.T) {
	// Late on day 1, Early on day 2 = forbidden.
	items := makeWorkItems(2)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, Skills: []string{"general"}, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Priority: 100, RequiredSkill: "general", Duration: 480, EarliestStart: 840, LatestFinish: 1320, Day: 1, ShiftType: "late"},
		{WorkItemID: items[1].ID(), Priority: 100, RequiredSkill: "general", Duration: 480, EarliestStart: 360, LatestFinish: 840, Day: 2, ShiftType: "early"},
	}

	a1, _ := assignment.New("R1", items[0].ID())
	a2, _ := assignment.New("R1", items[1].ID())
	assignments := []assignment.Assignment{a1, a2}

	ctx := NewContextWithTravel(items, capacities, priorities, nil)
	ctx = ctx.WithForbiddenSuccessions([]ForbiddenSuccession{
		{PrecedingShift: "late", SuccessorShift: "early"},
	})

	if scheduleFeasible(assignments, capacities, priorities, ctx) {
		t.Error("expected schedule to be infeasible due to forbidden succession late->early")
	}
}

func TestForbiddenSuccession_AllowedWhenNotForbidden(t *testing.T) {
	// Early on day 1, Late on day 2 = allowed.
	items := makeWorkItems(2)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, Skills: []string{"general"}, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Priority: 100, RequiredSkill: "general", Duration: 480, EarliestStart: 360, LatestFinish: 840, Day: 1, ShiftType: "early"},
		{WorkItemID: items[1].ID(), Priority: 100, RequiredSkill: "general", Duration: 480, EarliestStart: 840, LatestFinish: 1320, Day: 2, ShiftType: "late"},
	}

	a1, _ := assignment.New("R1", items[0].ID())
	a2, _ := assignment.New("R1", items[1].ID())
	assignments := []assignment.Assignment{a1, a2}

	ctx := NewContextWithTravel(items, capacities, priorities, nil)
	ctx = ctx.WithForbiddenSuccessions([]ForbiddenSuccession{
		{PrecedingShift: "late", SuccessorShift: "early"},
	})

	if !scheduleFeasible(assignments, capacities, priorities, ctx) {
		t.Error("expected schedule to be feasible; early->late is not forbidden")
	}
}

func TestSameDayAssignment_Rejected(t *testing.T) {
	items := makeWorkItems(2)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, Skills: []string{"general"}, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Priority: 100, Duration: 480, EarliestStart: 360, LatestFinish: 840, Day: 1, ShiftType: "early"},
		{WorkItemID: items[1].ID(), Priority: 100, Duration: 480, EarliestStart: 840, LatestFinish: 1320, Day: 1, ShiftType: "late"},
	}

	a1, _ := assignment.New("R1", items[0].ID())
	a2, _ := assignment.New("R1", items[1].ID())
	assignments := []assignment.Assignment{a1, a2}

	ctx := NewContext(items, capacities, priorities)

	if scheduleFeasible(assignments, capacities, priorities, ctx) {
		t.Error("expected infeasible: two items on same day for same resource")
	}
}

func TestSameDayAssignment_DifferentNursesAllowed(t *testing.T) {
	items := makeWorkItems(2)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, Skills: []string{"general"}, ShiftStart: 360, ShiftEnd: 1320},
		{ResourceID: "R2", Capacity: 960, Available: true, Skills: []string{"general"}, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Priority: 100, Duration: 480, EarliestStart: 360, LatestFinish: 840, Day: 1, ShiftType: "early"},
		{WorkItemID: items[1].ID(), Priority: 100, Duration: 480, EarliestStart: 360, LatestFinish: 840, Day: 1, ShiftType: "early"},
	}

	a1, _ := assignment.New("R1", items[0].ID())
	a2, _ := assignment.New("R2", items[1].ID())
	assignments := []assignment.Assignment{a1, a2}

	ctx := NewContext(items, capacities, priorities)

	if !scheduleFeasible(assignments, capacities, priorities, ctx) {
		t.Error("expected feasible: different nurses can work same day")
	}
}

func TestSameDayAssignment_SameNurseDifferentDaysAllowed(t *testing.T) {
	items := makeWorkItems(2)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, Skills: []string{"general"}, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Priority: 100, Duration: 480, EarliestStart: 360, LatestFinish: 840, Day: 1, ShiftType: "early"},
		{WorkItemID: items[1].ID(), Priority: 100, Duration: 480, EarliestStart: 360, LatestFinish: 840, Day: 2, ShiftType: "early"},
	}

	a1, _ := assignment.New("R1", items[0].ID())
	a2, _ := assignment.New("R1", items[1].ID())
	assignments := []assignment.Assignment{a1, a2}

	ctx := NewContext(items, capacities, priorities)

	if !scheduleFeasible(assignments, capacities, priorities, ctx) {
		t.Error("expected feasible: same nurse on different days is allowed")
	}
}

// --- Soft Constraint Tests ---

func TestRawTotalAssignments_UnderMin(t *testing.T) {
	items := makeWorkItems(1)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ContractID: "c1", ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 1},
	}

	a, _ := assignment.New("R1", items[0].ID())
	assignments := []assignment.Assignment{a}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithContracts([]Contract{
		{ID: "c1", MinAssignments: 5, MaxAssignments: 10},
	})

	violations := rawTotalAssignments(assignments, ctx)
	// 1 assignment, min is 5, so 4 violations.
	if violations != 4 {
		t.Errorf("expected 4 violations, got %d", violations)
	}
}

func TestRawTotalAssignments_OverMax(t *testing.T) {
	items := makeWorkItems(3)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ContractID: "c1", ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 1},
		{WorkItemID: items[1].ID(), Day: 2},
		{WorkItemID: items[2].ID(), Day: 3},
	}

	a1, _ := assignment.New("R1", items[0].ID())
	a2, _ := assignment.New("R1", items[1].ID())
	a3, _ := assignment.New("R1", items[2].ID())
	assignments := []assignment.Assignment{a1, a2, a3}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithContracts([]Contract{
		{ID: "c1", MinAssignments: 1, MaxAssignments: 2},
	})

	violations := rawTotalAssignments(assignments, ctx)
	// 3 assignments, max is 2, so 1 violation.
	if violations != 1 {
		t.Errorf("expected 1 violation, got %d", violations)
	}
}

func TestRawConsecutiveWorkingDays_ExceedsMax(t *testing.T) {
	items := makeWorkItems(4)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ContractID: "c1", ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 1},
		{WorkItemID: items[1].ID(), Day: 2},
		{WorkItemID: items[2].ID(), Day: 3},
		{WorkItemID: items[3].ID(), Day: 4},
	}

	var assignments []assignment.Assignment
	for _, item := range items {
		a, _ := assignment.New("R1", item.ID())
		assignments = append(assignments, a)
	}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithContracts([]Contract{
		{ID: "c1", MaxConsecutiveWorkingDays: 3},
	})

	violations := rawConsecutiveWorkingDays(assignments, ctx)
	// 4 consecutive, max 3 → 1 violation.
	if violations != 1 {
		t.Errorf("expected 1 violation, got %d", violations)
	}
}

func TestRawConsecutiveWorkingDays_BelowMin(t *testing.T) {
	items := makeWorkItems(1)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ContractID: "c1", ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 1},
	}

	a, _ := assignment.New("R1", items[0].ID())
	assignments := []assignment.Assignment{a}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithContracts([]Contract{
		{ID: "c1", MinConsecutiveWorkingDays: 2},
	})

	violations := rawConsecutiveWorkingDays(assignments, ctx)
	// 1 consecutive, min 2 → 1 violation.
	if violations != 1 {
		t.Errorf("expected 1 violation, got %d", violations)
	}
}

func TestRawDayRequests_DayOffViolated(t *testing.T) {
	items := makeWorkItems(1)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 1},
	}

	a, _ := assignment.New("R1", items[0].ID())
	assignments := []assignment.Assignment{a}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithRequests([]Request{
		{ResourceID: "R1", Day: 1, Type: "dayOff", Weight: 10},
	})

	violations := rawDayRequests(assignments, ctx)
	if violations != 1 {
		t.Errorf("expected 1 dayOff violation, got %d", violations)
	}
}

func TestRawDayRequests_DayOnViolated(t *testing.T) {
	items := makeWorkItems(1)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 2},
	}

	// Assign to day 2 — R1 requested day 1 on.
	a, _ := assignment.New("R1", items[0].ID())
	assignments := []assignment.Assignment{a}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithRequests([]Request{
		{ResourceID: "R1", Day: 1, Type: "dayOn", Weight: 10},
	})

	violations := rawDayRequests(assignments, ctx)
	// R1 not working day 1 → 1 violation.
	if violations != 1 {
		t.Errorf("expected 1 dayOn violation, got %d", violations)
	}
}

func TestRawShiftRequests_ShiftOffViolated(t *testing.T) {
	items := makeWorkItems(1)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 1, ShiftType: "early"},
	}

	a, _ := assignment.New("R1", items[0].ID())
	assignments := []assignment.Assignment{a}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithRequests([]Request{
		{ResourceID: "R1", Day: 1, ShiftType: "early", Type: "shiftOff", Weight: 10},
	})

	violations := rawShiftRequests(assignments, ctx)
	if violations != 1 {
		t.Errorf("expected 1 shiftOff violation, got %d", violations)
	}
}

func TestRawShiftRequests_ShiftOnViolated(t *testing.T) {
	items := makeWorkItems(1)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 1, ShiftType: "late"},
	}

	// R1 assigned late but requested early on day 1.
	a, _ := assignment.New("R1", items[0].ID())
	assignments := []assignment.Assignment{a}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithRequests([]Request{
		{ResourceID: "R1", Day: 1, ShiftType: "early", Type: "shiftOn", Weight: 10},
	})

	violations := rawShiftRequests(assignments, ctx)
	if violations != 1 {
		t.Errorf("expected 1 shiftOn violation, got %d", violations)
	}
}

func TestRawOptimalCoverage_GapPenalised(t *testing.T) {
	items := makeWorkItems(1)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 1, ShiftType: "early", RequiredSkill: "general", DemandGroup: "day1-early-general"},
	}

	a, _ := assignment.New("R1", items[0].ID())
	assignments := []assignment.Assignment{a}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithCoverageRequirements([]CoverageRequirement{
		{Day: 1, ShiftType: "early", Skill: "general", Minimum: 1, Optimal: 3},
	})

	violations := rawOptimalCoverage(assignments, ctx)
	// 1 assigned, min 1, optimal 3 → gap of 2.
	if violations != 2 {
		t.Errorf("expected 2 coverage gap violations, got %d", violations)
	}
}

func TestRawWorkingWeekends_ExceedsMax(t *testing.T) {
	// Days 6,7 = weekend 1; days 13,14 = weekend 2.
	items := makeWorkItems(2)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ContractID: "c1", ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 6},
		{WorkItemID: items[1].ID(), Day: 13},
	}

	a1, _ := assignment.New("R1", items[0].ID())
	a2, _ := assignment.New("R1", items[1].ID())
	assignments := []assignment.Assignment{a1, a2}

	ctx := NewContext(items, capacities, priorities)
	ctx = ctx.WithContracts([]Contract{
		{ID: "c1", MaxWorkingWeekends: 1},
	})

	violations := rawWorkingWeekends(assignments, ctx)
	// 2 working weekends, max 1 → 1 violation.
	if violations != 1 {
		t.Errorf("expected 1 weekend violation, got %d", violations)
	}
}

func TestRawCompleteWeekend_IncompleteViolation(t *testing.T) {
	// Works Saturday (day 6) but not Sunday (day 7).
	items := makeWorkItems(1)
	capacities := []ResourceInput{
		{ResourceID: "R1", Capacity: 960, Available: true, ContractID: "c1", ShiftStart: 360, ShiftEnd: 1320},
	}
	priorities := []WorkItemInput{
		{WorkItemID: items[0].ID(), Day: 6},
	}

	a, _ := assignment.New("R1", items[0].ID())
	assignments := []assignment.Assignment{a}

	ctx := NewContext(items, capacities, priorities)
	// Need max day >= 7 for weekend check.
	extraItems := makeWorkItemsFrom(1, 7) // day 7 work item (not assigned to R1)
	allPriorities := append(priorities, WorkItemInput{WorkItemID: extraItems[0].ID(), Day: 7})
	allItems := append(items, extraItems...)
	ctx = NewContext(allItems, capacities, allPriorities)
	ctx = ctx.WithContracts([]Contract{
		{ID: "c1", CompleteWeekend: true},
	})

	violations := rawCompleteWeekend(assignments, ctx)
	if violations != 1 {
		t.Errorf("expected 1 incomplete weekend violation, got %d", violations)
	}
}

// --- Helpers ---

func makeWorkItems(n int) []workitem.WorkItem {
	return makeWorkItemsFrom(n, 0)
}

func makeWorkItemsFrom(n int, startIdx int) []workitem.WorkItem {
	items := make([]workitem.WorkItem, n)
	for i := 0; i < n; i++ {
		id := "WI-" + itoa(startIdx+i+1)
		items[i], _ = workitem.New(id, "test", []byte(`{}`))
	}
	return items
}
