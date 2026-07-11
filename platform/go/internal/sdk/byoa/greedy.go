// Package byoa registers example custom search modes (BYOA demos).
package byoa

import (
	"math/rand"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk"
)

func init() {
	if err := sdk.RegisterSearch("greedy", runGreedyHillClimb); err != nil {
		panic("sdk/byoa greedy: " + err.Error())
	}
}

// runGreedyHillClimb accepts only strictly improving moves — BYOA demo for any searchdef.Problem.
func runGreedyHillClimb(problem searchdef.Problem, config optimisation.SearchConfig) optimisation.SearchResult {
	start := time.Now()
	rng := rand.New(rand.NewSource(config.Seed))

	sol, err := problem.CreateInitialSolution()
	if err != nil {
		return optimisation.SearchResult{}
	}
	current := problem.Evaluate(sol)
	initial := current
	bestSol := problem.CloneSolution(sol)
	best := current

	var candidates, accepted, improved int

	for i := 0; i < config.Iterations; i++ {
		candidates++
		trial := problem.CloneSolution(sol)
		move := problem.TryMove(trial, rng)
		if !move.Valid {
			continue
		}
		trialCost := problem.Evaluate(trial)
		if trialCost < current {
			sol = trial
			current = trialCost
			accepted++
			if trialCost < best {
				best = trialCost
				bestSol = problem.CloneSolution(sol)
				improved++
			}
		}
	}

	return optimisation.SearchResult{
		BestSolution:   bestSol,
		BestPenalty:    best,
		InitialPenalty: initial,
		DurationMs:     time.Since(start).Milliseconds(),
		Candidates:     candidates,
		Accepted:       accepted,
		Improved:       improved,
	}
}
