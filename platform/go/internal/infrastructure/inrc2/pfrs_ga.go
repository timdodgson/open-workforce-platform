package inrc2

import (
	"math/rand"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultPFRSPopulationSize = 8
	defaultPFRSEliteCount     = 1
	defaultPFRSTournamentSize = 3
	defaultPFRSMutationMoves  = 3
	defaultPFRSWarmupMoves    = 2
)

// DefaultPortfolioStrategies is the standard PFRS portfolio when none is specified.
var DefaultPortfolioStrategies = []string{"sa", "lahc", "tabu", "ga"}

type pfrsGAIndividual struct {
	roster  *Roster
	fitness int
}

// gaWorker runs a population-based search within PFRS using the same candidate
// budget semantics as SA/LAHC (IterationsPerWorker = scored swap evaluations).
func gaWorker(startRoster *Roster, sc Scenario, wd WeekData, hist History,
	nurseSkills []map[string]bool, forbidden map[string]bool, histLastShift []string,
	config PFRSConfig, workerID int, parentWorkerID int, globalBest *int64, bestMu *sync.Mutex,
	bestRoster **Roster, branchChan chan<- *Roster, stats *PFRSStats, statsMu *sync.Mutex,
	liveCandidates *int64, auditChan chan<- WorkerAudit, bestUpdateChan chan<- BestUpdateEvent, discoveryChan chan<- DiscoveryEvent, pfrsStart time.Time) {

	ws := NewScoringWorkspace(sc, wd, hist)
	rng := rand.New(rand.NewSource(config.Seed + int64(workerID)))

	popSize := config.GAPopulationSize
	if popSize <= 0 {
		popSize = defaultPFRSPopulationSize
	}
	eliteCount := config.GAEliteCount
	if eliteCount <= 0 {
		eliteCount = defaultPFRSEliteCount
	}
	if eliteCount >= popSize {
		eliteCount = popSize / 4
		if eliteCount < 1 {
			eliteCount = 1
		}
	}
	tourneySize := config.GATournamentSize
	if tourneySize <= 0 {
		tourneySize = defaultPFRSTournamentSize
	}
	mutationMoves := config.GAMutationMoves
	if mutationMoves <= 0 {
		mutationMoves = defaultPFRSMutationMoves
	}

	audit := newWorkerAuditState(workerID, parentWorkerID, scorePenaltyWithMode(startRoster, ws, config.ScoringMode))
	audit.algorithm = "ga"

	pop := make([]pfrsGAIndividual, popSize)
	for i := range pop {
		pop[i].roster = startRoster.Clone()
		pop[i].fitness = scorePenaltyWithMode(pop[i].roster, ws, config.ScoringMode)
	}

	localBest := pop[0].fitness
	localBestRoster := pop[0].roster.Clone()
	lastBranchCandidate := 0
	candidates := 0
	accepted := 0
	rejected := 0
	attempts := 0
	budget := config.IterationsPerWorker

	// Light warmup on the first individual only — keeps startup cheap.
	pfrsGAGreedyMutate(pop[0].roster, &pop[0].fitness, defaultPFRSWarmupMoves, ws, config.ScoringMode, sc,
		nurseSkills, forbidden, histLastShift, rng, &audit, liveCandidates, &candidates, &accepted, &rejected, budget)
	if pop[0].fitness < localBest {
		localBest = pop[0].fitness
		localBestRoster = pop[0].roster.Clone()
	}

	for candidates < budget {
		sort.Slice(pop, func(i, j int) bool {
			return pop[i].fitness < pop[j].fitness
		})

		if pop[0].fitness < localBest {
			previousLocalBest := localBest
			localBest = pop[0].fitness
			localBestRoster = pop[0].roster.Clone()
			audit.bestPenalty = localBest
			audit.bestIteration = candidates
			candidates, accepted, rejected, lastBranchCandidate = pfrsGANotifyImprovement(
				localBest, previousLocalBest, localBestRoster, config, workerID, globalBest, bestMu, bestRoster,
				branchChan, candidates, accepted, rejected, lastBranchCandidate, discoveryChan, bestUpdateChan, pfrsStart, &audit,
			)
		}

		next := make([]pfrsGAIndividual, 0, popSize)
		for e := 0; e < eliteCount && e < len(pop); e++ {
			// Retain elite roster pointers — offspring are cloned before mutation.
			next = append(next, pop[e])
		}

		for len(next) < popSize && candidates < budget {
			attempts++
			parent := pfrsGATournamentSelect(pop, tourneySize, rng)
			childRoster := parent.roster.Clone()
			childFitness := parent.fitness
			pfrsGAGreedyMutate(childRoster, &childFitness, mutationMoves, ws, config.ScoringMode, sc,
				nurseSkills, forbidden, histLastShift, rng, &audit, liveCandidates, &candidates, &accepted, &rejected, budget)
			next = append(next, pfrsGAIndividual{roster: childRoster, fitness: childFitness})
		}
		pop = next
	}

	statsMu.Lock()
	stats.TotalIterations += candidates
	stats.CandidatesEvaluated += candidates
	stats.ImprovementsAccepted += accepted
	stats.InvalidMovesRejected += rejected
	statsMu.Unlock()

	audit.iterations = candidates
	audit.attempts = attempts
	audit.candidates = candidates
	audit.accepted = accepted
	audit.rejected = rejected
	if auditChan != nil {
		sort.Slice(pop, func(i, j int) bool { return pop[i].fitness < pop[j].fitness })
		auditChan <- audit.toAudit(pop[0].fitness)
	}
}

