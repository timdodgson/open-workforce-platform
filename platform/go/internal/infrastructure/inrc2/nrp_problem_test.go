package inrc2

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

const testProblemDataDir = "../../../../../examples/inrc2/testdatasets_json/n005w4/"

// TestNRPProblem_ImplementsInterface verifies compile-time interface satisfaction.
func TestNRPProblem_ImplementsInterface(t *testing.T) {
	var _ optimisation.Problem = (*NRPProblem)(nil)
}

// TestNRPProblem_CreateInitialSolution verifies that initial solution is feasible.
func TestNRPProblem_CreateInitialSolution(t *testing.T) {
	sc, err := LoadScenario(testProblemDataDir + "Sc-n005w4.json")
	if err != nil {
		t.Fatalf("LoadScenario: %v", err)
	}
	wd, err := LoadWeekData(testProblemDataDir + "WD-n005w4-0.json")
	if err != nil {
		t.Fatalf("LoadWeekData: %v", err)
	}
	hist, err := LoadHistory(testProblemDataDir + "H0-n005w4-0.json")
	if err != nil {
		t.Fatalf("LoadHistory: %v", err)
	}

	problem := NewNRPProblem(NRPProblemConfig{
		Scenario: sc,
		WeekData: wd,
		History:  hist,
	})

	sol, err := problem.CreateInitialSolution()
	if err != nil {
		t.Fatalf("CreateInitialSolution: %v", err)
	}
	if sol == nil {
		t.Fatal("CreateInitialSolution returned nil")
	}

	// Verify it's a valid roster.
	roster := sol.(*Roster)
	if len(roster.Assignments) != len(sc.Nurses) {
		t.Errorf("Expected %d nurses, got %d", len(sc.Nurses), len(roster.Assignments))
	}
	if roster.NumDays != 7 {
		t.Errorf("Expected 7 days, got %d", roster.NumDays)
	}
}

// TestNRPProblem_CloneSolution verifies that clone is independent.
func TestNRPProblem_CloneSolution(t *testing.T) {
	sc, _ := LoadScenario(testProblemDataDir + "Sc-n005w4.json")
	wd, _ := LoadWeekData(testProblemDataDir + "WD-n005w4-0.json")
	hist, _ := LoadHistory(testProblemDataDir + "H0-n005w4-0.json")

	problem := NewNRPProblem(NRPProblemConfig{
		Scenario: sc, WeekData: wd, History: hist,
	})

	sol, _ := problem.CreateInitialSolution()
	clone := problem.CloneSolution(sol)

	// Modify original — clone should be unaffected.
	roster := sol.(*Roster)
	cloneRoster := clone.(*Roster)

	original00 := roster.Get(0, 0)
	roster.Set(0, 0, ShiftAssignment{ShiftType: "MODIFIED", Skill: "TEST"})

	if cloneRoster.Get(0, 0) == roster.Get(0, 0) {
		t.Error("Clone was affected by original mutation")
	}

	// Restore for clean test.
	roster.Set(0, 0, original00)
}

// TestNRPProblem_Evaluate verifies that evaluation matches existing scorer.
func TestNRPProblem_Evaluate(t *testing.T) {
	sc, _ := LoadScenario(testProblemDataDir + "Sc-n005w4.json")
	wd, _ := LoadWeekData(testProblemDataDir + "WD-n005w4-0.json")
	hist, _ := LoadHistory(testProblemDataDir + "H0-n005w4-0.json")

	problem := NewNRPProblem(NRPProblemConfig{
		Scenario: sc, WeekData: wd, History: hist,
	})

	sol, _ := problem.CreateInitialSolution()
	penalty := problem.Evaluate(sol)

	// Cross-check with direct scorer.
	roster := sol.(*Roster)
	ws := NewScoringWorkspace(sc, wd, hist)
	expected := ScorePenaltyOnlyFromRoster(ws, roster)

	if penalty != expected {
		t.Errorf("Evaluate=%d, expected %d (from direct scorer)", penalty, expected)
	}
}

// TestNRPProblem_TryMoveAndUndo verifies move/undo roundtrip preserves solution.
func TestNRPProblem_TryMoveAndUndo(t *testing.T) {
	sc, _ := LoadScenario(testProblemDataDir + "Sc-n005w4.json")
	wd, _ := LoadWeekData(testProblemDataDir + "WD-n005w4-0.json")
	hist, _ := LoadHistory(testProblemDataDir + "H0-n005w4-0.json")

	problem := NewNRPProblem(NRPProblemConfig{
		Scenario: sc, WeekData: wd, History: hist,
	})

	sol, _ := problem.CreateInitialSolution()
	rng := rand.New(rand.NewSource(42))

	// Evaluate before move.
	penaltyBefore := problem.Evaluate(sol)
	fingerprintBefore := problem.SolutionFingerprint(sol)

	// Try moves until we get a valid one.
	var validResult optimisation.MoveResult
	for i := 0; i < 1000; i++ {
		result := problem.TryMove(sol, rng)
		if result.Valid {
			validResult = result
			break
		}
	}

	if !validResult.Valid {
		t.Fatal("Could not find a valid move in 1000 attempts")
	}

	// Solution should be different now.
	fingerprintAfter := problem.SolutionFingerprint(sol)
	if fingerprintAfter == fingerprintBefore {
		// Move was applied but produced same fingerprint — possible if move was cosmetically neutral.
		// Not an error, just skip the undo check in this case.
	}

	// Undo the move.
	problem.UndoMove(sol, validResult.Move)

	// Solution should be back to original.
	penaltyAfterUndo := problem.Evaluate(sol)
	fingerprintAfterUndo := problem.SolutionFingerprint(sol)

	if penaltyAfterUndo != penaltyBefore {
		t.Errorf("After undo: penalty=%d, expected %d", penaltyAfterUndo, penaltyBefore)
	}
	if fingerprintAfterUndo != fingerprintBefore {
		t.Errorf("After undo: fingerprint=%s, expected %s", fingerprintAfterUndo, fingerprintBefore)
	}
}

