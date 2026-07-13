package inrc2

import (
	"fmt"
	"os"
	"strings"
)

// Mid-horizon rebalancing (4+4 style) — Phase 0 telemetry + Phase 1 selection bias.
// Does NOT change official INRC-II scoring. S7 = total assignments, S8 = working weekends.

// MidHorizonNurseExposure is remaining contract capacity for one nurse at a checkpoint.
type MidHorizonNurseExposure struct {
	Nurse                        string
	AssignmentsCompleted         int
	AssignmentsRemainingMinimum  int
	AssignmentsRemainingMaximum  int
	WeekendsWorked               int
	WeekendsRemainingAllowance   int
	ProjectedS7Penalty           int
	ProjectedS8Penalty           int
	RemainingAssignmentsFeasible bool
	RemainingWeekendsFeasible    bool
}

// MidHorizonExposure aggregates projected end-horizon S7/S8 exposure from a history.
type MidHorizonExposure struct {
	ProjectedS7Penalty           int
	ProjectedS8Penalty           int
	RemainingAssignmentsFeasible bool
	RemainingWeekendsFeasible    bool
	Nurses                       []MidHorizonNurseExposure
}

// ProjectedTotal returns S7+S8 projected soft penalty.
func (e MidHorizonExposure) ProjectedTotal() int {
	return e.ProjectedS7Penalty + e.ProjectedS8Penalty
}

// MidHorizonPathSnapshot is one beam path's checkpoint record (telemetry + selection audit).
type MidHorizonPathSnapshot struct {
	Week                         int
	PathID                       int
	ParentID                     int
	Seed                         int64
	CurrentObjective             int
	ProjectedS7Penalty           int
	ProjectedS8Penalty           int
	ProjectedFinalObjective      int
	RemainingAssignmentsFeasible bool
	RemainingWeekendsFeasible    bool
	SelectionScore               int
	Retained                     bool
	Winning                      bool
	Nurses                       []MidHorizonNurseExposure
}

// ResolveMidHorizonWeek returns the 1-indexed checkpoint week, or 0 if disabled.
// If week is unset but weight > 0, defaults to numWeeks/2 (e.g. 4 for an 8-week horizon).
func ResolveMidHorizonWeek(numWeeks, configuredWeek int, weight float64) int {
	if configuredWeek > 0 {
		if configuredWeek >= numWeeks {
			return 0 // last week already scores S7/S8 officially
		}
		return configuredWeek
	}
	if weight <= 0 {
		return 0
	}
	mid := numWeeks / 2
	if mid < 1 || mid >= numWeeks {
		return 0
	}
	return mid
}

