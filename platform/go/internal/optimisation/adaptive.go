package optimisation

import (
	"math"
	"math/rand"
	"time"
)

// --- Adaptive Hyper-Heuristic ---
//
// A single continuous search that uses SA as its primary acceptance criterion
// but adaptively switches to LAHC when stagnation is detected, then reheats
// SA when LAHC produces an improvement.
//
// Design:
//   - Primary mode: SA with adaptive cooling (same as standalone SA)
//   - Stagnation detection: no global best for N candidates
//   - On stagnation: switch to LAHC acceptance for a burst
//   - If LAHC finds improvement: reheat SA and switch back
//   - If LAHC burst expires without improvement: reheat SA anyway, switch back
//
// This gives the search SA's convergence properties + LAHC's ability to
// escape plateaus without the cost of running separate parallel strategies.
//
// The full iteration budget is used by ONE search on ONE solution.
// No budget splitting. No switching overhead beyond acceptance criterion change.

// strategyState holds mutable state for the adaptive search.
// (Retained as package-level type for future telemetry extensions.)
type strategyState struct {
	mode            string
	reward          float64
	totalIter       int
	totalImprove    int
	totalImproveAmt int
	share           float64
	shareHistory    []float64
	temperature     float64
	coolingRate     float64
	fitnessArray    []int
	lahcIdx         int
	tabuList        *genericTabuList
}

func runAdaptive(problem Problem, config SearchConfig) SearchResult {
	start := time.Now()
	rng := rand.New(rand.NewSource(config.Seed))

	sol, err := problem.CreateInitialSolution()
	if err != nil {
		return SearchResult{}
	}

	currentPenalty := problem.Evaluate(sol)
	bestPenalty := currentPenalty
	bestSolution := problem.CloneSolution(sol)
	initialPenalty := currentPenalty

	// SA parameters.
	temperature := config.InitialTemperature
	if temperature <= 0 {
		temperature = 100.0
	}
	minTemp := config.MinTemperature
	if minTemp <= 0 {
		minTemp = 0.0001
	}
	coolingRate := config.CoolingRate
	if config.CoolingMode == "adaptive" && config.Iterations > 0 {
		ratio := minTemp / temperature
		exponent := 1.0 / float64(config.Iterations)
		coolingRate = 1.0 - math.Pow(ratio, exponent)
	}

	// LAHC parameters.
	lahcLen := config.LateAcceptanceLength
	if lahcLen <= 0 {
		lahcLen = 1000
	}
	fitnessArray := make([]int, lahcLen)
	for i := range fitnessArray {
		fitnessArray[i] = currentPenalty
	}

	// Adaptive parameters.
	stagnationThreshold := config.AdaptiveWindow
	if stagnationThreshold <= 0 {
		stagnationThreshold = config.Iterations / 20 // 5% of budget
	}
	lahcBurstLength := stagnationThreshold // LAHC burst same length as detection window
	reheatFactor := 0.5                    // reheat to 50% of initial temperature

	// State.
	candidates := 0
	accepted := 0
	rejected := 0
	improved := 0
	var discoveries []Discovery

	usingSA := true // true = SA acceptance, false = LAHC acceptance
	lastImprovementAt := 0
	hasEverImproved := false // don't trigger stagnation until first improvement
	lahcBurstStart := 0
	lahcIdx := 0
	reheats := 0

	for candidates < config.Iterations {
		result := problem.TryMove(sol, rng)
		if !result.Valid {
			rejected++
			continue
		}
		candidates++

		newPenalty := problem.Evaluate(sol)
		delta := float64(newPenalty - currentPenalty)

		// Acceptance criterion depends on current mode.
		accept := false
		if usingSA {
			// SA: Metropolis acceptance.
			if delta <= 0 {
				accept = true
			} else if temperature > 0 {
				prob := math.Exp(-delta / temperature)
				accept = rng.Float64() < prob
			}
		} else {
			// LAHC: accept if <= current OR <= fitness history.
			v := lahcIdx % lahcLen
			if newPenalty <= currentPenalty || newPenalty <= fitnessArray[v] {
				accept = true
			}
			fitnessArray[v] = currentPenalty
			lahcIdx++
		}

		if accept {
			currentPenalty = newPenalty
			accepted++

			if currentPenalty < bestPenalty {
				oldBest := bestPenalty
				bestPenalty = currentPenalty
				bestSolution = problem.CloneSolution(sol)
				improved++
				lastImprovementAt = candidates

				discoveries = append(discoveries, Discovery{
					ElapsedMs:   time.Since(start).Milliseconds(),
					Candidate:   candidates,
					OldBest:     oldBest,
					NewBest:     bestPenalty,
					Improvement: oldBest - bestPenalty,
				})

				hasEverImproved = true

				// If LAHC found an improvement, switch back to SA with reheat.
				if !usingSA {
					usingSA = true
					temperature = config.InitialTemperature * reheatFactor
					reheats++
				}
			}
		} else {
			problem.UndoMove(sol, result.Move)
		}

		// SA cooling (always cool, even during LAHC — keeps temperature state fresh).
		if usingSA {
			temperature *= (1 - coolingRate)
			if temperature < minTemp {
				temperature = minTemp
			}
		}

		// Stagnation detection: switch to LAHC if no improvement for threshold.
		// Only trigger after we've seen at least one improvement (SA needs time to descend first).
		if usingSA && hasEverImproved && (candidates-lastImprovementAt) > stagnationThreshold {
			usingSA = false
			lahcBurstStart = candidates
			// Reset LAHC fitness array to current penalty.
			for i := range fitnessArray {
				fitnessArray[i] = currentPenalty
			}
		}

		// LAHC burst expiry: switch back to SA with reheat even without improvement.
		if !usingSA && (candidates-lahcBurstStart) > lahcBurstLength {
			usingSA = true
			temperature = config.InitialTemperature * reheatFactor
			reheats++
			lastImprovementAt = candidates // reset stagnation counter
		}
	}

	_ = reheats // available for future telemetry

	return SearchResult{
		BestSolution:   bestSolution,
		BestPenalty:    bestPenalty,
		InitialPenalty: initialPenalty,
		FinalPenalty:   currentPenalty,
		Candidates:     candidates,
		Accepted:       accepted,
		Rejected:       rejected,
		Improved:       improved,
		DurationMs:     time.Since(start).Milliseconds(),
		Discoveries:    discoveries,
	}
}
