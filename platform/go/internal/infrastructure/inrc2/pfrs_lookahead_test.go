package inrc2_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func buildTestScenario() inrc2.Scenario {
	return inrc2.Scenario{
		ID:            "test",
		NumberOfWeeks: 8,
		Contracts: []inrc2.Contract{
			{
				ID:                             "FullTime",
				MinimumNumberOfAssignments:     30,
				MaximumNumberOfAssignments:     40,
				MaximumNumberOfWorkingWeekends: 4,
			},
		},
		Nurses: []inrc2.Nurse{
			{ID: "Alice", Contract: "FullTime"},
			{ID: "Bob", Contract: "FullTime"},
		},
	}
}

func TestLookaheadPenalty_ZeroWeight_ReturnsZero(t *testing.T) {
	sc := buildTestScenario()
	hist := inrc2.History{
		Week: 4,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 30, NumberOfWorkingWeekends: 4},
		},
	}
	penalty := inrc2.LookaheadPenalty(sc, hist, 0.0)
	if penalty != 0 {
		t.Errorf("expected 0 with zero weight, got %d", penalty)
	}
}

func TestLookaheadPenalty_AtHorizon_ReturnsZero(t *testing.T) {
	sc := buildTestScenario()
	hist := inrc2.History{
		Week: 8, // already at horizon
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 50, NumberOfWorkingWeekends: 6},
		},
	}
	penalty := inrc2.LookaheadPenalty(sc, hist, 1.0)
	if penalty != 0 {
		t.Errorf("expected 0 at horizon (official scorer handles it), got %d", penalty)
	}
}

func TestLookaheadPenalty_NurseAlreadyOverWeekendMax(t *testing.T) {
	sc := buildTestScenario()
	// After week 4, nurse has already worked 5 weekends (max is 4). Guaranteed penalty.
	hist := inrc2.History{
		Week: 4,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 20, NumberOfWorkingWeekends: 5},
			{Nurse: "Bob", NumberOfAssignments: 20, NumberOfWorkingWeekends: 3},
		},
	}
	penalty := inrc2.LookaheadPenalty(sc, hist, 1.0)
	// Alice is 1 over max weekends. Penalty: 1 * 30 * effectiveWeight.
	// effectiveWeight = 1.0 * (4/8) = 0.5.
	// So: 30 * 0.5 = 15.
	if penalty < 10 {
		t.Errorf("expected meaningful penalty for nurse already over weekend max, got %d", penalty)
	}
}

func TestLookaheadPenalty_HealthyTrajectory_LowPenalty(t *testing.T) {
	sc := buildTestScenario()
	// After week 4, both nurses are on perfect track.
	// Alice: 20 assignments (projects to 40 = max, fine). 2 weekends (projects to 4 = max, fine).
	hist := inrc2.History{
		Week: 4,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 20, NumberOfWorkingWeekends: 2},
			{Nurse: "Bob", NumberOfAssignments: 18, NumberOfWorkingWeekends: 2},
		},
	}
	penalty := inrc2.LookaheadPenalty(sc, hist, 1.0)
	// Both nurses project to exactly their limits or below. Should be 0 or very low.
	if penalty > 50 {
		t.Errorf("expected low/zero penalty for healthy trajectory, got %d", penalty)
	}
}

func TestLookaheadPenalty_TimeScaling_LaterWeeksStronger(t *testing.T) {
	sc := buildTestScenario()
	// Same overshoot at week 2 vs week 6 — week 6 should produce higher penalty.
	histEarly := inrc2.History{
		Week: 2,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 12, NumberOfWorkingWeekends: 2},
			{Nurse: "Bob", NumberOfAssignments: 12, NumberOfWorkingWeekends: 2},
		},
	}
	histLate := inrc2.History{
		Week: 6,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 36, NumberOfWorkingWeekends: 5},
			{Nurse: "Bob", NumberOfAssignments: 36, NumberOfWorkingWeekends: 5},
		},
	}
	penaltyEarly := inrc2.LookaheadPenalty(sc, histEarly, 1.0)
	penaltyLate := inrc2.LookaheadPenalty(sc, histLate, 1.0)

	// Late penalty should be significantly higher (higher time factor + more overshoot).
	if penaltyLate <= penaltyEarly {
		t.Errorf("expected late-week penalty (%d) > early-week penalty (%d)", penaltyLate, penaltyEarly)
	}
}

func TestLookaheadPenalty_MinAssignment_OnlyIfImpossible(t *testing.T) {
	sc := buildTestScenario()
	// After week 7, nurse has only 10 assignments. Min is 30.
	// Remaining: 1 week * 7 max = 7. Best case total = 17. Still under 30.
	// Guaranteed undershoot: 30 - 17 = 13.
	hist := inrc2.History{
		Week: 7,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 10, NumberOfWorkingWeekends: 2},
			{Nurse: "Bob", NumberOfAssignments: 28, NumberOfWorkingWeekends: 3},
		},
	}
	penalty := inrc2.LookaheadPenalty(sc, hist, 1.0)
	// Alice is guaranteed to undershoot by 13. Bob is fine.
	// Penalty: 13 * 20 * effectiveWeight(7/8=0.875) = 13 * 20 * 0.875 = 227.5 → 227.
	if penalty < 100 {
		t.Errorf("expected significant penalty for guaranteed min undershoot, got %d", penalty)
	}
}

func TestLookaheadPenalty_AssignmentOvershoot_GuaranteedVsProjected(t *testing.T) {
	sc := buildTestScenario()
	// After week 5, nurse has 30 assignments. Max is 40.
	// Projected: 30 * 8/5 = 48. Overshoot: 8.
	// Guaranteed min future: 3 remaining weeks * 3 = 9. Guaranteed total: 39. Under max.
	// So this is projected-only (not guaranteed) — should get half-weight.
	hist := inrc2.History{
		Week: 5,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 30, NumberOfWorkingWeekends: 3},
			{Nurse: "Bob", NumberOfAssignments: 25, NumberOfWorkingWeekends: 2},
		},
	}
	penalty := inrc2.LookaheadPenalty(sc, hist, 1.0)
	// Alice projects over but isn't guaranteed. Should see moderate penalty.
	// Bob is fine. Should be relatively low total.
	if penalty <= 0 {
		t.Errorf("expected some penalty for projected overshoot, got %d", penalty)
	}
}
