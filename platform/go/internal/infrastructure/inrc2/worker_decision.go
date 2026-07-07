package inrc2

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// --- Worker Decision Engine ---
//
// Evaluates whether a worker looks worth running at spawn time (PFRS WorkerAssist layer).
// Modes via CLI --worker-decision-mode:
//   shadow: predictions recorded; all workers still run (worker_decisions.csv)
//   assist: recommendations acted upon with safety overrides (worker_assist.csv)
//   adaptive: assist with live-updating engine (same AssistMode path as assist in PFRS)

// Recommendation is the decision engine's suggested action.
type Recommendation string

const (
	RecRun            Recommendation = "run"
	RecSkip           Recommendation = "skip"
	RecReduceBudget   Recommendation = "reduce_budget"
	RecIncreaseBudget Recommendation = "increase_budget"
	RecChangeAlgo     Recommendation = "change_algorithm"
)

// WorkerDecisionInput captures the spawn context for decision-making.
type WorkerDecisionInput struct {
	Algorithm          string
	Week               int
	Depth              int
	ParentObjective    int
	GlobalBest         int
	BeamRank           int
	Entropy            float64
	BeamHealth         float64
	RecentImprovRate   float64
	FamilyID           int
	AllocatedIters     int
	WorkerCount        int
	ActiveFamilies     int
	DistanceFromBest   int

	// Lineage context: is this worker descended from a recent global best?
	IsGlobalBestLineage bool
	// How many generations since the last global best in this worker's lineage.
	GenerationsSinceGlobalBest int
	// Whether the parent itself produced a global best.
	ParentProducedGlobalBest bool
}

// WorkerDecision is the engine's output for a single worker spawn.
type WorkerDecision struct {
	Recommendation     Recommendation
	Confidence         float64 // 0.0 to 1.0
	ReasonCodes        []string
	SuggestedAlgorithm string // only if RecChangeAlgo
	SuggestedBudget    int    // only if RecReduceBudget/RecIncreaseBudget
}

// WorkerDecisionRecord combines the prediction with actual outcome for analysis.
type WorkerDecisionRecord struct {
	// Spawn context.
	WorkerID         int
	Week             int
	Depth            int
	Algorithm        string
	ParentObjective  int
	GlobalBest       int
	DistanceFromBest int
	BeamRank         int
	Entropy          float64
	BeamHealth       float64
	RecentImprovRate float64
	AllocatedIters   int

	// Decision.
	Recommendation     Recommendation
	Confidence         float64
	ReasonCodes        string // comma-separated
	SuggestedAlgorithm string
	SuggestedBudget    int

	// Actual outcome (filled after worker completes).
	Improved           bool
	ProducedGlobalBest bool
	ImprovementAmount  int
	FinalObjective     int
	RuntimeMs          int64
	ROI                float64
}

// WorkerDecisionEngine is the interface for worker spawn evaluation.
type WorkerDecisionEngine interface {
	// Evaluate returns a decision for the given spawn context.
	Evaluate(input WorkerDecisionInput) WorkerDecision
}

// --- Rule-Based Implementation ---

// RuleBasedWorkerDecisionEngine uses simple heuristic rules.
type RuleBasedWorkerDecisionEngine struct{}

// NewRuleBasedEngine creates a rule-based decision engine.
func NewRuleBasedEngine() *RuleBasedWorkerDecisionEngine {
	return &RuleBasedWorkerDecisionEngine{}
}

