package inrc2

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// tabuMove represents a swap that is forbidden. We store the day and the two
// nurse indices involved. A move is tabu if the same (day, nurseA, nurseB)
// pair was used within the last TabuTenure iterations.
type tabuMove struct {
	day    int
	nurseA int
	nurseB int
}

// tabuList is a fixed-size circular buffer of recent moves.
type tabuList struct {
	moves  []tabuMove
	tenure int
	head   int
	size   int
}

func newTabuList(tenure int) *tabuList {
	return &tabuList{
		moves:  make([]tabuMove, tenure),
		tenure: tenure,
	}
}

func (t *tabuList) add(m tabuMove) {
	t.moves[t.head] = m
	t.head = (t.head + 1) % t.tenure
	if t.size < t.tenure {
		t.size++
	}
}

func (t *tabuList) contains(m tabuMove) bool {
	limit := t.size
	for i := 0; i < limit; i++ {
		stored := t.moves[i]
		// A move is tabu if it involves the same day and same pair (in either order).
		if stored.day == m.day &&
			((stored.nurseA == m.nurseA && stored.nurseB == m.nurseB) ||
				(stored.nurseA == m.nurseB && stored.nurseB == m.nurseA)) {
			return true
		}
	}
	return false
}

// tabuWorker implements Tabu Search within the PFRS framework.
// At each iteration it generates a random swap, checks if it's tabu,
// and accepts it if: (a) it's not tabu, OR (b) it satisfies the aspiration
// criterion (achieves a new best-ever penalty).
//
// Unlike SA/LAHC, Tabu Search always accepts moves (even worsening ones)
// as long as they're not forbidden. This forces systematic exploration.
func tabuWorker(startRoster *Roster, sc Scenario, wd WeekData, hist History,
	nurseSkills []map[string]bool, forbidden map[string]bool, histLastShift []string,
	config PFRSConfig, workerID int, parentWorkerID int, globalBest *int64, bestMu *sync.Mutex,
	bestRoster **Roster, branchChan chan<- *Roster, stats *PFRSStats, statsMu *sync.Mutex,
	liveCandidates *int64, auditChan chan<- WorkerAudit, bestUpdateChan chan<- BestUpdateEvent, discoveryChan chan<- DiscoveryEvent, pfrsStart time.Time) {

	ws := NewScoringWorkspace(sc, wd, hist)
	rng := rand.New(rand.NewSource(config.Seed + int64(workerID)))
	roster := startRoster.Clone()
	currentPenalty := scorePenaltyWithMode(roster, ws, config.ScoringMode)
	localBest := currentPenalty
	localBestRoster := roster.Clone()

	audit := newWorkerAuditState(workerID, parentWorkerID, currentPenalty)
	audit.algorithm = "tabu"
	plateau := newPlateauObserver(workerID, parentWorkerID, 0, 0, 0)

	numNurses := len(roster.NurseIDs)
	tenure := config.TabuTenure
	if tenure <= 0 {
		tenure = 7
	}
	tl := newTabuList(tenure)

	candidates := 0
	accepted := 0
	rejected := 0
	attempts := 0
	tabuRejected := 0
	lastBranchCandidate := 0

	// Best-move tabu: sample multiple swaps per iteration, pick the best feasible non-tabu one.
	neighbourhoodSize := 10 // Number of swaps to evaluate per iteration.

	for candidates < config.IterationsPerWorker {
		// Evaluate a neighbourhood of random swaps.
		type moveCandidate struct {
			day, nurseA, nurseB int
			aOld, bOld          ShiftAssignment
			penalty             int
			isTabu              bool
		}

		var bestMove *moveCandidate
		var bestTabuMove *moveCandidate // best move that's tabu (for aspiration)

		for n := 0; n < neighbourhoodSize; n++ {
			attempts++

			day := rng.Intn(7)
			nurseA := rng.Intn(numNurses)
			nurseB := rng.Intn(numNurses)
			if nurseA == nurseB {
				nurseB = (nurseA + 1) % numNurses
			}

			aOld := roster.Get(nurseA, day)
			bOld := roster.Get(nurseB, day)

			rejectReason := swapNurses(roster, nurseA, nurseB, day, sc, nurseSkills, forbidden, histLastShift)
			if rejectReason >= 0 {
				rejected++
				audit.recordReject(rejectReason)
				continue
			}

			// Score this swap.
			penalty := scorePenaltyWithMode(roster, ws, config.ScoringMode)
			move := tabuMove{day: day, nurseA: nurseA, nurseB: nurseB}
			isTabu := tl.contains(move)

			mc := &moveCandidate{day: day, nurseA: nurseA, nurseB: nurseB, aOld: aOld, bOld: bOld, penalty: penalty, isTabu: isTabu}

			if !isTabu {
				if bestMove == nil || penalty < bestMove.penalty {
					bestMove = mc
				}
			} else {
				if bestTabuMove == nil || penalty < bestTabuMove.penalty {
					bestTabuMove = mc
				}
			}

			// Undo swap — we're just evaluating, not committing yet.
			roster.Set(nurseA, day, aOld)
			roster.Set(nurseB, day, bOld)
		}

		// Pick the best move to commit.
		var chosen *moveCandidate

		if bestMove != nil {
			chosen = bestMove
		} else if bestTabuMove != nil && bestTabuMove.penalty < localBest {
			// Aspiration: accept tabu move if it achieves new local best.
			chosen = bestTabuMove
		}

		if chosen == nil {
			// All neighbourhood swaps were hard-rejected. Count as candidate with no progress.
			candidates++
			atomic.AddInt64(liveCandidates, 1)
			tabuRejected++
			audit.rejectedByProb++
			plateau.observe(candidates, 0, currentPenalty, localBest, atomic.LoadInt64(globalBest))
			continue
		}

		// Commit the chosen swap.
		candidates++
		atomic.AddInt64(liveCandidates, 1)

		// Re-apply the chosen swap.
		swapNurses(roster, chosen.nurseA, chosen.nurseB, chosen.day, sc, nurseSkills, forbidden, histLastShift)

		delta := chosen.penalty - currentPenalty
		if delta <= 0 {
			audit.acceptedBetter++
		} else {
			audit.acceptedWorse++
		}

		currentPenalty = chosen.penalty
		accepted++
		tl.add(tabuMove{day: chosen.day, nurseA: chosen.nurseA, nurseB: chosen.nurseB})

		if chosen.isTabu {
			tabuRejected-- // It was aspiration-accepted, not truly rejected.
		}

		if currentPenalty < localBest {
			previousLocalBest := localBest
			localBest = currentPenalty
			localBestRoster = roster.Clone()

			audit.bestPenalty = localBest
			audit.bestIteration = candidates
			audit.tempAtBest = 0

			plateau.recordImprovement(candidates)

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
					Temperature:        0,
					CurrentPenalty:     localBest,
					PreviousBest:       previousLocalBest,
					NewBest:            localBest,
					Improvement:        previousLocalBest - localBest,
					EventType:          eventType,
					AcceptedWorseCount: audit.acceptedWorse,
					HardRejectCount:    audit.rejected,
					SoftRejectCount:    tabuRejected,
				}:
				default:
				}
			}
		}

		// Plateau observation.
		plateau.observe(candidates, 0, currentPenalty, localBest, atomic.LoadInt64(globalBest))
	}

	// Update shared stats.
	statsMu.Lock()
	stats.TotalIterations += candidates
	stats.CandidatesEvaluated += candidates
	stats.ImprovementsAccepted += accepted
	stats.InvalidMovesRejected += rejected
	statsMu.Unlock()

	// Emit audit record.
	audit.iterations = candidates
	audit.attempts = attempts
	audit.candidates = candidates
	audit.accepted = accepted
	audit.rejected = rejected
	audit.finalTemp = 0
	if auditChan != nil {
		wa := audit.toAudit(currentPenalty)
		wa.Plateaus = plateau.events
		auditChan <- wa
	}
}
