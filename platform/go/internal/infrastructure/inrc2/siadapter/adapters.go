// Package siadapter maps between NRP worker telemetry shapes and generic SI CSV contracts.
// Lives under inrc2 because it adapts PFRS-specific recorder types to optimisation CSV schemas.
package siadapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

// AdaptSearchAssistToWorkerDecisions maps search-level checkpoints to worker_decisions.csv rows.
func AdaptSearchAssistToWorkerDecisions(records []optimisation.SearchAssistRecord) []inrc2.WorkerDecisionRecord {
	out := make([]inrc2.WorkerDecisionRecord, 0, len(records))
	for i, r := range records {
		out = append(out, inrc2.WorkerDecisionRecord{
			WorkerID:         i,
			Algorithm:        r.Algorithm,
			ParentObjective:  r.CurrentPenalty,
			GlobalBest:       r.BestPenalty,
			DistanceFromBest: r.CurrentPenalty - r.BestPenalty,
			Recommendation:   inrc2.Recommendation(r.RecommendedAction),
			Confidence:       float64(r.Confidence),
			ReasonCodes:      r.Reasons,
			Improved:         r.FinalBestPenalty < r.InitialPenalty,
			FinalObjective:   r.FinalBestPenalty,
			RuntimeMs:        r.RuntimeMs,
		})
	}
	return out
}

// AdaptSearchAssistToWorkerAssist maps search-level checkpoints to worker_assist.csv rows.
func AdaptSearchAssistToWorkerAssist(records []optimisation.SearchAssistRecord) []inrc2.AssistRecord {
	out := make([]inrc2.AssistRecord, 0, len(records))
	for i, r := range records {
		outcome := inrc2.AssistRejected
		if r.Accepted {
			outcome = inrc2.AssistAccepted
		}
		out = append(out, inrc2.AssistRecord{
			WorkerID:         i,
			Algorithm:        r.Algorithm,
			ParentObjective:  r.CurrentPenalty,
			GlobalBest:       r.BestPenalty,
			DistanceFromBest: r.CurrentPenalty - r.BestPenalty,
			Recommendation:   inrc2.Recommendation(r.RecommendedAction),
			Confidence:       float64(r.Confidence),
			ReasonCodes:      r.Reasons,
			SafetyTriggered:  r.SafetyTriggered,
			SafetyRule:       r.SafetyRule,
			Outcome:          outcome,
			FinalAction:      inrc2.Recommendation(r.FinalAction),
			Improved:         r.FinalBestPenalty < r.InitialPenalty,
			FinalObjective:   r.FinalBestPenalty,
			RuntimeMs:        r.RuntimeMs,
		})
	}
	return out
}

// AdaptWorkerDecisionsToSearchAssist maps NRP worker decisions to generic_search_assist.csv rows.
func AdaptWorkerDecisionsToSearchAssist(records []inrc2.WorkerDecisionRecord, defaultBudget int) []optimisation.SearchAssistRecord {
	out := make([]optimisation.SearchAssistRecord, 0, len(records))
	for i, r := range records {
		budget := workerBudget(r.AllocatedIters, r.SuggestedBudget, defaultBudget)
		plateau := r.DistanceFromBest
		if plateau <= 0 && !r.Improved {
			plateau = 1
		}
		rec := optimisation.SearchAssistRecord{
			Algorithm:         r.Algorithm,
			Checkpoint:        i,
			CurrentPenalty:    r.ParentObjective,
			BestPenalty:       r.GlobalBest,
			InitialPenalty:    r.ParentObjective,
			RecommendedAction: optimisation.SearchAction(r.Recommendation),
			Confidence:        optimisation.Confidence(r.Confidence),
			Reasons:           r.ReasonCodes,
			Accepted:          true,
			FinalAction:       optimisation.SearchAction(r.Recommendation),
			FinalBestPenalty:  r.FinalObjective,
			RuntimeMs:         r.RuntimeMs,
		}
		applyWorkerBudgetFields(&rec, budget, plateau)
		out = append(out, rec)
	}
	return out
}

// AdaptWorkerAssistToSearchAssist maps NRP worker assist to generic_search_assist.csv rows.
func AdaptWorkerAssistToSearchAssist(records []inrc2.AssistRecord, defaultBudget int) []optimisation.SearchAssistRecord {
	out := make([]optimisation.SearchAssistRecord, 0, len(records))
	for i, r := range records {
		budget := assistBudget(r, defaultBudget)
		plateau := r.DistanceFromBest
		if plateau <= 0 && !r.Improved {
			plateau = 1
		}
		rec := optimisation.SearchAssistRecord{
			Algorithm:         r.Algorithm,
			Checkpoint:        i,
			CurrentPenalty:    r.ParentObjective,
			BestPenalty:       r.GlobalBest,
			InitialPenalty:    r.ParentObjective,
			RecommendedAction: optimisation.SearchAction(r.Recommendation),
			Confidence:        optimisation.Confidence(r.Confidence),
			Reasons:           r.ReasonCodes,
			SafetyTriggered:   r.SafetyTriggered,
			SafetyRule:        r.SafetyRule,
			Accepted:          r.Outcome == inrc2.AssistAccepted,
			FinalAction:       optimisation.SearchAction(r.FinalAction),
			FinalBestPenalty:  r.FinalObjective,
			RuntimeMs:         r.RuntimeMs,
		}
		applyWorkerBudgetFields(&rec, budget, plateau)
		out = append(out, rec)
	}
	return out
}

