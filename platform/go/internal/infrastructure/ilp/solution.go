package ilp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// ExtractSolutions converts ILP solution variable values into per-week INRC-II Solutions.
// Each week gets its own Solution struct suitable for validation with the official scorer.
//
// Variable format: x_<nurseIdx>_<dayIdx>_<shiftType>_<skill>
func ExtractSolutions(sc inrc2.Scenario, weeks int, solValues map[string]float64) []inrc2.Solution {
	solutions := make([]inrc2.Solution, weeks)
	for w := 0; w < weeks; w++ {
		solutions[w] = inrc2.Solution{
			Scenario: sc.ID,
			Week:     w,
		}
	}

	for varName, val := range solValues {
		if val < 0.5 {
			continue
		}
		nurseIdx, dayIdx, shiftType, skill, ok := parseXVar(varName)
		if !ok {
			continue
		}
		if nurseIdx < 0 || nurseIdx >= len(sc.Nurses) {
			continue
		}

		week := dayIdx / 7
		dayInWeek := dayIdx % 7
		if week >= weeks {
			continue
		}

		nurse := sc.Nurses[nurseIdx]
		solutions[week].Assignments = append(solutions[week].Assignments, inrc2.Assignment{
			Nurse:     nurse.ID,
			Day:       inrc2.DayName(dayInWeek),
			ShiftType: shiftType,
			Skill:     skill,
		})
	}

	return solutions
}

// parseXVar parses "x_<nurseIdx>_<dayIdx>_<shiftType>_<skill>" variable names.
func parseXVar(name string) (nurseIdx, dayIdx int, shiftType, skill string, ok bool) {
	if !strings.HasPrefix(name, "x_") {
		return 0, 0, "", "", false
	}
	rest := name[2:] // remove "x_"

	// Split into parts: nurseIdx, dayIdx, shiftType, skill.
	// Problem: shiftType and skill may not contain underscores, but we split on _.
	// Format is exactly 4 underscore-separated fields after "x_".
	parts := strings.SplitN(rest, "_", 4)
	if len(parts) < 4 {
		return 0, 0, "", "", false
	}

	ni, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, "", "", false
	}
	di, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, "", "", false
	}

	return ni, di, parts[2], parts[3], true
}

// ValidateILPSolution runs each week's solution through the official INRC-II scorer
// and returns the total penalty and hard violations.
func ValidateILPSolution(sc inrc2.Scenario, weekDataFiles []string, initialHist inrc2.History, solutions []inrc2.Solution) (totalPenalty int, totalHardViolations int, perWeek []inrc2.ScoreResult, err error) {
	weeks := len(solutions)
	perWeek = make([]inrc2.ScoreResult, weeks)

	currentHist := initialHist
	for w := 0; w < weeks; w++ {
		wd, loadErr := inrc2.LoadWeekData(weekDataFiles[w])
		if loadErr != nil {
			return 0, 0, nil, fmt.Errorf("failed to load week %d data: %w", w+1, loadErr)
		}

		result := inrc2.Score(sc, wd, currentHist, solutions[w])
		perWeek[w] = result
		totalPenalty += result.SoftPenalty
		totalHardViolations += result.HardViolations

		// Update history for next week.
		currentHist = inrc2.UpdateHistory(sc, currentHist, solutions[w])
	}

	return totalPenalty, totalHardViolations, perWeek, nil
}
