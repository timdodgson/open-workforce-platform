package inrc2

import (
	"math"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// --- PFRS Configuration ---

// PFRSProgress holds a snapshot of PFRS execution state for progress reporting.
type PFRSProgress struct {
	WorkersStarted      int
	TotalWorkers        int // MaxTotalWorkers from config (0 = unlimited)
	ActiveWorkers       int
	QueueDepth          int
	CandidatesEvaluated int
	BestPenalty         int
	ElapsedMs           int64
}

// ProgressFunc is the callback signature for PFRS progress reporting.
// Called periodically during execution. Must be safe to call from a goroutine.
type ProgressFunc func(PFRSProgress)

// PFRSConfig holds all tunables for the Parallel Feasible Roster Search.
type PFRSConfig struct {
	Mode                 string  // "sa" or "lahc"
	IterationsPerWorker  int
	MaxConcurrentWorkers int
	MaxTotalWorkers      int
	BranchOnGlobalBest   bool
	InitialTemperature   float64
	CoolingRate          float64 // used when CoolingMode = "fixed-rate"
	CoolingMode          string  // "adaptive" or "fixed-rate"
	MinTemperature       float64
	LateAcceptanceLength int
	Seed                 int64
	Deterministic        bool
	ScoringMode          string // "official-penalty" or "soft-violation-count"
	OnProgress           ProgressFunc // optional progress callback
	ProgressIntervalMs   int64        // how often to call OnProgress (0 = disabled)
	OnAudit              AuditFunc    // optional audit callback; called once after execution completes

	// Reheat: reset temperature when stagnation is detected.
	ReheatEnabled              bool    // default true — enable stagnation-triggered reheating
	ReheatThreshold            int     // candidates without local best improvement before reheat (default 50000)
	ReheatFactor               float64 // fraction of InitialTemperature to reheat to (default 1.0 = full reset)
	ReheatMinCandidateFraction float64 // minimum fraction of budget before reheat is eligible (default 0.20)
}

// DefaultPFRSConfig returns sensible defaults matching Tim's dissertation parameters.
func DefaultPFRSConfig() PFRSConfig {
	return PFRSConfig{
		Mode:                 "sa",
		IterationsPerWorker:  500000,
		MaxConcurrentWorkers: 4,
		MaxTotalWorkers:      0, // 0 = unlimited (queue-based)
		BranchOnGlobalBest:   true,
		InitialTemperature:   100.0,
		CoolingRate:          0, // ignored in adaptive mode
		CoolingMode:          "adaptive",
		MinTemperature:       0.0001,
		LateAcceptanceLength: 1000,
		Seed:                 42,
		Deterministic:        true,
		ScoringMode:          "official-penalty",
		ReheatEnabled:              true,
		ReheatThreshold:            50000,
		ReheatFactor:               1.0,
		ReheatMinCandidateFraction: 0.20,
	}
}

// EffectiveCoolingRate computes the actual cooling rate used by SA workers.
// For "adaptive" mode, it derives the rate from initial/min temperature and iteration budget.
// For "fixed-rate" mode, it returns the configured CoolingRate.
func (c PFRSConfig) EffectiveCoolingRate() float64 {
	if c.CoolingMode == "fixed-rate" {
		return c.CoolingRate
	}
	// Adaptive: rate = 1 - pow(minTemp / initialTemp, 1 / iterations)
	if c.IterationsPerWorker <= 0 || c.InitialTemperature <= 0 || c.MinTemperature <= 0 {
		return c.CoolingRate // fallback
	}
	ratio := c.MinTemperature / c.InitialTemperature
	exponent := 1.0 / float64(c.IterationsPerWorker)
	return 1.0 - math.Pow(ratio, exponent)
}

// --- PFRS Statistics ---

// PFRSStats captures execution metrics for the PFRS algorithm.
type PFRSStats struct {
	WorkersStarted       int
	BranchesCreated      int
	BranchesDropped      int // only when MaxTotalWorkers > 0 and reached
	BestUpdates          int
	MaxQueueDepth        int
	MaxConcurrentSeen    int
	InvalidMovesRejected int
	TotalIterations      int
	CandidatesEvaluated  int
	ImprovementsAccepted int
	DurationMs           int64
	FinalPenalty         int
}

// --- Swap Operator ---

// swapNurses swaps two nurses' assignments on the same day.
// Returns RejectNoop/RejectSkill/RejectSuccession/RejectHistory if rejected, or -1 if accepted.
// histLastShift provides each nurse's last shift from history for day-0 succession checks.
func swapNurses(roster *Roster, nurseA, nurseB, dayIdx int, sc Scenario, nurseSkills []map[string]bool, forbidden map[string]bool, histLastShift []string) RejectReason {
	if roster == nil {
		return RejectNoop
	}
	aAssign := roster.Get(nurseA, dayIdx)
	bAssign := roster.Get(nurseB, dayIdx)

	// If both are off or both have the same assignment, swap is a no-op.
	if aAssign == bAssign {
		return RejectNoop
	}

	// Check skill validity after swap.
	// Nurse A would get B's assignment.
	if bAssign.ShiftType != "" && bAssign.Skill != "" {
		if !nurseSkills[nurseA][bAssign.Skill] {
			return RejectSkill
		}
	}
	// Nurse B would get A's assignment.
	if aAssign.ShiftType != "" && aAssign.Skill != "" {
		if !nurseSkills[nurseB][aAssign.Skill] {
			return RejectSkill
		}
	}

	// Check forbidden succession for nurse A with B's shift.
	reason := successionRejectReason(roster, nurseA, dayIdx, bAssign.ShiftType, forbidden, histLastShift)
	if reason >= 0 {
		return reason
	}
	// Check forbidden succession for nurse B with A's shift.
	reason = successionRejectReason(roster, nurseB, dayIdx, aAssign.ShiftType, forbidden, histLastShift)
	if reason >= 0 {
		return reason
	}

	// Apply swap.
	roster.Set(nurseA, dayIdx, bAssign)
	roster.Set(nurseB, dayIdx, aAssign)
	return RejectReason(-1) // accepted
}

// successionRejectReason checks if placing newShift on nurseIdx at dayIdx
// would violate forbidden successions. Returns the specific rejection reason
// or -1 if valid. Safe against nil roster or out-of-bounds access.
func successionRejectReason(roster *Roster, nurseIdx, dayIdx int, newShift string, forbidden map[string]bool, histLastShift []string) RejectReason {
	if roster == nil {
		return RejectSuccession // cannot validate — reject safely
	}

	// Check previous day → this day.
	if dayIdx > 0 {
		if nurseIdx < len(roster.Assignments) && dayIdx-1 < len(roster.Assignments[nurseIdx]) {
			prevShift := roster.Assignments[nurseIdx][dayIdx-1].ShiftType
			if prevShift != "" && newShift != "" {
				if forbidden[prevShift+"|"+newShift] {
					return RejectSuccession
				}
			}
		}
	} else {
		// Day 0: check history last shift → new shift.
		if nurseIdx < len(histLastShift) {
			prevShift := histLastShift[nurseIdx]
			if prevShift != "" && newShift != "" {
				if forbidden[prevShift+"|"+newShift] {
					return RejectHistory
				}
			}
		}
	}

	// Check this day → next day.
	if dayIdx < roster.NumDays-1 {
		if nurseIdx < len(roster.Assignments) && dayIdx+1 < len(roster.Assignments[nurseIdx]) {
			nextShift := roster.Assignments[nurseIdx][dayIdx+1].ShiftType
			if newShift != "" && nextShift != "" {
				if forbidden[newShift+"|"+nextShift] {
					return RejectSuccession
				}
			}
		}
	}

	return RejectReason(-1) // valid
}

// --- Scoring ---

// scorePenaltyWithMode computes the score based on the configured mode.
func scorePenaltyWithMode(roster *Roster, ws *ScoringWorkspace, mode string) int {
	if mode == "soft-violation-count" {
		sol := RosterToSolution(roster, ws.Sc, ws.Hist.Week)
		result := ScoreWith(ws, sol)
		return len(result.SoftDetails)
	}
	return ScorePenaltyOnlyFromRoster(ws, roster)
}

// ScorePenaltyOnlyFromRoster computes the official INRC-II soft penalty directly
// from a Roster without converting to Solution. Zero allocations in the hot path.
func ScorePenaltyOnlyFromRoster(ws *ScoringWorkspace, roster *Roster) int {
	sc := ws.Sc
	wd := ws.Wd
	numNurses := len(sc.Nurses)

	// Reuse pre-allocated nurse buffer — zero it first.
	nurses := ws.nurseBuffer
	for i := range nurses {
		nurses[i] = scoringNurseWeek{}
	}

	// Build assignment matrix directly from roster (no Solution conversion).
	for ni := 0; ni < numNurses && ni < len(roster.Assignments); ni++ {
		for d := 0; d < 7 && d < len(roster.Assignments[ni]); d++ {
			a := roster.Assignments[ni][d]
			if a.ShiftType != "" {
				nurses[ni].days[d] = assignEntry{shiftType: a.ShiftType, skill: a.Skill}
				nurses[ni].has[d] = true
			}
		}
	}

	penalty := 0

	// S1: Optimal coverage (30 per unit).
	for _, req := range wd.Requirements {
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			dayReq := req.RequirementForDay(dayIdx)
			if dayReq.Optimal <= dayReq.Minimum {
				continue
			}
			count := 0
			for ni := 0; ni < numNurses && ni < len(roster.Assignments); ni++ {
				if nurses[ni].has[dayIdx] && nurses[ni].days[dayIdx].shiftType == req.ShiftType && nurses[ni].days[dayIdx].skill == req.Skill {
					count++
				}
			}
			if count >= dayReq.Minimum && count < dayReq.Optimal {
				penalty += (dayReq.Optimal - count) * 30
			}
		}
	}

	// Per-nurse soft constraints — fully index-based, no allocations.
	for ni := 0; ni < numNurses; ni++ {
		contract := ws.ContractByIdx[ni]
		nh := ws.HistByIdx[ni]
		has := nurses[ni].has
		days := nurses[ni].days

		// Count working days — reuse pre-allocated buffer.
		workDays := ws.workDaysBuf[:0]
		for d := 0; d < 7; d++ {
			if has[d] {
				workDays = append(workDays, d)
			}
		}

		// S2: Consecutive working days.
		penalty += scoreConsecutiveWorkingDays(workDays, contract, nh)

		// S3: Consecutive days off.
		penalty += scoreConsecutiveDaysOff(workDays, contract, nh)

		// S4: Consecutive same shift type.
		penalty += scorePenaltyConsecutiveShift(days, has, sc.ShiftTypes, nh)

		// S6: Complete weekend.
		if contract.CompleteWeekends == 1 {
			satWorked := has[5]
			sunWorked := has[6]
			if satWorked != sunWorked {
				penalty += 30
			}
		}

		// S7: Total assignments.
		totalAssign := nh.NumberOfAssignments + len(workDays)
		penalty += scoreTotalAssignments(totalAssign, contract, sc.NumberOfWeeks, ws.Hist.Week)

		// S8: Total working weekends.
		weekendWorked := has[5] || has[6]
		totalWeekends := nh.NumberOfWorkingWeekends
		if weekendWorked {
			totalWeekends++
		}
		penalty += scoreTotalWorkingWeekends(totalWeekends, contract, sc.NumberOfWeeks, ws.Hist.Week)
	}

	// S5: Shift-off requests (10 per violation).
	for _, req := range wd.ShiftOffRequests {
		dayIdx := DayIndex(req.Day)
		if dayIdx < 0 || dayIdx > 6 {
			continue
		}
		ni, ok := ws.NurseIndex[req.Nurse]
		if !ok {
			continue
		}
		if !nurses[ni].has[dayIdx] {
			continue
		}
		entry := nurses[ni].days[dayIdx]
		if req.ShiftType == "Any" || req.ShiftType == entry.shiftType {
			penalty += 10
		}
	}

	return penalty
}

