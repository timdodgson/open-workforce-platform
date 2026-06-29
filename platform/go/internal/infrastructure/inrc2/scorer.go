package inrc2

import "sort"

// ScoreResult contains the official INRC-II scoring breakdown.
type ScoreResult struct {
	HardViolations     int
	SoftPenalty        int
	TotalObjective     int
	HardDetails        []Violation
	SoftDetails        []SoftPenaltyDetail
}

// Violation describes a single hard constraint violation.
type Violation struct {
	Code    string
	Nurse   string
	Day     int // -1 if not day-specific
	Message string
}

// SoftPenaltyDetail describes a soft constraint penalty contribution.
type SoftPenaltyDetail struct {
	Constraint string
	Nurse      string
	Penalty    int
}

// assignEntry represents what a nurse is doing on a given day.
type assignEntry struct {
	shiftType string
	skill     string
}

// Score evaluates an INRC-II solution against the scenario, week data, and history.
// It returns hard violations and soft penalties using official INRC-II rules.
//
// Official INRC-II penalty weights (from the specification):
//   - Insufficient staffing for optimal coverage: 30 per unit
//   - Consecutive assignments (working days): 30 per day
//   - Consecutive days off: 30 per day
//   - Preferences (shift off requests): 10 per request
//   - Complete weekend: 30 per weekend
//   - Total assignments: 20 per assignment
//   - Total working weekends: 30 per weekend
//   - Consecutive same shift type: 15 per assignment
func Score(sc Scenario, wd WeekData, hist History, sol Solution) ScoreResult {
	var result ScoreResult

	// Build lookups.
	nurseContract := make(map[string]Contract)
	contractOf := make(map[string]string)
	for _, n := range sc.Nurses {
		contractOf[n.ID] = n.Contract
	}
	for _, c := range sc.Contracts {
		for _, n := range sc.Nurses {
			if n.Contract == c.ID {
				nurseContract[n.ID] = c
			}
		}
	}

	nurseSkills := make(map[string]map[string]bool)
	for _, n := range sc.Nurses {
		skills := make(map[string]bool, len(n.Skills))
		for _, s := range n.Skills {
			skills[s] = true
		}
		nurseSkills[n.ID] = skills
	}

	// Build forbidden succession lookup: preceding -> set of forbidden successors.
	forbidden := make(map[string]map[string]bool)
	for _, fs := range sc.ForbiddenShiftTypeSuccessions {
		if len(fs.SucceedingShiftTypes) > 0 {
			set := make(map[string]bool, len(fs.SucceedingShiftTypes))
			for _, s := range fs.SucceedingShiftTypes {
				set[s] = true
			}
			forbidden[fs.PrecedingShiftType] = set
		}
	}

	// Build assignment matrix: nurse -> day -> (shiftType, skill).
	nurseDay := make(map[string]map[int]assignEntry)
	for _, a := range sol.Assignments {
		dayIdx := DayIndex(a.Day)
		if dayIdx < 0 {
			continue
		}
		if nurseDay[a.Nurse] == nil {
			nurseDay[a.Nurse] = make(map[int]assignEntry)
		}
		if _, exists := nurseDay[a.Nurse][dayIdx]; exists {
			// H1: duplicate assignment on same day.
			result.HardDetails = append(result.HardDetails, Violation{
				Code: "H1_SingleAssignment", Nurse: a.Nurse, Day: dayIdx,
				Message: "nurse assigned multiple shifts on same day",
			})
			result.HardViolations++
		}
		nurseDay[a.Nurse][dayIdx] = assignEntry{shiftType: a.ShiftType, skill: a.Skill}
	}

	// H2: Skill requirement check.
	for _, a := range sol.Assignments {
		if a.Skill != "" {
			skills := nurseSkills[a.Nurse]
			if !skills[a.Skill] {
				result.HardDetails = append(result.HardDetails, Violation{
					Code: "H2_Skill", Nurse: a.Nurse, Day: DayIndex(a.Day),
					Message: "nurse lacks required skill: " + a.Skill,
				})
				result.HardViolations++
			}
		}
	}

	// H3: Minimum coverage.
	for _, req := range wd.Requirements {
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			dayReq := req.RequirementForDay(dayIdx)
			if dayReq.Minimum == 0 {
				continue
			}
			count := 0
			for _, a := range sol.Assignments {
				if DayIndex(a.Day) == dayIdx && a.ShiftType == req.ShiftType && a.Skill == req.Skill {
					count++
				}
			}
			if count < dayReq.Minimum {
				deficit := dayReq.Minimum - count
				for i := 0; i < deficit; i++ {
					result.HardDetails = append(result.HardDetails, Violation{
						Code: "H3_MinCoverage", Day: dayIdx,
						Message: DayName(dayIdx) + " " + req.ShiftType + "/" + req.Skill + " under minimum",
					})
					result.HardViolations++
				}
			}
		}
	}

	// H4: Forbidden shift succession.
	// Check history -> day 0 and day-to-day within week.
	nurseHist := make(map[string]NurseHistory)
	for _, nh := range hist.NurseHistory {
		nurseHist[nh.Nurse] = nh
	}

	for _, nurse := range sc.Nurses {
		schedule := nurseDay[nurse.ID]

		// History -> Monday.
		// Only applies if nurse was working on the day immediately before this week
		// (i.e., no days off between last shift and this Monday).
		nh := nurseHist[nurse.ID]
		if nh.LastAssignedShiftType != "" && nh.LastAssignedShiftType != "None" && nh.NumberOfConsecutiveDaysOff == 0 {
			if entry, ok := schedule[0]; ok {
				if succ, exists := forbidden[nh.LastAssignedShiftType]; exists && succ[entry.shiftType] {
					result.HardDetails = append(result.HardDetails, Violation{
						Code: "H4_Succession", Nurse: nurse.ID, Day: 0,
						Message: nh.LastAssignedShiftType + " -> " + entry.shiftType + " (from history)",
					})
					result.HardViolations++
				}
			}
		}

		// Within week.
		for d := 0; d < 6; d++ {
			entryD, okD := schedule[d]
			entryNext, okNext := schedule[d+1]
			if okD && okNext {
				if succ, exists := forbidden[entryD.shiftType]; exists && succ[entryNext.shiftType] {
					result.HardDetails = append(result.HardDetails, Violation{
						Code: "H4_Succession", Nurse: nurse.ID, Day: d + 1,
						Message: entryD.shiftType + " -> " + entryNext.shiftType,
					})
					result.HardViolations++
				}
			}
		}
	}

	// --- Soft Constraints ---

	// S1: Optimal coverage (30 per unit).
	for _, req := range wd.Requirements {
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			dayReq := req.RequirementForDay(dayIdx)
			if dayReq.Optimal <= dayReq.Minimum {
				continue
			}
			count := 0
			for _, a := range sol.Assignments {
				if DayIndex(a.Day) == dayIdx && a.ShiftType == req.ShiftType && a.Skill == req.Skill {
					count++
				}
			}
			if count >= dayReq.Minimum && count < dayReq.Optimal {
				gap := dayReq.Optimal - count
				penalty := gap * 30
				result.SoftPenalty += penalty
				result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
					Constraint: "S1_OptimalCoverage", Penalty: penalty,
				})
			}
		}
	}

	// Per-nurse soft constraints.
	for _, nurse := range sc.Nurses {
		contract := nurseContract[nurse.ID]
		nh := nurseHist[nurse.ID]
		schedule := nurseDay[nurse.ID]

		// Build working days array for this week (0-indexed, Mon=0).
		var workDays []int
		for d := 0; d < 7; d++ {
			if _, ok := schedule[d]; ok {
				workDays = append(workDays, d)
			}
		}

		// S2: Consecutive working days (30 per violation).
		penalty := scoreConsecutiveWorkingDays(workDays, contract, nh)
		if penalty > 0 {
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S2_ConsecutiveWorkingDays", Nurse: nurse.ID, Penalty: penalty,
			})
		}

		// S3: Consecutive days off (30 per violation).
		penalty = scoreConsecutiveDaysOff(workDays, contract, nh)
		if penalty > 0 {
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S3_ConsecutiveDaysOff", Nurse: nurse.ID, Penalty: penalty,
			})
		}

		// S4: Consecutive same shift type (15 per violation).
		penalty = scoreConsecutiveShiftType(schedule, sc.ShiftTypes, nh)
		if penalty > 0 {
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S4_ConsecutiveShiftType", Nurse: nurse.ID, Penalty: penalty,
			})
		}

		// S6: Complete weekend (30 per violation).
		if contract.CompleteWeekends == 1 {
			penalty = scoreCompleteWeekend(schedule)
			if penalty > 0 {
				result.SoftPenalty += penalty
				result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
					Constraint: "S6_CompleteWeekend", Nurse: nurse.ID, Penalty: penalty,
				})
			}
		}

		// S7: Total assignments (20 per violation).
		totalAssign := nh.NumberOfAssignments + len(workDays)
		penalty = scoreTotalAssignments(totalAssign, contract, sc.NumberOfWeeks, hist.Week)
		if penalty > 0 {
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S7_TotalAssignments", Nurse: nurse.ID, Penalty: penalty,
			})
		}

		// S8: Total working weekends (30 per violation).
		weekendWorked := schedule[5] != (assignEntry{}) || schedule[6] != (assignEntry{})
		totalWeekends := nh.NumberOfWorkingWeekends
		if weekendWorked {
			totalWeekends++
		}
		penalty = scoreTotalWorkingWeekends(totalWeekends, contract, sc.NumberOfWeeks, hist.Week)
		if penalty > 0 {
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S8_TotalWorkingWeekends", Nurse: nurse.ID, Penalty: penalty,
			})
		}
	}

	// S5: Preferences / shift-off requests (10 per violation).
	for _, req := range wd.ShiftOffRequests {
		dayIdx := DayIndex(req.Day)
		if dayIdx < 0 {
			continue
		}
		schedule := nurseDay[req.Nurse]
		if schedule == nil {
			continue
		}
		entry, assigned := schedule[dayIdx]
		if !assigned {
			continue
		}
		// "Any" means any shift type.
		if req.ShiftType == "Any" || req.ShiftType == entry.shiftType {
			result.SoftPenalty += 10
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S5_ShiftOffRequest", Nurse: req.Nurse, Penalty: 10,
			})
		}
	}

	result.TotalObjective = result.SoftPenalty
	return result
}