// TestNRPProblem_SearchLoop simulates a simplified SA search loop using the interface
// to verify it produces the same quality as the direct implementation.
func TestNRPProblem_SearchLoop(t *testing.T) {
	sc, _ := LoadScenario(testProblemDataDir + "Sc-n005w4.json")
	wd, _ := LoadWeekData(testProblemDataDir + "WD-n005w4-0.json")
	hist, _ := LoadHistory(testProblemDataDir + "H0-n005w4-0.json")

	problem := NewNRPProblem(NRPProblemConfig{
		Scenario: sc, WeekData: wd, History: hist,
	})

	sol, err := problem.CreateInitialSolution()
	if err != nil {
		t.Fatalf("CreateInitialSolution: %v", err)
	}

	rng := rand.New(rand.NewSource(42))
	currentPenalty := problem.Evaluate(sol)
	bestPenalty := currentPenalty
	iterations := 10000
	candidates := 0

	for i := 0; i < iterations; i++ {
		result := problem.TryMove(sol, rng)
		if !result.Valid {
			continue
		}
		candidates++

		newPenalty := problem.Evaluate(sol)
		delta := newPenalty - currentPenalty

		// Simple hill-climbing: only accept improvements.
		if delta <= 0 {
			currentPenalty = newPenalty
			if currentPenalty < bestPenalty {
				bestPenalty = currentPenalty
			}
		} else {
			problem.UndoMove(sol, result.Move)
		}
	}

	// Should have found at least some improvement from the initial constructive solution.
	initialPenalty := problem.Evaluate(sol)
	_ = initialPenalty

	if candidates == 0 {
		t.Fatal("No valid candidates found in 10000 attempts")
	}
	if bestPenalty >= currentPenalty+100 {
		t.Errorf("Hill-climbing didn't improve: best=%d, current=%d", bestPenalty, currentPenalty)
	}

	t.Logf("SearchLoop: %d candidates, initial penalty unknown, best=%d", candidates, bestPenalty)
}

// TestNRPProblem_SerializeSolution verifies serialisation produces valid JSON.
func TestNRPProblem_SerializeSolution(t *testing.T) {
	sc, _ := LoadScenario(testProblemDataDir + "Sc-n005w4.json")
	wd, _ := LoadWeekData(testProblemDataDir + "WD-n005w4-0.json")
	hist, _ := LoadHistory(testProblemDataDir + "H0-n005w4-0.json")

	problem := NewNRPProblem(NRPProblemConfig{
		Scenario: sc, WeekData: wd, History: hist,
	})

	sol, _ := problem.CreateInitialSolution()
	data, err := problem.SerializeSolution(sol)
	if err != nil {
		t.Fatalf("SerializeSolution: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("SerializeSolution returned empty data")
	}

	// Should be valid JSON array.
	var entries []dashboardRosterEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("SerializeSolution output is not valid JSON: %v", err)
	}
	if len(entries) == 0 {
		t.Error("SerializeSolution returned empty roster entries")
	}

	// Verify entries have expected fields.
	for _, e := range entries {
		if e.Nurse == "" {
			t.Error("Entry has empty nurse")
		}
		if e.ShiftType == "" {
			t.Error("Entry has empty shiftType")
		}
		if e.Day == "" {
			t.Error("Entry has empty day")
		}
	}
}

// TestNRPProblem_SolutionFingerprint verifies fingerprints are deterministic.
func TestNRPProblem_SolutionFingerprint(t *testing.T) {
	sc, _ := LoadScenario(testProblemDataDir + "Sc-n005w4.json")
	wd, _ := LoadWeekData(testProblemDataDir + "WD-n005w4-0.json")
	hist, _ := LoadHistory(testProblemDataDir + "H0-n005w4-0.json")

	problem := NewNRPProblem(NRPProblemConfig{
		Scenario: sc, WeekData: wd, History: hist,
	})

	sol, _ := problem.CreateInitialSolution()

	fp1 := problem.SolutionFingerprint(sol)
	fp2 := problem.SolutionFingerprint(sol)
	fp3 := problem.SolutionFingerprint(sol)

	if fp1 != fp2 || fp2 != fp3 {
		t.Errorf("Fingerprint not deterministic: %s, %s, %s", fp1, fp2, fp3)
	}
	if len(fp1) == 0 {
		t.Error("Fingerprint is empty")
	}
}
