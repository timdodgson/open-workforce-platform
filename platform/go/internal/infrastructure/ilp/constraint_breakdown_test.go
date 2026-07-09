package ilp

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestBuildConstraintBreakdown(t *testing.T) {
	perWeek := []inrc2.ScoreResult{
		{
			SoftPenalty: 100,
			SoftDetails: []inrc2.SoftPenaltyDetail{
				{Constraint: "S1_OptimalCoverage", Penalty: 60},
				{Constraint: "S6_CompleteWeekend", Penalty: 40},
			},
		},
		{
			SoftPenalty: 50,
			SoftDetails: []inrc2.SoftPenaltyDetail{
				{Constraint: "S1_OptimalCoverage", Penalty: 50},
			},
		},
	}

	out := BuildConstraintBreakdown(perWeek)
	if out.TotalPenalty != 150 {
		t.Fatalf("totalPenalty=%d want 150", out.TotalPenalty)
	}
	if out.NumWeeks != 2 {
		t.Fatalf("numWeeks=%d want 2", out.NumWeeks)
	}
	s1 := out.Constraints[0]
	if s1.ID != "S1" || s1.Penalty != 110 || s1.Violations != 2 {
		t.Fatalf("S1: %+v", s1)
	}
	s6 := out.Constraints[5]
	if s6.ID != "S6" || s6.Penalty != 40 || s6.Violations != 1 {
		t.Fatalf("S6: %+v", s6)
	}
}
