package optimisation

import (
	"sort"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/domain/assignment"
)

// --- INRC-II Soft Constraint Scoring Functions ---
// Each function returns a raw violation count (positive = violations).
// The weight (negative for penalties) is applied by ObjectiveScore.

// rawOptimalCoverage counts the number of coverage gaps between minimum and optimal.
// For each coverage requirement where assigned count >= minimum but < optimal,
// each gap unit counts as one violation.
func rawOptimalCoverage(assignments []assignment.Assignment, ctx OptimisationContext) int {
	reqs := ctx.CoverageRequirements()
	if len(reqs) == 0 {
		return 0
	}

	workItems := ctx.WorkItems()

	// Build lookup: demandGroup -> count of assigned items in that group.
	groupOf := make(map[string]string, len(workItems))
	for _, wi := range workItems {
		if wi.DemandGroup != "" {
			groupOf[wi.WorkItemID] = wi.DemandGroup
		}
	}

	assignedInGroup := make(map[string]int)
	for _, a := range assignments {
		if g, ok := groupOf[a.WorkItemID()]; ok {
			assignedInGroup[g]++
		}
	}

	// Count gaps between assigned and optimal for each coverage requirement.
	violations := 0
	for _, req := range reqs {
		group := demandGroupKey(req.Day, req.ShiftType, req.Skill)
		assigned := assignedInGroup[group]
		if assigned >= req.Minimum && assigned < req.Optimal {
			violations += req.Optimal - assigned
		}
	}

	return violations
}

// rawConsecutiveWorkingDays counts violations of min/max consecutive working day rules.
// Returns total number of violations across all resources.
func rawConsecutiveWorkingDays(assignments []assignment.Assignment, ctx OptimisationContext) int {
	contracts := ctx.Contracts()
	if len(contracts) == 0 {
		return 0
	}

	resources := ctx.Resources()
	contractOf := buildContractLookup(contracts)
	resourceDays := buildResourceDayMap(assignments, ctx)

	violations := 0
	for _, rc := range resources {
		if !rc.Available {
			continue
		}
		contract, ok := contractOf[rc.ContractID]
		if !ok {
			continue
		}
		days := resourceDays[rc.ResourceID]
		if len(days) == 0 {
			continue
		}
		violations += countConsecutiveWorkingDayViolations(days, contract)
	}

	return violations
}

// rawConsecutiveDaysOff counts violations of min/max consecutive days off rules.
func rawConsecutiveDaysOff(assignments []assignment.Assignment, ctx OptimisationContext) int {
	contracts := ctx.Contracts()
	if len(contracts) == 0 {
		return 0
	}

	resources := ctx.Resources()
	contractOf := buildContractLookup(contracts)
	resourceDays := buildResourceDayMap(assignments, ctx)

	// Determine planning horizon (max day across all work items).
	maxDay := findMaxDay(ctx.WorkItems())
	if maxDay == 0 {
		return 0
	}

	violations := 0
	for _, rc := range resources {
		if !rc.Available {
			continue
		}
		contract, ok := contractOf[rc.ContractID]
		if !ok {
			continue
		}
		days := resourceDays[rc.ResourceID]
		violations += countConsecutiveDaysOffViolations(days, maxDay, contract)
	}

	return violations
}

// rawConsecutiveShiftType counts violations of min/max consecutive same-shift-type rules.
func rawConsecutiveShiftType(assignments []assignment.Assignment, ctx OptimisationContext) int {
	shiftTypes := ctx.ShiftTypes()
	if len(shiftTypes) == 0 {
		return 0
	}

	shiftInfo := make(map[string]ShiftTypeInfo, len(shiftTypes))
	for _, st := range shiftTypes {
		shiftInfo[st.ID] = st
	}

	resources := ctx.Resources()
	workItems := ctx.WorkItems()

	// Build resource -> day -> shift type.
	shiftTypeOf := make(map[string]string, len(workItems))
	dayOf := make(map[string]int, len(workItems))
	for _, wi := range workItems {
		shiftTypeOf[wi.WorkItemID] = wi.ShiftType
		dayOf[wi.WorkItemID] = wi.Day
	}

	byResource := make(map[string][]string)
	for _, a := range assignments {
		byResource[a.ResourceID()] = append(byResource[a.ResourceID()], a.WorkItemID())
	}

	violations := 0
	for _, rc := range resources {
		if !rc.Available {
			continue
		}
		itemIDs := byResource[rc.ResourceID]
		if len(itemIDs) == 0 {
			continue
		}
		// Build day -> shift type for this resource.
		dayShift := make(map[int]string)
		for _, id := range itemIDs {
			d := dayOf[id]
			st := shiftTypeOf[id]
			if d > 0 && st != "" {
				dayShift[d] = st
			}
		}
		violations += countConsecutiveShiftTypeViolations(dayShift, shiftInfo)
	}

	return violations
}

