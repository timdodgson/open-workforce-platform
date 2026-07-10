package inrc2

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func TestSolveSingleWeek_constructive(t *testing.T) {
	bundle, err := LoadInstanceBundle("n005w4")
	if err != nil {
		t.Skipf("instance not available: %v", err)
	}
	if len(bundle.WeekFiles) == 0 {
		t.Skip("no week files")
	}

	wd, err := LoadWeekData(bundle.WeekFiles[0])
	if err != nil {
		t.Fatalf("load week: %v", err)
	}

	profile, _ := optimisation.GetProfile("research")
	out, err := SolveSingleWeek(bundle.Scenario, wd, bundle.History, WeekSolveParams{
		Algorithm:  "constructive",
		AlgProfile: profile,
	})
	if err != nil {
		t.Fatalf("SolveSingleWeek: %v", err)
	}
	if len(out.Solution.Assignments) == 0 {
		t.Error("expected assignments")
	}
	if out.PFRSStats != nil {
		t.Error("constructive path should not set PFRS stats")
	}
}
