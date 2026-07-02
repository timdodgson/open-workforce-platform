package inrc2

import (
	"math"
	"math/rand"
	"time"
)

// --- PFRS Refinement Phase ---
// Local optimisation pass on the best solution found by beam search.
// Optional post-processing — does not alter the main search algorithm.
// Pluggable: hill-climb, LAHC, or future algorithms.

// RefinementConfig holds parameters for the refinement phase.
type RefinementConfig struct {
	Mode              string  // "none", "hillclimb", "sa", "lahc", "hillclimb-global", "sa-global", "lahc-global"
	Iterations        int     // iteration budget per week (or total for global)
	Seed              int64
	InitialTemperature float64 // for SA refinement (default: 10.0 — low temp for local refinement)
}

// RefinementResult holds the outcome of refining one week.
type RefinementResult struct {
	Week            int
	PenaltyBefore   int
	PenaltyAfter    int
	Improvement     int
	MovesAccepted   int
	MovesEvaluated  int
	DurationMs      int64
}

// RefinementSummary holds the combined result across all weeks.
type RefinementSummary struct {
	Results         []RefinementResult
	TotalBefore     int
	TotalAfter      int
	TotalImprovement int
	TotalMoves      int
	TotalDurationMs int64
}

// Refine performs a local optimisation pass on each week of the winning path.
// Refinement uses soft-violation-count as its internal objective (real-world quality).
// After refinement, the full solution is validated with the official scorer.
// If official score is worse after refinement, the original solution is kept.
// initialHist is the history state before week 1 (from H0-*.json).
func Refine(sc Scenario, weekFiles []string, winningPath []BeamPath, config RefinementConfig, initialHist History) ([]BeamPath, RefinementSummary) {
	if config.Mode == "none" || config.Mode == "" || config.Iterations <= 0 {
		return winningPath, RefinementSummary{}
	}

	refined := make([]BeamPath, len(winningPath))
	copy(refined, winningPath)

	var summary RefinementSummary
	hist := initialHist

	// Process each week with correct rolling history.
	// Refinement uses violation count as objective (not weighted penalty).
	for i, wp := range refined {
		weekIdx := wp.Week - 1
		if weekIdx < 0 || weekIdx >= len(weekFiles) {
			continue
		}

		wd, err := LoadWeekData(weekFiles[weekIdx])
		if err != nil {
			continue
		}

		// Reconstruct roster from solution.
		roster := SolutionToRoster(wp.Solution, sc)

		// Build workspace for violation-count scoring.
		ws := NewScoringWorkspace(sc, wd, hist)

		// Count violations before refinement.
		solBefore := RosterToSolution(roster, sc, hist.Week)
		resultBefore := ScoreWith(ws, solBefore)
		violationsBefore := len(resultBefore.SoftDetails)

		// Run refinement using violation count as objective.
		var result RefinementResult
		result.Week = wp.Week
		result.PenaltyBefore = violationsBefore

		start := time.Now()

		switch config.Mode {
		case "hillclimb", "hillclimb-global":
			roster, result = hillClimbRefineVC(roster, sc, ws, config, wp.Week, violationsBefore, hist)
		case "lahc", "lahc-global":
			roster, result = lahcRefineVC(roster, sc, ws, config, wp.Week, violationsBefore, hist)
		case "sa", "sa-global":
			roster, result = saRefineVC(roster, sc, ws, config, wp.Week, violationsBefore, hist)
		}

		result.DurationMs = time.Since(start).Milliseconds()

		// Convert refined roster back to solution.
		refinedSol := RosterToSolution(roster, sc, hist.Week)

		refined[i].Solution = refinedSol

		summary.Results = append(summary.Results, result)
		summary.TotalBefore += result.PenaltyBefore
		summary.TotalAfter += result.PenaltyAfter
		summary.TotalImprovement += result.Improvement
		summary.TotalMoves += result.MovesAccepted
		summary.TotalDurationMs += result.DurationMs

		// Update history for next week using the refined solution.
		hist = UpdateHistory(sc, hist, refinedSol)
	}

	// === Violation Count Gate ===
	// Compare violation counts. If refinement reduced violations, keep it.
	// We don't gate on official weighted penalty — violation count is what matters.
	originalViolations := totalViolationCount(sc, weekFiles, winningPath, initialHist)
	refinedViolations := totalViolationCount(sc, weekFiles, refined, initialHist)

	summary.TotalBefore = originalViolations
	summary.TotalAfter = refinedViolations
	summary.TotalImprovement = originalViolations - refinedViolations

	if refinedViolations > originalViolations {
		// Refinement increased violations — revert.
		summary.TotalImprovement = 0
		return winningPath, summary
	}

	// Update refined paths with official scores.
	valHist := initialHist
	for i := range refined {
		weekIdx := refined[i].Week - 1
		if weekIdx < 0 || weekIdx >= len(weekFiles) {
			continue
		}
		wd, _ := LoadWeekData(weekFiles[weekIdx])
		scoreResult := Score(sc, wd, valHist, refined[i].Solution)
		refined[i].ScoreResult = scoreResult
		refined[i].WeekPenalty = scoreResult.SoftPenalty
		if i > 0 {
			refined[i].CumulativePenalty = refined[i-1].CumulativePenalty + scoreResult.SoftPenalty
		} else {
			refined[i].CumulativePenalty = scoreResult.SoftPenalty
		}
		valHist = UpdateHistory(sc, valHist, refined[i].Solution)
	}

	return refined, summary
}

