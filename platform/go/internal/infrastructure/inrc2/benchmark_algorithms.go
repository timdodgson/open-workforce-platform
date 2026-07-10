package inrc2

import (
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2/legacysearch"
)

// AlgorithmBenchmarkResult aggregates one algorithm's scores across all weeks.
type AlgorithmBenchmarkResult struct {
	Algorithm    string
	TotalPenalty int
	TotalHard    int
	TotalSoft    int
	TotalAssign  int
	TotalMs      int64
	TotalCands   int
}

// AlgorithmBenchmarkParams configures a multi-algorithm INRC-II week-by-week benchmark.
type AlgorithmBenchmarkParams struct {
	Scenario   Scenario
	WeekFiles  []string
	History    History
	NumWeeks   int
	Algorithms []string
	AlgProfile legacysearch.AlgorithmProfile
	PFRSConfig PFRSConfig
	OnWeekStart func(week int, algorithm string) // optional CLI progress
}

// RunAlgorithmBenchmark runs each algorithm across all weeks and returns per-algorithm totals.
// History advances using the single algorithm's solution when only one algorithm is run;
// otherwise constructive is used for fair cross-algorithm comparison.
func RunAlgorithmBenchmark(p AlgorithmBenchmarkParams) map[string]*AlgorithmBenchmarkResult {
	results := make(map[string]*AlgorithmBenchmarkResult, len(p.Algorithms))
	for _, alg := range p.Algorithms {
		results[alg] = &AlgorithmBenchmarkResult{Algorithm: alg}
	}

	currentHist := p.History
	for w := 0; w < p.NumWeeks; w++ {
		wd, err := LoadWeekData(p.WeekFiles[w])
		if err != nil {
			continue
		}

		var weekSolForHistory Solution
		hasSolForHistory := false

		for _, alg := range p.Algorithms {
			if p.OnWeekStart != nil {
				p.OnWeekStart(w, alg)
			}
			var sol Solution
			var scoreResult ScoreResult
			var durationMs int64
			var candidatesEval int

			if alg == "parallel-feasible-roster-search" {
				pfrsSol, pfrsStats, pfrsScore, err := SolveWeekPFRS(p.Scenario, wd, currentHist, p.PFRSConfig)
				if err != nil {
					continue
				}
				sol = pfrsSol
				scoreResult = pfrsScore
				durationMs = pfrsStats.DurationMs
				candidatesEval = pfrsStats.CandidatesEvaluated
			} else {
				owpSol, planResult, err := SolveWeek(p.Scenario, wd, currentHist, alg, p.AlgProfile)
				if err != nil {
					continue
				}
				sol = owpSol
				scoreResult = Score(p.Scenario, wd, currentHist, sol)
				stats := planResult.Statistics()
				durationMs = stats.DurationMs
				candidatesEval = stats.CandidatesEvaluated
			}

			r := results[alg]
			r.TotalPenalty += scoreResult.SoftPenalty
			r.TotalHard += scoreResult.HardViolations
			r.TotalSoft += len(scoreResult.SoftDetails)
			r.TotalAssign += len(sol.Assignments)
			r.TotalMs += durationMs
			r.TotalCands += candidatesEval

			if len(p.Algorithms) == 1 {
				weekSolForHistory = sol
				hasSolForHistory = true
			}
		}

		if len(p.Algorithms) == 1 && hasSolForHistory {
			currentHist = UpdateHistory(p.Scenario, currentHist, weekSolForHistory)
		} else {
			sol, _, _ := SolveWeek(p.Scenario, wd, currentHist, "constructive", p.AlgProfile)
			currentHist = UpdateHistory(p.Scenario, currentHist, sol)
		}
	}

	return results
}

// RankAlgorithmBenchmarkResults splits and ranks valid (hard=0) vs invalid results.
func RankAlgorithmBenchmarkResults(algorithms []string, results map[string]*AlgorithmBenchmarkResult) (valid, invalid []*AlgorithmBenchmarkResult) {
	for _, alg := range algorithms {
		r := results[alg]
		if r == nil {
			continue
		}
		if r.TotalHard == 0 {
			valid = append(valid, r)
		} else {
			invalid = append(invalid, r)
		}
	}

	sortAlgorithmResults(valid, true)
	sortAlgorithmResults(invalid, false)
	return valid, invalid
}

func sortAlgorithmResults(results []*AlgorithmBenchmarkResult, valid bool) {
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if algorithmResultLess(results[j], results[i], valid) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func algorithmResultLess(a, b *AlgorithmBenchmarkResult, valid bool) bool {
	if valid {
		if a.TotalPenalty != b.TotalPenalty {
			return a.TotalPenalty < b.TotalPenalty
		}
		if a.TotalSoft != b.TotalSoft {
			return a.TotalSoft < b.TotalSoft
		}
		if a.TotalMs != b.TotalMs {
			return a.TotalMs < b.TotalMs
		}
		return a.Algorithm < b.Algorithm
	}
	if a.TotalHard != b.TotalHard {
		return a.TotalHard < b.TotalHard
	}
	return a.TotalPenalty < b.TotalPenalty
}
