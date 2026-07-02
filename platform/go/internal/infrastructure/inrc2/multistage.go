package inrc2

// ScoreMultiStage evaluates a complete multi-stage INRC-II solution sequence
// against the official validator's scoring rules.
//
// This function produces the same total as the official validator by:
// - Evaluating S1 (optimal coverage) per week
// - Evaluating S2/S3/S4 (consecutive constraints) across the full horizon
// - Evaluating S5 (preferences) per week
// - Evaluating S6 (complete weekends) per week
// - Evaluating S7 (total assignments) once at end of horizon
// - Evaluating S8 (working weekends) once at end of horizon
// - Evaluating hard constraints per week
func ScoreMultiStage(sc Scenario, weeks []WeekData, hist History, solutions []Solution) ScoreResult {
	var result ScoreResult

	// Build full schedule: nurse -> absolute day (0-indexed across all weeks) -> assignment.
	nurseSchedule := make(map[string]map[int]assignEntry)
	for _, nurse := range sc.Nurses {
		nurseSchedule[nurse.ID] = make(map[int]assignEntry)
	}

	numWeeks := len(solutions)
	currentHist := hist

	// Per-week evaluation for S1, S5, S6, and hard constraints.
	for weekIdx := 0; weekIdx < numWeeks; weekIdx++ {
		sol := solutions[weekIdx]
		wd := weeks[weekIdx]

		// Populate full schedule for this week.
		dayOffset := weekIdx * 7
		for _, a := range sol.Assignments {
			dayIdx := DayIndex(a.Day)
			if dayIdx < 0 {
				continue
			}
			absDay := dayOffset + dayIdx
			nurseSchedule[a.Nurse][absDay] = assignEntry{shiftType: a.ShiftType, skill: a.Skill}
		}

		// Score single-week constraints using existing scorer for hard + S1 + S5.
		weekResult := Score(sc, wd, currentHist, sol)
		result.HardViolations += weekResult.HardViolations
		result.HardDetails = append(result.HardDetails, weekResult.HardDetails...)

		// S1: Optimal coverage (already in weekResult).
		for _, d := range weekResult.SoftDetails {
			if d.Constraint == "S1_OptimalCoverage" || d.Constraint == "S5_ShiftOffRequest" {
				result.SoftPenalty += d.Penalty
				result.SoftDetails = append(result.SoftDetails, d)
			}
		}

		// S6: Complete weekends per week.
		for _, d := range weekResult.SoftDetails {
			if d.Constraint == "S6_CompleteWeekend" {
				result.SoftPenalty += d.Penalty
				result.SoftDetails = append(result.SoftDetails, d)
			}
		}

		// Update history for next week.
		currentHist = UpdateHistory(sc, currentHist, sol)
	}

	// --- Full-horizon evaluation ---

	// Build contract lookup.
	contractOf := make(map[string]Contract)
	for _, c := range sc.Contracts {
		contractOf[c.ID] = c
	}
	nurseContract := make(map[string]Contract)
	for _, n := range sc.Nurses {
		nurseContract[n.ID] = contractOf[n.Contract]
	}

	// Build shift type limits.
	shiftLimits := make(map[string]ShiftType)
	for _, st := range sc.ShiftTypes {
		shiftLimits[st.ID] = st
	}

	totalDays := numWeeks * 7

	for _, nurse := range sc.Nurses {
		contract := nurseContract[nurse.ID]
		schedule := nurseSchedule[nurse.ID]

		// Count total assignments.
		totalAssign := hist.nurseAssignments(nurse.ID) + len(schedule)

		// Count working weekends (Sat/Sun pairs across full horizon).
		totalWeekends := hist.nurseWeekends(nurse.ID)
		for w := 0; w < numWeeks; w++ {
			satDay := w*7 + 5
			sunDay := w*7 + 6
			_, satWorked := schedule[satDay]
			_, sunWorked := schedule[sunDay]
			if satWorked || sunWorked {
				totalWeekends++
			}
		}

		// S7: Total assignments (end of horizon).
		if contract.MinimumNumberOfAssignments > 0 && totalAssign < contract.MinimumNumberOfAssignments {
			penalty := (contract.MinimumNumberOfAssignments - totalAssign) * 20
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S7_TotalAssignments", Nurse: nurse.ID, Penalty: penalty,
			})
		}
		if contract.MaximumNumberOfAssignments > 0 && totalAssign > contract.MaximumNumberOfAssignments {
			penalty := (totalAssign - contract.MaximumNumberOfAssignments) * 20
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S7_TotalAssignments", Nurse: nurse.ID, Penalty: penalty,
			})
		}

		// S8: Working weekends (end of horizon).
		if contract.MaximumNumberOfWorkingWeekends > 0 && totalWeekends > contract.MaximumNumberOfWorkingWeekends {
			penalty := (totalWeekends - contract.MaximumNumberOfWorkingWeekends) * 30
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S8_TotalWorkingWeekends", Nurse: nurse.ID, Penalty: penalty,
			})
		}

		// S2: Consecutive working days across full horizon (including history).
		penalty := scoreFullHorizonConsecutiveWorkingDays(schedule, totalDays, contract, hist, nurse.ID)
		if penalty > 0 {
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S2_ConsecutiveWorkingDays", Nurse: nurse.ID, Penalty: penalty,
			})
		}

		// S3: Consecutive days off across full horizon (including history).
		penalty = scoreFullHorizonConsecutiveDaysOff(schedule, totalDays, contract, hist, nurse.ID)
		if penalty > 0 {
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S3_ConsecutiveDaysOff", Nurse: nurse.ID, Penalty: penalty,
			})
		}

		// S4: Consecutive shift type across full horizon (including history).
		penalty = scoreFullHorizonConsecutiveShiftType(schedule, totalDays, shiftLimits, hist, nurse.ID)
		if penalty > 0 {
			result.SoftPenalty += penalty
			result.SoftDetails = append(result.SoftDetails, SoftPenaltyDetail{
				Constraint: "S4_ConsecutiveShiftType", Nurse: nurse.ID, Penalty: penalty,
			})
		}
	}

	result.TotalObjective = result.SoftPenalty
	return result
}