func pfrsGATournamentSelect(pop []pfrsGAIndividual, size int, rng *rand.Rand) pfrsGAIndividual {
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

func pfrsGAGreedyMutate(roster *Roster, fitness *int, maxMoves int, ws *ScoringWorkspace, scoringMode string, sc Scenario,
	nurseSkills []map[string]bool, forbidden map[string]bool, histLastShift []string, rng *rand.Rand,
	audit *workerAuditState, liveCandidates *int64, candidates, accepted, rejected *int, budget int) {
	current := *fitness
	numNurses := len(roster.NurseIDs)
	for m := 0; m < maxMoves && *candidates < budget; m++ {
		day := rng.Intn(7)
		nurseA := rng.Intn(numNurses)
		nurseB := rng.Intn(numNurses)
		if nurseA == nurseB {
			nurseB = (nurseA + 1) % numNurses
		}
		aOld := roster.Get(nurseA, day)
		bOld := roster.Get(nurseB, day)
		if rejectReason := swapNurses(roster, nurseA, nurseB, day, sc, nurseSkills, forbidden, histLastShift); rejectReason >= 0 {
			*rejected++
			audit.recordReject(rejectReason)
			continue
		}
		*candidates++
		atomic.AddInt64(liveCandidates, 1)
		newFit := scorePenaltyWithMode(roster, ws, scoringMode)
		if newFit <= current {
			current = newFit
			*accepted++
		} else {
			roster.Set(nurseA, day, aOld)
			roster.Set(nurseB, day, bOld)
		}
	}
	*fitness = current
}

func pfrsGANotifyImprovement(
	localBest, previousLocalBest int,
	localBestRoster *Roster,
	config PFRSConfig,
	workerID int,
	globalBest *int64,
	bestMu *sync.Mutex,
	bestRoster **Roster,
	branchChan chan<- *Roster,
	candidates, accepted, rejected, lastBranchCandidate int,
	discoveryChan chan<- DiscoveryEvent,
	bestUpdateChan chan<- BestUpdateEvent,
	pfrsStart time.Time,
	audit *workerAuditState,
) (int, int, int, int) {
	isGlobalBest := false
	gb := atomic.LoadInt64(globalBest)
	if int64(localBest) < gb {
		bestMu.Lock()
		if int64(localBest) < atomic.LoadInt64(globalBest) {
			oldGlobal := int(atomic.LoadInt64(globalBest))
			atomic.StoreInt64(globalBest, int64(localBest))
			*bestRoster = localBestRoster.Clone()
			isGlobalBest = true
			if bestUpdateChan != nil {
				bestUpdateChan <- BestUpdateEvent{
					TimestampMs: time.Since(pfrsStart).Milliseconds(),
					WorkerID:    workerID,
					OldPenalty:  oldGlobal,
					NewPenalty:  localBest,
					Iteration:   candidates,
				}
			}
			if config.BranchOnGlobalBest && branchChan != nil {
				cooldown := config.BranchCooldown
				if cooldown <= 0 || (candidates-lastBranchCandidate) >= cooldown {
					select {
					case branchChan <- localBestRoster.Clone():
						lastBranchCandidate = candidates
					default:
					}
				}
			}
		}
		bestMu.Unlock()
	}

	if discoveryChan != nil {
		eventType := "LOCAL_BEST"
		if isGlobalBest {
			eventType = "GLOBAL_BEST"
		}
		select {
		case discoveryChan <- DiscoveryEvent{
			TimestampMs:        time.Since(pfrsStart).Milliseconds(),
			WorkerID:           workerID,
			Candidate:          candidates,
			CurrentPenalty:     localBest,
			PreviousBest:       previousLocalBest,
			NewBest:            localBest,
			Improvement:        previousLocalBest - localBest,
			EventType:          eventType,
			AcceptedWorseCount: audit.acceptedWorse,
			HardRejectCount:    audit.rejected,
		}:
		default:
		}
	}
	return candidates, accepted, rejected, lastBranchCandidate
}