// EvaluateMidHorizon computes remaining S7/S8 capacity and projected end penalties.
// Unlike LookaheadPenalty, this is not time-damped — intended for an explicit mid checkpoint.
func EvaluateMidHorizon(sc Scenario, hist History) MidHorizonExposure {
	out := MidHorizonExposure{
		RemainingAssignmentsFeasible: true,
		RemainingWeekendsFeasible:    true,
	}
	currentWeek := hist.Week
	totalWeeks := sc.NumberOfWeeks
	if currentWeek <= 0 || currentWeek >= totalWeeks {
		return out
	}
	remainingWeeks := totalWeeks - currentWeek

	contractMap := make(map[string]Contract, len(sc.Contracts))
	for _, c := range sc.Contracts {
		contractMap[c.ID] = c
	}
	nurseContract := make(map[string]string, len(sc.Nurses))
	for _, n := range sc.Nurses {
		nurseContract[n.ID] = n.Contract
	}

	for _, nh := range hist.NurseHistory {
		contract, ok := contractMap[nurseContract[nh.Nurse]]
		if !ok {
			continue
		}
		ne := MidHorizonNurseExposure{
			Nurse:                        nh.Nurse,
			AssignmentsCompleted:         nh.NumberOfAssignments,
			WeekendsWorked:               nh.NumberOfWorkingWeekends,
			RemainingAssignmentsFeasible: true,
			RemainingWeekendsFeasible:    true,
		}

		// --- S7 remaining capacity ---
		if contract.MinimumNumberOfAssignments > 0 {
			ne.AssignmentsRemainingMinimum = contract.MinimumNumberOfAssignments - nh.NumberOfAssignments
			if ne.AssignmentsRemainingMinimum < 0 {
				ne.AssignmentsRemainingMinimum = 0
			}
		}
		if contract.MaximumNumberOfAssignments > 0 {
			ne.AssignmentsRemainingMaximum = contract.MaximumNumberOfAssignments - nh.NumberOfAssignments
			if ne.AssignmentsRemainingMaximum < 0 {
				// Already over max — remaining max is 0; overage is guaranteed penalty.
				over := nh.NumberOfAssignments - contract.MaximumNumberOfAssignments
				ne.AssignmentsRemainingMaximum = 0
				ne.ProjectedS7Penalty += over * 20
				ne.RemainingAssignmentsFeasible = false
			}
		}

		// Impossible to hit minimum even working every day remaining.
		if contract.MinimumNumberOfAssignments > 0 {
			maxFuture := remainingWeeks * 7
			bestCase := nh.NumberOfAssignments + maxFuture
			if bestCase < contract.MinimumNumberOfAssignments {
				undershoot := contract.MinimumNumberOfAssignments - bestCase
				ne.ProjectedS7Penalty += undershoot * 20
				ne.RemainingAssignmentsFeasible = false
			}
		}

		// Guaranteed max bust even at minimal future workload (~3/week coverage floor).
		if contract.MaximumNumberOfAssignments > 0 && nh.NumberOfAssignments <= contract.MaximumNumberOfAssignments {
			minFuture := remainingWeeks * 3
			guaranteed := nh.NumberOfAssignments + minFuture
			if guaranteed > contract.MaximumNumberOfAssignments {
				ne.ProjectedS7Penalty += (guaranteed - contract.MaximumNumberOfAssignments) * 20
				ne.RemainingAssignmentsFeasible = false
			} else {
				// Soft rate projection (half weight) when trending over but not guaranteed.
				projected := float64(nh.NumberOfAssignments) * float64(totalWeeks) / float64(currentWeek)
				if overshoot := projected - float64(contract.MaximumNumberOfAssignments); overshoot > 0 {
					ne.ProjectedS7Penalty += int(overshoot * 20.0 * 0.5)
				}
			}
		}

		// Soft undershoot trend for min (half weight) when catch-up is still possible but behind pace.
		if contract.MinimumNumberOfAssignments > 0 && ne.RemainingAssignmentsFeasible {
			pace := float64(contract.MinimumNumberOfAssignments) * float64(currentWeek) / float64(totalWeeks)
			if undershoot := pace - float64(nh.NumberOfAssignments); undershoot > 0 {
				ne.ProjectedS7Penalty += int(undershoot * 20.0 * 0.5)
			}
		}

		// --- S8 remaining weekends ---
		if contract.MaximumNumberOfWorkingWeekends > 0 {
			ne.WeekendsRemainingAllowance = contract.MaximumNumberOfWorkingWeekends - nh.NumberOfWorkingWeekends
			if ne.WeekendsRemainingAllowance < 0 {
				over := nh.NumberOfWorkingWeekends - contract.MaximumNumberOfWorkingWeekends
				ne.WeekendsRemainingAllowance = 0
				ne.ProjectedS8Penalty += over * 30
				ne.RemainingWeekendsFeasible = false
			} else {
				projected := float64(nh.NumberOfWorkingWeekends) * float64(totalWeeks) / float64(currentWeek)
				if overshoot := projected - float64(contract.MaximumNumberOfWorkingWeekends); overshoot > 0 {
					ne.ProjectedS8Penalty += int(overshoot * 30.0 * 0.5)
				}
			}
		}

		if !ne.RemainingAssignmentsFeasible {
			out.RemainingAssignmentsFeasible = false
		}
		if !ne.RemainingWeekendsFeasible {
			out.RemainingWeekendsFeasible = false
		}
		out.ProjectedS7Penalty += ne.ProjectedS7Penalty
		out.ProjectedS8Penalty += ne.ProjectedS8Penalty
		out.Nurses = append(out.Nurses, ne)
	}
	return out
}

// MidHorizonSelectionBias returns λ × projected(S7+S8) when this week is the checkpoint.
func MidHorizonSelectionBias(sc Scenario, hist History, week int, midWeek int, weight float64) int {
	if midWeek <= 0 || weight <= 0 || week != midWeek {
		return 0
	}
	exp := EvaluateMidHorizon(sc, hist)
	return int(float64(exp.ProjectedTotal()) * weight)
}