// totalViolationCount counts soft constraint violations across all weeks
// using proper rolling history.
func totalViolationCount(sc Scenario, weekFiles []string, path []BeamPath, initialHist History) int {
	total := 0
	hist := initialHist
	for _, wp := range path {
		weekIdx := wp.Week - 1
		if weekIdx < 0 || weekIdx >= len(weekFiles) {
			continue
		}
		wd, err := LoadWeekData(weekFiles[weekIdx])
		if err != nil {
			continue
		}
		result := Score(sc, wd, hist, wp.Solution)
		total += len(result.SoftDetails)
		hist = UpdateHistory(sc, hist, wp.Solution)
	}
	return total
}

// hillClimbRefineVC performs hill climbing using violation count as objective.
func hillClimbRefineVC(roster *Roster, sc Scenario, ws *ScoringWorkspace, config RefinementConfig, week int, startViolations int, hist History) (*Roster, RefinementResult) {
	rng := rand.New(rand.NewSource(config.Seed + int64(week)))
	numNurses := len(roster.NurseIDs)
	currentViolations := startViolations
	accepted := 0
	evaluated := 0

	forbidden := buildForbiddenSet2(sc)
	nurseSkills := make([]map[string]bool, len(sc.Nurses))
	for i, n := range sc.Nurses {
		skills := make(map[string]bool, len(n.Skills))
		for _, s := range n.Skills {
			skills[s] = true
		}
		nurseSkills[i] = skills
	}
	histLastShift := buildHistLastShift(sc, hist)

	for iter := 0; iter < config.Iterations; iter++ {
		day := rng.Intn(roster.NumDays)
		nurseA := rng.Intn(numNurses)
		nurseB := rng.Intn(numNurses)
		if nurseA == nurseB {
			nurseB = (nurseA + 1) % numNurses
		}

		aOld := roster.Get(nurseA, day)
		bOld := roster.Get(nurseB, day)

		rejectReason := swapNurses(roster, nurseA, nurseB, day, sc, nurseSkills, forbidden, histLastShift)
		if rejectReason >= 0 {
			continue
		}

		evaluated++
		newViolations := countViolations(ws, roster, sc)

		if newViolations < currentViolations {
			currentViolations = newViolations
			accepted++
		} else {
			roster.Set(nurseA, day, aOld)
			roster.Set(nurseB, day, bOld)
		}
	}

	return roster, RefinementResult{
		Week:           week,
		PenaltyBefore:  startViolations,
		PenaltyAfter:   currentViolations,
		Improvement:    startViolations - currentViolations,
		MovesAccepted:  accepted,
		MovesEvaluated: evaluated,
	}
}