// ScorePenaltyOnly computes the official INRC-II soft penalty without building
// detail slices. This eliminates the allocation pressure that causes stack/GC
// crashes when called millions of times from concurrent workers.
//
// Zero heap allocations in the per-nurse loop — all lookups are index-based.
func ScorePenaltyOnly(ws *ScoringWorkspace, sol Solution) int {
	sc := ws.Sc
	wd := ws.Wd
	numNurses := len(sc.Nurses)

	// Reuse pre-allocated nurse buffer — zero it first.
	nurses := ws.nurseBuffer
	for i := range nurses {
		nurses[i] = scoringNurseWeek{}
	}

	for _, a := range sol.Assignments {
		dayIdx := DayIndex(a.Day)
		if dayIdx < 0 || dayIdx > 6 {
			continue
		}
		ni, ok := ws.NurseIndex[a.Nurse]
		if !ok {
			continue
		}
		nurses[ni].days[dayIdx] = assignEntry{shiftType: a.ShiftType, skill: a.Skill}
		nurses[ni].has[dayIdx] = true
	}

	penalty := 0

	// S1: Optimal coverage (30 per unit).
	for _, req := range wd.Requirements {
		for dayIdx := 0; dayIdx < 7; dayIdx++ {
			dayReq := req.RequirementForDay(dayIdx)
			if dayReq.Optimal <= dayReq.Minimum {
				continue
			}
			count := 0
			for _, a := range sol.Assignments {
				if DayIndex(a.Day) == dayIdx && a.ShiftType == req.ShiftType && a.Skill == req.Skill {
					count++
				}
			}
			if count >= dayReq.Minimum && count < dayReq.Optimal {
				penalty += (dayReq.Optimal - count) * 30
			}
		}
	}

	// Per-nurse soft constraints — fully index-based, no map reads in this loop.
	for ni := 0; ni < numNurses; ni++ {
		contract := ws.ContractByIdx[ni]
		nh := ws.HistByIdx[ni]
		has := nurses[ni].has
		days := nurses[ni].days

		// Count working days — reuse pre-allocated buffer.
		workDays := ws.workDaysBuf[:0]
		for d := 0; d < 7; d++ {
			if has[d] {
				workDays = append(workDays, d)
			}
		}

		// S2: Consecutive working days.
		penalty += scoreConsecutiveWorkingDays(workDays, contract, nh)

		// S3: Consecutive days off.
		penalty += scoreConsecutiveDaysOff(workDays, contract, nh)

		// S4: Consecutive same shift type.
		penalty += scorePenaltyConsecutiveShift(days, has, sc.ShiftTypes, nh)

		// S6: Complete weekend.
		if contract.CompleteWeekends == 1 {
			satWorked := has[5]
			sunWorked := has[6]
			if satWorked != sunWorked {
				penalty += 30
			}
		}

		// S7: Total assignments.
		totalAssign := nh.NumberOfAssignments + len(workDays)
		penalty += scoreTotalAssignments(totalAssign, contract, sc.NumberOfWeeks, ws.Hist.Week)

		// S8: Total working weekends.
		weekendWorked := has[5] || has[6]
		totalWeekends := nh.NumberOfWorkingWeekends
		if weekendWorked {
			totalWeekends++
		}
		penalty += scoreTotalWorkingWeekends(totalWeekends, contract, sc.NumberOfWeeks, ws.Hist.Week)
	}

	// S5: Shift-off requests (10 per violation).
	for _, req := range wd.ShiftOffRequests {
		dayIdx := DayIndex(req.Day)
		if dayIdx < 0 || dayIdx > 6 {
			continue
		}
		ni, ok := ws.NurseIndex[req.Nurse]
		if !ok {
			continue
		}
		if !nurses[ni].has[dayIdx] {
			continue
		}
		entry := nurses[ni].days[dayIdx]
		if req.ShiftType == "Any" || req.ShiftType == entry.shiftType {
			penalty += 10
		}
	}

	return penalty
}

