package inrc2

import (
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// --- PFRS Configuration ---

// PFRSConfig holds all tunables for the Parallel Feasible Roster Search.
type PFRSConfig struct {
	Mode                 string  // "sa" or "lahc"
	IterationsPerWorker  int
	MaxConcurrentWorkers int
	MaxTotalWorkers      int
	BranchOnGlobalBest   bool
	InitialTemperature   float64
	CoolingRate          float64
	MinTemperature       float64
	LateAcceptanceLength int
	Seed                 int64
	Deterministic        bool
	ScoringMode          string // "official-penalty" or "soft-violation-count"
}

// DefaultPFRSConfig returns sensible defaults matching Tim's dissertation parameters.
func DefaultPFRSConfig() PFRSConfig {
	return PFRSConfig{
		Mode:                 "sa",
		IterationsPerWorker:  30000,
		MaxConcurrentWorkers: 4,
		MaxTotalWorkers:      16,
		BranchOnGlobalBest:   true,
		InitialTemperature:   1.0,
		CoolingRate:          0.0009,
		MinTemperature:       0.0001,
		LateAcceptanceLength: 1000,
		Seed:                 42,
		Deterministic:        true,
		ScoringMode:          "official-penalty",
	}
}

// --- PFRS Statistics ---

// PFRSStats captures execution metrics for the PFRS algorithm.
type PFRSStats struct {
	WorkersStarted       int
	BranchesCreated      int
	BestUpdates          int
	InvalidMovesRejected int
	TotalIterations      int
	CandidatesEvaluated  int
	ImprovementsAccepted int
	DurationMs           int64
	FinalPenalty         int
}

// --- Swap Operator ---

// swapNurses swaps two nurses' assignments on the same day.
// Returns true if the swap is hard-valid (preserves feasibility).
// histLastShift provides each nurse's last shift from history for day-0 succession checks.
func swapNurses(roster *Roster, nurseA, nurseB, dayIdx int, sc Scenario, nurseSkills []map[string]bool, forbidden map[string]bool, histLastShift []string) bool {
	aAssign := roster.Get(nurseA, dayIdx)
	bAssign := roster.Get(nurseB, dayIdx)

	// If both are off or both have the same assignment, swap is a no-op.
	if aAssign == bAssign {
		return false
	}

	// Check skill validity after swap.
	// Nurse A would get B's assignment.
	if bAssign.ShiftType != "" && bAssign.Skill != "" {
		if !nurseSkills[nurseA][bAssign.Skill] {
			return false
		}
	}
	// Nurse B would get A's assignment.
	if aAssign.ShiftType != "" && aAssign.Skill != "" {
		if !nurseSkills[nurseB][aAssign.Skill] {
			return false
		}
	}

	// Check forbidden succession for nurse A with B's shift.
	if !isSuccessionValidAfterSwap(roster, nurseA, dayIdx, bAssign.ShiftType, forbidden, histLastShift) {
		return false
	}
	// Check forbidden succession for nurse B with A's shift.
	if !isSuccessionValidAfterSwap(roster, nurseB, dayIdx, aAssign.ShiftType, forbidden, histLastShift) {
		return false
	}

	// Apply swap.
	roster.Set(nurseA, dayIdx, bAssign)
	roster.Set(nurseB, dayIdx, aAssign)
	return true
}

// isSuccessionValidAfterSwap checks if placing newShift on nurseIdx at dayIdx
// would violate forbidden successions with the previous and next day.
// histLastShift provides each nurse's last shift from history for day-0 validation.
func isSuccessionValidAfterSwap(roster *Roster, nurseIdx, dayIdx int, newShift string, forbidden map[string]bool, histLastShift []string) bool {
	// Check previous day → this day.
	if dayIdx > 0 {
		prevShift := roster.Get(nurseIdx, dayIdx-1).ShiftType
		if prevShift != "" && newShift != "" {
			if forbidden[prevShift+"|"+newShift] {
				return false
			}
		}
	} else {
		// Day 0: check history last shift → new shift.
		if nurseIdx < len(histLastShift) {
			prevShift := histLastShift[nurseIdx]
			if prevShift != "" && newShift != "" {
				if forbidden[prevShift+"|"+newShift] {
					return false
				}
			}
		}
	}

	// Check this day → next day.
	if dayIdx < roster.NumDays-1 {
		nextShift := roster.Get(nurseIdx, dayIdx+1).ShiftType
		if newShift != "" && nextShift != "" {
			if forbidden[newShift+"|"+nextShift] {
				return false
			}
		}
	}

	return true
}

// --- Scoring ---

// scorePenaltyWithMode computes the score based on the configured mode.
// "official-penalty" returns the official soft penalty.
// "soft-violation-count" returns the count of soft constraint violations.
func scorePenaltyWithMode(roster *Roster, sc Scenario, wd WeekData, hist History, mode string) int {
	sol := RosterToSolution(roster, sc, hist.Week)
	result := Score(sc, wd, hist, sol)
	if mode == "soft-violation-count" {
		return len(result.SoftDetails)
	}
	return result.SoftPenalty
}

// --- SA Worker ---

func saWorker(startRoster *Roster, sc Scenario, wd WeekData, hist History,
	nurseSkills []map[string]bool, forbidden map[string]bool, histLastShift []string,
	config PFRSConfig, workerID int, globalBest *int64, bestMu *sync.Mutex,
	bestRoster **Roster, branchChan chan<- *Roster, stats *PFRSStats, statsMu *sync.Mutex) {

	rng := rand.New(rand.NewSource(config.Seed + int64(workerID)))
	roster := startRoster.Clone()
	currentPenalty := scorePenaltyWithMode(roster, sc, wd, hist, config.ScoringMode)
	localBest := currentPenalty
	localBestRoster := roster.Clone()

	numNurses := len(roster.NurseIDs)
	temperature := config.InitialTemperature
	iterations := 0
	candidates := 0
	accepted := 0
	rejected := 0

	for iter := 0; iter < config.IterationsPerWorker; iter++ {
		if temperature < config.MinTemperature {
			break
		}
		iterations++

		// Generate a random swap: pick a day, pick two nurses.
		day := rng.Intn(7)
		nurseA := rng.Intn(numNurses)
		nurseB := rng.Intn(numNurses)
		if nurseA == nurseB {
			nurseB = (nurseA + 1) % numNurses
		}

		candidates++

		// Try swap.
		aOld := roster.Get(nurseA, day)
		bOld := roster.Get(nurseB, day)

		if !swapNurses(roster, nurseA, nurseB, day, sc, nurseSkills, forbidden, histLastShift) {
			rejected++
			// Linear cooling: temperature -= coolingRate per iteration.
			temperature -= config.CoolingRate
			continue
		}

		// Score the new roster.
		newPenalty := scorePenaltyWithMode(roster, sc, wd, hist, config.ScoringMode)
		delta := float64(newPenalty - currentPenalty)

		// Metropolis acceptance (minimisation: delta < 0 is improvement).
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

			if currentPenalty < localBest {
				localBest = currentPenalty
				localBestRoster = roster.Clone()

				// Check global best.
				gb := atomic.LoadInt64(globalBest)
				if int64(localBest) < gb {
					bestMu.Lock()
					if int64(localBest) < atomic.LoadInt64(globalBest) {
						atomic.StoreInt64(globalBest, int64(localBest))
						*bestRoster = localBestRoster.Clone()
						// Signal branch.
						if config.BranchOnGlobalBest && branchChan != nil {
							select {
							case branchChan <- localBestRoster.Clone():
							default:
							}
						}
					}
					bestMu.Unlock()
				}
			}
		} else {
			// Undo swap.
			roster.Set(nurseA, day, aOld)
			roster.Set(nurseB, day, bOld)
		}

		// Linear cooling (ramp descent per dissertation).
		temperature -= config.CoolingRate
	}

	// Update shared stats.
	statsMu.Lock()
	stats.TotalIterations += iterations
	stats.CandidatesEvaluated += candidates
	stats.ImprovementsAccepted += accepted
	stats.InvalidMovesRejected += rejected
	statsMu.Unlock()
}

