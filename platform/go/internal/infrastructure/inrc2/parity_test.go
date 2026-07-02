package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// TestValidatorParity_n005w4_H0_WD_1233 compares our scorer against the official
// validator output for the included n005w4 sample (Solution_H_0-WD_1-2-3-3).
//
// Official validator total: 1695
// Official hard violations: 0
func TestValidatorParity_n005w4_H0_WD_1233(t *testing.T) {
	sc, err := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}

	// Load all 4 week data files in sequence.
	weekFiles := []string{
		testDataDir + "WD-n005w4-1.json",
		testDataDir + "WD-n005w4-2.json",
		testDataDir + "WD-n005w4-3.json",
		testDataDir + "WD-n005w4-3.json", // Week 4 reuses WD-3
	}

	var weeks []inrc2.WeekData
	for _, f := range weekFiles {
		wd, err := inrc2.LoadWeekData(f)
		if err != nil {
			t.Fatalf("load week data %s: %v", f, err)
		}
		weeks = append(weeks, wd)
	}

	// Load initial history.
	hist, err := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")
	if err != nil {
		t.Fatalf("load history: %v", err)
	}

	// Load all 4 solutions.
	solDir := testDataDir + "Solution_H_0-WD_1-2-3-3/"
	solFiles := []string{
		solDir + "Sol-n005w4-1-0.json",
		solDir + "Sol-n005w4-2-1.json",
		solDir + "Sol-n005w4-3-2.json",
		solDir + "Sol-n005w4-3-3.json",
	}

	var solutions []inrc2.Solution
	for _, f := range solFiles {
		sol, err := inrc2.LoadSolution(f)
		if err != nil {
			t.Fatalf("load solution %s: %v", f, err)
		}
		solutions = append(solutions, sol)
	}

	// Run multi-stage scorer.
	result := inrc2.ScoreMultiStage(sc, weeks, hist, solutions)

	// Official validator: 0 hard violations.
	if result.HardViolations != 0 {
		t.Errorf("HARD VIOLATIONS: expected 0, got %d", result.HardViolations)
		for _, v := range result.HardDetails {
			t.Logf("  [%s] nurse=%s day=%d: %s", v.Code, v.Nurse, v.Day, v.Message)
		}
	}

	// Official validator total: 1695.
	const expectedTotal = 1695
	if result.TotalObjective != expectedTotal {
		t.Errorf("TOTAL SCORE: expected %d, got %d (diff: %d)",
			expectedTotal, result.TotalObjective, result.TotalObjective-expectedTotal)

		// Print breakdown for debugging.
		t.Log("Breakdown:")
		for _, d := range result.SoftDetails {
			t.Logf("  [%s] nurse=%s penalty=%d", d.Constraint, d.Nurse, d.Penalty)
		}
	}

	// Per-category comparison from official validator verbose output.
	categories := map[string]int{
		"S7_TotalAssignments":       320,
		"S2_ConsecutiveWorkingDays": 0,  // Will sum below
		"S4_ConsecutiveShiftType":   0,  // Will sum below
		"S3_ConsecutiveDaysOff":     330,
		"S5_ShiftOffRequest":        70,
		"S8_TotalWorkingWeekends":   210,
		"S6_CompleteWeekend":        60,
		"S1_OptimalCoverage":        240,
	}

	// The official validator groups S2 and S4 together as "Consecutive constraints: 465"
	// S2 (consecutive working days) + S4 (consecutive shift type) = 465
	expectedConsecutive := 465

	// Sum up our scores by category.
	actual := make(map[string]int)
	for _, d := range result.SoftDetails {
		actual[d.Constraint] += d.Penalty
	}

	t.Logf("Category comparison:")
	t.Logf("  S1_OptimalCoverage: expected=%d, got=%d", categories["S1_OptimalCoverage"], actual["S1_OptimalCoverage"])
	t.Logf("  S3_ConsecutiveDaysOff: expected=%d, got=%d", categories["S3_ConsecutiveDaysOff"], actual["S3_ConsecutiveDaysOff"])
	t.Logf("  S5_ShiftOffRequest: expected=%d, got=%d", categories["S5_ShiftOffRequest"], actual["S5_ShiftOffRequest"])
	t.Logf("  S6_CompleteWeekend: expected=%d, got=%d", categories["S6_CompleteWeekend"], actual["S6_CompleteWeekend"])
	t.Logf("  S7_TotalAssignments: expected=%d, got=%d", categories["S7_TotalAssignments"], actual["S7_TotalAssignments"])
	t.Logf("  S8_TotalWorkingWeekends: expected=%d, got=%d", categories["S8_TotalWorkingWeekends"], actual["S8_TotalWorkingWeekends"])
	t.Logf("  S2+S4 (Consecutive): expected=%d, got=%d", expectedConsecutive, actual["S2_ConsecutiveWorkingDays"]+actual["S4_ConsecutiveShiftType"])
}
