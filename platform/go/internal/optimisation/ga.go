package optimisation

import (
	"math/rand"
	"sort"
	"time"
)

// gaIndividual is one member of the evolutionary population.
type gaIndividual struct {
	sol     Solution
	fitness int
}

// runGA executes a population-based genetic algorithm using only the Problem interface.
//
// Each generation: elitism preserves the best individuals; offspring are built via
// two-parent tournament selection, dual-parent crossover (neighbourhood blending),
// and greedy-biased mutation. Total valid move evaluations approximate Iterations.
func runGA(problem Problem, config SearchConfig) SearchResult {
	start := time.Now()
	rng := rand.New(rand.NewSource(config.Seed))

	popSize := config.GAPopulationSize
	if popSize <= 0 {
		popSize = 32
	}
	eliteCount := config.GAEliteCount
	if eliteCount <= 0 {
		eliteCount = 2
	}
	if eliteCount >= popSize {
		eliteCount = popSize / 4
		if eliteCount < 1 {
			eliteCount = 1
		}
	}
	tourneySize := config.GATournamentSize
	if tourneySize <= 0 {
		tourneySize = 3
	}
	mutationMoves := config.GAMutationMoves
	if mutationMoves <= 0 {
		mutationMoves = 5
	}
	crossoverMoves := config.GACrossoverMoves
	if crossoverMoves <= 0 {
		crossoverMoves = 3
	}

	generations := config.Iterations / popSize
	if generations < 1 {
		generations = 1
	}

	initSol, err := problem.CreateInitialSolution()
	if err != nil {
		return SearchResult{}
	}
	initialPenalty := problem.Evaluate(initSol)

	pop := make([]gaIndividual, 0, popSize)
	for i := 0; i < popSize; i++ {
		ind := gaIndividual{sol: problem.CloneSolution(initSol)}
		gaWarmup(problem, &ind, 8, rng)
		pop = append(pop, ind)
	}

	bestPenalty := pop[0].fitness
	bestSolution := problem.CloneSolution(pop[0].sol)
	for _, ind := range pop[1:] {
		if ind.fitness < bestPenalty {
			bestPenalty = ind.fitness
			bestSolution = problem.CloneSolution(ind.sol)
		}
	}

	candidates := 0
	accepted := 0
	rejected := 0
	improved := 0
	var discoveries []Discovery

	hooks := newSearchHooks(config)

	for gen := 0; gen < generations; gen++ {
		sort.Slice(pop, func(i, j int) bool {
			return pop[i].fitness < pop[j].fitness
		})

		if pop[0].fitness < bestPenalty {
			oldBest := bestPenalty
			bestPenalty = pop[0].fitness
			bestSolution = problem.CloneSolution(pop[0].sol)
			improved++
			if hooks != nil {
				hooks.OnImprovement(candidates)
			}
			discoveries = append(discoveries, Discovery{
				ElapsedMs:   time.Since(start).Milliseconds(),
				Candidate:   candidates,
				OldBest:     oldBest,
				NewBest:     bestPenalty,
				Improvement: oldBest - bestPenalty,
			})
		}

		newPop := make([]gaIndividual, 0, popSize)
		for e := 0; e < eliteCount; e++ {
			newPop = append(newPop, gaIndividual{
				sol:     problem.CloneSolution(pop[e].sol),
				fitness: pop[e].fitness,
			})
		}

		for len(newPop) < popSize {
			p1 := gaTournamentSelect(pop, tourneySize, rng)
			p2 := gaTournamentSelect(pop, tourneySize, rng)
			child := gaCrossover(problem, p1, p2, crossoverMoves, rng, &candidates, &accepted, &rejected)
			gaMutate(problem, &child, mutationMoves, rng, &candidates, &accepted, &rejected)
			newPop = append(newPop, child)
		}
		pop = newPop

		if hooks != nil && hooks.ShouldCheckpoint(candidates) {
			currentPenalty := pop[0].fitness
			action := hooks.RunCheckpoint("ga", candidates, currentPenalty, bestPenalty, initialPenalty, 0)
			var stop bool
			var sol Solution = pop[0].sol
			var current int = currentPenalty
			sol, stop = applyCheckpointAction(action, hooks, candidates, problem, sol, &current, bestSolution, bestPenalty, nil, config)
			if stop {
				break
			}
		}
	}

	assistRecords, policyDecisions := finalizeSearchHooks(hooks, bestPenalty, candidates)

	sort.Slice(pop, func(i, j int) bool {
		return pop[i].fitness < pop[j].fitness
	})

	return SearchResult{
		BestSolution:    bestSolution,
		BestPenalty:     bestPenalty,
		InitialPenalty:  initialPenalty,
		FinalPenalty:    pop[0].fitness,
		Candidates:      candidates,
		Accepted:        accepted,
		Rejected:        rejected,
		Improved:        improved,
		DurationMs:      time.Since(start).Milliseconds(),
		Discoveries:     discoveries,
		AssistRecords:   assistRecords,
		PolicyDecisions: policyDecisions,
	}
}

func gaWarmup(problem Problem, ind *gaIndividual, moves int, rng *rand.Rand) {
	current := problem.Evaluate(ind.sol)
	for i := 0; i < moves; i++ {
		result := problem.TryMove(ind.sol, rng)
		if !result.Valid {
			continue
		}
		newFit := problem.Evaluate(ind.sol)
		if newFit <= current {
			current = newFit
		} else {
			problem.UndoMove(ind.sol, result.Move)
		}
	}
	ind.fitness = current
}

func gaTournamentSelect(pop []gaIndividual, size int, rng *rand.Rand) gaIndividual {
	if size > len(pop) {
		size = len(pop)
	}
	best := pop[rng.Intn(len(pop))]
	for i := 1; i < size; i++ {
		candidate := pop[rng.Intn(len(pop))]
		if candidate.fitness < best.fitness {
			best = candidate
		}
	}
	return best
}

func gaCrossover(
	problem Problem,
	a, b gaIndividual,
	moves int,
	rng *rand.Rand,
	candidates, accepted, rejected *int,
) gaIndividual {
	better, worse := a, b
	if b.fitness < a.fitness {
		better, worse = b, a
	}

	childSol := problem.CloneSolution(better.sol)
	shadow := problem.CloneSolution(worse.sol)

	for i := 0; i < moves; i++ {
		sr := problem.TryMove(shadow, rng)
		if !sr.Valid {
			*rejected++
		} else {
			*candidates++
			shadowFit := problem.Evaluate(shadow)
			_ = shadowFit
			*accepted++

			cr := problem.TryMove(childSol, rng)
			if !cr.Valid {
				*rejected++
				continue
			}
			*candidates++
			childFit := problem.Evaluate(childSol)
			if childFit > shadowFit && rng.Float64() < 0.5 {
				problem.UndoMove(childSol, cr.Move)
			} else {
				*accepted++
			}
		}
	}

	return gaIndividual{sol: childSol, fitness: problem.Evaluate(childSol)}
}

func gaMutate(
	problem Problem,
	ind *gaIndividual,
	moves int,
	rng *rand.Rand,
	candidates, accepted, rejected *int,
) {
	current := ind.fitness
	for i := 0; i < moves; i++ {
		result := problem.TryMove(ind.sol, rng)
		if !result.Valid {
			*rejected++
			continue
		}
		*candidates++
		newFit := problem.Evaluate(ind.sol)
		if newFit <= current {
			current = newFit
			*accepted++
		} else {
			problem.UndoMove(ind.sol, result.Move)
		}
	}
	ind.fitness = current
}