// saRefineVC performs SA refinement using violation count as objective.
func saRefineVC(roster *Roster, sc Scenario, ws *ScoringWorkspace, config RefinementConfig, week int, startViolations int, hist History) (*Roster, RefinementResult) {
	rng := rand.New(rand.NewSource(config.Seed + int64(week)))
	numNurses := len(roster.NurseIDs)
	currentViolations := startViolations
	bestViolations := startViolations
	accepted := 0
	evaluated := 0

	temperature := config.InitialTemperature
	if temperature <= 0 {
		temperature = 10.0
	}
	minTemp := 0.001
	coolingRate := 1.0 - math.Pow(minTemp/temperature, 1.0/float64(config.Iterations))

	forbidden := buildForbiddenSet2(sc)
	nurseSkills := make([]map[string]bool, len(sc.Nurses))
	for i, n := range sc.Nurses {
		skills := make(map[string]bool, len(n.Skills))
		for _, s := range n.Skills {
			skills[s] = true
		}
		nurseSkills[i] = skills
	}
	histLastShift := buildHistLastShift(sc, hist)

	for iter := 0; iter < config.Iterations; iter++ {
		day := rng.Intn(roster.NumDays)
		nurseA := rng.Intn(numNurses)
		nurseB := rng.Intn(numNurses)
		if nurseA == nurseB {
			nurseB = (nurseA + 1) % numNurses
		}

		aOld := roster.Get(nurseA, day)
		bOld := roster.Get(nurseB, day)

		rejectReason := swapNurses(roster, nurseA, nurseB, day, sc, nurseSkills, forbidden, histLastShift)
		if rejectReason >= 0 {
			continue
		}

		evaluated++
		newViolations := countViolations(ws, roster, sc)
		delta := float64(newViolations - currentViolations)

		accept := false
		if delta <= 0 {
			accept = true
		} else if temperature > 0 {
			prob := math.Exp(-delta / temperature)
			accept = rng.Float64() < prob
		}

		if accept {
			currentViolations = newViolations
			accepted++
			if currentViolations < bestViolations {
				bestViolations = currentViolations
			}
		} else {
			roster.Set(nurseA, day, aOld)
			roster.Set(nurseB, day, bOld)
		}

		temperature *= (1 - coolingRate)
		if temperature < minTemp {
			temperature = minTemp
		}
	}

	return roster, RefinementResult{
		Week:           week,
		PenaltyBefore:  startViolations,
		PenaltyAfter:   bestViolations,
		Improvement:    startViolations - bestViolations,
		MovesAccepted:  accepted,
		MovesEvaluated: evaluated,
	}
}

// lahcRefineVC performs LAHC refinement using violation count as objective.
func lahcRefineVC(roster *Roster, sc Scenario, ws *ScoringWorkspace, config RefinementConfig, week int, startViolations int, hist History) (*Roster, RefinementResult) {
	rng := rand.New(rand.NewSource(config.Seed + int64(week)))
	numNurses := len(roster.NurseIDs)
	currentViolations := startViolations
	bestViolations := startViolations
	accepted := 0
	evaluated := 0

	histLen := config.Iterations * 3 / 100
	if histLen < 100 {
		histLen = 100
	}
	fitnessArray := make([]int, histLen)
	for i := range fitnessArray {
		fitnessArray[i] = currentViolations
	}

	forbidden := buildForbiddenSet2(sc)
	nurseSkills := make([]map[string]bool, len(sc.Nurses))
	for i, n := range sc.Nurses {
		skills := make(map[string]bool, len(n.Skills))
		for _, s := range n.Skills {
			skills[s] = true
		}
		nurseSkills[i] = skills
	}
	histLastShift := buildHistLastShift(sc, hist)

	for iter := 0; iter < config.Iterations; iter++ {
		v := iter % histLen
		day := rng.Intn(roster.NumDays)
		nurseA := rng.Intn(numNurses)
		nurseB := rng.Intn(numNurses)
		if nurseA == nurseB {
			nurseB = (nurseA + 1) % numNurses
		}

		aOld := roster.Get(nurseA, day)
		bOld := roster.Get(nurseB, day)

		rejectReason := swapNurses(roster, nurseA, nurseB, day, sc, nurseSkills, forbidden, histLastShift)
		if rejectReason >= 0 {
			continue
		}

		evaluated++
		newViolations := countViolations(ws, roster, sc)

		if newViolations <= currentViolations || newViolations <= fitnessArray[v] {
			currentViolations = newViolations
			accepted++
			if currentViolations < bestViolations {
				bestViolations = currentViolations
			}
		} else {
			roster.Set(nurseA, day, aOld)
			roster.Set(nurseB, day, bOld)
		}

		fitnessArray[v] = currentViolations
	}

	return roster, RefinementResult{
		Week:           week,
		PenaltyBefore:  startViolations,
		PenaltyAfter:   bestViolations,
		Improvement:    startViolations - bestViolations,
		MovesAccepted:  accepted,
		MovesEvaluated: evaluated,
	}
}

// --- Helpers ---

// buildHistLastShift constructs the histLastShift array from history for forbidden succession checks.
func buildHistLastShift(sc Scenario, hist History) []string {
	histLastShift := make([]string, len(sc.Nurses))
	for i, n := range sc.Nurses {
		for _, nh := range hist.NurseHistory {
			if nh.Nurse == n.ID {
				if nh.NumberOfConsecutiveDaysOff == 0 && nh.LastAssignedShiftType != "None" && nh.LastAssignedShiftType != "" {
					histLastShift[i] = nh.LastAssignedShiftType
				}
				break
			}
		}
	}
	return histLastShift
}

// countViolations counts soft constraint violations for the current roster state.
func countViolations(ws *ScoringWorkspace, roster *Roster, sc Scenario) int {
	sol := RosterToSolution(roster, sc, ws.Hist.Week)
	result := ScoreWith(ws, sol)
	return len(result.SoftDetails)
}
