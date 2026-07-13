package inrc2_test

import (
	"fmt"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// TestOfficialRevalidate_MatchesJavaValidator_n005w4 proves Final Official Penalty
// uses ScoreMultiStage and matches the official Java validator total (1695).
func TestOfficialRevalidate_MatchesJavaValidator_n005w4(t *testing.T) {
	sc, err := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	if err != nil {
		t.Fatal(err)
	}
	hist, err := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")
	if err != nil {
		t.Fatal(err)
	}
	weekFiles := []string{
		testDataDir + "WD-n005w4-1.json",
		testDataDir + "WD-n005w4-2.json",
		testDataDir + "WD-n005w4-3.json",
		testDataDir + "WD-n005w4-3.json",
	}
	solFiles := []string{
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-1-0.json",
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-2-1.json",
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-3-2.json",
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-3-3.json",
	}

	var path []inrc2.BeamPath
	for i, f := range solFiles {
		sol, err := inrc2.LoadSolution(f)
		if err != nil {
			t.Fatal(err)
		}
		path = append(path, inrc2.BeamPath{ID: i + 1, Week: i + 1, Solution: sol})
	}

	updated, finalPenalty, _ := inrc2.OfficialRevalidateBeamPath(sc, weekFiles, path, hist)
	const want = 1695
	if finalPenalty != want {
		t.Fatalf("OfficialRevalidateBeamPath=%d want %d (Java validator)", finalPenalty, want)
	}
	if len(updated) != 4 {
		t.Fatalf("expected 4 weeks, got %d", len(updated))
	}
	sumWeeks := 0
	for _, wp := range updated {
		sumWeeks += wp.WeekPenalty
	}
	if sumWeeks != want {
		t.Fatalf("attributed week penalties sum=%d want %d", sumWeeks, want)
	}
	if updated[len(updated)-1].CumulativePenalty != want {
		t.Fatalf("final cumulative=%d want %d", updated[len(updated)-1].CumulativePenalty, want)
	}
}

func TestOfficialRevalidate_MatchesJavaValidator_n012w8(t *testing.T) {
	base := "../../../../../examples/inrc2/testdatasets_json/n012w8/"
	sc, err := inrc2.LoadScenario(base + "Sc-n012w8.json")
	if err != nil {
		t.Fatal(err)
	}
	hist, err := inrc2.LoadHistory(base + "H0-n012w8-0.json")
	if err != nil {
		t.Fatal(err)
	}
	weekSeq := []int{3, 5, 0, 2, 0, 4, 5, 2}
	solFiles := []string{
		"Sol-n012w8-3-0.json",
		"Sol-n012w8-5-1.json",
		"Sol-n012w8-0-2.json",
		"Sol-n012w8-2-3.json",
		"Sol-n012w8-0-4.json",
		"Sol-n012w8-4-5.json",
		"Sol-n012w8-5-6.json",
		"Sol-n012w8-2-7.json",
	}
	solDir := base + "Solution_H_0-WD_3-5-0-2-0-4-5-2/"

	weekFiles := make([]string, len(weekSeq))
	var path []inrc2.BeamPath
	for i, w := range weekSeq {
		weekFiles[i] = fmt.Sprintf("%sWD-n012w8-%d.json", base, w)
		sol, err := inrc2.LoadSolution(solDir + solFiles[i])
		if err != nil {
			t.Fatal(err)
		}
		path = append(path, inrc2.BeamPath{ID: i + 1, Week: i + 1, Solution: sol})
	}

	_, finalPenalty, _ := inrc2.OfficialRevalidateBeamPath(sc, weekFiles, path, hist)
	const want = 3295 // official Java validator.jar total
	if finalPenalty != want {
		t.Fatalf("OfficialRevalidateBeamPath=%d want %d (Java validator)", finalPenalty, want)
	}
}

// TestWeeklyScoreSum_DiffersFromOfficial documents that per-week Score summation
// overcounts consecutive constraints; search may still use it as a heuristic.
func TestWeeklyScoreSum_DiffersFromOfficial(t *testing.T) {
	sc, err := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	if err != nil {
		t.Fatal(err)
	}
	hist, err := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")
	if err != nil {
		t.Fatal(err)
	}
	weekFiles := []string{
		testDataDir + "WD-n005w4-1.json",
		testDataDir + "WD-n005w4-2.json",
		testDataDir + "WD-n005w4-3.json",
		testDataDir + "WD-n005w4-3.json",
	}
	solFiles := []string{
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-1-0.json",
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-2-1.json",
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-3-2.json",
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-3-3.json",
	}
	var weeks []inrc2.WeekData
	var sols []inrc2.Solution
	for _, f := range weekFiles {
		wd, err := inrc2.LoadWeekData(f)
		if err != nil {
			t.Fatal(err)
		}
		weeks = append(weeks, wd)
	}
	for _, f := range solFiles {
		sol, err := inrc2.LoadSolution(f)
		if err != nil {
			t.Fatal(err)
		}
		sols = append(sols, sol)
	}
	ms := inrc2.ScoreMultiStage(sc, weeks, hist, sols)
	h := hist
	weeklySum := 0
	for i := range sols {
		r := inrc2.Score(sc, weeks[i], h, sols[i])
		weeklySum += r.SoftPenalty
		h = inrc2.UpdateHistory(sc, h, sols[i])
	}
	if weeklySum == ms.TotalObjective {
		t.Fatalf("expected weekly Score sum to differ from official MultiStage; both=%d", weeklySum)
	}
	t.Logf("weekly heuristic sum=%d official MultiStage=%d overscore=%d", weeklySum, ms.TotalObjective, weeklySum-ms.TotalObjective)
}

func TestScoreMultiStage_PrefixOmitsS7S8(t *testing.T) {
	sc, err := inrc2.LoadScenario(testDataDir + "Sc-n005w4.json")
	if err != nil {
		t.Fatal(err)
	}
	hist, err := inrc2.LoadHistory(testDataDir + "H0-n005w4-0.json")
	if err != nil {
		t.Fatal(err)
	}
	weeks := make([]inrc2.WeekData, 2)
	sols := make([]inrc2.Solution, 2)
	for i, f := range []string{
		testDataDir + "WD-n005w4-1.json",
		testDataDir + "WD-n005w4-2.json",
	} {
		wd, err := inrc2.LoadWeekData(f)
		if err != nil {
			t.Fatal(err)
		}
		weeks[i] = wd
	}
	for i, f := range []string{
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-1-0.json",
		testDataDir + "Solution_H_0-WD_1-2-3-3/Sol-n005w4-2-1.json",
	} {
		sol, err := inrc2.LoadSolution(f)
		if err != nil {
			t.Fatal(err)
		}
		sols[i] = sol
	}
	ms := inrc2.ScoreMultiStage(sc, weeks, hist, sols)
	for _, d := range ms.SoftDetails {
		if d.Constraint == "S7_TotalAssignments" || d.Constraint == "S8_TotalWorkingWeekends" {
			t.Fatalf("prefix score must not include %s before full horizon", d.Constraint)
		}
	}
}
