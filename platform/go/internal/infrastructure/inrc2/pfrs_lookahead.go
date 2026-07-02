package inrc2

// --- Amortized Global Constraint Look-Ahead ---
// Estimates future penalty from contractual caps (S7: total assignments, S8: weekends).
// Pure heuristic for beam path ranking. Does NOT change official scoring.
// Does NOT modify optimiser, acceptance, scoring, or branching behaviour.

// LookaheadPenalty computes an estimated future penalty for a given history state.
//
// Design principles (per Gemini's analysis):
// 1. Time-scaled weight: confidence increases as horizon approaches.
//    Weight = baseWeight * (currentWeek / totalWeeks). Weak early, strong late.
// 2. Asymmetric: aggressive on max cap overshoots, relaxed on min undershoots.
//    Min assignments are easy to catch up; max overruns are irreversible.
// 3. Floor-based trigger for max: only penalise if the nurse is *guaranteed*
//    to bust the cap even at minimum future workload. Prevents false positives
//    from early-week variance.
//
// weight: base scaling factor (0.0 = disabled, 1.0 = full estimate at horizon).
// Returns the estimated additional penalty that will materialise at the horizon.
func LookaheadPenalty(sc Scenario, hist History, weight float64) int {
	if weight <= 0 {
		return 0
	}

	currentWeek := hist.Week // after solving week W, hist.Week = W
	if currentWeek <= 0 {
		return 0
	}
	totalWeeks := sc.NumberOfWeeks
	if currentWeek >= totalWeeks {
		return 0 // already at horizon, official scorer handles it
	}

	// Time-scaled confidence: ramps from weak (early) to full (late).
	// At week 1: factor = 1/8 = 0.125. At week 7: factor = 7/8 = 0.875.
	timeFactor := float64(currentWeek) / float64(totalWeeks)
	effectiveWeight := weight * timeFactor

	remainingWeeks := totalWeeks - currentWeek

	// Build contract lookup.
	contractMap := make(map[string]Contract, len(sc.Contracts))
	for _, c := range sc.Contracts {
		contractMap[c.ID] = c
	}

	// Build nurse -> contract lookup.
	nurseContract := make(map[string]string, len(sc.Nurses))
	for _, n := range sc.Nurses {
		nurseContract[n.ID] = n.Contract
	}

	totalPenalty := 0.0

	for _, nh := range hist.NurseHistory {
		contract, ok := contractMap[nurseContract[nh.Nurse]]
		if !ok {
			continue
		}

		// --- S7: Total Assignments (20 per unit over/under) ---
		currentAssignments := nh.NumberOfAssignments

		// MAX assignments: floor-based trigger.
		// If nurse has already exceeded their max, guaranteed penalty.
		// If nurse would bust even working minimum shifts per remaining week:
		// minimum future = remainingWeeks * (minConsecutiveWorkDays is not relevant here,
		// but minimum coverage demand will force some assignments).
		// Simplification: use 0 minimum future (most conservative — only penalise if ALREADY over).
		// More aggressive: project at current rate.
		if contract.MaximumNumberOfAssignments > 0 {
			// Project: current rate * totalWeeks / currentWeek.
			projected := float64(currentAssignments) * float64(totalWeeks) / float64(currentWeek)
			overshoot := projected - float64(contract.MaximumNumberOfAssignments)

			if overshoot > 0 {
				// Floor check: is the nurse guaranteed to bust even if they work minimally?
				// Minimum future assignments: assume at least 3 per remaining week (coverage demands force this).
				minFutureAssignments := remainingWeeks * 3
				guaranteedTotal := currentAssignments + minFutureAssignments
				guaranteedOvershoot := guaranteedTotal - contract.MaximumNumberOfAssignments

				if guaranteedOvershoot > 0 {
					// Guaranteed bust — full penalty weight.
					totalPenalty += float64(guaranteedOvershoot) * 20.0
				} else {
					// Projected bust but not guaranteed — use projected overshoot at reduced confidence.
					totalPenalty += overshoot * 20.0 * 0.5
				}
			}
		}

		// MIN assignments: relaxed (easy to catch up by assigning more shifts later).
		// Only penalise if mathematically impossible to reach minimum.
		if contract.MinimumNumberOfAssignments > 0 {
			// Maximum future assignments: assume at most 7 per remaining week.
			maxFutureAssignments := remainingWeeks * 7
			bestCaseTotal := currentAssignments + maxFutureAssignments
			if bestCaseTotal < contract.MinimumNumberOfAssignments {
				// Impossible to reach minimum — guaranteed undershoot.
				undershoot := contract.MinimumNumberOfAssignments - bestCaseTotal
				totalPenalty += float64(undershoot) * 20.0
			}
			// Otherwise: no penalty. Can easily catch up.
		}

		// --- S8: Total Working Weekends (30 per weekend over max) ---
		currentWeekends := nh.NumberOfWorkingWeekends

		if contract.MaximumNumberOfWorkingWeekends > 0 {
			// Already over: guaranteed penalty.
			if currentWeekends > contract.MaximumNumberOfWorkingWeekends {
				alreadyOver := currentWeekends - contract.MaximumNumberOfWorkingWeekends
				totalPenalty += float64(alreadyOver) * 30.0
			} else {
				// Project: will they bust at current rate?
				projected := float64(currentWeekends) * float64(totalWeeks) / float64(currentWeek)
				overshoot := projected - float64(contract.MaximumNumberOfWorkingWeekends)
				if overshoot > 0 {
					// Not yet guaranteed, but trending over.
					totalPenalty += overshoot * 30.0 * 0.5
				}
			}
		}
	}

	return int(totalPenalty * effectiveWeight)
}

