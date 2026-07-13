package inrc2

import "fmt"

// weekLocalSoftConstraints are scored per week by the official validator and by ScoreMultiStage.
// Consecutive (S2/S3/S4) and horizon totals (S7/S8) must be evaluated once across the full horizon.
func isWeekLocalSoftConstraint(name string) bool {
	switch name {
	case "S1_OptimalCoverage", "S5_ShiftOffRequest", "S6_CompleteWeekend":
		return true
	default:
		return false
	}
}

// weekLocalSoftPenalty sums only S1/S5/S6 from a weekly Score result.
func weekLocalSoftPenalty(r ScoreResult) int {
	total := 0
	for _, d := range r.SoftDetails {
		if isWeekLocalSoftConstraint(d.Constraint) {
			total += d.Penalty
		}
	}
	return total
}

// loadPathWeeksAndSolutions loads week data and solutions for a beam path in week order.
func loadPathWeeksAndSolutions(weekFiles []string, path []BeamPath) (weeks []WeekData, sols []Solution, err error) {
	if len(path) == 0 {
		return nil, nil, fmt.Errorf("empty beam path")
	}
	weeks = make([]WeekData, len(path))
	sols = make([]Solution, len(path))
	for i, wp := range path {
		weekIdx := wp.Week - 1
		if weekIdx < 0 || weekIdx >= len(weekFiles) {
			return nil, nil, fmt.Errorf("path week %d out of range for %d week files", wp.Week, len(weekFiles))
		}
		wd, loadErr := LoadWeekData(weekFiles[weekIdx])
		if loadErr != nil {
			return nil, nil, fmt.Errorf("load week %d: %w", wp.Week, loadErr)
		}
		weeks[i] = wd
		sols[i] = wp.Solution
	}
	return weeks, sols, nil
}

// ScoreBeamPathOfficial scores a beam path with ScoreMultiStage (Java validator aligned).
// WeekPenalty is attributed so week-local soft (S1/S5/S6) stays on its week and horizon
// soft (S2/S3/S4/S7/S8) is placed on the final week — weeks sum to TotalObjective.
func ScoreBeamPathOfficial(sc Scenario, weekFiles []string, path []BeamPath, initialHist History) (updated []BeamPath, ms ScoreResult, err error) {
	weeks, sols, err := loadPathWeeksAndSolutions(weekFiles, path)
	if err != nil {
		return path, ScoreResult{}, err
	}

	ms = ScoreMultiStage(sc, weeks, initialHist, sols)
	updated = make([]BeamPath, len(path))
	copy(updated, path)

	hist := initialHist
	locals := make([]int, len(updated))
	localSum := 0
	for i := range updated {
		wr := Score(sc, weeks[i], hist, sols[i])
		locals[i] = weekLocalSoftPenalty(wr)
		localSum += locals[i]
		updated[i].ScoreResult = wr
		hist = UpdateHistory(sc, hist, sols[i])
	}

	horizon := ms.TotalObjective - localSum
	if horizon < 0 {
		horizon = 0
	}

	cum := 0
	for i := range updated {
		pen := locals[i]
		if i == len(updated)-1 {
			pen += horizon
			// Attach full multi-stage soft details on the final week for audit.
			updated[i].ScoreResult = ms
		}
		updated[i].WeekPenalty = pen
		cum += pen
		updated[i].CumulativePenalty = cum
	}

	return updated, ms, nil
}

// OfficialValidateBeamPath returns the official (ScoreMultiStage) soft penalty and soft-detail count.
func OfficialValidateBeamPath(sc Scenario, weekFiles []string, path []BeamPath, initialHist History) (totalPenalty, totalViolations int) {
	_, ms, err := ScoreBeamPathOfficial(sc, weekFiles, path, initialHist)
	if err != nil {
		return 0, 0
	}
	return ms.TotalObjective, len(ms.SoftDetails)
}

// OfficialRevalidateBeamPath re-scores a beam path with ScoreMultiStage and updates path penalties.
func OfficialRevalidateBeamPath(sc Scenario, weekFiles []string, path []BeamPath, initialHist History) (updated []BeamPath, finalPenalty int, totalViolations int) {
	updated, ms, err := ScoreBeamPathOfficial(sc, weekFiles, path, initialHist)
	if err != nil {
		return path, 0, 0
	}
	return updated, ms.TotalObjective, len(ms.SoftDetails)
}