// Helper methods on History for multi-stage access.
func (h History) nurseAssignments(nurseID string) int {
	for _, nh := range h.NurseHistory {
		if nh.Nurse == nurseID {
			return nh.NumberOfAssignments
		}
	}
	return 0
}

func (h History) nurseWeekends(nurseID string) int {
	for _, nh := range h.NurseHistory {
		if nh.Nurse == nurseID {
			return nh.NumberOfWorkingWeekends
		}
	}
	return 0
}

func (h History) nurseHistoryFor(nurseID string) NurseHistory {
	for _, nh := range h.NurseHistory {
		if nh.Nurse == nurseID {
			return nh
		}
	}
	return NurseHistory{}
}

// scoreFullHorizonConsecutiveWorkingDays evaluates S2 across the entire planning horizon.
func scoreFullHorizonConsecutiveWorkingDays(schedule map[int]assignEntry, totalDays int, contract Contract, hist History, nurseID string) int {
	min := contract.MinimumNumberOfConsecutiveWorkingDays
	max := contract.MaximumNumberOfConsecutiveWorkingDays
	if min == 0 && max == 0 {
		return 0
	}

	nh := hist.nurseHistoryFor(nurseID)
	penalty := 0

	// Start with historical streak.
	streak := 0
	if nh.NumberOfConsecutiveWorkingDays > 0 {
		if _, ok := schedule[0]; ok {
			streak = nh.NumberOfConsecutiveWorkingDays
		} else {
			// History streak ended before this horizon — check min.
			if min > 0 && nh.NumberOfConsecutiveWorkingDays < min {
				penalty += (min - nh.NumberOfConsecutiveWorkingDays) * 30
			}
		}
	}

	for d := 0; d < totalDays; d++ {
		if _, ok := schedule[d]; ok {
			streak++
		} else {
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

	// End of horizon: check max only (min not penalised at end per INRC-II rules).
	if streak > 0 && max > 0 && streak > max {
		penalty += (streak - max) * 30
	}

	return penalty
}

// scoreFullHorizonConsecutiveDaysOff evaluates S3 across the entire planning horizon.
func scoreFullHorizonConsecutiveDaysOff(schedule map[int]assignEntry, totalDays int, contract Contract, hist History, nurseID string) int {
	min := contract.MinimumNumberOfConsecutiveDaysOff
	max := contract.MaximumNumberOfConsecutiveDaysOff
	if min == 0 && max == 0 {
		return 0
	}

	nh := hist.nurseHistoryFor(nurseID)
	penalty := 0

	// Start with historical days-off streak.
	streak := 0
	if nh.NumberOfConsecutiveDaysOff > 0 {
		if _, ok := schedule[0]; !ok {
			streak = nh.NumberOfConsecutiveDaysOff
		} else {
			// History off-streak ended — check min.
			if min > 0 && nh.NumberOfConsecutiveDaysOff < min {
				penalty += (min - nh.NumberOfConsecutiveDaysOff) * 30
			}
		}
	}

	for d := 0; d < totalDays; d++ {
		if _, ok := schedule[d]; !ok {
			streak++
		} else {
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

	// End of horizon: check max only.
	if streak > 0 && max > 0 && streak > max {
		penalty += (streak - max) * 30
	}

	return penalty
}

// scoreFullHorizonConsecutiveShiftType evaluates S4 across the entire planning horizon.
func scoreFullHorizonConsecutiveShiftType(schedule map[int]assignEntry, totalDays int, shiftLimits map[string]ShiftType, hist History, nurseID string) int {
	nh := hist.nurseHistoryFor(nurseID)
	penalty := 0

	currentType := ""
	streak := 0

	// Initialize from history.
	if nh.LastAssignedShiftType != "" && nh.LastAssignedShiftType != "None" {
		if entry, ok := schedule[0]; ok && entry.shiftType == nh.LastAssignedShiftType {
			currentType = nh.LastAssignedShiftType
			streak = nh.NumberOfConsecutiveAssignments
		} else if entry, ok := schedule[0]; ok && entry.shiftType != nh.LastAssignedShiftType {
			// History shift type streak ended — check min.
			if st, exists := shiftLimits[nh.LastAssignedShiftType]; exists {
				if st.MinimumNumberOfConsecutiveAssignments > 0 && nh.NumberOfConsecutiveAssignments < st.MinimumNumberOfConsecutiveAssignments {
					penalty += (st.MinimumNumberOfConsecutiveAssignments - nh.NumberOfConsecutiveAssignments) * 15
				}
			}
			_ = entry // use the variable
		} else if _, ok := schedule[0]; !ok {
			// Day off on first day — history shift streak ended.
			if st, exists := shiftLimits[nh.LastAssignedShiftType]; exists {
				if st.MinimumNumberOfConsecutiveAssignments > 0 && nh.NumberOfConsecutiveAssignments < st.MinimumNumberOfConsecutiveAssignments {
					penalty += (st.MinimumNumberOfConsecutiveAssignments - nh.NumberOfConsecutiveAssignments) * 15
				}
			}
		}
	}

	for d := 0; d < totalDays; d++ {
		entry, assigned := schedule[d]
		if !assigned {
			// Day off ends streak.
			if streak > 0 && currentType != "" {
				if st, exists := shiftLimits[currentType]; exists {
					if st.MaximumNumberOfConsecutiveAssignments > 0 && streak > st.MaximumNumberOfConsecutiveAssignments {
						penalty += (streak - st.MaximumNumberOfConsecutiveAssignments) * 15
					}
					if st.MinimumNumberOfConsecutiveAssignments > 0 && streak < st.MinimumNumberOfConsecutiveAssignments {
						penalty += (st.MinimumNumberOfConsecutiveAssignments - streak) * 15
					}
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
				if st, exists := shiftLimits[currentType]; exists {
					if st.MaximumNumberOfConsecutiveAssignments > 0 && streak > st.MaximumNumberOfConsecutiveAssignments {
						penalty += (streak - st.MaximumNumberOfConsecutiveAssignments) * 15
					}
					if st.MinimumNumberOfConsecutiveAssignments > 0 && streak < st.MinimumNumberOfConsecutiveAssignments {
						penalty += (st.MinimumNumberOfConsecutiveAssignments - streak) * 15
					}
				}
			}
			currentType = entry.shiftType
			streak = 1
		}
	}

	// End of horizon: check max only.
	if streak > 0 && currentType != "" {
		if st, exists := shiftLimits[currentType]; exists {
			if st.MaximumNumberOfConsecutiveAssignments > 0 && streak > st.MaximumNumberOfConsecutiveAssignments {
				penalty += (streak - st.MaximumNumberOfConsecutiveAssignments) * 15
			}
		}
	}

	return penalty
}
