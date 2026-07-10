package inrc2

// OfficialValidateBeamPath scores each week on a beam path using the official scorer.
func OfficialValidateBeamPath(sc Scenario, weekFiles []string, path []BeamPath, initialHist History) (totalPenalty, totalViolations int) {
	weekData := make([]WeekData, len(weekFiles))
	weekLoaded := make([]bool, len(weekFiles))
	for i, wf := range weekFiles {
		wd, err := LoadWeekData(wf)
		if err != nil {
			continue
		}
		weekData[i] = wd
		weekLoaded[i] = true
	}

	valHist := initialHist
	for _, wp := range path {
		weekIdx := wp.Week - 1
		if weekIdx < 0 || weekIdx >= len(weekFiles) || !weekLoaded[weekIdx] {
			continue
		}
		result := Score(sc, weekData[weekIdx], valHist, wp.Solution)
		totalPenalty += result.SoftPenalty
		totalViolations += len(result.SoftDetails)
		valHist = UpdateHistory(sc, valHist, wp.Solution)
	}
	return totalPenalty, totalViolations
}
