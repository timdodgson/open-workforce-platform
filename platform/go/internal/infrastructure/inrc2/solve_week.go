package inrc2

import (
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// WeekSolveParams configures a single-week INRC-II solve.
type WeekSolveParams struct {
	Algorithm  string
	AlgProfile optimisation.AlgorithmProfile
	PFRSConfig PFRSConfig
}

// WeekSolveResult holds the solution and scoring outcome for one week.
type WeekSolveResult struct {
	Solution  Solution
	Score     ScoreResult
	PFRSStats *PFRSStats
	PFRSMode  string
}

// SolveSingleWeek runs PFRS or a standard algorithm for one week.
func SolveSingleWeek(sc Scenario, wd WeekData, hist History, p WeekSolveParams) (WeekSolveResult, error) {
	if p.Algorithm == "parallel-feasible-roster-search" {
		sol, stats, score, err := SolveWeekPFRS(sc, wd, hist, p.PFRSConfig)
		if err != nil {
			return WeekSolveResult{}, err
		}
		return WeekSolveResult{Solution: sol, Score: score, PFRSStats: &stats, PFRSMode: p.PFRSConfig.Mode}, nil
	}

	sol, _, err := SolveWeek(sc, wd, hist, p.Algorithm, p.AlgProfile)
	if err != nil {
		return WeekSolveResult{}, fmt.Errorf("solve week: %w", err)
	}
	return WeekSolveResult{
		Solution: sol,
		Score:    Score(sc, wd, hist, sol),
	}, nil
}