// rawWorkingWeekends counts violations of max working weekends per resource.
func rawWorkingWeekends(assignments []assignment.Assignment, ctx OptimisationContext) int {
	contracts := ctx.Contracts()
	if len(contracts) == 0 {
		return 0
	}

	resources := ctx.Resources()
	contractOf := buildContractLookup(contracts)
	resourceDays := buildResourceDayMap(assignments, ctx)

	// Weekend start defaults to 6 (Saturday in Mon-start week).
	weekendStart := 6 // This would ideally come from context but we use convention.

	violations := 0
	for _, rc := range resources {
		if !rc.Available {
			continue
		}
		contract, ok := contractOf[rc.ContractID]
		if !ok {
			continue
		}
		days := resourceDays[rc.ResourceID]
		workingWeekends := countWorkingWeekends(days, weekendStart)
		if workingWeekends > contract.MaxWorkingWeekends {
			violations += workingWeekends - contract.MaxWorkingWeekends
		}
	}

	return violations
}

// rawCompleteWeekend counts violations of the complete weekend rule.
// A violation occurs when a nurse works Saturday XOR Sunday but not both.
func rawCompleteWeekend(assignments []assignment.Assignment, ctx OptimisationContext) int {
	contracts := ctx.Contracts()
	if len(contracts) == 0 {
		return 0
	}

	resources := ctx.Resources()
	contractOf := buildContractLookup(contracts)
	resourceDays := buildResourceDayMap(assignments, ctx)
	maxDay := findMaxDay(ctx.WorkItems())
	weekendStart := 6

	violations := 0
	for _, rc := range resources {
		if !rc.Available {
			continue
		}
		contract, ok := contractOf[rc.ContractID]
		if !ok || !contract.CompleteWeekend {
			continue
		}
		days := resourceDays[rc.ResourceID]
		violations += countIncompleteWeekends(days, maxDay, weekendStart)
	}

	return violations
}

// rawTotalAssignments counts violations of min/max total assignments per resource.
func rawTotalAssignments(assignments []assignment.Assignment, ctx OptimisationContext) int {
	contracts := ctx.Contracts()
	if len(contracts) == 0 {
		return 0
	}

	resources := ctx.Resources()
	contractOf := buildContractLookup(contracts)

	// Count assignments per resource.
	countOf := make(map[string]int)
	for _, a := range assignments {
		countOf[a.ResourceID()]++
	}

	violations := 0
	for _, rc := range resources {
		if !rc.Available {
			continue
		}
		contract, ok := contractOf[rc.ContractID]
		if !ok {
			continue
		}
		count := countOf[rc.ResourceID]
		if contract.MinAssignments > 0 && count < contract.MinAssignments {
			violations += contract.MinAssignments - count
		}
		if contract.MaxAssignments > 0 && count > contract.MaxAssignments {
			violations += count - contract.MaxAssignments
		}
	}

	return violations
}

// rawShiftRequests counts violations of shift-on and shift-off requests.
func rawShiftRequests(assignments []assignment.Assignment, ctx OptimisationContext) int {
	requests := ctx.Requests()
	if len(requests) == 0 {
		return 0
	}

	workItems := ctx.WorkItems()
	dayOf := make(map[string]int, len(workItems))
	shiftTypeOf := make(map[string]string, len(workItems))
	for _, wi := range workItems {
		dayOf[wi.WorkItemID] = wi.Day
		shiftTypeOf[wi.WorkItemID] = wi.ShiftType
	}

	// Build resource -> set of (day, shiftType) assigned.
	type dayShiftKey struct {
		day       int
		shiftType string
	}
	resourceAssigned := make(map[string]map[dayShiftKey]bool)
	for _, a := range assignments {
		if resourceAssigned[a.ResourceID()] == nil {
			resourceAssigned[a.ResourceID()] = make(map[dayShiftKey]bool)
		}
		d := dayOf[a.WorkItemID()]
		st := shiftTypeOf[a.WorkItemID()]
		if d > 0 && st != "" {
			resourceAssigned[a.ResourceID()][dayShiftKey{d, st}] = true
		}
	}

	violations := 0
	for _, req := range requests {
		if req.ShiftType == "" {
			continue // day-level requests handled by rawDayRequests
		}
		switch req.Type {
		case "shiftOn":
			// Violated if nurse is NOT assigned this shift on this day.
			assigned := resourceAssigned[req.ResourceID]
			if assigned == nil || !assigned[dayShiftKey{req.Day, req.ShiftType}] {
				violations++
			}
		case "shiftOff":
			// Violated if nurse IS assigned this shift on this day.
			assigned := resourceAssigned[req.ResourceID]
			if assigned != nil && assigned[dayShiftKey{req.Day, req.ShiftType}] {
				violations++
			}
		}
	}

	return violations
}

