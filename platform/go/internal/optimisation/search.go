package optimisation

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// --- Generic Metaheuristic Search Engine ---
//
// Implements SA and LAHC using only the Problem interface.
// No knowledge of NRP, CVRP, or any specific problem domain.

// SearchConfig holds parameters for the generic search engine.
type SearchConfig struct {
	Mode                 string   // "sa", "lahc", "tabu", or "portfolio"
	Iterations           int      // total candidate iterations per strategy
	InitialTemperature   float64  // SA: starting temperature
	MinTemperature       float64  // SA: minimum temperature
	CoolingMode          string   // SA: "adaptive" or "fixed-rate"
	CoolingRate          float64  // SA: used when CoolingMode = "fixed-rate"
	LateAcceptanceLength int      // LAHC: fitness array length (default 1000)
	TabuTenure           int      // Tabu: number of iterations a move stays forbidden (default 7)
	TabuNeighbourhood    int      // Tabu: moves sampled per iteration for best-move (default 100)
	Portfolio            []string // Portfolio: list of modes to run (e.g. ["sa","lahc","tabu"])
	Seed                 int64
}

// DefaultSearchConfig returns sensible defaults for SA.
func DefaultSearchConfig() SearchConfig {
	return SearchConfig{
		Mode:                 "sa",
		Iterations:           500000,
		InitialTemperature:   100.0,
		MinTemperature:       0.0001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 1000,
		TabuTenure:           7,
		TabuNeighbourhood:    100,
		Seed:                 42,
	}
}

// SearchResult captures the output of a search run.
type SearchResult struct {
	BestSolution   Solution
	BestPenalty    int
	InitialPenalty int
	FinalPenalty   int // penalty at end of search (current, not necessarily best)
	Candidates     int
	Accepted       int
	Rejected       int // hard-rejected (TryMove returned invalid)
	Improved       int // moves that improved the best penalty
	DurationMs     int64
	Discoveries    []Discovery // every global best improvement
}

// Discovery records a single global best improvement during search.
type Discovery struct {
	ElapsedMs   int64
	Candidate   int
	OldBest     int
	NewBest     int
	Improvement int
}

// RunSearch executes a metaheuristic search using only the Problem interface.
// Supports SA, LAHC, Tabu, and Portfolio modes.
func RunSearch(problem Problem, config SearchConfig) SearchResult {
	switch config.Mode {
	case "lahc":
		return runLAHC(problem, config)
	case "tabu":
		return runTabu(problem, config)
	case "portfolio":
		return runPortfolio(problem, config)
	default:
		return runSA(problem, config)
	}
}

// --- Simulated Annealing ---

