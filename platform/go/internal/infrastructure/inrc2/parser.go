// Package inrc2 provides parsing, writing, and scoring for official
// INRC-II competition files in JSON format.
//
// This package is the single entry point for all INRC-II competition
// compliance logic. It does NOT embed nurse rostering rules into
// optimisation algorithms.
package inrc2

import (
	"encoding/json"
	"fmt"
	"os"
)

// --- Scenario ---

// Scenario represents an official INRC-II scenario file (Sc-*.json).
type Scenario struct {
	ID                          string                   `json:"id"`
	NumberOfWeeks               int                      `json:"numberOfWeeks"`
	Skills                      []string                 `json:"skills"`
	ShiftTypes                  []ShiftType              `json:"shiftTypes"`
	ForbiddenShiftTypeSuccessions []ForbiddenSuccession  `json:"forbiddenShiftTypeSuccessions"`
	Contracts                   []Contract               `json:"contracts"`
	Nurses                      []Nurse                  `json:"nurses"`
}

// ShiftType defines a shift type with consecutive assignment limits.
type ShiftType struct {
	ID                                 string `json:"id"`
	MinimumNumberOfConsecutiveAssignments int    `json:"minimumNumberOfConsecutiveAssignments"`
	MaximumNumberOfConsecutiveAssignments int    `json:"maximumNumberOfConsecutiveAssignments"`
}

// ForbiddenSuccession defines which shift types may not follow a preceding type.
type ForbiddenSuccession struct {
	PrecedingShiftType  string   `json:"precedingShiftType"`
	SucceedingShiftTypes []string `json:"succeedingShiftTypes"`
}

// Contract defines nurse employment constraints.
type Contract struct {
	ID                                string `json:"id"`
	MinimumNumberOfAssignments        int    `json:"minimumNumberOfAssignments"`
	MaximumNumberOfAssignments        int    `json:"maximumNumberOfAssignments"`
	MinimumNumberOfConsecutiveWorkingDays int `json:"minimumNumberOfConsecutiveWorkingDays"`
	MaximumNumberOfConsecutiveWorkingDays int `json:"maximumNumberOfConsecutiveWorkingDays"`
	MinimumNumberOfConsecutiveDaysOff int    `json:"minimumNumberOfConsecutiveDaysOff"`
	MaximumNumberOfConsecutiveDaysOff int    `json:"maximumNumberOfConsecutiveDaysOff"`
	MaximumNumberOfWorkingWeekends    int    `json:"maximumNumberOfWorkingWeekends"`
	CompleteWeekends                  int    `json:"completeWeekends"`
}

// Nurse defines a nurse in the scenario.
type Nurse struct {
	ID       string   `json:"id"`
	Contract string   `json:"contract"`
	Skills   []string `json:"skills"`
}

// --- Week Data ---

// WeekData represents an official INRC-II week data file (WD-*.json).
type WeekData struct {
	Scenario         string          `json:"scenario"`
	Requirements     []Requirement   `json:"requirements"`
	ShiftOffRequests []ShiftRequest  `json:"shiftOffRequests"`
}

// Requirement defines staffing needs for a shift/skill combination across a week.
type Requirement struct {
	ShiftType              string      `json:"shiftType"`
	Skill                  string      `json:"skill"`
	RequirementOnMonday    DayRequirement `json:"requirementOnMonday"`
	RequirementOnTuesday   DayRequirement `json:"requirementOnTuesday"`
	RequirementOnWednesday DayRequirement `json:"requirementOnWednesday"`
	RequirementOnThursday  DayRequirement `json:"requirementOnThursday"`
	RequirementOnFriday    DayRequirement `json:"requirementOnFriday"`
	RequirementOnSaturday  DayRequirement `json:"requirementOnSaturday"`
	RequirementOnSunday    DayRequirement `json:"requirementOnSunday"`
}

// DayRequirement defines minimum and optimal staffing for a single day.
type DayRequirement struct {
	Minimum int `json:"minimum"`
	Optimal int `json:"optimal"`
}

// ShiftRequest represents a shift-off (or shift-on) request.
type ShiftRequest struct {
	Nurse     string `json:"nurse"`
	ShiftType string `json:"shiftType"`
	Day       string `json:"day"`
}

// --- History ---

