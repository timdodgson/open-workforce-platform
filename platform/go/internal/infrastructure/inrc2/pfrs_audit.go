package inrc2

import "time"

// --- PFRS Audit Instrumentation ---
// Pure observation layer. Does not alter algorithm behaviour, scoring, or acceptance decisions.

// RejectReason categorises why a swap was rejected by hard constraints.
type RejectReason int

const (
	RejectNoop       RejectReason = iota // same assignment, no-op
	RejectSkill                          // nurse lacks required skill
	RejectSuccession                     // forbidden shift succession (day > 0)
	RejectHistory                        // forbidden succession from history (day 0)
)

// WorkerAudit captures the search trajectory of a single PFRS worker.
type WorkerAudit struct {
	WorkerID       int
	ParentWorkerID int // -1 for initial worker
	StartPenalty   int
	FinalPenalty   int
	BestPenalty    int
	BestIteration  int // candidate iteration at which local best was found
	Iterations     int // same as CandidatesEval (scored moves)
	Attempts       int // total swap attempts (candidates + hard rejected)
	CandidatesEval int
	Accepted       int
	Rejected       int // hard-rejected (did not reach scoring)
	AcceptanceRate float64 // Accepted / CandidatesEval
	HardRejectRate float64 // Rejected / Attempts
	DurationMs     int64
	ImprovedParent bool // did this worker find a penalty < parent's start?
	ProducedGlobal bool // did this worker produce the final global best?

	// SA-specific.
	FinalTemperature float64 // temperature at end of run
	TempAtBest       float64 // temperature when best solution was found
	AcceptedBetter   int     // accepted moves with delta <= 0
	AcceptedWorse    int     // accepted moves with delta > 0 (by probability)
	RejectedByProb   int     // valid moves rejected by Metropolis probability

	// LAHC-specific.
	AcceptedByCurrent int // accepted because newPenalty <= currentPenalty
	AcceptedByLate    int // accepted because newPenalty <= fitnessArray[v] (but > current)
	RejectedByLate    int // valid moves rejected by both criteria

	// Rejection breakdown.
	RejectedNoop       int // same assignment
	RejectedSkill      int // skill mismatch
	RejectedSuccession int // forbidden succession (day > 0)
	RejectedHistory    int // forbidden succession from history (day 0)
}

// BranchEvent records when a new worker was spawned from a global-best update.
type BranchEvent struct {
	TimestampMs  int64
	ParentWorker int
	ChildWorker  int
	Penalty      int
	Depth        int // branch depth (0 = initial, 1 = child of initial, etc.)
}

// BestUpdateEvent records each time the global best is updated.
type BestUpdateEvent struct {
	TimestampMs int64
	WorkerID    int
	OldPenalty  int
	NewPenalty  int
	Iteration   int
}

// PFRSAudit holds the complete audit trail for one PFRS execution.
type PFRSAudit struct {
	Workers     []WorkerAudit
	Branches    []BranchEvent
	BestUpdates []BestUpdateEvent

	// Branching summary (computed at delivery).
	MaxBranchDepth     int
	WinningWorkerID    int // worker that produced final global best
	WinningBranchDepth int // depth of the winning worker
}

// AuditFunc is the callback signature for receiving the audit trail.
// Called once after PFRS completes. Must be safe to call from the RunPFRS goroutine.
type AuditFunc func(PFRSAudit)

// --- Internal mutable state used during worker execution ---

// workerAuditState holds mutable per-worker audit counters during execution.
type workerAuditState struct {
	workerID       int
	parentWorkerID int
	startPenalty   int
	bestPenalty    int
	bestIteration  int
	iterations     int
	attempts       int
	candidates     int
	accepted       int
	rejected       int
	startTime      time.Time

	// SA-specific.
	finalTemp      float64
	tempAtBest     float64
	acceptedBetter int
	acceptedWorse  int
	rejectedByProb int

	// LAHC-specific.
	acceptedByCurrent int
	acceptedByLate    int
	rejectedByLate    int

	// Rejection breakdown.
	rejectedNoop       int
	rejectedSkill      int
	rejectedSuccession int
	rejectedHistory    int
}

func newWorkerAuditState(workerID, parentWorkerID, startPenalty int) workerAuditState {
	return workerAuditState{
		workerID:       workerID,
		parentWorkerID: parentWorkerID,
		startPenalty:   startPenalty,
		bestPenalty:    startPenalty,
		startTime:      time.Now(),
	}
}

func (w *workerAuditState) recordReject(reason RejectReason) {
	w.rejected++
	switch reason {
	case RejectNoop:
		w.rejectedNoop++
	case RejectSkill:
		w.rejectedSkill++
	case RejectSuccession:
		w.rejectedSuccession++
	case RejectHistory:
		w.rejectedHistory++
	}
}

func (w *workerAuditState) toAudit(finalPenalty int) WorkerAudit {
	rate := 0.0
	if w.candidates > 0 {
		rate = float64(w.accepted) / float64(w.candidates)
	}
	hardRejectRate := 0.0
	if w.attempts > 0 {
		hardRejectRate = float64(w.rejected) / float64(w.attempts)
	}
	return WorkerAudit{
		WorkerID:           w.workerID,
		ParentWorkerID:     w.parentWorkerID,
		StartPenalty:       w.startPenalty,
		FinalPenalty:       finalPenalty,
		BestPenalty:        w.bestPenalty,
		BestIteration:      w.bestIteration,
		Iterations:         w.iterations,
		Attempts:           w.attempts,
		CandidatesEval:     w.candidates,
		Accepted:           w.accepted,
		Rejected:           w.rejected,
		AcceptanceRate:     rate,
		HardRejectRate:     hardRejectRate,
		DurationMs:         time.Since(w.startTime).Milliseconds(),
		FinalTemperature:   w.finalTemp,
		TempAtBest:         w.tempAtBest,
		AcceptedBetter:     w.acceptedBetter,
		AcceptedWorse:      w.acceptedWorse,
		RejectedByProb:     w.rejectedByProb,
		AcceptedByCurrent:  w.acceptedByCurrent,
		AcceptedByLate:     w.acceptedByLate,
		RejectedByLate:     w.rejectedByLate,
		RejectedNoop:       w.rejectedNoop,
		RejectedSkill:      w.rejectedSkill,
		RejectedSuccession: w.rejectedSuccession,
		RejectedHistory:    w.rejectedHistory,
	}
}

