package inrc2

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// --- Worker Assist (PFRS) ---
//
// When PFRSConfig.AssistMode is true (CLI --worker-decision-mode assist or adaptive),
// the WorkerDecisionEngine's recommendations are acted upon by the optimiser.
// Safety rules protect against critical failures (missed global bests).
// Every decision is logged: accepted, rejected, or overridden (worker_assist.csv).

// AssistOutcome describes what the optimiser did with the AI's recommendation.
type AssistOutcome string

const (
	AssistAccepted   AssistOutcome = "accepted"
	AssistRejected   AssistOutcome = "rejected"   // safety override prevented action
	AssistOverridden AssistOutcome = "overridden"  // optimiser chose differently
)

// AssistRecord captures one assist decision and its outcome.
type AssistRecord struct {
	// Worker context.
	WorkerID        int
	Week            int
	Depth           int
	Algorithm       string
	ParentObjective int
	GlobalBest      int
	DistanceFromBest int

	// AI recommendation.
	Recommendation     Recommendation
	Confidence         float64
	ReasonCodes        string
	SuggestedAlgorithm string
	SuggestedBudget    int

	// Safety evaluation.
	SafetyTriggered bool
	SafetyRule      string // which safety rule fired (empty if none)

	// Final decision.
	Outcome          AssistOutcome
	FinalAction      Recommendation // what the optimiser actually did
	FinalBudget      int
	FinalAlgorithm   string

	// Actual outcome (filled after worker completes).
	Improved           bool
	ProducedGlobalBest bool
	ImprovementAmount  int
	FinalObjective     int
	RuntimeMs          int64
}

// AssistRecorder collects all assist mode decisions for logging and analysis.
type AssistRecorder struct {
	mu      sync.Mutex
	records []AssistRecord
}

// NewAssistRecorder creates a new assist mode recorder.
func NewAssistRecorder() *AssistRecorder {
	return &AssistRecorder{}
}

// RecordDecision logs an assist decision at spawn time. Returns the index for outcome update.
func (ar *AssistRecorder) RecordDecision(record AssistRecord) int {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	idx := len(ar.records)
	ar.records = append(ar.records, record)
	return idx
}

// RecordOutcome updates a recorded assist decision with the actual result.
func (ar *AssistRecorder) RecordOutcome(idx int, improved bool, producedGlobal bool, improvement int, finalObj int, runtimeMs int64) {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	if idx < 0 || idx >= len(ar.records) {
		return
	}
	r := &ar.records[idx]
	r.Improved = improved
	r.ProducedGlobalBest = producedGlobal
	r.ImprovementAmount = improvement
	r.FinalObjective = finalObj
	r.RuntimeMs = runtimeMs
}

// Records returns all collected assist records.
func (ar *AssistRecorder) Records() []AssistRecord {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	return append([]AssistRecord{}, ar.records...)
}

// --- Safety Rules ---
//
// Hard safety rules that override AI recommendations.
// These ensure the optimiser never makes a decision that could miss a global best.

// AssistSafetyResult describes the outcome of safety evaluation.
type AssistSafetyResult struct {
	Safe       bool
	Rule       string
	Override   Recommendation
	OverBudget int
}

// EvaluateSafety checks whether an AI recommendation is safe to follow.
// Returns whether the recommendation should be overridden.
func EvaluateSafety(input WorkerDecisionInput, decision WorkerDecision) AssistSafetyResult {
	// Rule 1: Never skip workers in global-best lineage.
	if decision.Recommendation == RecSkip && (input.IsGlobalBestLineage || input.ParentProducedGlobalBest) {
		return AssistSafetyResult{
			Safe:       false,
			Rule:       "global_best_lineage",
			Override:   RecRun,
			OverBudget: input.AllocatedIters,
		}
	}

	// Rule 2: Never skip workers spawned from the current global best.
	if decision.Recommendation == RecSkip && input.DistanceFromBest == 0 {
		return AssistSafetyResult{
			Safe:       false,
			Rule:       "spawned_from_global_best",
			Override:   RecIncreaseBudget,
			OverBudget: input.AllocatedIters * 2,
		}
	}

	// Rule 3: Never skip with high uncertainty (low confidence).
	if decision.Recommendation == RecSkip && decision.Confidence < 0.65 {
		return AssistSafetyResult{
			Safe:       false,
			Rule:       "high_uncertainty",
			Override:   RecRun,
			OverBudget: input.AllocatedIters,
		}
	}

	// Rule 4: Never skip workers the model predicts likely to produce a global best.
	// (Encoded via IsGlobalBestLineage / distance — already covered above.)

	return AssistSafetyResult{Safe: true}
}

// --- CSV Output ---

// WriteWorkerAssistCSV writes worker_assist.csv.
func WriteWorkerAssistCSV(path string, records []AssistRecord) error {
	if len(records) == 0 {
		return nil
	}

	header := strings.Join([]string{
		"worker_id", "week", "depth", "algorithm",
		"parent_objective", "global_best", "distance_from_best",
		"recommendation", "confidence", "reason_codes",
		"suggested_algorithm", "suggested_budget",
		"safety_triggered", "safety_rule",
		"outcome", "final_action", "final_budget", "final_algorithm",
		"improved", "produced_global_best", "improvement_amount",
		"final_objective", "runtime_ms",
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
		safety := 0
		if r.SafetyTriggered {
			safety = 1
		}

		row := fmt.Sprintf("%d,%d,%d,%s,%d,%d,%d,%s,%.2f,%s,%s,%d,%d,%s,%s,%s,%d,%s,%d,%d,%d,%d,%d",
			r.WorkerID, r.Week, r.Depth, r.Algorithm,
			r.ParentObjective, r.GlobalBest, r.DistanceFromBest,
			r.Recommendation, r.Confidence, r.ReasonCodes,
			r.SuggestedAlgorithm, r.SuggestedBudget,
			safety, r.SafetyRule,
			r.Outcome, r.FinalAction, r.FinalBudget, r.FinalAlgorithm,
			improved, pgb, r.ImprovementAmount,
			r.FinalObjective, r.RuntimeMs,
		)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}