// BuildMidHorizonPathSnapshot builds a telemetry row for one beam path at the checkpoint.
func BuildMidHorizonPathSnapshot(path BeamPath, sc Scenario, weight float64, retained, winning bool) MidHorizonPathSnapshot {
	exp := EvaluateMidHorizon(sc, path.History)
	projectedTotal := exp.ProjectedTotal()
	selection := path.CumulativePenalty + int(float64(projectedTotal)*weight)
	if weight <= 0 {
		selection = path.CumulativePenalty + projectedTotal
	}
	return MidHorizonPathSnapshot{
		Week:                         path.Week,
		PathID:                       path.ID,
		ParentID:                     path.ParentID,
		Seed:                         path.Seed,
		CurrentObjective:             path.CumulativePenalty,
		ProjectedS7Penalty:           exp.ProjectedS7Penalty,
		ProjectedS8Penalty:           exp.ProjectedS8Penalty,
		ProjectedFinalObjective:      path.CumulativePenalty + projectedTotal,
		RemainingAssignmentsFeasible: exp.RemainingAssignmentsFeasible,
		RemainingWeekendsFeasible:    exp.RemainingWeekendsFeasible,
		SelectionScore:               selection,
		Retained:                     retained,
		Winning:                      winning,
		Nurses:                       exp.Nurses,
	}
}

// MidHorizonCSVHeader is the path-level checkpoint CSV header.
func MidHorizonCSVHeader() string {
	return strings.Join([]string{
		"week", "path_id", "parent_id", "seed",
		"current_objective", "projected_s7_penalty", "projected_s8_penalty",
		"projected_final_objective", "remaining_assignments_feasible", "remaining_weekends_feasible",
		"selection_score", "retained", "winning",
	}, ",")
}

// MidHorizonNurseCSVHeader is the nurse-detail checkpoint CSV header.
func MidHorizonNurseCSVHeader() string {
	return strings.Join([]string{
		"week", "path_id", "nurse",
		"assignments_completed", "assignments_remaining_min", "assignments_remaining_max",
		"weekends_worked", "weekends_remaining_allowance",
		"projected_s7_penalty", "projected_s8_penalty",
		"remaining_assignments_feasible", "remaining_weekends_feasible",
	}, ",")
}

func midHorizonPathCSVRow(s MidHorizonPathSnapshot) string {
	af, wf, ret, win := 0, 0, 0, 0
	if s.RemainingAssignmentsFeasible {
		af = 1
	}
	if s.RemainingWeekendsFeasible {
		wf = 1
	}
	if s.Retained {
		ret = 1
	}
	if s.Winning {
		win = 1
	}
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d",
		s.Week, s.PathID, s.ParentID, s.Seed,
		s.CurrentObjective, s.ProjectedS7Penalty, s.ProjectedS8Penalty,
		s.ProjectedFinalObjective, af, wf,
		s.SelectionScore, ret, win)
}

func midHorizonNurseCSVRow(week, pathID int, n MidHorizonNurseExposure) string {
	af, wf := 0, 0
	if n.RemainingAssignmentsFeasible {
		af = 1
	}
	if n.RemainingWeekendsFeasible {
		wf = 1
	}
	return fmt.Sprintf("%d,%d,%s,%d,%d,%d,%d,%d,%d,%d,%d,%d",
		week, pathID, n.Nurse,
		n.AssignmentsCompleted, n.AssignmentsRemainingMinimum, n.AssignmentsRemainingMaximum,
		n.WeekendsWorked, n.WeekendsRemainingAllowance,
		n.ProjectedS7Penalty, n.ProjectedS8Penalty, af, wf)
}

// WriteMidHorizonCSV writes path-level mid-horizon snapshots.
func WriteMidHorizonCSV(path string, snaps []MidHorizonPathSnapshot) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, MidHorizonCSVHeader())
	for _, s := range snaps {
		fmt.Fprintln(f, midHorizonPathCSVRow(s))
	}
	return nil
}

// WriteMidHorizonNurseCSV writes nurse-level detail for checkpoint paths.
func WriteMidHorizonNurseCSV(path string, snaps []MidHorizonPathSnapshot) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintln(f, MidHorizonNurseCSVHeader())
	for _, s := range snaps {
		for _, n := range s.Nurses {
			fmt.Fprintln(f, midHorizonNurseCSVRow(s.Week, s.PathID, n))
		}
	}
	return nil
}
