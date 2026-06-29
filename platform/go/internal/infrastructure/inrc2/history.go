package inrc2

// UpdateHistory computes the new history state after solving a week.
// This carries forward all counters needed for multi-stage evaluation.
func UpdateHistory(sc Scenario, prevHist History, sol Solution) History {
	// Build assignment matrix: nurse -> day -> shiftType.
	nurseSchedule := make(map[string]map[int]string)
	for _, a := range sol.Assignments {
		if nurseSchedule[a.Nurse] == nil {
			nurseSchedule[a.Nurse] = make(map[int]string)
		}
		nurseSchedule[a.Nurse][DayIndex(a.Day)] = a.ShiftType
	}

	// Build previous history lookup.
	prevNurseHist := make(map[string]NurseHistory)
	for _, nh := range prevHist.NurseHistory {
		prevNurseHist[nh.Nurse] = nh
	}

	newHist := History{
		Week:     prevHist.Week + 1,
		Scenario: sc.ID,
	}

	for _, nurse := range sc.Nurses {
		prev := prevNurseHist[nurse.ID]
		schedule := nurseSchedule[nurse.ID]

		// Count assignments this week.
		weekAssignments := len(schedule)

		// Determine if weekend was worked (Sat=5 or Sun=6).
		_, satWorked := schedule[5]
		_, sunWorked := schedule[6]
		weekendWorked := satWorked || sunWorked

		// Compute trailing state from end of week (Sunday backward).
		lastShift := "None"
		consecutiveAssignments := 0
		consecutiveWorkingDays := 0
		consecutiveDaysOff := 0

		// Find the last assigned day (scanning backward from Sunday).
		lastWorkDay := -1
		for d := 6; d >= 0; d-- {
			if _, ok := schedule[d]; ok {
				lastWorkDay = d
				break
			}
		}

		if lastWorkDay == -1 {
			// Nurse didn't work at all this week.
			// Days off streak = previous streak + 7.
			consecutiveDaysOff = prev.NumberOfConsecutiveDaysOff + 7
			lastShift = "None"
		} else if lastWorkDay == 6 {
			// Nurse worked on Sunday — count trailing working streak.
			lastShift = schedule[6]
			consecutiveAssignments = 0
			consecutiveWorkingDays = 0

			// Count consecutive same-shift from end.
			for d := 6; d >= 0; d-- {
				st, ok := schedule[d]
				if !ok {
					break
				}
				consecutiveWorkingDays++
				if st == lastShift {
					consecutiveAssignments++
				} else {
					break
				}
			}

			// If streak extends all week and nurse was working before.
			if consecutiveWorkingDays == 7 && prev.NumberOfConsecutiveWorkingDays > 0 {
				consecutiveWorkingDays += prev.NumberOfConsecutiveWorkingDays
			}
			if consecutiveAssignments == 7 && prev.LastAssignedShiftType == lastShift && prev.NumberOfConsecutiveAssignments > 0 {
				consecutiveAssignments += prev.NumberOfConsecutiveAssignments
			}
		} else {
			// Nurse's last work day is before Sunday — trailing days off.
			consecutiveDaysOff = 6 - lastWorkDay
			lastShift = schedule[lastWorkDay]

			// Count consecutive same-shift ending on lastWorkDay.
			consecutiveAssignments = 0
			consecutiveWorkingDays = 0
			for d := lastWorkDay; d >= 0; d-- {
				st, ok := schedule[d]
				if !ok {
					break
				}
				consecutiveWorkingDays++
				if consecutiveAssignments == 0 || st == lastShift {
					if st == lastShift {
						consecutiveAssignments++
					} else {
						break
					}
				}
			}
		}

		newHist.NurseHistory = append(newHist.NurseHistory, NurseHistory{
			Nurse:                          nurse.ID,
			NumberOfAssignments:            prev.NumberOfAssignments + weekAssignments,
			NumberOfWorkingWeekends:        prev.NumberOfWorkingWeekends + boolToInt(weekendWorked),
			LastAssignedShiftType:          lastShift,
			NumberOfConsecutiveAssignments: consecutiveAssignments,
			NumberOfConsecutiveWorkingDays: consecutiveWorkingDays,
			NumberOfConsecutiveDaysOff:     consecutiveDaysOff,
		})
	}

	return newHist
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