// rawDayRequests counts violations of day-on and day-off requests.
func rawDayRequests(assignments []assignment.Assignment, ctx OptimisationContext) int {
	requests := ctx.Requests()
	if len(requests) == 0 {
		return 0
	}

	workItems := ctx.WorkItems()
	dayOf := make(map[string]int, len(workItems))
	for _, wi := range workItems {
		dayOf[wi.WorkItemID] = wi.Day
	}

	// Build resource -> set of days worked.
	resourceWorkDays := make(map[string]map[int]bool)
	for _, a := range assignments {
		if resourceWorkDays[a.ResourceID()] == nil {
			resourceWorkDays[a.ResourceID()] = make(map[int]bool)
		}
		d := dayOf[a.WorkItemID()]
		if d > 0 {
			resourceWorkDays[a.ResourceID()][d] = true
		}
	}

	violations := 0
	for _, req := range requests {
		if req.ShiftType != "" {
			continue // shift-level requests handled by rawShiftRequests
		}
		switch req.Type {
		case "dayOn":
			// Violated if nurse is NOT working on this day.
			days := resourceWorkDays[req.ResourceID]
			if days == nil || !days[req.Day] {
				violations++
			}
		case "dayOff":
			// Violated if nurse IS working on this day.
			days := resourceWorkDays[req.ResourceID]
			if days != nil && days[req.Day] {
				violations++
			}
		}
	}

	return violations
}

// --- Helper functions for NRP objective scoring ---

// demandGroupKey constructs the demand group identifier matching the NRP adapter.
func demandGroupKey(day int, shiftType, skill string) string {
	return "day" + itoa(day) + "-" + shiftType + "-" + skill
}

// itoa converts an int to string without fmt dependency.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf) - 1
	for n > 0 {
		buf[i] = byte('0' + n%10)
		n /= 10
		i--
	}
	if neg {
		buf[i] = '-'
		i--
	}
	return string(buf[i+1:])
}

// buildContractLookup creates a map from contract ID to Contract.
func buildContractLookup(contracts []Contract) map[string]Contract {
	m := make(map[string]Contract, len(contracts))
	for _, c := range contracts {
		m[c.ID] = c
	}
	return m
}

// buildResourceDayMap builds a map from resource ID to sorted slice of working days.
func buildResourceDayMap(assignments []assignment.Assignment, ctx OptimisationContext) map[string][]int {
	workItems := ctx.WorkItems()
	dayOf := make(map[string]int, len(workItems))
	for _, wi := range workItems {
		dayOf[wi.WorkItemID] = wi.Day
	}

	daySet := make(map[string]map[int]bool)
	for _, a := range assignments {
		d := dayOf[a.WorkItemID()]
		if d > 0 {
			if daySet[a.ResourceID()] == nil {
				daySet[a.ResourceID()] = make(map[int]bool)
			}
			daySet[a.ResourceID()][d] = true
		}
	}

	result := make(map[string][]int, len(daySet))
	for resID, days := range daySet {
		sorted := make([]int, 0, len(days))
		for d := range days {
			sorted = append(sorted, d)
		}
		sort.Ints(sorted)
		result[resID] = sorted
	}
	return result
}

// findMaxDay returns the highest day number across all work items.
func findMaxDay(workItems []WorkItemInput) int {
	max := 0
	for _, wi := range workItems {
		if wi.Day > max {
			max = wi.Day
		}
	}
	return max
}

// countConsecutiveWorkingDayViolations counts violations of min/max consecutive working days.
func countConsecutiveWorkingDayViolations(workingDays []int, contract Contract) int {
	if len(workingDays) == 0 {
		return 0
	}

	violations := 0
	streak := 1

	for i := 1; i < len(workingDays); i++ {
		if workingDays[i] == workingDays[i-1]+1 {
			streak++
		} else {
			// End of streak.
			if contract.MinConsecutiveWorkingDays > 0 && streak < contract.MinConsecutiveWorkingDays {
				violations += contract.MinConsecutiveWorkingDays - streak
			}
			if contract.MaxConsecutiveWorkingDays > 0 && streak > contract.MaxConsecutiveWorkingDays {
				violations += streak - contract.MaxConsecutiveWorkingDays
			}
			streak = 1
		}
	}

	// Final streak.
	if contract.MinConsecutiveWorkingDays > 0 && streak < contract.MinConsecutiveWorkingDays {
		violations += contract.MinConsecutiveWorkingDays - streak
	}
	if contract.MaxConsecutiveWorkingDays > 0 && streak > contract.MaxConsecutiveWorkingDays {
		violations += streak - contract.MaxConsecutiveWorkingDays
	}

	return violations
}