// Evaluate applies heuristic rules to the spawn context.
//
// Design principles (updated):
// - SKIP only fires when multiple independent negative signals agree.
// - No worker can be skipped if it is near a recent global best lineage.
// - No worker can be skipped if uncertainty is high (low confidence).
// - Parent-gap rules are advisory only (reduce_budget), never skip rules by themselves.
func (e *RuleBasedWorkerDecisionEngine) Evaluate(input WorkerDecisionInput) WorkerDecision {
	var reasons []string
	rec := RecRun
	confidence := 0.5
	suggestedAlgo := ""
	suggestedBudget := 0

	// --- Protective rules (prevent skip) ---

	// Protection 1: Never skip workers in global best lineage.
	if input.IsGlobalBestLineage || input.ParentProducedGlobalBest {
		rec = RecIncreaseBudget
		suggestedBudget = input.AllocatedIters * 2
		confidence = 0.8
		reasons = append(reasons, "global_best_lineage_protected")
		return WorkerDecision{
			Recommendation:     rec,
			Confidence:         confidence,
			ReasonCodes:        reasons,
			SuggestedAlgorithm: suggestedAlgo,
			SuggestedBudget:    suggestedBudget,
		}
	}

	// Protection 2: Never skip workers spawned from the current global best.
	if input.DistanceFromBest == 0 && input.ParentObjective == input.GlobalBest {
		rec = RecIncreaseBudget
		suggestedBudget = input.AllocatedIters * 2
		confidence = 0.7
		reasons = append(reasons, "spawned_from_global_best")
		return WorkerDecision{
			Recommendation:     rec,
			Confidence:         confidence,
			ReasonCodes:        reasons,
			SuggestedAlgorithm: suggestedAlgo,
			SuggestedBudget:    suggestedBudget,
		}
	}

	// --- Advisory signals (accumulate negative evidence) ---

	negativeSignals := 0
	var negativeReasons []string

	// Signal 1: Parent gap is large (advisory only — not a skip by itself).
	if input.GlobalBest > 0 && input.DistanceFromBest > 0 {
		gapPct := float64(input.DistanceFromBest) / float64(input.GlobalBest) * 100
		if gapPct > 50 {
			negativeSignals += 2
			negativeReasons = append(negativeReasons, fmt.Sprintf("parent_gap_%.0f_pct", gapPct))
		} else if gapPct > 25 {
			negativeSignals++
			negativeReasons = append(negativeReasons, fmt.Sprintf("parent_gap_%.0f_pct_reduce", gapPct))
		}
	}

	// Signal 2: Worker is deep in the tree with no recent improvement in lineage.
	if input.Depth > 0 && input.GenerationsSinceGlobalBest > 10 {
		negativeSignals++
		negativeReasons = append(negativeReasons, "stale_lineage")
	}

	// Signal 3: High worker count with diminishing returns.
	if input.WorkerCount > 50 && input.DistanceFromBest > 0 {
		gapPct := float64(input.DistanceFromBest) / float64(input.GlobalBest) * 100
		if gapPct > 35 {
			negativeSignals++
			negativeReasons = append(negativeReasons, "crowded_and_distant")
		}
	}

	// --- Decision logic ---
	// SKIP only when multiple independent negative signals agree (threshold: 3+).
	if negativeSignals >= 3 {
		rec = RecSkip
		confidence = 0.6 + float64(negativeSignals-3)*0.05
		if confidence > 0.85 {
			confidence = 0.85
		}
		reasons = append(reasons, negativeReasons...)
	} else if negativeSignals >= 1 {
		// Advisory: reduce budget but still run.
		rec = RecReduceBudget
		suggestedBudget = input.AllocatedIters / 2
		confidence = 0.55
		reasons = append(reasons, negativeReasons...)
	} else {
		// No negative signals — apply positive recommendations.

		// Prefer exploration if entropy is low.
		if input.Entropy > 0 && input.Entropy < 1.0 {
			rec = RecChangeAlgo
			suggestedAlgo = "lahc"
			confidence = 0.5
			reasons = append(reasons, "low_entropy_explore")
		}

		// Prefer exploitation if recent improvement rate is high.
		if input.RecentImprovRate > 2.0 {
			rec = RecIncreaseBudget
			suggestedBudget = input.AllocatedIters * 2
			confidence = 0.6
			reasons = append(reasons, "high_improv_rate_exploit")
		}
	}

	// Default reason if none triggered.
	if len(reasons) == 0 {
		reasons = append(reasons, "default_run")
	}

	return WorkerDecision{
		Recommendation:     rec,
		Confidence:         confidence,
		ReasonCodes:        reasons,
		SuggestedAlgorithm: suggestedAlgo,
		SuggestedBudget:    suggestedBudget,
	}
}