// scorePenaltyConsecutiveShift evaluates S4 using fixed arrays instead of maps.
func scorePenaltyConsecutiveShift(days [7]assignEntry, has [7]bool, shiftTypes []ShiftType, nh NurseHistory) int {
	penalty := 0
	currentType := ""
	streak := 0

	// Initialize from history.
	if nh.LastAssignedShiftType != "" && nh.LastAssignedShiftType != "None" {
		if has[0] && days[0].shiftType == nh.LastAssignedShiftType {
			currentType = nh.LastAssignedShiftType
			streak = nh.NumberOfConsecutiveAssignments
		}
	}

	for d := 0; d < 7; d++ {
		if !has[d] {
			if streak > 0 && currentType != "" {
				penalty += penaltyForShiftStreak(currentType, streak, shiftTypes)
			}
			currentType = ""
			streak = 0
			continue
		}
		if days[d].shiftType == currentType {
			streak++
		} else {
			if streak > 0 && currentType != "" {
				penalty += penaltyForShiftStreak(currentType, streak, shiftTypes)
			}
			currentType = days[d].shiftType
			streak = 1
		}
	}

	// End of week: check max only.
	if streak > 0 && currentType != "" {
		for _, st := range shiftTypes {
			if st.ID == currentType {
				if st.MaximumNumberOfConsecutiveAssignments > 0 && streak > st.MaximumNumberOfConsecutiveAssignments {
					penalty += (streak - st.MaximumNumberOfConsecutiveAssignments) * 15
				}
				break
			}
		}
	}

	return penalty
}