// countConsecutiveDaysOffViolations counts violations of min/max consecutive days off.
func countConsecutiveDaysOffViolations(workingDays []int, maxDay int, contract Contract) int {
	if maxDay == 0 {
		return 0
	}

	// Build set of working days.
	working := make(map[int]bool, len(workingDays))
	for _, d := range workingDays {
		working[d] = true
	}

	violations := 0
	streak := 0

	for d := 1; d <= maxDay; d++ {
		if !working[d] {
			streak++
		} else {
			if streak > 0 {
				if contract.MinConsecutiveDaysOff > 0 && streak < contract.MinConsecutiveDaysOff {
					violations += contract.MinConsecutiveDaysOff - streak
				}
				if contract.MaxConsecutiveDaysOff > 0 && streak > contract.MaxConsecutiveDaysOff {
					violations += streak - contract.MaxConsecutiveDaysOff
				}
			}
			streak = 0
		}
	}

	// Final streak.
	if streak > 0 {
		if contract.MinConsecutiveDaysOff > 0 && streak < contract.MinConsecutiveDaysOff {
			violations += contract.MinConsecutiveDaysOff - streak
		}
		if contract.MaxConsecutiveDaysOff > 0 && streak > contract.MaxConsecutiveDaysOff {
			violations += streak - contract.MaxConsecutiveDaysOff
		}
	}

	return violations
}

// countConsecutiveShiftTypeViolations counts violations for a resource's shift pattern.
func countConsecutiveShiftTypeViolations(dayShift map[int]string, shiftInfo map[string]ShiftTypeInfo) int {
	if len(dayShift) == 0 {
		return 0
	}

	// Get sorted days.
	days := make([]int, 0, len(dayShift))
	for d := range dayShift {
		days = append(days, d)
	}
	sort.Ints(days)

	violations := 0
	currentType := dayShift[days[0]]
	streak := 1

	for i := 1; i < len(days); i++ {
		st := dayShift[days[i]]
		if st == currentType && days[i] == days[i-1]+1 {
			streak++
		} else {
			// End of streak for currentType.
			if info, ok := shiftInfo[currentType]; ok {
				if info.MinConsecutiveAssignments > 0 && streak < info.MinConsecutiveAssignments {
					violations += info.MinConsecutiveAssignments - streak
				}
				if info.MaxConsecutiveAssignments > 0 && streak > info.MaxConsecutiveAssignments {
					violations += streak - info.MaxConsecutiveAssignments
				}
			}
			currentType = st
			streak = 1
		}
	}

	// Final streak.
	if info, ok := shiftInfo[currentType]; ok {
		if info.MinConsecutiveAssignments > 0 && streak < info.MinConsecutiveAssignments {
			violations += info.MinConsecutiveAssignments - streak
		}
		if info.MaxConsecutiveAssignments > 0 && streak > info.MaxConsecutiveAssignments {
			violations += streak - info.MaxConsecutiveAssignments
		}
	}

	return violations
}

// countWorkingWeekends returns how many weekends the resource works.
// A weekend is worked if the nurse has an assignment on weekendStart or weekendStart+1.
func countWorkingWeekends(workingDays []int, weekendStart int) int {
	weekends := make(map[int]bool) // key = weekend number
	for _, d := range workingDays {
		// Determine which weekend this day belongs to.
		if d >= weekendStart {
			weekNum := (d - weekendStart) / 7
			dayInWeek := (d - weekendStart) % 7
			if dayInWeek == 0 || dayInWeek == 1 {
				weekends[weekNum] = true
			}
		}
	}
	return len(weekends)
}

// countIncompleteWeekends counts weekends where only Saturday or only Sunday is worked.
func countIncompleteWeekends(workingDays []int, maxDay, weekendStart int) int {
	working := make(map[int]bool, len(workingDays))
	for _, d := range workingDays {
		working[d] = true
	}

	violations := 0
	// Iterate through possible weekends.
	for sat := weekendStart; sat <= maxDay; sat += 7 {
		sun := sat + 1
		workedSat := working[sat]
		workedSun := sun <= maxDay && working[sun]
		if workedSat != workedSun {
			violations++
		}
	}

	return violations
}