// --- Shadow Mode Recorder ---

// ShadowRecorder collects decisions and outcomes for analysis.
type ShadowRecorder struct {
	mu      sync.Mutex
	records []WorkerDecisionRecord
}

// NewShadowRecorder creates a new shadow mode recorder.
func NewShadowRecorder() *ShadowRecorder {
	return &ShadowRecorder{}
}

// RecordDecision stores a decision at spawn time. Returns the index for later outcome update.
func (sr *ShadowRecorder) RecordDecision(workerID int, input WorkerDecisionInput, decision WorkerDecision) int {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	idx := len(sr.records)
	sr.records = append(sr.records, WorkerDecisionRecord{
		WorkerID:           workerID,
		Week:               input.Week,
		Depth:              input.Depth,
		Algorithm:          input.Algorithm,
		ParentObjective:    input.ParentObjective,
		GlobalBest:         input.GlobalBest,
		DistanceFromBest:   input.DistanceFromBest,
		BeamRank:           input.BeamRank,
		Entropy:            input.Entropy,
		BeamHealth:         input.BeamHealth,
		RecentImprovRate:   input.RecentImprovRate,
		AllocatedIters:     input.AllocatedIters,
		Recommendation:     decision.Recommendation,
		Confidence:         decision.Confidence,
		ReasonCodes:        strings.Join(decision.ReasonCodes, ";"),
		SuggestedAlgorithm: decision.SuggestedAlgorithm,
		SuggestedBudget:    decision.SuggestedBudget,
	})
	return idx
}

// RecordOutcome updates a previously recorded decision with the actual result.
func (sr *ShadowRecorder) RecordOutcome(idx int, improved bool, producedGlobal bool, improvement int, finalObj int, runtimeMs int64) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if idx < 0 || idx >= len(sr.records) {
		return
	}
	r := &sr.records[idx]
	r.Improved = improved
	r.ProducedGlobalBest = producedGlobal
	r.ImprovementAmount = improvement
	r.FinalObjective = finalObj
	r.RuntimeMs = runtimeMs
	if runtimeMs > 0 {
		r.ROI = float64(improvement) / float64(runtimeMs) * 1000
	}
}

// Records returns all collected records.
func (sr *ShadowRecorder) Records() []WorkerDecisionRecord {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return append([]WorkerDecisionRecord{}, sr.records...)
}

// --- CSV Output ---

// WriteWorkerDecisionsCSV writes worker_decisions.csv.
func WriteWorkerDecisionsCSV(path string, records []WorkerDecisionRecord) error {
	if len(records) == 0 {
		return nil
	}

	header := strings.Join([]string{
		"worker_id", "week", "depth", "algorithm",
		"parent_objective", "global_best", "distance_from_best",
		"beam_rank", "entropy", "beam_health", "recent_improv_rate", "allocated_iters",
		"recommendation", "confidence", "reason_codes",
		"suggested_algorithm", "suggested_budget",
		"improved", "produced_global_best", "improvement_amount",
		"final_objective", "runtime_ms", "roi",
	}, ",")

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")

	for _, r := range records {
		improved := 0
		if r.Improved {
			improved = 1
		}
		pgb := 0
		if r.ProducedGlobalBest {
			pgb = 1
		}

		row := fmt.Sprintf("%d,%d,%d,%s,%d,%d,%d,%d,%.4f,%.4f,%.4f,%d,%s,%.2f,%s,%s,%d,%d,%d,%d,%d,%d,%.4f",
			r.WorkerID, r.Week, r.Depth, r.Algorithm,
			r.ParentObjective, r.GlobalBest, r.DistanceFromBest,
			r.BeamRank, r.Entropy, r.BeamHealth, r.RecentImprovRate, r.AllocatedIters,
			r.Recommendation, r.Confidence, r.ReasonCodes,
			r.SuggestedAlgorithm, r.SuggestedBudget,
			improved, pgb, r.ImprovementAmount,
			r.FinalObjective, r.RuntimeMs, r.ROI,
		)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