// --- LAHC Worker ---

func lahcWorker(startRoster *Roster, sc Scenario, wd WeekData, hist History,
	nurseSkills []map[string]bool, forbidden map[string]bool, histLastShift []string,
	config PFRSConfig, workerID int, globalBest *int64, bestMu *sync.Mutex,
	bestRoster **Roster, branchChan chan<- *Roster, stats *PFRSStats, statsMu *sync.Mutex) {

	rng := rand.New(rand.NewSource(config.Seed + int64(workerID)))
	roster := startRoster.Clone()
	currentPenalty := scorePenaltyWithMode(roster, sc, wd, hist, config.ScoringMode)

	// Initialize LAHC fitness array.
	histLen := config.LateAcceptanceLength
	if histLen <= 0 {
		histLen = 1000
	}
	fitnessArray := make([]int, histLen)
	for i := range fitnessArray {
		fitnessArray[i] = currentPenalty
	}

	localBest := currentPenalty
	localBestRoster := roster.Clone()
	numNurses := len(roster.NurseIDs)
	iterations := 0
	candidates := 0
	accepted := 0
	rejected := 0

	for iter := 0; iter < config.IterationsPerWorker; iter++ {
		iterations++
		v := iter % histLen

		day := rng.Intn(7)
		nurseA := rng.Intn(numNurses)
		nurseB := rng.Intn(numNurses)
		if nurseA == nurseB {
			nurseB = (nurseA + 1) % numNurses
		}

		candidates++
		aOld := roster.Get(nurseA, day)
		bOld := roster.Get(nurseB, day)

		if !swapNurses(roster, nurseA, nurseB, day, sc, nurseSkills, forbidden, histLastShift) {
			rejected++
			continue
		}

		newPenalty := scorePenaltyWithMode(roster, sc, wd, hist, config.ScoringMode)

		// LAHC acceptance: accept if better than current OR better than fitnessArray[v].
		if newPenalty <= currentPenalty || newPenalty <= fitnessArray[v] {
			currentPenalty = newPenalty
			accepted++

			if currentPenalty < localBest {
				localBest = currentPenalty
				localBestRoster = roster.Clone()

				gb := atomic.LoadInt64(globalBest)
				if int64(localBest) < gb {
					bestMu.Lock()
					if int64(localBest) < atomic.LoadInt64(globalBest) {
						atomic.StoreInt64(globalBest, int64(localBest))
						*bestRoster = localBestRoster.Clone()
						if config.BranchOnGlobalBest && branchChan != nil {
							select {
							case branchChan <- localBestRoster.Clone():
							default:
							}
						}
					}
					bestMu.Unlock()
				}
			}
		} else {
			// Undo.
			roster.Set(nurseA, day, aOld)
			roster.Set(nurseB, day, bOld)
		}

		fitnessArray[v] = currentPenalty
	}

	statsMu.Lock()
	stats.TotalIterations += iterations
	stats.CandidatesEvaluated += candidates
	stats.ImprovementsAccepted += accepted
	stats.InvalidMovesRejected += rejected
	statsMu.Unlock()
}

