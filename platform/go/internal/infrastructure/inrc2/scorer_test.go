package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestScore_OfficialSolution_NoHardViolations(t *testing.T) {
	sc, err := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	if err != nil {
		t.Fatalf("load scenario: %v", err)
	}
	wd, err := inrc2.LoadWeekData(testDataDir + "WD-n005w4-1.json")
	if err != nil {
		t.Fatalf("load week data: %v", err)
	}
	hist, err := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")
	if err != nil {
		t.Fatalf("load history: %v", err)
	}
	sol, err := inrc2.LoadSolution(testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-1-0.json")
	if err != nil {
		t.Fatalf("load solution: %v", err)
	}

	result := inrc2.Score(sc, wd, hist, sol)

	// Official solutions should have 0 hard violations.
	if result.HardViolations != 0 {
		t.Errorf("expected 0 hard violations, got %d", result.HardViolations)
		for _, v := range result.HardDetails {
			t.Logf("  [%s] nurse=%s day=%d: %s", v.Code, v.Nurse, v.Day, v.Message)
		}
	}

	// Soft penalty should be > 0 (the official solution has some soft violations).
	t.Logf("Soft penalty: %d", result.SoftPenalty)
	for _, d := range result.SoftDetails {
		t.Logf("  [%s] nurse=%s penalty=%d", d.Constraint, d.Nurse, d.Penalty)
	}
}

func TestScore_DetectsSingleAssignmentViolation(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Create a solution with duplicate assignment.
	sol := inrc2.Solution{
		Scenario: "n005w4",
		Week:     0,
		Assignments: []inrc2.Assignment{
			{Nurse: "Patrick", Day: "Mon", ShiftType: "Early", Skill: "Nurse"},
			{Nurse: "Patrick", Day: "Mon", ShiftType: "Late", Skill: "Nurse"},
		},
	}

	result := inrc2.Score(sc, wd, hist, sol)
	if result.HardViolations == 0 {
		t.Error("expected hard violation for double assignment on same day")
	}

	found := false
	for _, v := range result.HardDetails {
		if v.Code == "H1_SingleAssignment" {
			found = true
		}
	}
	if !found {
		t.Error("expected H1_SingleAssignment violation")
	}
}

func TestScore_DetectsSkillViolation(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Sara only has "Nurse" skill, assign as HeadNurse.
	sol := inrc2.Solution{
		Scenario: "n005w4",
		Week:     0,
		Assignments: []inrc2.Assignment{
			{Nurse: "Sara", Day: "Mon", ShiftType: "Early", Skill: "HeadNurse"},
		},
	}

	result := inrc2.Score(sc, wd, hist, sol)
	found := false
	for _, v := range result.HardDetails {
		if v.Code == "H2_Skill" {
			found = true
		}
	}
	if !found {
		t.Error("expected H2_Skill violation for Sara assigned as HeadNurse")
	}
}

func TestScore_DetectsForbiddenSuccessionFromHistory(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Patrick's last shift was Night. Night -> Early is forbidden.
	sol := inrc2.Solution{
		Scenario: "n005w4",
		Week:     0,
		Assignments: []inrc2.Assignment{
			{Nurse: "Patrick", Day: "Mon", ShiftType: "Early", Skill: "Nurse"},
		},
	}

	result := inrc2.Score(sc, wd, hist, sol)
	found := false
	for _, v := range result.HardDetails {
		if v.Code == "H4_Succession" && v.Nurse == "Patrick" {
			found = true
		}
	}
	if !found {
		t.Error("expected H4_Succession for Patrick (Night->Early from history)")
	}
}

func TestScore_ShiftOffRequestPenalty(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Sara has a shift-off request for Thursday (Any).
	// Assign Sara on Thursday.
	sol := inrc2.Solution{
		Scenario: "n005w4",
		Week:     0,
		Assignments: []inrc2.Assignment{
			{Nurse: "Sara", Day: "Thu", ShiftType: "Night", Skill: "Nurse"},
		},
	}

	result := inrc2.Score(sc, wd, hist, sol)
	found := false
	for _, d := range result.SoftDetails {
		if d.Constraint == "S5_ShiftOffRequest" && d.Nurse == "Sara" {
			found = true
		}
	}
	if !found {
		t.Error("expected S5_ShiftOffRequest penalty for Sara on Thursday")
	}
}
