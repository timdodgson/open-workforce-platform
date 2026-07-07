package optimisation

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// --- Generic Search Assist Hooks ---
//
// Adds optional AI advisory hooks to the generic search engine (SA, LAHC, Tabu).
// In off mode: no hooks, zero overhead.
// In shadow mode: records recommendations without changing behaviour.
// In assist mode: applies safe recommendations (early stop, budget adjust).

// SearchAssistConfig holds the configuration for search-level AI assistance.
type SearchAssistConfig struct {
	// Mode: "off", "shadow", "assist", or "adaptive".
	Mode string

	// CheckpointInterval: how often (in candidates) to call the assist engine.
	// Default: 10000 (every 10K candidates).
	CheckpointInterval int

	// MinBudgetFraction: minimum fraction of budget that must be consumed before
	// early stop is allowed. Default: 0.20.
	MinBudgetFraction float64

	// StagnationWindow: candidates without improvement before stagnation is flagged.
	// Default: 50000.
	StagnationWindow int

	// RecentImprovWindow: how far back to look for "recent improvement" (protects against
	// immediate stop after a discovery). Default: 5000.
	RecentImprovWindow int
}

// DefaultSearchAssistConfig returns the default configuration.
func DefaultSearchAssistConfig() SearchAssistConfig {
	return SearchAssistConfig{
		Mode:               "off",
		CheckpointInterval: 10000,
		MinBudgetFraction:  0.20,
		StagnationWindow:   50000,
		RecentImprovWindow: 5000,
	}
}

// SearchAssistRecord captures one assist checkpoint and its outcome.
type SearchAssistRecord struct {
	// Context at checkpoint.
	Algorithm       string
	Checkpoint      int // which checkpoint number (0, 1, 2, ...)
	Candidates      int
	IterationsTotal int
	CurrentPenalty  int
	BestPenalty     int
	InitialPenalty  int
	Temperature     float64
	PlateauLength   int
	ImprovementRate float64

	// Recommendation.
	RecommendedAction SearchAction
	Confidence        Confidence
	Reasons           string

	// Safety evaluation.
	SafetyTriggered bool
	SafetyRule      string

	// Final decision.
	Accepted    bool
	FinalAction SearchAction

	// Actual outcome (filled at end of search).
	FinalBestPenalty int
	TotalCandidates  int
	RuntimeMs        int64
}

// SearchAssistRecorder collects all assist decisions for a search run.
type SearchAssistRecorder struct {
	mu      sync.Mutex
	records []SearchAssistRecord
}

// NewSearchAssistRecorder creates a new recorder.
func NewSearchAssistRecorder() *SearchAssistRecorder {
	return &SearchAssistRecorder{}
}

// Record adds a checkpoint record.
func (r *SearchAssistRecorder) Record(rec SearchAssistRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.records = append(r.records, rec)
}

// FinaliseAll updates all records with the final search outcome.
func (r *SearchAssistRecorder) FinaliseAll(finalBest int, totalCandidates int, runtimeMs int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.records {
		r.records[i].FinalBestPenalty = finalBest
		r.records[i].TotalCandidates = totalCandidates
		r.records[i].RuntimeMs = runtimeMs
	}
}

// Records returns all collected records.
func (r *SearchAssistRecorder) Records() []SearchAssistRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]SearchAssistRecord{}, r.records...)
}

// --- Rule-Based SearchAssist Implementation ---

// RuleBasedSearchAssist implements SearchAssist using simple heuristic rules.
type RuleBasedSearchAssist struct {
	config SearchAssistConfig
}

// NewRuleBasedSearchAssist creates a rule-based search assist engine.
func NewRuleBasedSearchAssist(config SearchAssistConfig) *RuleBasedSearchAssist {
	return &RuleBasedSearchAssist{config: config}
}

// Checkpoint evaluates search progress and returns a recommendation.
func (a *RuleBasedSearchAssist) Checkpoint(progress SearchProgress) *SearchRecommendation {
	// Rule 1: Early stop if stagnating and past minimum budget.
	budgetUsed := float64(progress.IterationsComplete) / float64(progress.IterationsTotal)
	if budgetUsed >= a.config.MinBudgetFraction && progress.PlateauLength >= a.config.StagnationWindow {
		return &SearchRecommendation{
			Action:     SearchEarlyStop,
			Confidence: Confidence(0.6 + budgetUsed*0.2), // higher confidence later in search
			Reasons:    []string{fmt.Sprintf("stagnation_%d_cands", progress.PlateauLength), fmt.Sprintf("budget_%.0f_pct", budgetUsed*100)},
		}
	}

	// Rule 2: Reduce budget if acceptance rate is very low and no recent improvements.
	if budgetUsed >= 0.5 && progress.ImprovementRate < 0.01 && progress.PlateauLength > a.config.StagnationWindow/2 {
		newBudget := progress.IterationsComplete + (progress.IterationsTotal-progress.IterationsComplete)/2
		return &SearchRecommendation{
			Action:     SearchAdjustBudget,
			Confidence: 0.55,
			Reasons:    []string{"low_improvement_rate", "half_remaining"},
			NewBudget:  newBudget,
		}
	}

	// Rule 3: Extend budget if still improving near the end.
	if budgetUsed >= 0.9 && progress.ImprovementRate > 1.0 {
		newBudget := int(float64(progress.IterationsTotal) * 1.25)
		return &SearchRecommendation{
			Action:     SearchAdjustBudget,
			Confidence: 0.6,
			Reasons:    []string{"still_improving_near_end", "extend_25_pct"},
			NewBudget:  newBudget,
		}
	}

	return nil // continue as normal
}

