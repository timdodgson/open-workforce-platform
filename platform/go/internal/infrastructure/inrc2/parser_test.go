package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

const testDataDir = "../../../../../examples/inrc2/testdatasets_json/n005w4/"

func TestLoadScenario(t *testing.T) {
	sc, err := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	if err != nil {
		t.Fatalf("failed to load scenario: %v", err)
	}

	if sc.ID != "n005w4" {
		t.Errorf("expected id n005w4, got %s", sc.ID)
	}
	if sc.NumberOfWeeks != 4 {
		t.Errorf("expected 4 weeks, got %d", sc.NumberOfWeeks)
	}
	if len(sc.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(sc.Skills))
	}
	if len(sc.ShiftTypes) != 3 {
		t.Errorf("expected 3 shift types, got %d", len(sc.ShiftTypes))
	}
	if len(sc.Contracts) != 2 {
		t.Errorf("expected 2 contracts, got %d", len(sc.Contracts))
	}
	if len(sc.Nurses) != 5 {
		t.Errorf("expected 5 nurses, got %d", len(sc.Nurses))
	}

	// Verify forbidden successions.
	if len(sc.ForbiddenShiftTypeSuccessions) != 3 {
		t.Errorf("expected 3 forbidden succession entries, got %d", len(sc.ForbiddenShiftTypeSuccessions))
	}
	// Night -> [Early, Late] is forbidden.
	for _, fs := range sc.ForbiddenShiftTypeSuccessions {
		if fs.PrecedingShiftType == "Night" {
			if len(fs.SucceedingShiftTypes) != 2 {
				t.Errorf("expected 2 forbidden successors for Night, got %d", len(fs.SucceedingShiftTypes))
			}
		}
	}

	// Verify contract details.
	for _, c := range sc.Contracts {
		if c.ID == "FullTime" {
			if c.MinimumNumberOfAssignments != 15 {
				t.Errorf("expected FullTime minAssignments=15, got %d", c.MinimumNumberOfAssignments)
			}
			if c.MaximumNumberOfAssignments != 22 {
				t.Errorf("expected FullTime maxAssignments=22, got %d", c.MaximumNumberOfAssignments)
			}
		}
	}
}

func TestLoadWeekData(t *testing.T) {
	wd, err := inrc2.LoadWeekData(testDataDir + "WD-n005w4-0.json")
	if err != nil {
		t.Fatalf("failed to load week data: %v", err)
	}

	if wd.Scenario != "n005w4" {
		t.Errorf("expected scenario n005w4, got %s", wd.Scenario)
	}
	if len(wd.Requirements) != 6 {
		t.Errorf("expected 6 requirements (3 shifts × 2 skills), got %d", len(wd.Requirements))
	}
	if len(wd.ShiftOffRequests) != 3 {
		t.Errorf("expected 3 shift off requests, got %d", len(wd.ShiftOffRequests))
	}

	// Verify a specific requirement.
	for _, req := range wd.Requirements {
		if req.ShiftType == "Early" && req.Skill == "Nurse" {
			monday := req.RequirementForDay(0)
			if monday.Minimum != 1 || monday.Optimal != 2 {
				t.Errorf("expected Early/Nurse Monday min=1 opt=2, got min=%d opt=%d",
					monday.Minimum, monday.Optimal)
			}
		}
	}
}

func TestLoadHistory(t *testing.T) {
	h, err := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}

	if h.Week != 0 {
		t.Errorf("expected week 0, got %d", h.Week)
	}
	if h.Scenario != "n005w4" {
		t.Errorf("expected scenario n005w4, got %s", h.Scenario)
	}
	if len(h.NurseHistory) != 5 {
		t.Errorf("expected 5 nurse histories, got %d", len(h.NurseHistory))
	}

	// Verify Patrick's history.
	for _, nh := range h.NurseHistory {
		if nh.Nurse == "Patrick" {
			if nh.LastAssignedShiftType != "Night" {
				t.Errorf("expected Patrick last shift Night, got %s", nh.LastAssignedShiftType)
			}
			if nh.NumberOfConsecutiveWorkingDays != 4 {
				t.Errorf("expected Patrick 4 consecutive working days, got %d", nh.NumberOfConsecutiveWorkingDays)
			}
		}
	}
}

func TestLoadSolution(t *testing.T) {
	sol, err := inrc2.LoadSolution(testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-1-0.json")
	if err != nil {
		t.Fatalf("failed to load solution: %v", err)
	}

	if sol.Scenario != "n005w4" {
		t.Errorf("expected scenario n005w4, got %s", sol.Scenario)
	}
	if sol.Week != 0 {
		t.Errorf("expected week 0, got %d", sol.Week)
	}
	if len(sol.Assignments) != 25 {
		t.Errorf("expected 25 assignments, got %d", len(sol.Assignments))
	}

	// Verify first assignment.
	first := sol.Assignments[0]
	if first.Nurse != "Patrick" || first.Day != "Mon" || first.ShiftType != "Night" || first.Skill != "Nurse" {
		t.Errorf("unexpected first assignment: %+v", first)
	}
}

func TestDayIndex(t *testing.T) {
	tests := []struct {
		day    string
		expect int
	}{
		{"Mon", 0}, {"Monday", 0},
		{"Tue", 1}, {"Tuesday", 1},
		{"Wed", 2}, {"Wednesday", 2},
		{"Thu", 3}, {"Thursday", 3},
		{"Fri", 4}, {"Friday", 4},
		{"Sat", 5}, {"Saturday", 5},
		{"Sun", 6}, {"Sunday", 6},
		{"Unknown", -1},
	}
	for _, tt := range tests {
		got := inrc2.DayIndex(tt.day)
		if got != tt.expect {
			t.Errorf("DayIndex(%q) = %d, want %d", tt.day, got, tt.expect)
		}
	}
}
