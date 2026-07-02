package inrc2

import (
	"fmt"
	"os"
	"strings"
)

// --- Roster CSV Export ---
// Exports the final winning roster assignments for dashboard visualisation.
// Pure output — does not alter algorithm behaviour.

// RosterCSVHeader returns the header for roster output.
func RosterCSVHeader() string {
	return "week,nurse,day,day_index,shift_type,skill"
}

// WriteRosterCSV writes the winning path's weekly solutions to a CSV file.
func WriteRosterCSV(path string, sc Scenario, winningPath []BeamPath) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, RosterCSVHeader())

	for _, wp := range winningPath {
		week := wp.Week
		for _, a := range wp.Solution.Assignments {
			dayIdx := DayIndex(a.Day)
			line := fmt.Sprintf("%d,%s,%s,%d,%s,%s",
				week, a.Nurse, a.Day, dayIdx, a.ShiftType, a.Skill)
			fmt.Fprintln(f, line)
		}
	}
	return nil
}

// WriteRosterJSON writes the winning roster as a JSON array for easy dashboard consumption.
func WriteRosterJSON(path string, sc Scenario, winningPath []BeamPath) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "[")
	first := true
	for _, wp := range winningPath {
		week := wp.Week
		for _, a := range wp.Solution.Assignments {
			if !first {
				fmt.Fprintln(f, ",")
			}
			// Find nurse contract and skills.
			contract := ""
			var skills []string
			for _, n := range sc.Nurses {
				if n.ID == a.Nurse {
					contract = n.Contract
					skills = n.Skills
					break
				}
			}
			fmt.Fprintf(f, `  {"week":%d,"nurse":%q,"day":%q,"dayIndex":%d,"shiftType":%q,"skill":%q,"contract":%q,"nurseSkills":[%s]}`,
				week, a.Nurse, a.Day, DayIndex(a.Day), a.ShiftType, a.Skill, contract,
				func() string {
					parts := make([]string, len(skills))
					for i, s := range skills {
						parts[i] = fmt.Sprintf("%q", s)
					}
					return strings.Join(parts, ",")
				}())
			first = false
		}
	}
	fmt.Fprintln(f, "\n]")
	return nil
}