// --- Per-Week Budget Strategy ---
// Divides S7/S8 contractual limits by total weeks to get a per-week budget.
// Penalises paths whose nurses exceed their per-week budget immediately.
// Simpler and more direct than projection-based look-ahead.
// Pure beam ranking heuristic — does NOT change official scoring.

// BudgetPenalty computes how much a path's history exceeds fair per-week budgets.
// For each nurse: if their cumulative assignments or weekends exceed what the budget
// allows at this point in the horizon, apply penalty proportional to the overshoot.
//
// weight: scaling factor (0.0 = disabled).
func BudgetPenalty(sc Scenario, hist History, weight float64) int {
	if weight <= 0 {
		return 0
	}

	currentWeek := hist.Week
	if currentWeek <= 0 {
		return 0
	}
	totalWeeks := sc.NumberOfWeeks
	if currentWeek >= totalWeeks {
		return 0 // at horizon, official scorer handles it
	}

	// Build contract lookup.
	contractMap := make(map[string]Contract, len(sc.Contracts))
	for _, c := range sc.Contracts {
		contractMap[c.ID] = c
	}
	nurseContract := make(map[string]string, len(sc.Nurses))
	for _, n := range sc.Nurses {
		nurseContract[n.ID] = n.Contract
	}

	totalPenalty := 0.0

	for _, nh := range hist.NurseHistory {
		contract, ok := contractMap[nurseContract[nh.Nurse]]
		if !ok {
			continue
		}

		// S7: Total Assignments budget.
		// Fair budget at this point: max * (currentWeek / totalWeeks).
		if contract.MaximumNumberOfAssignments > 0 {
			budgetNow := float64(contract.MaximumNumberOfAssignments) * float64(currentWeek) / float64(totalWeeks)
			overshoot := float64(nh.NumberOfAssignments) - budgetNow
			if overshoot > 0 {
				totalPenalty += overshoot * 20.0 // matches S7 penalty rate
			}
		}
		if contract.MinimumNumberOfAssignments > 0 {
			budgetNow := float64(contract.MinimumNumberOfAssignments) * float64(currentWeek) / float64(totalWeeks)
			undershoot := budgetNow - float64(nh.NumberOfAssignments)
			if undershoot > 0 {
				totalPenalty += undershoot * 20.0
			}
		}

		// S8: Total Working Weekends budget.
		if contract.MaximumNumberOfWorkingWeekends > 0 {
			budgetNow := float64(contract.MaximumNumberOfWorkingWeekends) * float64(currentWeek) / float64(totalWeeks)
			overshoot := float64(nh.NumberOfWorkingWeekends) - budgetNow
			if overshoot > 0 {
				totalPenalty += overshoot * 30.0 // matches S8 penalty rate
			}
		}
	}

	return int(totalPenalty * weight)
}