func runSA(problem Problem, config SearchConfig) SearchResult {
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

	// Compute effective cooling rate.
	coolingRate := config.CoolingRate
	if config.CoolingMode == "adaptive" && config.Iterations > 0 &&
		config.InitialTemperature > 0 && config.MinTemperature > 0 {
		ratio := config.MinTemperature / config.InitialTemperature
		exponent := 1.0 / float64(config.Iterations)
		coolingRate = 1.0 - math.Pow(ratio, exponent)
	}

	temperature := config.InitialTemperature
	candidates := 0
	accepted := 0
	rejected := 0
	improved := 0
	var discoveries []Discovery

	for candidates < config.Iterations {
		result := problem.TryMove(sol, rng)
		if !result.Valid {
			rejected++
			continue
		}
		candidates++

		newPenalty := problem.Evaluate(sol)
		delta := float64(newPenalty - currentPenalty)

		// Metropolis acceptance.
		accept := false
		if delta <= 0 {
			accept = true
		} else if temperature > 0 {
			prob := math.Exp(-delta / temperature)
			accept = rng.Float64() < prob
		}

		if accept {
			currentPenalty = newPenalty
			accepted++

			if currentPenalty < bestPenalty {
				oldBest := bestPenalty
				bestPenalty = currentPenalty
				bestSolution = problem.CloneSolution(sol)
				improved++
				discoveries = append(discoveries, Discovery{
					ElapsedMs:   time.Since(start).Milliseconds(),
					Candidate:   candidates,
					OldBest:     oldBest,
					NewBest:     bestPenalty,
					Improvement: oldBest - bestPenalty,
				})
			}
		} else {
			problem.UndoMove(sol, result.Move)
		}

		// Cool temperature.
		temperature *= (1 - coolingRate)
		if temperature < config.MinTemperature {
			temperature = config.MinTemperature
		}
	}

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

// --- Late Acceptance Hill Climbing ---

func runLAHC(problem Problem, config SearchConfig) SearchResult {
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

	// Initialize fitness array.
	histLen := config.LateAcceptanceLength
	if histLen <= 0 {
		histLen = 1000
	}
	fitnessArray := make([]int, histLen)
	for i := range fitnessArray {
		fitnessArray[i] = currentPenalty
	}

	candidates := 0
	accepted := 0
	rejected := 0
	improved := 0
	var discoveries []Discovery

	for candidates < config.Iterations {
		result := problem.TryMove(sol, rng)
		if !result.Valid {
			rejected++
			continue
		}
		candidates++

		v := candidates % histLen
		newPenalty := problem.Evaluate(sol)

		// LAHC acceptance: accept if better than current OR better than fitness[v].
		accept := newPenalty <= currentPenalty || newPenalty <= fitnessArray[v]

		if accept {
			currentPenalty = newPenalty
			accepted++

			if currentPenalty < bestPenalty {
				oldBest := bestPenalty
				bestPenalty = currentPenalty
				bestSolution = problem.CloneSolution(sol)
				improved++
				discoveries = append(discoveries, Discovery{
					ElapsedMs:   time.Since(start).Milliseconds(),
					Candidate:   candidates,
					OldBest:     oldBest,
					NewBest:     bestPenalty,
					Improvement: oldBest - bestPenalty,
				})
			}
		} else {
			problem.UndoMove(sol, result.Move)
		}

		fitnessArray[v] = currentPenalty
	}

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

// --- Tabu Search ---

// genericTabuEntry stores a move signature in the generic tabu list.
type genericTabuEntry struct {
	signature string
}

// genericTabuList is a fixed-size circular buffer of recently accepted moves.
type genericTabuList struct {
	entries []genericTabuEntry
	tenure  int
	head    int
	size    int
}

func newGenericTabuList(tenure int) *genericTabuList {
	return &genericTabuList{
		entries: make([]genericTabuEntry, tenure),
		tenure:  tenure,
	}
}

func (tl *genericTabuList) add(sig string) {
	tl.entries[tl.head] = genericTabuEntry{signature: sig}
	tl.head = (tl.head + 1) % tl.tenure
	if tl.size < tl.tenure {
		tl.size++
	}
}

func (tl *genericTabuList) contains(sig string) bool {
	for i := 0; i < tl.size; i++ {
		if tl.entries[i].signature == sig {
			return true
		}
	}
	return false
}

// MoveSignature extracts a string signature from a Move for tabu comparison.
// This is a simple approach: use the Move's fmt representation.
// Problem-specific implementations can provide better signatures via type assertion.
func moveSignature(m Move) string {
	// Use fmt.Sprint which will produce a stable string for any concrete type.
	return fmt.Sprint(m)
}

func runTabu(problem Problem, config SearchConfig) SearchResult {
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

	tenure := config.TabuTenure
	if tenure <= 0 {
		tenure = 7
	}
	tl := newGenericTabuList(tenure)

	neighbourhoodSize := config.TabuNeighbourhood
	if neighbourhoodSize <= 0 {
		neighbourhoodSize = 100
	}

	iterations := 0
	accepted := 0
	rejected := 0
	improved := 0
	var discoveries []Discovery

	for iterations < config.Iterations {
		// Best-move neighbourhood evaluation.
		// For each candidate: TryMove → Evaluate → UndoMove.
		// Track the best admissible move's penalty and the solution state (via clone).
		var bestMovePenalty int
		var bestMoveSig string
		var bestMoveSolution Solution
		foundAdmissible := false

		for s := 0; s < neighbourhoodSize; s++ {
			result := problem.TryMove(sol, rng)
			if !result.Valid {
				rejected++
				continue
			}

			penalty := problem.Evaluate(sol)
			sig := moveSignature(result.Move)
			isTabu := tl.contains(sig)
			admissible := !isTabu || penalty < bestPenalty

			if admissible && (!foundAdmissible || penalty < bestMovePenalty) {
				bestMovePenalty = penalty
				bestMoveSig = sig
				// Clone the solution in its current (move-applied) state.
				bestMoveSolution = problem.CloneSolution(sol)
				foundAdmissible = true
			}

			// Undo — return to base state for next candidate evaluation.
			problem.UndoMove(sol, result.Move)
		}

		iterations++

		if !foundAdmissible {
			// No admissible move found in neighbourhood — skip this iteration.
			continue
		}

		// Commit the best move: replace solution with the cloned best state.
		sol = bestMoveSolution
		currentPenalty = bestMovePenalty
		accepted++
		tl.add(bestMoveSig)

		if currentPenalty < bestPenalty {
			oldBest := bestPenalty
			bestPenalty = currentPenalty
			bestSolution = problem.CloneSolution(sol)
			improved++
			discoveries = append(discoveries, Discovery{
				ElapsedMs:   time.Since(start).Milliseconds(),
				Candidate:   iterations,
				OldBest:     oldBest,
				NewBest:     bestPenalty,
				Improvement: oldBest - bestPenalty,
			})
		}
	}

	return SearchResult{
		BestSolution:   bestSolution,
		BestPenalty:    bestPenalty,
		InitialPenalty: initialPenalty,
		FinalPenalty:   currentPenalty,
		Candidates:     iterations,
		Accepted:       accepted,
		Rejected:       rejected,
		Improved:       improved,
		DurationMs:     time.Since(start).Milliseconds(),
		Discoveries:    discoveries,
	}
}

// --- Portfolio Mode ---

// PortfolioEntry captures the result from one strategy within a portfolio run.
type PortfolioEntry struct {
	Mode       string
	Seed       int64
	Result     SearchResult
}

// PortfolioResult captures the full portfolio output including per-strategy breakdown.
type PortfolioResult struct {
	Winner     string          // mode of the best strategy
	Entries    []PortfolioEntry
	BestResult SearchResult    // the winning strategy's result
}

// RunPortfolio runs multiple strategies sequentially and returns the best.
// Use this directly when you need the full PortfolioResult with per-strategy breakdown.
func RunPortfolio(problem Problem, config SearchConfig) PortfolioResult {
	strategies := config.Portfolio
	if len(strategies) == 0 {
		strategies = []string{"sa", "lahc", "tabu"}
	}

	var entries []PortfolioEntry
	var bestIdx int
	bestPenalty := int(^uint(0) >> 1) // max int

	for i, mode := range strategies {
		// Derive a deterministic seed per strategy.
		derivedSeed := config.Seed + int64(i)*7919 // prime offset for separation

		stratConfig := config
		stratConfig.Mode = mode
		stratConfig.Seed = derivedSeed

		result := RunSearch(problem, stratConfig)

		entries = append(entries, PortfolioEntry{
			Mode:   mode,
			Seed:   derivedSeed,
			Result: result,
		})

		if result.BestPenalty < bestPenalty {
			bestPenalty = result.BestPenalty
			bestIdx = i
		}
	}

	return PortfolioResult{
		Winner:     entries[bestIdx].Mode,
		Entries:    entries,
		BestResult: entries[bestIdx].Result,
	}
}

// runPortfolio adapts the portfolio to the RunSearch interface (returns single SearchResult).
func runPortfolio(problem Problem, config SearchConfig) SearchResult {
	pr := RunPortfolio(problem, config)
	return pr.BestResult
}
