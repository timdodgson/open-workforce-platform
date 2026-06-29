// Package inrc2 — Parallel Feasible Roster Search (PFRS)
//
// Tim's Algorithm: starts from a hard-feasible roster and optimises
// soft constraints using nurse/day shift swaps. Never accepts hard-invalid moves.
package inrc2

import (
	"fmt"
	"sort"
)

// --- Roster Representation ---

// ShiftAssignment represents what a nurse does on a specific day.
// Empty string means day off.
type ShiftAssignment struct {
	ShiftType string // e.g. "Early", "Late", "Night", "" = off
	Skill     string // e.g. "HeadNurse", "Nurse"
}

// Roster is the INRC-II-specific internal representation.
// roster[nurseIndex][dayIndex] = ShiftAssignment
type Roster struct {
	Assignments [][]ShiftAssignment // [nurse][day]
	NurseIDs    []string            // nurse ID by index
	NumDays     int
}

// NewRoster creates an empty roster for the given nurses and days.
func NewRoster(nurseIDs []string, numDays int) *Roster {
	assignments := make([][]ShiftAssignment, len(nurseIDs))
	for i := range assignments {
		assignments[i] = make([]ShiftAssignment, numDays)
	}
	return &Roster{
		Assignments: assignments,
		NurseIDs:    nurseIDs,
		NumDays:     numDays,
	}
}

// Clone creates a deep copy of the roster.
func (r *Roster) Clone() *Roster {
	clone := &Roster{
		Assignments: make([][]ShiftAssignment, len(r.Assignments)),
		NurseIDs:    r.NurseIDs, // shared (immutable)
		NumDays:     r.NumDays,
	}
	for i := range r.Assignments {
		clone.Assignments[i] = make([]ShiftAssignment, r.NumDays)
		copy(clone.Assignments[i], r.Assignments[i])
	}
	return clone
}

// Get returns the assignment for a nurse on a day.
func (r *Roster) Get(nurseIdx, dayIdx int) ShiftAssignment {
	return r.Assignments[nurseIdx][dayIdx]
}

// Set assigns a shift to a nurse on a day.
func (r *Roster) Set(nurseIdx, dayIdx int, a ShiftAssignment) {
	r.Assignments[nurseIdx][dayIdx] = a
}

// IsOff returns true if the nurse has no shift on the given day.
func (r *Roster) IsOff(nurseIdx, dayIdx int) bool {
	return r.Assignments[nurseIdx][dayIdx].ShiftType == ""
}

// --- Feasible Builder ---

// BuilderDiagnostic describes why a feasible roster could not be built.
type BuilderDiagnostic struct {
	Day       int
	ShiftType string
	Skill     string
	Message   string
}

// rankedSlot represents one coverage demand slot with priority metadata.
type rankedSlot struct {
	day           int
	shiftType     string
	skill         string
	eligibleCount int
	skillRarity   int
}

// BuildFeasibleRoster constructs a hard-feasible initial roster using deterministic
// greedy construction with bounded backtracking.
//
// Returns the roster or an error with diagnostics if no feasible solution can be built.
func BuildFeasibleRoster(sc Scenario, wd WeekData, hist History) (*Roster, error) {
	nurseIDs := make([]string, len(sc.Nurses))
	for i, n := range sc.Nurses {
		nurseIDs[i] = n.ID
	}

	roster := NewRoster(nurseIDs, 7)

	// Build nurse index lookup.
	nurseIdx := make(map[string]int, len(sc.Nurses))
	for i, n := range sc.Nurses {
		nurseIdx[n.ID] = i
	}

	// Build nurse skill lookup.
	nurseSkills := make([]map[string]bool, len(sc.Nurses))
	for i, n := range sc.Nurses {
		skills := make(map[string]bool, len(n.Skills))
		for _, s := range n.Skills {
			skills[s] = true
		}
		nurseSkills[i] = skills
	}

	// Skill rarity: how many nurses have this skill.
	skillCount := make(map[string]int)
	for _, n := range sc.Nurses {
		for _, s := range n.Skills {
			skillCount[s]++
		}
	}

	// Build forbidden succession set.
	forbidden := buildForbiddenSet2(sc)

	// Build history lookup for last shift type.
	lastShift := make([]string, len(sc.Nurses))
	for i, n := range sc.Nurses {
		for _, nh := range hist.NurseHistory {
			if nh.Nurse == n.ID {
				if nh.NumberOfConsecutiveDaysOff == 0 && nh.LastAssignedShiftType != "None" && nh.LastAssignedShiftType != "" {
					lastShift[i] = nh.LastAssignedShiftType
				}
				break
			}
		}
	}

	// Collect all coverage slots for the week.
	var allSlots []rankedSlot
	for _, req := range wd.Requirements {
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			dayReq := req.RequirementForDay(dayIdx)
			for i := 0; i < dayReq.Minimum; i++ {
				eligible := 0
				for ni := range sc.Nurses {
					if nurseSkills[ni][req.Skill] {
						eligible++
					}
				}
				allSlots = append(allSlots, rankedSlot{
					day:           dayIdx,
					shiftType:     req.ShiftType,
					skill:         req.Skill,
					eligibleCount: eligible,
					skillRarity:   skillCount[req.Skill],
				})
			}
		}
	}

	// Sort: rarest skill first, then fewest eligible nurses, then by day/shift for determinism.
	sort.SliceStable(allSlots, func(i, j int) bool {
		if allSlots[i].skillRarity != allSlots[j].skillRarity {
			return allSlots[i].skillRarity < allSlots[j].skillRarity
		}
		if allSlots[i].eligibleCount != allSlots[j].eligibleCount {
			return allSlots[i].eligibleCount < allSlots[j].eligibleCount
		}
		if allSlots[i].day != allSlots[j].day {
			return allSlots[i].day < allSlots[j].day
		}
		if allSlots[i].shiftType != allSlots[j].shiftType {
			return allSlots[i].shiftType < allSlots[j].shiftType
		}
		return allSlots[i].skill < allSlots[j].skill
	})

	// Group slots by day for processing.
	dayMap := make(map[int][]rankedSlot)
	for _, rs := range allSlots {
		dayMap[rs.day] = append(dayMap[rs.day], rs)
	}

	// Process day by day (0-6).
	for dayIdx := 0; dayIdx < 7; dayIdx++ {
		slots := dayMap[dayIdx]
		if len(slots) == 0 {
			continue
		}

		// Sort this day's slots by difficulty.
		sort.SliceStable(slots, func(i, j int) bool {
			if slots[i].skillRarity != slots[j].skillRarity {
				return slots[i].skillRarity < slots[j].skillRarity
			}
			return slots[i].eligibleCount < slots[j].eligibleCount
		})

		// Try greedy assignment for this day.
		ok := assignDay(roster, dayIdx, slots, sc, nurseSkills, forbidden, lastShift, hist, 0)
		if !ok {
			// Bounded backtracking: try different nurse orderings (up to 100 attempts).
			found := false
			for attempt := 1; attempt <= 100; attempt++ {
				// Clear this day's assignments.
				for ni := range roster.Assignments {
					roster.Assignments[ni][dayIdx] = ShiftAssignment{}
				}
				ok = assignDay(roster, dayIdx, slots, sc, nurseSkills, forbidden, lastShift, hist, attempt)
				if ok {
					found = true
					break
				}
			}
			if !found {
				// Report diagnostic.
				return nil, fmt.Errorf("no hard-feasible initial roster found: day=%s slots=%d unmet demand",
					DayName(dayIdx), len(slots))
			}
		}

		// Update lastShift for next day's succession check.
		for ni := range sc.Nurses {
			a := roster.Get(ni, dayIdx)
			if a.ShiftType != "" {
				lastShift[ni] = a.ShiftType
			}
		}
	}

	return roster, nil
}