// WeekAuditRow holds the per-week metrics for CSV export.
type WeekAuditRow struct {
	// Run-level context (repeated per row for flat CSV).
	Instance           string
	Seed               int64
	Mode               string
	IterationsPerWorker int
	MaxTotalWorkers    int
	MaxConcurrent      int
	InitialTemperature float64
	CoolingRate        float64
	CoolingMode        string
	EffectiveCoolingRate float64
	MinTemperature     float64
	LateAcceptanceLen  int

	// Per-week metrics.
	Week             int
	StartPenalty     int
	FinalPenalty     int
	Improvement      int
	HardViolations   int
	SoftViolations   int
	Candidates       int
	Accepted         int
	Rejected         int
	AcceptanceRate   float64
	BestIteration    int
	BestWorkerID     int
	WorkersStarted   int
	BranchesCreated  int
	BranchesDropped  int
	MaxQueueDepth    int
	MaxConcurrentSeen int
	DurationMs       int64

	// SA-specific.
	SAFinalTemp      float64
	SATempAtBest     float64
	SAAcceptedBetter int
	SAAcceptedWorse  int
	SARejectedByProb int

	// LAHC-specific.
	LAHCAcceptedByCurrent int
	LAHCAcceptedByLate    int
	LAHCRejectedByLate    int

	// Branching.
	BranchesQueued     int
	BranchesStarted    int
	BranchesCompleted  int
	WinningBranchDepth int
	WorkersImproved    int // workers that improved over their parent
	WorkersProducedBest int // workers that produced the global best

	// Rejection breakdown.
	RejectedNoop       int
	RejectedSkill      int
	RejectedSuccession int
	RejectedHistory    int
}

// BuildWeekAuditRow constructs a CSV row from PFRS execution results.
func BuildWeekAuditRow(instance string, config PFRSConfig, week int,
	startPenalty int, stats PFRSStats, scoreResult ScoreResult, audit PFRSAudit) WeekAuditRow {

	row := WeekAuditRow{
		Instance:           instance,
		Seed:               config.Seed,
		Mode:               config.Mode,
		IterationsPerWorker: config.IterationsPerWorker,
		MaxTotalWorkers:    config.MaxTotalWorkers,
		MaxConcurrent:      config.MaxConcurrentWorkers,
		InitialTemperature: config.InitialTemperature,
		CoolingRate:          config.CoolingRate,
		CoolingMode:          config.CoolingMode,
		EffectiveCoolingRate: config.EffectiveCoolingRate(),
		MinTemperature:       config.MinTemperature,
		LateAcceptanceLen:  config.LateAcceptanceLength,

		Week:              week,
		StartPenalty:      startPenalty,
		FinalPenalty:      stats.FinalPenalty,
		Improvement:       startPenalty - stats.FinalPenalty,
		HardViolations:    scoreResult.HardViolations,
		SoftViolations:    len(scoreResult.SoftDetails),
		Candidates:        stats.CandidatesEvaluated,
		Accepted:          stats.ImprovementsAccepted,
		Rejected:          stats.InvalidMovesRejected,
		WorkersStarted:    stats.WorkersStarted,
		BranchesCreated:   stats.BranchesCreated,
		BranchesDropped:   stats.BranchesDropped,
		MaxQueueDepth:     stats.MaxQueueDepth,
		MaxConcurrentSeen: stats.MaxConcurrentSeen,
		DurationMs:        stats.DurationMs,

		BranchesQueued:    stats.BranchesCreated + stats.BranchesDropped,
		BranchesStarted:   stats.BranchesCreated,
		BranchesCompleted: stats.BranchesCreated, // all started workers run to completion
		WinningBranchDepth: audit.WinningBranchDepth,
	}

	// Compute acceptance rate.
	if row.Candidates > 0 {
		row.AcceptanceRate = float64(row.Accepted) / float64(row.Candidates)
	}

	// Aggregate per-worker metrics.
	for _, w := range audit.Workers {
		// SA-specific aggregation.
		row.SAAcceptedBetter += w.AcceptedBetter
		row.SAAcceptedWorse += w.AcceptedWorse
		row.SARejectedByProb += w.RejectedByProb

		// LAHC-specific aggregation.
		row.LAHCAcceptedByCurrent += w.AcceptedByCurrent
		row.LAHCAcceptedByLate += w.AcceptedByLate
		row.LAHCRejectedByLate += w.RejectedByLate

		// Rejection breakdown.
		row.RejectedNoop += w.RejectedNoop
		row.RejectedSkill += w.RejectedSkill
		row.RejectedSuccession += w.RejectedSuccession
		row.RejectedHistory += w.RejectedHistory

		// Best worker tracking.
		if w.BestPenalty == stats.FinalPenalty {
			row.BestWorkerID = w.WorkerID
			row.BestIteration = w.BestIteration
			row.SAFinalTemp = w.FinalTemperature
			row.SATempAtBest = w.TempAtBest
		}

		if w.ImprovedParent {
			row.WorkersImproved++
		}
		if w.ProducedGlobal {
			row.WorkersProducedBest++
		}
	}

	return row
}