// --- Soft constraint scoring helpers ---

// scoreConsecutiveWorkingDays evaluates S2 for a nurse's week, considering history.
// Returns penalty at 30 per violation unit.
func scoreConsecutiveWorkingDays(workDays []int, contract Contract, nh NurseHistory) int {
	if len(workDays) == 0 {
		return 0
	}
	sort.Ints(workDays)

	penalty := 0
	min := contract.MinimumNumberOfConsecutiveWorkingDays
	max := contract.MaximumNumberOfConsecutiveWorkingDays

	// Start with historical streak if nurse was working leading into this week.
	streak := 0
	if nh.NumberOfConsecutiveWorkingDays > 0 && len(workDays) > 0 && workDays[0] == 0 {
		streak = nh.NumberOfConsecutiveWorkingDays
	}

	prevDay := -2 // sentinel
	for _, d := range workDays {
		if d == prevDay+1 {
			streak++
		} else {
			// End of previous streak (if any).
			if streak > 0 && prevDay >= 0 {
				if max > 0 && streak > max {
					penalty += (streak - max) * 30
				}
			}
			if streak > 0 && streak < min && prevDay >= 0 {
				// Only penalise if streak ended within the week (not at week boundary).
				penalty += (min - streak) * 30
			}
			// Start new streak.
			if d == 0 && nh.NumberOfConsecutiveWorkingDays > 0 {
				streak = nh.NumberOfConsecutiveWorkingDays + 1
			} else {
				streak = 1
			}
		}
		prevDay = d
	}

	// Final streak: only check max at end of week (min checked only if streak
	// ended within the week, i.e., nurse has day off after last working day).
	if streak > 0 {
		if max > 0 && streak > max {
			penalty += (streak - max) * 30
		}
		// If the last working day is before Sunday, the streak ended.
		if prevDay < 6 && min > 0 && streak < min {
			penalty += (min - streak) * 30
		}
	}

	return penalty
}