// --- Safety Evaluation ---

// EvaluateSearchSafety checks whether a search assist recommendation is safe.
func EvaluateSearchSafety(progress SearchProgress, rec *SearchRecommendation, config SearchAssistConfig) (safe bool, rule string) {
	if rec == nil {
		return true, ""
	}

	// Safety 1: Never stop before minimum budget.
	if rec.Action == SearchEarlyStop {
		budgetUsed := float64(progress.IterationsComplete) / float64(progress.IterationsTotal)
		if budgetUsed < config.MinBudgetFraction {
			return false, "below_min_budget"
		}
	}

	// Safety 2: Never stop immediately after a recent improvement.
	if rec.Action == SearchEarlyStop && progress.PlateauLength < config.RecentImprovWindow {
		return false, "recent_improvement"
	}

	// Safety 3: Never reduce budget below minimum fraction.
	if rec.Action == SearchAdjustBudget && rec.NewBudget > 0 {
		minAllowed := int(float64(progress.IterationsTotal) * config.MinBudgetFraction)
		if rec.NewBudget < minAllowed {
			return false, "budget_below_minimum"
		}
	}

	return true, ""
}

// --- CSV Output ---

// WriteSearchAssistCSV writes generic_search_assist.csv.
func WriteSearchAssistCSV(path string, records []SearchAssistRecord) error {
	if len(records) == 0 {
		return nil
	}

	header := strings.Join([]string{
		"algorithm", "checkpoint", "candidates", "iterations_total",
		"current_penalty", "best_penalty", "initial_penalty",
		"temperature", "plateau_length", "improvement_rate",
		"recommended_action", "confidence", "reasons",
		"safety_triggered", "safety_rule",
		"accepted", "final_action",
		"final_best_penalty", "total_candidates", "runtime_ms",
	}, ",")

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteString("\n")

	for _, r := range records {
		accepted := 0
		if r.Accepted {
			accepted = 1
		}
		safety := 0
		if r.SafetyTriggered {
			safety = 1
		}

		row := fmt.Sprintf("%s,%d,%d,%d,%d,%d,%d,%.6f,%d,%.4f,%s,%.2f,%s,%d,%s,%d,%s,%d,%d,%d",
			r.Algorithm, r.Checkpoint, r.Candidates, r.IterationsTotal,
			r.CurrentPenalty, r.BestPenalty, r.InitialPenalty,
			r.Temperature, r.PlateauLength, r.ImprovementRate,
			r.RecommendedAction, r.Confidence, r.Reasons,
			safety, r.SafetyRule,
			accepted, r.FinalAction,
			r.FinalBestPenalty, r.TotalCandidates, r.RuntimeMs,
		)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// --- Search Hook Runner ---
//
// SearchHookRunner encapsulates the assist logic that gets called from within
// the search loops. It handles shadow vs assist mode, safety, and recording.

// SearchHookRunner manages assist hooks for a single search run.
type SearchHookRunner struct {
	assist   *RuleBasedSearchAssist
	recorder *SearchAssistRecorder
	config   SearchAssistConfig
	mode     string // "off", "shadow", "assist", "adaptive"

	// Adaptive engine (nil for non-adaptive modes).
	adaptiveAssist *AdaptiveSearchAssist

	// Tracking state.
	checkpointNum   int
	lastImproveAt   int // candidate number of last improvement
	iterationsTotal int // may be adjusted by assist
	startTime       time.Time
}

// NewSearchHookRunner creates a hook runner. Returns nil if mode is "off".
func NewSearchHookRunner(mode string, config SearchAssistConfig, iterationsTotal int) *SearchHookRunner {
	if mode == "" || mode == "off" {
		return nil
	}

	var adaptiveAssist *AdaptiveSearchAssist
	if mode == "adaptive" {
		adaptiveAssist = NewAdaptiveSearchAssist(config)
	}

	return &SearchHookRunner{
		assist:          NewRuleBasedSearchAssist(config),
		recorder:        NewSearchAssistRecorder(),
		config:          config,
		mode:            mode,
		iterationsTotal: iterationsTotal,
		adaptiveAssist:  adaptiveAssist,
		startTime:       time.Now(),
	}
}

// OnImprovement records that an improvement happened at this candidate count.
func (h *SearchHookRunner) OnImprovement(candidates int) {
	if h == nil {
		return
	}
	h.lastImproveAt = candidates

	// Feed adaptive engine with live improvement data.
	if h.adaptiveAssist != nil {
		improvRate := 0.0
		if candidates > 0 {
			gap := candidates - h.lastImproveAt
			if gap > 0 {
				improvRate = 10000.0 / float64(gap)
			} else {
				improvRate = 10.0 // fresh improvement
			}
		}
		h.adaptiveAssist.RecordImprovement(candidates, improvRate)
	}
}

// ShouldCheckpoint returns true if enough candidates have passed for a checkpoint.
func (h *SearchHookRunner) ShouldCheckpoint(candidates int) bool {
	if h == nil {
		return false
	}
	interval := h.config.CheckpointInterval
	if interval <= 0 {
		interval = 10000
	}
	return candidates > 0 && candidates%interval == 0
}

// RunCheckpoint evaluates the search state and returns an action.
// In shadow mode: always returns SearchContinue (but records the recommendation).
// In assist mode: returns the action to take (may be early stop or budget change).
func (h *SearchHookRunner) RunCheckpoint(algorithm string, candidates int, currentPenalty int, bestPenalty int, initialPenalty int, temperature float64) SearchAction {
	if h == nil {
		return SearchContinue
	}

	plateauLength := candidates - h.lastImproveAt
	improvRate := 0.0
	if candidates > 0 {
		// Simple approximation: improvements per 10K based on plateau.
		if plateauLength < candidates {
			improvRate = 10000.0 / float64(plateauLength+1)
		}
	}

	progress := SearchProgress{
		Algorithm:          algorithm,
		IterationsComplete: candidates,
		IterationsTotal:    h.iterationsTotal,
		CurrentPenalty:     currentPenalty,
		BestPenalty:        bestPenalty,
		InitialPenalty:     initialPenalty,
		ImprovementRate:    improvRate,
		Temperature:        temperature,
		PlateauLength:      plateauLength,
	}

	// Get recommendation from assist engine.
	var rec *SearchRecommendation
	if h.mode == "adaptive" && h.adaptiveAssist != nil {
		rec = h.adaptiveAssist.Checkpoint(progress)
	} else {
		rec = h.assist.Checkpoint(progress)
	}

	recommendedAction := SearchContinue
	confidence := Confidence(0)
	reasons := ""
	if rec != nil {
		recommendedAction = rec.Action
		confidence = rec.Confidence
		reasons = strings.Join(rec.Reasons, ";")
	}

	// Safety evaluation.
	safe, safetyRule := EvaluateSearchSafety(progress, rec, h.config)

	// Determine final action.
	finalAction := SearchContinue
	accepted := false

	if (h.mode == "assist" || h.mode == "adaptive") && rec != nil && safe {
		finalAction = rec.Action
		accepted = true
		// Update iterations total if budget was adjusted.
		if rec.Action == SearchAdjustBudget && rec.NewBudget > 0 {
			h.iterationsTotal = rec.NewBudget
		}
	}

	// Record the checkpoint.
	record := SearchAssistRecord{
		Algorithm:         algorithm,
		Checkpoint:        h.checkpointNum,
		Candidates:        candidates,
		IterationsTotal:   h.iterationsTotal,
		CurrentPenalty:    currentPenalty,
		BestPenalty:       bestPenalty,
		InitialPenalty:    initialPenalty,
		Temperature:       temperature,
		PlateauLength:     plateauLength,
		ImprovementRate:   improvRate,
		RecommendedAction: recommendedAction,
		Confidence:        confidence,
		Reasons:           reasons,
		SafetyTriggered:   !safe,
		SafetyRule:        safetyRule,
		Accepted:          accepted,
		FinalAction:       finalAction,
	}
	h.recorder.Record(record)
	h.checkpointNum++

	return finalAction
}

// GetIterationsTotal returns the current (possibly adjusted) iteration budget.
func (h *SearchHookRunner) GetIterationsTotal() int {
	if h == nil {
		return 0
	}
	return h.iterationsTotal
}

// Finalise records the search outcome on all records and returns the recorder.
func (h *SearchHookRunner) Finalise(bestPenalty int, totalCandidates int) *SearchAssistRecorder {
	if h == nil {
		return nil
	}
	runtimeMs := time.Since(h.startTime).Milliseconds()
	h.recorder.FinaliseAll(bestPenalty, totalCandidates, runtimeMs)
	return h.recorder
}