// assignDay attempts to assign all coverage slots for a single day.
// attempt > 0 rotates the nurse preference order for backtracking diversity.
func assignDay(roster *Roster, dayIdx int, slots []rankedSlot, sc Scenario,
	nurseSkills []map[string]bool, forbidden map[string]bool, lastShift []string,
	hist History, attempt int) bool {

	numNurses := len(sc.Nurses)

	for _, slot := range slots {
		assigned := false

		// Build candidate list: nurses eligible for this slot.
		type candidate struct {
			nurseIdx       int
			assignCount    int
			hasRequiredDay bool
		}

		var candidates []candidate
		for ni := 0; ni < numNurses; ni++ {
			// Must have required skill.
			if !nurseSkills[ni][slot.skill] {
				continue
			}
			// Must not already be assigned this day.
			if !roster.IsOff(ni, dayIdx) {
				continue
			}
			// Must not violate forbidden succession from previous day.
			prevShift := ""
			if dayIdx > 0 {
				prevShift = roster.Get(ni, dayIdx-1).ShiftType
			} else {
				prevShift = lastShift[ni]
			}
			if prevShift != "" && forbidden[prevShift+"|"+slot.shiftType] {
				continue
			}

			// Count current assignments for tie-breaking.
			count := 0
			for d := 0; d < 7; d++ {
				if !roster.IsOff(ni, d) {
					count++
				}
			}

			candidates = append(candidates, candidate{
				nurseIdx:    ni,
				assignCount: count,
			})
		}

		if len(candidates) == 0 {
			return false // No eligible nurse — day cannot be completed.
		}

		// Sort candidates: lowest assignment count first, then by nurse index for determinism.
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].assignCount != candidates[j].assignCount {
				return candidates[i].assignCount < candidates[j].assignCount
			}
			return candidates[i].nurseIdx < candidates[j].nurseIdx
		})

		// Rotate selection by attempt for backtracking diversity.
		selectIdx := attempt % len(candidates)
		chosen := candidates[selectIdx]

		roster.Set(chosen.nurseIdx, dayIdx, ShiftAssignment{
			ShiftType: slot.shiftType,
			Skill:     slot.skill,
		})
		assigned = true

		if !assigned {
			return false
		}
	}

	// Verify: check forbidden succession into next day won't be immediately broken.
	// (Full validation happens when processing next day.)
	return true
}

// buildForbiddenSet2 creates the forbidden succession lookup for PFRS.
func buildForbiddenSet2(sc Scenario) map[string]bool {
	set := make(map[string]bool)
	for _, fs := range sc.ForbiddenShiftTypeSuccessions {
		for _, succ := range fs.SucceedingShiftTypes {
			set[fs.PrecedingShiftType+"|"+succ] = true
		}
	}
	return set
}

// --- Roster to Solution Conversion ---

// RosterToSolution converts a Roster back to an official INRC-II Solution.
func RosterToSolution(roster *Roster, sc Scenario, week int) Solution {
	var assignments []Assignment
	for ni, nurseID := range roster.NurseIDs {
		for d := 0; d < roster.NumDays; d++ {
			a := roster.Get(ni, d)
			if a.ShiftType != "" {
				assignments = append(assignments, Assignment{
					Nurse:     nurseID,
					Day:       DayName(d),
					ShiftType: a.ShiftType,
					Skill:     a.Skill,
				})
			}
		}
	}
	return Solution{
		Scenario:    sc.ID,
		Week:        week,
		Assignments: assignments,
	}
}
