package inrc2_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func TestResolveMidHorizonWeek(t *testing.T) {
	if got := inrc2.ResolveMidHorizonWeek(8, 4, 0); got != 4 {
		t.Fatalf("explicit week: got %d want 4", got)
	}
	if got := inrc2.ResolveMidHorizonWeek(8, 0, 1.0); got != 4 {
		t.Fatalf("auto week with weight: got %d want 4", got)
	}
	if got := inrc2.ResolveMidHorizonWeek(8, 0, 0); got != 0 {
		t.Fatalf("disabled: got %d want 0", got)
	}
	if got := inrc2.ResolveMidHorizonWeek(8, 8, 1.0); got != 0 {
		t.Fatalf("last week disabled: got %d want 0", got)
	}
	if got := inrc2.ResolveMidHorizonWeek(4, 0, 1.0); got != 2 {
		t.Fatalf("4-week auto: got %d want 2", got)
	}
}

func TestEvaluateMidHorizon_WeekendExhausted(t *testing.T) {
	sc := buildTestScenario()
	hist := inrc2.History{
		Week: 4,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 20, NumberOfWorkingWeekends: 5},
			{Nurse: "Bob", NumberOfAssignments: 20, NumberOfWorkingWeekends: 2},
		},
	}
	exp := inrc2.EvaluateMidHorizon(sc, hist)
	if exp.RemainingWeekendsFeasible {
		t.Fatal("expected remaining weekends infeasible for Alice over max")
	}
	if exp.ProjectedS8Penalty < 30 {
		t.Fatalf("expected S8 penalty >= 30 for 1 weekend over, got %d", exp.ProjectedS8Penalty)
	}
	var alice *inrc2.MidHorizonNurseExposure
	for i := range exp.Nurses {
		if exp.Nurses[i].Nurse == "Alice" {
			alice = &exp.Nurses[i]
			break
		}
	}
	if alice == nil {
		t.Fatal("missing Alice nurse exposure")
	}
	if alice.WeekendsRemainingAllowance != 0 {
		t.Fatalf("Alice weekends remaining: got %d want 0", alice.WeekendsRemainingAllowance)
	}
	if alice.AssignmentsRemainingMinimum != 10 { // 30-20
		t.Fatalf("Alice assignments remaining min: got %d want 10", alice.AssignmentsRemainingMinimum)
	}
}

func TestEvaluateMidHorizon_Healthy(t *testing.T) {
	sc := buildTestScenario()
	hist := inrc2.History{
		Week: 4,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 20, NumberOfWorkingWeekends: 2},
			{Nurse: "Bob", NumberOfAssignments: 18, NumberOfWorkingWeekends: 2},
		},
	}
	exp := inrc2.EvaluateMidHorizon(sc, hist)
	if !exp.RemainingAssignmentsFeasible || !exp.RemainingWeekendsFeasible {
		t.Fatalf("expected feasible remaining capacity, got assign=%v weekends=%v",
			exp.RemainingAssignmentsFeasible, exp.RemainingWeekendsFeasible)
	}
	if exp.ProjectedTotal() > 100 {
		t.Fatalf("expected modest projected exposure for healthy path, got %d", exp.ProjectedTotal())
	}
}

func TestEvaluateMidHorizon_ImpossibleMin(t *testing.T) {
	sc := buildTestScenario()
	hist := inrc2.History{
		Week: 7,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 10, NumberOfWorkingWeekends: 2},
			{Nurse: "Bob", NumberOfAssignments: 28, NumberOfWorkingWeekends: 3},
		},
	}
	exp := inrc2.EvaluateMidHorizon(sc, hist)
	if exp.RemainingAssignmentsFeasible {
		t.Fatal("expected assignments infeasible for Alice")
	}
	if exp.ProjectedS7Penalty < 200 {
		t.Fatalf("expected large S7 penalty, got %d", exp.ProjectedS7Penalty)
	}
}

func TestMidHorizonSelectionBias_OnlyAtCheckpoint(t *testing.T) {
	sc := buildTestScenario()
	hist := inrc2.History{
		Week: 4,
		NurseHistory: []inrc2.NurseHistory{
			{Nurse: "Alice", NumberOfAssignments: 20, NumberOfWorkingWeekends: 5},
			{Nurse: "Bob", NumberOfAssignments: 20, NumberOfWorkingWeekends: 2},
		},
	}
	at := inrc2.MidHorizonSelectionBias(sc, hist, 4, 4, 1.0)
	off := inrc2.MidHorizonSelectionBias(sc, hist, 3, 4, 1.0)
	if at <= 0 {
		t.Fatalf("expected bias at checkpoint, got %d", at)
	}
	if off != 0 {
		t.Fatalf("expected zero bias off checkpoint, got %d", off)
	}
}

func TestWriteMidHorizonCSV(t *testing.T) {
	dir := t.TempDir()
	snaps := []inrc2.MidHorizonPathSnapshot{
		{
			Week: 4, PathID: 7, ParentID: 2, Seed: 42,
			CurrentObjective: 120, ProjectedS7Penalty: 220, ProjectedS8Penalty: 120,
			ProjectedFinalObjective: 460, RemainingAssignmentsFeasible: true,
			RemainingWeekendsFeasible: false, SelectionScore: 460, Retained: true, Winning: true,
			Nurses: []inrc2.MidHorizonNurseExposure{
				{Nurse: "Alice", AssignmentsCompleted: 13, AssignmentsRemainingMinimum: 17,
					AssignmentsRemainingMaximum: 25, WeekendsWorked: 3, WeekendsRemainingAllowance: 1,
					ProjectedS7Penalty: 0, ProjectedS8Penalty: 0,
					RemainingAssignmentsFeasible: true, RemainingWeekendsFeasible: true},
			},
		},
	}
	pathCSV := filepath.Join(dir, "mid_horizon.csv")
	nurseCSV := filepath.Join(dir, "mid_horizon_nurses.csv")
	if err := inrc2.WriteMidHorizonCSV(pathCSV, snaps); err != nil {
		t.Fatalf("WriteMidHorizonCSV: %v", err)
	}
	if err := inrc2.WriteMidHorizonNurseCSV(nurseCSV, snaps); err != nil {
		t.Fatalf("WriteMidHorizonNurseCSV: %v", err)
	}
	raw, err := os.ReadFile(pathCSV)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.Contains(text, "projected_final_objective") {
		t.Fatal("missing header")
	}
	if !strings.Contains(text, ",460,") {
		t.Fatalf("missing projected final objective row:\n%s", text)
	}
	nraw, err := os.ReadFile(nurseCSV)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(nraw), "Alice") {
		t.Fatalf("missing nurse row:\n%s", nraw)
	}
}