// --- Parallel Branching Coordinator ---

// RunPFRS executes the Parallel Feasible Roster Search algorithm.
func RunPFRS(sc Scenario, wd WeekData, hist History, config PFRSConfig) (Solution, PFRSStats, error) {
	startTime := time.Now()

	// Build feasible initial roster.
	initialRoster, err := BuildFeasibleRoster(sc, wd, hist)
	if err != nil {
		return Solution{}, PFRSStats{}, err
	}

	// Pre-compute nurse skills and forbidden set.
	nurseSkills := make([]map[string]bool, len(sc.Nurses))
	for i, n := range sc.Nurses {
		skills := make(map[string]bool, len(n.Skills))
		for _, s := range n.Skills {
			skills[s] = true
		}
		nurseSkills[i] = skills
	}
	forbidden := buildForbiddenSet2(sc)

	// Pre-compute history last shift for day-0 succession validation in swaps.
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

	// Initialize global best.
	initialPenalty := scorePenaltyWithMode(initialRoster, sc, wd, hist, config.ScoringMode)
	var globalBest int64 = int64(initialPenalty)
	var bestRoster *Roster = initialRoster.Clone()
	var bestMu sync.Mutex
	var stats PFRSStats
	var statsMu sync.Mutex

	// Branch channel (buffered to prevent blocking).
	branchChan := make(chan *Roster, config.MaxTotalWorkers)

	// Worker tracking.
	var wg sync.WaitGroup
	sem := make(chan struct{}, config.MaxConcurrentWorkers)
	totalWorkers := 0

	// Launch initial worker.
	launchWorker := func(startFrom *Roster, workerID int) {
		wg.Add(1)
		go func() {
			sem <- struct{}{}
			defer wg.Done()
			defer func() { <-sem }()

			if config.Mode == "lahc" {
				lahcWorker(startFrom, sc, wd, hist, nurseSkills, forbidden, histLastShift,
					config, workerID, &globalBest, &bestMu, &bestRoster, branchChan, &stats, &statsMu)
			} else {
				saWorker(startFrom, sc, wd, hist, nurseSkills, forbidden, histLastShift,
					config, workerID, &globalBest, &bestMu, &bestRoster, branchChan, &stats, &statsMu)
			}
		}()
	}

	// Start first worker.
	totalWorkers++
	stats.WorkersStarted++
	launchWorker(initialRoster, 0)

	// Branch monitor: drain branchChan and spawn new workers.
	done := make(chan struct{})
	go func() {
		for branchRoster := range branchChan {
			statsMu.Lock()
			stats.BestUpdates++
			statsMu.Unlock()

			if totalWorkers >= config.MaxTotalWorkers {
				continue
			}
			totalWorkers++
			statsMu.Lock()
			stats.WorkersStarted++
			stats.BranchesCreated++
			statsMu.Unlock()
			launchWorker(branchRoster, totalWorkers)
		}
		close(done)
	}()

	// Wait for all workers to finish.
	wg.Wait()
	close(branchChan)
	<-done

	stats.DurationMs = time.Since(startTime).Milliseconds()
	stats.FinalPenalty = int(atomic.LoadInt64(&globalBest))

	// Convert best roster to solution.
	sol := RosterToSolution(bestRoster, sc, hist.Week)
	return sol, stats, nil
}