// History represents an official INRC-II history file (H0-*.json).
type History struct {
	Week         int            `json:"week"`
	Scenario     string         `json:"scenario"`
	NurseHistory []NurseHistory `json:"nurseHistory"`
}

// NurseHistory represents historical state for a single nurse.
type NurseHistory struct {
	Nurse                          string `json:"nurse"`
	NumberOfAssignments            int    `json:"numberOfAssignments"`
	NumberOfWorkingWeekends        int    `json:"numberOfWorkingWeekends"`
	LastAssignedShiftType          string `json:"lastAssignedShiftType"`
	NumberOfConsecutiveAssignments int    `json:"numberOfConsecutiveAssignments"`
	NumberOfConsecutiveWorkingDays int    `json:"numberOfConsecutiveWorkingDays"`
	NumberOfConsecutiveDaysOff     int    `json:"numberOfConsecutiveDaysOff"`
}

// --- Solution ---

// Solution represents an official INRC-II solution file (Sol-*.json).
type Solution struct {
	Scenario    string       `json:"scenario"`
	Week        int          `json:"week"`
	Assignments []Assignment `json:"assignments"`
}

// Assignment represents a single nurse-day-shift-skill assignment.
type Assignment struct {
	Nurse     string `json:"nurse"`
	Day       string `json:"day"`
	ShiftType string `json:"shiftType"`
	Skill     string `json:"skill"`
}

// --- Parsing Functions ---

// LoadScenario reads an official INRC-II scenario JSON file.
func LoadScenario(path string) (Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, fmt.Errorf("failed to read scenario file: %w", err)
	}
	var s Scenario
	if err := json.Unmarshal(data, &s); err != nil {
		return Scenario{}, fmt.Errorf("failed to parse scenario file: %w", err)
	}
	return s, nil
}

// LoadWeekData reads an official INRC-II week data JSON file.
func LoadWeekData(path string) (WeekData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return WeekData{}, fmt.Errorf("failed to read week data file: %w", err)
	}
	var wd WeekData
	if err := json.Unmarshal(data, &wd); err != nil {
		return WeekData{}, fmt.Errorf("failed to parse week data file: %w", err)
	}
	return wd, nil
}

// LoadHistory reads an official INRC-II history JSON file.
func LoadHistory(path string) (History, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return History{}, fmt.Errorf("failed to read history file: %w", err)
	}
	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		return History{}, fmt.Errorf("failed to parse history file: %w", err)
	}
	return h, nil
}

// LoadSolution reads an official INRC-II solution JSON file.
func LoadSolution(path string) (Solution, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Solution{}, fmt.Errorf("failed to read solution file: %w", err)
	}
	var sol Solution
	if err := json.Unmarshal(data, &sol); err != nil {
		return Solution{}, fmt.Errorf("failed to parse solution file: %w", err)
	}
	return sol, nil
}

// --- Helper: Day Indexing ---

// DayIndex converts a day name to a 0-based index (Mon=0, Sun=6).
func DayIndex(day string) int {
	switch day {
	case "Mon", "Monday":
		return 0
	case "Tue", "Tuesday":
		return 1
	case "Wed", "Wednesday":
		return 2
	case "Thu", "Thursday":
		return 3
	case "Fri", "Friday":
		return 4
	case "Sat", "Saturday":
		return 5
	case "Sun", "Sunday":
		return 6
	default:
		return -1
	}
}

// DayName converts a 0-based index to the short day name used in solutions.
func DayName(index int) string {
	switch index {
	case 0:
		return "Mon"
	case 1:
		return "Tue"
	case 2:
		return "Wed"
	case 3:
		return "Thu"
	case 4:
		return "Fri"
	case 5:
		return "Sat"
	case 6:
		return "Sun"
	default:
		return ""
	}
}

// RequirementForDay returns the day requirement for a given day index (0=Mon, 6=Sun).
func (r *Requirement) RequirementForDay(dayIndex int) DayRequirement {
	switch dayIndex {
	case 0:
		return r.RequirementOnMonday
	case 1:
		return r.RequirementOnTuesday
	case 2:
		return r.RequirementOnWednesday
	case 3:
		return r.RequirementOnThursday
	case 4:
		return r.RequirementOnFriday
	case 5:
		return r.RequirementOnSaturday
	case 6:
		return r.RequirementOnSunday
	default:
		return DayRequirement{}
	}
}