func penaltyForShiftStreak(shiftType string, streak int, shiftTypes []ShiftType) int {
	for _, st := range shiftTypes {
		if st.ID == shiftType {
			p := 0
			if st.MaximumNumberOfConsecutiveAssignments > 0 && streak > st.MaximumNumberOfConsecutiveAssignments {
				p += (streak - st.MaximumNumberOfConsecutiveAssignments) * 15
			}
			if st.MinimumNumberOfConsecutiveAssignments > 0 && streak < st.MinimumNumberOfConsecutiveAssignments {
				p += (st.MinimumNumberOfConsecutiveAssignments - streak) * 15
			}
			return p
		}
	}
	return 0
}

// --- SA Worker ---

func saWorker(startRoster *Roster, sc Scenario, wd WeekData, hist History,
	nurseSkills []map[string]bool, forbidden map[string]bool, histLastShift []string,
	config PFRSConfig, workerID int, parentWorkerID int, globalBest *int64, bestMu *sync.Mutex,
	bestRoster **Roster, branchChan chan<- *Roster, stats *PFRSStats, statsMu *sync.Mutex,
	liveCandidates *int64, auditChan chan<- WorkerAudit, bestUpdateChan chan<- BestUpdateEvent, discoveryChan chan<- DiscoveryEvent, pfrsStart time.Time) {

	// Each worker gets its own scoring workspace — no shared mutable state.
	ws := NewScoringWorkspace(sc, wd, hist)

	rng := rand.New(rand.NewSource(config.Seed + int64(workerID)))
	roster := startRoster.Clone()
	currentPenalty := scorePenaltyWithMode(roster, ws, config.ScoringMode)
	localBest := currentPenalty
	localBestRoster := roster.Clone()

	// Audit state — observation only.
	audit := newWorkerAuditState(workerID, parentWorkerID, currentPenalty)

	// Plateau detection — pure observation, no behaviour change.
	plateau := newPlateauObserver(workerID, parentWorkerID, 0, 0, config.InitialTemperature)

	numNurses := len(roster.NurseIDs)
	temperature := config.InitialTemperature
	coolingRate := config.EffectiveCoolingRate()
	candidates := 0
	accepted := 0
	rejected := 0
	attempts := 0

	// Reheat state: track stagnation for temperature reset.
	lastLocalBestCand := 0 // candidate number of last local best improvement
	reheatCount := 0
	hasProducedBranch := false
	reheatMinCandidate := int(float64(config.IterationsPerWorker) * config.ReheatMinCandidateFraction)

	for candidates < config.IterationsPerWorker {
		attempts++

		// Generate a random swap: pick a day, pick two nurses.
		day := rng.Intn(7)
		nurseA := rng.Intn(numNurses)
		nurseB := rng.Intn(numNurses)
		if nurseA == nurseB {
			nurseB = (nurseA + 1) % numNurses
		}

		// Try swap — hard constraint check.
		aOld := roster.Get(nurseA, day)
		bOld := roster.Get(nurseB, day)

		rejectReason := swapNurses(roster, nurseA, nurseB, day, sc, nurseSkills, forbidden, histLastShift)
		if rejectReason >= 0 {
			rejected++
			audit.recordReject(rejectReason)
			// Hard rejection: does NOT count as a candidate, does NOT cool temperature.
			continue
		}

		// Swap passed hard checks — this is a candidate iteration.
		candidates++
		atomic.AddInt64(liveCandidates, 1)

		// Score the new roster.
		newPenalty := scorePenaltyWithMode(roster, ws, config.ScoringMode)
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

			// SA-specific audit: categorise acceptance.
			if delta <= 0 {
				audit.acceptedBetter++
			} else {
				audit.acceptedWorse++
			}

			if currentPenalty < localBest {
				previousLocalBest := localBest
				localBest = currentPenalty
				localBestRoster = roster.Clone()
				lastLocalBestCand = candidates

				// Audit: track local best.
				audit.bestPenalty = localBest
				audit.bestIteration = candidates
				audit.tempAtBest = temperature

				// Plateau: reset stagnation counter on improvement.
				plateau.recordImprovement(candidates)

				// Discovery instrumentation: record local best event.
				isGlobalBest := false

				// Check global best.
				gb := atomic.LoadInt64(globalBest)
				if int64(localBest) < gb {
					bestMu.Lock()
					if int64(localBest) < atomic.LoadInt64(globalBest) {
						oldGlobal := int(atomic.LoadInt64(globalBest))
						atomic.StoreInt64(globalBest, int64(localBest))
						*bestRoster = localBestRoster.Clone()
						isGlobalBest = true

						// Audit: record best-update event.
						if bestUpdateChan != nil {
							bestUpdateChan <- BestUpdateEvent{
								TimestampMs: time.Since(pfrsStart).Milliseconds(),
								WorkerID:    workerID,
								OldPenalty:  oldGlobal,
								NewPenalty:  localBest,
								Iteration:   candidates,
							}
						}

						// Signal branch.
						if config.BranchOnGlobalBest && branchChan != nil {
							select {
							case branchChan <- localBestRoster.Clone():
								hasProducedBranch = true
							default:
							}
						}
					}
					bestMu.Unlock()
				}

				// Emit discovery event.
				if discoveryChan != nil {
					eventType := "LOCAL_BEST"
					if isGlobalBest {
						eventType = "GLOBAL_BEST"
					}
					discoveryChan <- DiscoveryEvent{
						TimestampMs:        time.Since(pfrsStart).Milliseconds(),
						WorkerID:           workerID,
						Candidate:          candidates,
						Temperature:        temperature,
						CurrentPenalty:     localBest,
						PreviousBest:       previousLocalBest,
						NewBest:            localBest,
						Improvement:        previousLocalBest - localBest,
						EventType:          eventType,
						AcceptedWorseCount: audit.acceptedWorse,
						HardRejectCount:    audit.rejected,
						SoftRejectCount:    audit.rejectedByProb,
					}
				}
			}
		} else {
			// SA-specific audit: rejected by probability.
			audit.rejectedByProb++
			// Undo swap.
			roster.Set(nurseA, day, aOld)
			roster.Set(nurseB, day, bOld)
		}

		// Geometric cooling — only on scored candidates.
		temperature *= (1 - coolingRate)
		if temperature < config.MinTemperature {
			temperature = config.MinTemperature
		}

		// Plateau observation — after cooling, before next candidate.
		plateau.observe(candidates, temperature, currentPenalty, localBest, atomic.LoadInt64(globalBest))

		// Reheat: if stagnation threshold exceeded and eligibility met, reset temperature.
		if config.ReheatEnabled && config.ReheatThreshold > 0 &&
			hasProducedBranch &&
			candidates >= reheatMinCandidate &&
			candidates-lastLocalBestCand >= config.ReheatThreshold {

			tempBefore := temperature
			temperature = config.InitialTemperature * config.ReheatFactor
			lastLocalBestCand = candidates // reset counter
			reheatCount++

			// Record reheat as a discovery event for dashboard visibility.
			if discoveryChan != nil {
				discoveryChan <- DiscoveryEvent{
					TimestampMs:        time.Since(pfrsStart).Milliseconds(),
					WorkerID:           workerID,
					Candidate:          candidates,
					Temperature:        tempBefore, // temperature before reheat
					CurrentPenalty:     currentPenalty,
					PreviousBest:       localBest,
					NewBest:            localBest,
					Improvement:        0,
					EventType:          "REHEAT",
					AcceptedWorseCount: audit.acceptedWorse,
					HardRejectCount:    audit.rejected,
					SoftRejectCount:    audit.rejectedByProb,
				}
			}
		}
	}

	// Update shared stats.
	statsMu.Lock()
	stats.TotalIterations += candidates
	stats.CandidatesEvaluated += candidates
	stats.ImprovementsAccepted += accepted
	stats.InvalidMovesRejected += rejected
	statsMu.Unlock()
	_ = reheatCount // tracked for future audit extension

	// Emit audit record.
	audit.iterations = candidates
	audit.attempts = attempts
	audit.candidates = candidates
	audit.accepted = accepted
	audit.rejected = rejected
	audit.finalTemp = temperature
	if auditChan != nil {
		wa := audit.toAudit(currentPenalty)
		wa.Plateaus = plateau.events
		auditChan <- wa
	}
}