// scoreConsecutiveDaysOff evaluates S3 for a nurse's week, considering history.
// Returns penalty at 30 per violation unit.
func scoreConsecutiveDaysOff(workDays []int, contract Contract, nh NurseHistory) int {
	// Build working set.
	working := make(map[int]bool, len(workDays))
	for _, d := range workDays {
		working[d] = true
	}

	min := contract.MinimumNumberOfConsecutiveDaysOff
	max := contract.MaximumNumberOfConsecutiveDaysOff
	penalty := 0

	// Start with historical days-off streak.
	streak := 0
	if nh.NumberOfConsecutiveDaysOff > 0 && !working[0] {
		streak = nh.NumberOfConsecutiveDaysOff
	}

	for d := 0; d < 7; d++ {
		if !working[d] {
			if streak == 0 && d == 0 && nh.NumberOfConsecutiveDaysOff > 0 {
				streak = nh.NumberOfConsecutiveDaysOff + 1
			} else {
				streak++
			}
		} else {
			// End of days-off streak.
			if streak > 0 {
				if max > 0 && streak > max {
					penalty += (streak - max) * 30
				}
				if min > 0 && streak < min {
					penalty += (min - streak) * 30
				}
			}
			streak = 0
		}
	}

	// Final streak: only check max at end of week.
	if streak > 0 {
		if max > 0 && streak > max {
			penalty += (streak - max) * 30
		}
		// Don't penalise min for trailing off days (carried to next week).
	}

	return penalty
}