func applyWorkerBudgetFields(rec *optimisation.SearchAssistRecord, budget int, plateauLength int) {
	if budget <= 0 {
		budget = 200000
	}
	rec.Candidates = budget
	rec.IterationsTotal = budget
	rec.PlateauLength = plateauLength
}

// BuildNRPPortfolioAssistRecords synthesises portfolio_assist.csv rows for NRP portfolio mode.
func BuildNRPPortfolioAssistRecords(instance string, seed int64, strategies []string, iterations int, bestPenalty int) []optimisation.PortfolioAssistRecord {
	if len(strategies) == 0 {
		return nil
	}
	perBudget := iterations
	if perBudget <= 0 {
		perBudget = 1
	}
	records := make([]optimisation.PortfolioAssistRecord, 0, len(strategies))
	for _, strat := range strategies {
		records = append(records, optimisation.PortfolioAssistRecord{
			Domain:            "nrp",
			Instance:          instance,
			Strategy:          strat,
			Seed:              seed,
			OriginalBudget:    perBudget,
			RecommendedBudget: perBudget,
			FinalBudget:       perBudget,
			Recommendation:    optimisation.PortfolioActionRun,
			Confidence:        1.0,
			ReasonCodes:       "nrp_portfolio_adapter",
			Accepted:          true,
			ResultObjective:   bestPenalty,
			StrategyWon:       false,
		})
	}
	if len(records) > 0 {
		records[0].StrategyWon = true
	}
	return records
}

// MergeSearchAssistRecords appends b into a without duplicates by checkpoint+algorithm.
func MergeSearchAssistRecords(a, b []optimisation.SearchAssistRecord) []optimisation.SearchAssistRecord {
	if len(b) == 0 {
		return a
	}
	seen := make(map[string]struct{}, len(a))
	for _, r := range a {
		seen[searchAssistKey(r)] = struct{}{}
	}
	for _, r := range b {
		k := searchAssistKey(r)
		if _, ok := seen[k]; ok {
			continue
		}
		a = append(a, r)
		seen[k] = struct{}{}
	}
	return a
}

func searchAssistKey(r optimisation.SearchAssistRecord) string {
	return r.Algorithm + ":" + strconv.Itoa(r.Checkpoint)
}

// EmitAdaptedWorkerCSVs writes worker-level CSVs from search records (CVRP/JSS/VRPTW parity).
func EmitAdaptedWorkerCSVs(outputDir string, assistMode string, searchRecords []optimisation.SearchAssistRecord) {
	if len(searchRecords) == 0 {
		return
	}
	if assistMode == "shadow" || assistMode == "" {
		rows := AdaptSearchAssistToWorkerDecisions(searchRecords)
		if len(rows) > 0 {
			_ = inrc2.WriteWorkerDecisionsCSV(filepath.Join(outputDir, "worker_decisions.csv"), rows)
		}
	}
	if assistMode == "assist" || assistMode == "adaptive" {
		rows := AdaptSearchAssistToWorkerAssist(searchRecords)
		if len(rows) > 0 {
			_ = inrc2.WriteWorkerAssistCSV(filepath.Join(outputDir, "worker_assist.csv"), rows)
		}
	}
}

// EmitAdaptedSearchCSV writes generic_search_assist.csv from NRP worker records.
func EmitAdaptedSearchCSV(outputDir string, defaultBudget int, decisionRecorder *inrc2.ShadowRecorder, assistRecorder *inrc2.AssistRecorder) bool {
	var records []optimisation.SearchAssistRecord
	if decisionRecorder != nil {
		records = MergeSearchAssistRecords(records, AdaptWorkerDecisionsToSearchAssist(decisionRecorder.Records(), defaultBudget))
	}
	if assistRecorder != nil {
		records = MergeSearchAssistRecords(records, AdaptWorkerAssistToSearchAssist(assistRecorder.Records(), defaultBudget))
	}
	if len(records) == 0 {
		return false
	}
	path := filepath.Join(outputDir, "generic_search_assist.csv")
	if err := optimisation.WriteSearchAssistCSV(path, records); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing generic_search_assist.csv: %v\n", err)
		return false
	}
	return true
}

func workerBudget(allocated, suggested, defaultBudget int) int {
	if allocated > 0 {
		return allocated
	}
	if suggested > 0 {
		return suggested
	}
	if defaultBudget > 0 {
		return defaultBudget
	}
	return 200000
}

func assistBudget(r inrc2.AssistRecord, defaultBudget int) int {
	if r.FinalBudget > 0 {
		return r.FinalBudget
	}
	if r.SuggestedBudget > 0 {
		return r.SuggestedBudget
	}
	return workerBudget(0, 0, defaultBudget)
}
