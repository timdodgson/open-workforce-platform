package ilp_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestExtractSolutions_BasicMapping(t *testing.T) {
	sc, _, _ := loadTestInstance(t)

	// Variable format: x_<nurseIdx>_<dayIdx>_<shiftType>_<skill>
	solValues := map[string]float64{
		"x_0_0_Early_HeadNurse": 1.0, // Patrick, Mon, Early, HeadNurse
		"x_1_1_Late_Nurse":     1.0, // Andrea, Tue, Late, Nurse
		"x_2_8_Night_HeadNurse": 1.0, // Stefaan, Day 8 = week 1 day 1
	}

	solutions := ilp.ExtractSolutions(sc, 2, solValues)

	if len(solutions) != 2 {
		t.Fatalf("expected 2 week solutions, got %d", len(solutions))
	}

	// Week 0 should have 2 assignments.
	if len(solutions[0].Assignments) != 2 {
		t.Errorf("week 0: expected 2 assignments, got %d", len(solutions[0].Assignments))
	}

	// Week 1 should have 1 assignment.
	if len(solutions[1].Assignments) != 1 {
		t.Errorf("week 1: expected 1 assignment, got %d", len(solutions[1].Assignments))
	}

	// Verify first assignment has correct skill.
	found := false
	for _, a := range solutions[0].Assignments {
		if a.Nurse == "Patrick" && a.Day == "Mon" && a.ShiftType == "Early" && a.Skill == "HeadNurse" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Patrick-Mon-Early-HeadNurse assignment in week 0")
	}
}

func TestExtractSolutions_IgnoresNonXVariables(t *testing.T) {
	sc, _, _ := loadTestInstance(t)

	solValues := map[string]float64{
		"x_0_0_Early_Nurse":    1.0,
		"s1_0_Early_Nurse":     1.0, // slack variable — should be ignored
		"s5_0_0":               1.0, // violation variable — should be ignored
		"s6_0_0":               1.0, // weekend violation — should be ignored
	}

	solutions := ilp.ExtractSolutions(sc, 1, solValues)

	if len(solutions[0].Assignments) != 1 {
		t.Errorf("expected 1 assignment (ignoring slack vars), got %d", len(solutions[0].Assignments))
	}
}

func TestExtractSolutions_CorrectDayMapping(t *testing.T) {
	sc, _, _ := loadTestInstance(t)

	// Test all 7 days in a week.
	solValues := map[string]float64{
		"x_0_0_Early_Nurse": 1.0, // Monday
		"x_0_1_Early_Nurse": 1.0, // Tuesday
		"x_0_2_Early_Nurse": 1.0, // Wednesday
		"x_0_3_Early_Nurse": 1.0, // Thursday
		"x_0_4_Early_Nurse": 1.0, // Friday
		"x_0_5_Early_Nurse": 1.0, // Saturday
		"x_0_6_Early_Nurse": 1.0, // Sunday
	}

	solutions := ilp.ExtractSolutions(sc, 1, solValues)

	expectedDays := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	if len(solutions[0].Assignments) != 7 {
		t.Fatalf("expected 7 assignments, got %d", len(solutions[0].Assignments))
	}

	daySet := make(map[string]bool)
	for _, a := range solutions[0].Assignments {
		daySet[a.Day] = true
	}
	for _, d := range expectedDays {
		if !daySet[d] {
			t.Errorf("missing day %s in assignments", d)
		}
	}
}

func TestExtractSolutions_SkillPreserved(t *testing.T) {
	sc, _, _ := loadTestInstance(t)

	solValues := map[string]float64{
		"x_0_0_Early_HeadNurse": 1.0,
		"x_0_1_Late_Nurse":     1.0,
	}

	solutions := ilp.ExtractSolutions(sc, 1, solValues)

	for _, a := range solutions[0].Assignments {
		if a.Day == "Mon" && a.Skill != "HeadNurse" {
			t.Errorf("Monday assignment should have HeadNurse skill, got %s", a.Skill)
		}
		if a.Day == "Tue" && a.Skill != "Nurse" {
			t.Errorf("Tuesday assignment should have Nurse skill, got %s", a.Skill)
		}
	}
}

func TestValidateILPSolution_DetectsHardViolations(t *testing.T) {
	sc, weekFiles, hist := loadTestInstance(t)

	// Night->Early is forbidden succession.
	solutions := []inrc2.Solution{{
		Scenario: sc.ID,
		Week:     0,
		Assignments: []inrc2.Assignment{
			{Nurse: "Andrea", Day: "Mon", ShiftType: "Night", Skill: "HeadNurse"},
			{Nurse: "Andrea", Day: "Tue", ShiftType: "Early", Skill: "HeadNurse"},
		},
	}}

	_, hardViolations, _, err := ilp.ValidateILPSolution(sc, weekFiles[:1], hist, solutions)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	if hardViolations == 0 {
		t.Error("expected hard violations for Night->Early succession, got 0")
	}
}
