package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func TestUpdateHistory_BasicProperties(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")
	sol, _ := inrc2.LoadSolution(testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-1-0.json")

	newHist := inrc2.UpdateHistory(sc, hist, sol)

	if newHist.Week != 1 {
		t.Errorf("expected week 1, got %d", newHist.Week)
	}
	if newHist.Scenario != "n005w4" {
		t.Errorf("expected scenario n005w4, got %s", newHist.Scenario)
	}
	if len(newHist.NurseHistory) != 5 {
		t.Errorf("expected 5 nurse histories, got %d", len(newHist.NurseHistory))
	}

	// Patrick has 6 assignments in the solution.
	for _, nh := range newHist.NurseHistory {
		if nh.Nurse == "Patrick" {
			if nh.NumberOfAssignments != 6 {
				t.Errorf("expected Patrick 6 total assignments, got %d", nh.NumberOfAssignments)
			}
			// Patrick works Sat+Sun → weekend worked.
			if nh.NumberOfWorkingWeekends != 1 {
				t.Errorf("expected Patrick 1 working weekend, got %d", nh.NumberOfWorkingWeekends)
			}
			// Patrick's last assignment is Sunday Late.
			if nh.LastAssignedShiftType != "Late" {
				t.Errorf("expected Patrick last shift Late, got %s", nh.LastAssignedShiftType)
			}
		}
	}
}

func TestUpdateHistory_NurseWithNoAssignments(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Empty solution — no nurse assigned.
	sol := inrc2.Solution{Scenario: "n005w4", Week: 0}

	newHist := inrc2.UpdateHistory(sc, hist, sol)

	for _, nh := range newHist.NurseHistory {
		if nh.Nurse == "Patrick" {
			// Patrick had 0 assignments, still 0.
			if nh.NumberOfAssignments != 0 {
				t.Errorf("expected 0 assignments, got %d", nh.NumberOfAssignments)
			}
			// Patrick had 0 consecutive days off, now has 7 more.
			if nh.NumberOfConsecutiveDaysOff != 7 {
				t.Errorf("expected 7 consecutive days off, got %d", nh.NumberOfConsecutiveDaysOff)
			}
			if nh.LastAssignedShiftType != "None" {
				t.Errorf("expected last shift None, got %s", nh.LastAssignedShiftType)
			}
		}
	}
}

func TestSolveWeek_ProducesSolution(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	sol, _, err := inrc2.SolveWeek(sc, wd, hist, "constructive", inrc2DefaultProfile())
	if err != nil {
		t.Fatalf("solve failed: %v", err)
	}

	if sol.Scenario != "n005w4" {
		t.Errorf("expected scenario n005w4, got %s", sol.Scenario)
	}
	if len(sol.Assignments) == 0 {
		t.Error("expected at least some assignments")
	}

	// All assignments should reference valid nurses.
	nurseSet := make(map[string]bool)
	for _, n := range sc.Nurses {
		nurseSet[n.ID] = true
	}
	for _, a := range sol.Assignments {
		if !nurseSet[a.Nurse] {
			t.Errorf("assignment references unknown nurse: %s", a.Nurse)
		}
	}
}

func TestMultiStage_HistoryCarriesForward(t *testing.T) {
	sc, _ := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	wd0, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	wd1, _ := inrc2.LoadWeekData(testDataDir + "WD-n005w4-1.json")
	hist, _ := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")

	// Solve week 0.
	sol0, _, err := inrc2.SolveWeek(sc, wd0, hist, "constructive", inrc2DefaultProfile())
	if err != nil {
		t.Fatalf("week 0: %v", err)
	}

	// Update history.
	hist1 := inrc2.UpdateHistory(sc, hist, sol0)
	if hist1.Week != 1 {
		t.Errorf("expected history week 1, got %d", hist1.Week)
	}

	// Solve week 1 using updated history.
	sol1, _, err := inrc2.SolveWeek(sc, wd1, hist1, "constructive", inrc2DefaultProfile())
	if err != nil {
		t.Fatalf("week 1: %v", err)
	}

	if sol1.Week != 1 {
		t.Errorf("expected solution week 1, got %d", sol1.Week)
	}
	if len(sol1.Assignments) == 0 {
		t.Error("expected assignments in week 1")
	}

	// Accumulated assignments should increase.
	hist2 := inrc2.UpdateHistory(sc, hist1, sol1)
	for _, nh := range hist2.NurseHistory {
		if nh.NumberOfAssignments < 0 {
			t.Errorf("negative assignments for %s", nh.Nurse)
		}
	}
}

func inrc2DefaultProfile() optimisation.AlgorithmProfile {
	return optimisation.DefaultProfile()
}