// scoreConsecutiveShiftType evaluates S4 for a nurse's week, considering history.
// Returns penalty at 15 per violation unit.
func scoreConsecutiveShiftType(schedule map[int]assignEntry, shiftTypes []ShiftType, nh NurseHistory) int {
	// Build shift type limits.
	minConsec := make(map[string]int)
	maxConsec := make(map[string]int)
	for _, st := range shiftTypes {
		minConsec[st.ID] = st.MinimumNumberOfConsecutiveAssignments
		maxConsec[st.ID] = st.MaximumNumberOfConsecutiveAssignments
	}

	penalty := 0
	currentType := ""
	streak := 0

	// Initialize from history.
	if nh.LastAssignedShiftType != "" && nh.LastAssignedShiftType != "None" {
		if entry, ok := schedule[0]; ok && entry.shiftType == nh.LastAssignedShiftType {
			currentType = nh.LastAssignedShiftType
			streak = nh.NumberOfConsecutiveAssignments
		}
	}

	for d := 0; d < 7; d++ {
		entry, assigned := schedule[d]
		if !assigned {
			// Day off ends streak.
			if streak > 0 && currentType != "" {
				if max, ok := maxConsec[currentType]; ok && max > 0 && streak > max {
					penalty += (streak - max) * 15
				}
				if min, ok := minConsec[currentType]; ok && min > 0 && streak < min {
					penalty += (min - streak) * 15
				}
			}
			currentType = ""
			streak = 0
			continue
		}

		if entry.shiftType == currentType {
			streak++
		} else {
			// End of previous type streak.
			if streak > 0 && currentType != "" {
				if max, ok := maxConsec[currentType]; ok && max > 0 && streak > max {
					penalty += (streak - max) * 15
				}
				if min, ok := minConsec[currentType]; ok && min > 0 && streak < min {
					penalty += (min - streak) * 15
				}
			}
			currentType = entry.shiftType
			streak = 1
		}
	}

	// Final streak at end of week: check max only (min carries to next week).
	if streak > 0 && currentType != "" {
		if max, ok := maxConsec[currentType]; ok && max > 0 && streak > max {
			penalty += (streak - max) * 15
		}
	}

	return penalty
}

// scoreCompleteWeekend evaluates S6. Returns penalty at 30 per incomplete weekend.
func scoreCompleteWeekend(schedule map[int]assignEntry) int {
	_, satWorked := schedule[5]
	_, sunWorked := schedule[6]
	if satWorked != sunWorked {
		return 30
	}
	return 0
}

// scoreTotalAssignments evaluates S7. Returns penalty at 20 per assignment over/under.
// Only penalised at the end of the planning horizon.
func scoreTotalAssignments(totalAssign int, contract Contract, totalWeeks, currentWeek int) int {
	// Only evaluate at the last week of the horizon.
	if currentWeek < totalWeeks-1 {
		return 0
	}
	penalty := 0
	if contract.MinimumNumberOfAssignments > 0 && totalAssign < contract.MinimumNumberOfAssignments {
		penalty += (contract.MinimumNumberOfAssignments - totalAssign) * 20
	}
	if contract.MaximumNumberOfAssignments > 0 && totalAssign > contract.MaximumNumberOfAssignments {
		penalty += (totalAssign - contract.MaximumNumberOfAssignments) * 20
	}
	return penalty
}

// scoreTotalWorkingWeekends evaluates S8. Returns penalty at 30 per weekend over max.
// Only penalised at the end of the planning horizon.
func scoreTotalWorkingWeekends(totalWeekends int, contract Contract, totalWeeks, currentWeek int) int {
	if currentWeek < totalWeeks-1 {
		return 0
	}
	if contract.MaximumNumberOfWorkingWeekends > 0 && totalWeekends > contract.MaximumNumberOfWorkingWeekends {
		return (totalWeekends - contract.MaximumNumberOfWorkingWeekends) * 30
	}
	return 0
}