// --- LAHC Worker ---

func lahcWorker(startRoster *Roster, sc Scenario, wd WeekData, hist History,
	nurseSkills []map[string]bool, forbidden map[string]bool, histLastShift []string,
	config PFRSConfig, workerID int, parentWorkerID int, globalBest *int64, bestMu *sync.Mutex,
	bestRoster **Roster, branchChan chan<- *Roster, stats *PFRSStats, statsMu *sync.Mutex,
	liveCandidates *int64, auditChan chan<- WorkerAudit, bestUpdateChan chan<- BestUpdateEvent, discoveryChan chan<- DiscoveryEvent, pfrsStart time.Time) {

	// Each worker gets its own scoring workspace — no shared mutable state.
	ws := NewScoringWorkspace(sc, wd, hist)

	rng := rand.New(rand.NewSource(config.Seed + int64(workerID)))
	roster := startRoster.Clone()
	currentPenalty := scorePenaltyWithMode(roster, ws, config.ScoringMode)

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
	candidates := 0
	accepted := 0
	rejected := 0
	attempts := 0

	// Audit state — observation only.
	audit := newWorkerAuditState(workerID, parentWorkerID, currentPenalty)

	for candidates < config.IterationsPerWorker {
		attempts++
		v := candidates % histLen

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
			// Hard rejection: does NOT count as a candidate.
			continue
		}

		// Swap passed hard checks — this is a candidate iteration.
		candidates++
		atomic.AddInt64(liveCandidates, 1)

		newPenalty := scorePenaltyWithMode(roster, ws, config.ScoringMode)

		// LAHC acceptance: accept if better than current OR better than fitnessArray[v].
		acceptedByCurrent := newPenalty <= currentPenalty
		acceptedByLate := newPenalty <= fitnessArray[v]

		if acceptedByCurrent || acceptedByLate {
			currentPenalty = newPenalty
			accepted++

			// LAHC-specific audit.
			if acceptedByCurrent {
				audit.acceptedByCurrent++
			} else {
				audit.acceptedByLate++
			}

			if currentPenalty < localBest {
				previousLocalBest := localBest
				localBest = currentPenalty
				localBestRoster = roster.Clone()

				// Audit: track local best.
				audit.bestPenalty = localBest
				audit.bestIteration = candidates

				// Discovery instrumentation: record local best event.
				isGlobalBest := false

				gb := atomic.LoadInt64(globalBest)
				if int64(localBest) < gb {
					bestMu.Lock()
					if int64(localBest) < atomic.LoadInt64(globalBest) {
						oldGlobal := int(atomic.LoadInt64(globalBest))
						atomic.StoreInt64(globalBest, int64(localBest))
						*bestRoster = localBestRoster.Clone()
						isGlobalBest = true

						// Audit: record best-update event.
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
							select {
							case branchChan <- localBestRoster.Clone():
							default:
							}
						}
					}
					bestMu.Unlock()
				}

				// Emit discovery event.
				if discoveryChan != nil {
					eventType := "LOCAL_BEST"
					if isGlobalBest {
						eventType = "GLOBAL_BEST"
					}
					discoveryChan <- DiscoveryEvent{
						TimestampMs:        time.Since(pfrsStart).Milliseconds(),
						WorkerID:           workerID,
						Candidate:          candidates,
						Temperature:        0, // LAHC has no temperature
						CurrentPenalty:     localBest,
						PreviousBest:       previousLocalBest,
						NewBest:            localBest,
						Improvement:        previousLocalBest - localBest,
						EventType:          eventType,
						AcceptedWorseCount: audit.acceptedWorse,
						HardRejectCount:    audit.rejected,
						SoftRejectCount:    audit.rejectedByLate,
					}
				}
			}
		} else {
			// LAHC-specific audit: rejected by late score.
			audit.rejectedByLate++
			// Undo.
			roster.Set(nurseA, day, aOld)
			roster.Set(nurseB, day, bOld)
		}

		fitnessArray[v] = currentPenalty
	}

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
	audit.finalTemp = 0 // LAHC has no temperature
	if auditChan != nil {
		auditChan <- audit.toAudit(currentPenalty)
	}
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
	initialWs := NewScoringWorkspace(sc, wd, hist)
	initialPenalty := scorePenaltyWithMode(initialRoster, initialWs, config.ScoringMode)
	var globalBest int64 = int64(initialPenalty)
	var bestRoster *Roster = initialRoster.Clone()
	var bestMu sync.Mutex
	var stats PFRSStats
	var statsMu sync.Mutex

	// Branch channel: unbuffered — the branch monitor consumes immediately.
	// Workers send new global-best rosters here. The monitor decides whether to queue them.
	branchChan := make(chan *Roster, 256) // Large buffer so workers never block.

	// Audit channels — only allocated if OnAudit is set.
	// Drain concurrently to prevent deadlock when worker count exceeds buffer size.
	var auditChan chan WorkerAudit
	var bestUpdateChan chan BestUpdateEvent
	var workerAudits []WorkerAudit
	var bestUpdates []BestUpdateEvent
	var auditWg sync.WaitGroup
	if config.OnAudit != nil {
		auditChan = make(chan WorkerAudit, 256)
		bestUpdateChan = make(chan BestUpdateEvent, 256)

		// Drain audit channels concurrently so workers never block.
		auditWg.Add(2)
		go func() {
			defer auditWg.Done()
			for wa := range auditChan {
				workerAudits = append(workerAudits, wa)
			}
		}()
		go func() {
			defer auditWg.Done()
			for bu := range bestUpdateChan {
				bestUpdates = append(bestUpdates, bu)
			}
		}()
	}

	// Discovery channel — captures every local/global best improvement.
	var discoveryChan chan DiscoveryEvent
	var discoveries []DiscoveryEvent
	var discoveryWg sync.WaitGroup
	if config.OnAudit != nil {
		discoveryChan = make(chan DiscoveryEvent, 1024)
		discoveryWg.Add(1)
		go func() {
			defer discoveryWg.Done()
			for d := range discoveryChan {
				discoveries = append(discoveries, d)
			}
		}()
	}

	// === Worker Pool Coordinator ===
	// Fixed pool of NumCPU goroutines. Work items queued in a channel.
	// Every branch is preserved — nothing dropped, nothing discarded.
	// No goroutine-per-branch: eliminates Go deadlock detector false positives.

	maxConcurrent := config.MaxConcurrentWorkers
	cpus := runtime.NumCPU()
	if maxConcurrent > cpus {
		maxConcurrent = cpus
	}

	type workItem struct {
		roster         *Roster
		workerID       int
		parentWorkerID int
	}

	workQueue := make(chan workItem, 4096)
	var totalWorkers int64
	var activeWorkers int64
	var queueDepth int64
	var liveCandidates int64
	var pendingWork int64 // atomic: submitted - completed

	// Audit: collect branch events.
	var branchEvents []BranchEvent
	var branchEventsMu sync.Mutex

	// poolDone: signalled (non-blocking send) each time pendingWork hits zero.
	poolDone := make(chan struct{}, 1)

	// Fixed worker pool: NumCPU persistent goroutines.
	var poolWg sync.WaitGroup
	for i := 0; i < maxConcurrent; i++ {
		poolWg.Add(1)
		go func() {
			defer poolWg.Done()
			for item := range workQueue {
				atomic.AddInt64(&queueDepth, -1)
				atomic.AddInt64(&activeWorkers, 1)

				active := int(atomic.LoadInt64(&activeWorkers))
				statsMu.Lock()
				if active > stats.MaxConcurrentSeen {
					stats.MaxConcurrentSeen = active
				}
				statsMu.Unlock()

				if config.Mode == "lahc" {
					lahcWorker(item.roster, sc, wd, hist, nurseSkills, forbidden, histLastShift,
						config, item.workerID, item.parentWorkerID, &globalBest, &bestMu, &bestRoster,
						branchChan, &stats, &statsMu, &liveCandidates, auditChan, bestUpdateChan, discoveryChan, startTime)
				} else {
					saWorker(item.roster, sc, wd, hist, nurseSkills, forbidden, histLastShift,
						config, item.workerID, item.parentWorkerID, &globalBest, &bestMu, &bestRoster,
						branchChan, &stats, &statsMu, &liveCandidates, auditChan, bestUpdateChan, discoveryChan, startTime)
				}

				atomic.AddInt64(&activeWorkers, -1)
				if atomic.AddInt64(&pendingWork, -1) == 0 {
					select {
					case poolDone <- struct{}{}:
					default:
					}
				}
			}
		}()
	}

	// Submit work helper.
	submitWork := func(roster *Roster, wID int, parentID int) {
		atomic.AddInt64(&pendingWork, 1)
		atomic.AddInt64(&queueDepth, 1)
		workQueue <- workItem{roster: roster, workerID: wID, parentWorkerID: parentID}
	}

	// Start first work item.
	atomic.AddInt64(&totalWorkers, 1)
	statsMu.Lock()
	stats.WorkersStarted++
	statsMu.Unlock()
	submitWork(initialRoster, 0, -1)

	// Progress reporting.
	var progressStop chan struct{}
	if config.OnProgress != nil && config.ProgressIntervalMs > 0 {
		progressStop = make(chan struct{})
		go func() {
			ticker := time.NewTicker(time.Duration(config.ProgressIntervalMs) * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					statsMu.Lock()
					workersStarted := stats.WorkersStarted
					statsMu.Unlock()
					config.OnProgress(PFRSProgress{
						WorkersStarted:      workersStarted,
						TotalWorkers:        config.MaxTotalWorkers,
						ActiveWorkers:       int(atomic.LoadInt64(&activeWorkers)),
						QueueDepth:          int(atomic.LoadInt64(&queueDepth)),
						CandidatesEvaluated: int(atomic.LoadInt64(&liveCandidates)),
						BestPenalty:         int(atomic.LoadInt64(&globalBest)),
						ElapsedMs:           time.Since(startTime).Milliseconds(),
					})
				case <-progressStop:
					return
				}
			}
		}()
	}

	// Branch processor: reads branchChan, submits work items.
	// Runs until branchChan is closed.
	branchDone := make(chan struct{})
	go func() {
		for branchRoster := range branchChan {
			statsMu.Lock()
			stats.BestUpdates++
			statsMu.Unlock()

			if config.MaxTotalWorkers > 0 && int(atomic.LoadInt64(&totalWorkers)) >= config.MaxTotalWorkers {
				statsMu.Lock()
				stats.BranchesDropped++
				statsMu.Unlock()
				continue
			}

			currentQueue := int(atomic.LoadInt64(&queueDepth))
			statsMu.Lock()
			if currentQueue+1 > stats.MaxQueueDepth {
				stats.MaxQueueDepth = currentQueue + 1
			}
			statsMu.Unlock()

			parentID := int(atomic.LoadInt64(&totalWorkers)) - 1
			wID := int(atomic.AddInt64(&totalWorkers, 1))
			statsMu.Lock()
			stats.WorkersStarted++
			stats.BranchesCreated++
			statsMu.Unlock()

			if config.OnAudit != nil {
				branchEventsMu.Lock()
				branchEvents = append(branchEvents, BranchEvent{
					TimestampMs:  time.Since(startTime).Milliseconds(),
					ParentWorker: parentID,
					ChildWorker:  wID,
					Penalty:      int(atomic.LoadInt64(&globalBest)),
				})
				branchEventsMu.Unlock()
			}

			submitWork(branchRoster, wID, parentID)
		}
		close(branchDone)
	}()

	// Wait for all work to complete.
	// poolDone is signalled each time pendingWork reaches zero.
	// We re-check because the branch processor may add new work concurrently.
	for {
		<-poolDone
		// Brief pause to let branch processor submit any queued items.
		time.Sleep(time.Millisecond)
		if atomic.LoadInt64(&pendingWork) == 0 {
			break
		}
	}

	// All work done. Close branchChan (no workers running = no senders).
	close(branchChan)
	<-branchDone

	// Shut down worker pool.
	close(workQueue)
	poolWg.Wait()

	// Stop progress ticker.
	if progressStop != nil {
		close(progressStop)
	}

	stats.DurationMs = time.Since(startTime).Milliseconds()
	stats.FinalPenalty = int(atomic.LoadInt64(&globalBest))

	// Collect and deliver audit.
	if config.OnAudit != nil {
		// Close channels to signal drainer goroutines, then wait for them.
		close(auditChan)
		close(bestUpdateChan)
		auditWg.Wait()

		// Close discovery channel and wait for drain.
		close(discoveryChan)
		discoveryWg.Wait()

		// Compute branch depths from branch events.
		depthMap := map[int]int{0: 0} // workerID -> depth
		for _, be := range branchEvents {
			parentDepth := depthMap[be.ParentWorker]
			depthMap[be.ChildWorker] = parentDepth + 1
		}
		maxDepth := 0
		for _, d := range depthMap {
			if d > maxDepth {
				maxDepth = d
			}
		}

		// Determine winning worker and mark ImprovedParent / ProducedGlobal.
		winningWorkerID := 0
		for i := range workerAudits {
			wa := &workerAudits[i]
			if wa.BestPenalty < wa.StartPenalty {
				wa.ImprovedParent = true
			}
			if wa.BestPenalty == stats.FinalPenalty {
				wa.ProducedGlobal = true
				winningWorkerID = wa.WorkerID
			}
		}

		// Aggregate plateau events from all workers.
		var allPlateaus []PlateauEvent
		for _, wa := range workerAudits {
			allPlateaus = append(allPlateaus, wa.Plateaus...)
		}

		config.OnAudit(PFRSAudit{
			Workers:            workerAudits,
			Branches:           branchEvents,
			BestUpdates:        bestUpdates,
			Plateaus:           allPlateaus,
			Discoveries:        discoveries,
			MaxBranchDepth:     maxDepth,
			WinningWorkerID:    winningWorkerID,
			WinningBranchDepth: depthMap[winningWorkerID],
		})
	}

	// Convert best roster to solution.
	sol := RosterToSolution(bestRoster, sc, hist.Week)
	return sol, stats, nil
}
