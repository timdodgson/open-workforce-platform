package assist

import (
	"strings"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
)

// HookSnapshot exposes hook-runner state for policy engines in the parent package.
type HookSnapshot struct {
	ShadowMode        bool
	CheckpointNum     int
	IterationsTotal   int
	LastImproveAt     int
	MinBudgetFraction float64
}

// SearchHookRunner manages assist hooks for a single search run.
type SearchHookRunner struct {
	assist   *RuleBasedSearchAssist
	recorder *SearchAssistRecorder
	config   searchdef.SearchAssistConfig
	mode     string

	adaptiveAssist *AdaptiveSearchAssist

	checkpointNum   int
	lastImproveAt   int
	iterationsTotal int
	startTime       time.Time
}

// NewSearchHookRunner creates a hook runner. Returns nil if mode is "off".
func NewSearchHookRunner(mode string, config searchdef.SearchAssistConfig, iterationsTotal int) *SearchHookRunner {
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

// Snapshot returns policy-relevant hook state.
func (h *SearchHookRunner) Snapshot() HookSnapshot {
	if h == nil {
		return HookSnapshot{}
	}
	return HookSnapshot{
		ShadowMode:        h.mode == "shadow",
		CheckpointNum:     h.checkpointNum,
		IterationsTotal:   h.iterationsTotal,
		LastImproveAt:     h.lastImproveAt,
		MinBudgetFraction: h.config.MinBudgetFraction,
	}
}

// Mode returns the configured assist mode.
func (h *SearchHookRunner) Mode() string {
	if h == nil {
		return ""
	}
	return h.mode
}

// HasAdaptiveEngine reports whether the adaptive assist engine is active.
func (h *SearchHookRunner) HasAdaptiveEngine() bool {
	return h != nil && h.adaptiveAssist != nil
}

// OnImprovement records that an improvement happened at this candidate count.
func (h *SearchHookRunner) OnImprovement(candidates int) {
	if h == nil {
		return
	}
	h.lastImproveAt = candidates

	if h.adaptiveAssist != nil {
		improvRate := 0.0
		if candidates > 0 {
			gap := candidates - h.lastImproveAt
			if gap > 0 {
				improvRate = 10000.0 / float64(gap)
			} else {
				improvRate = 10.0
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
func (h *SearchHookRunner) RunCheckpoint(algorithm string, candidates int, currentPenalty int, bestPenalty int, initialPenalty int, temperature float64) searchdef.SearchAction {
	if h == nil {
		return searchdef.SearchContinue
	}

	plateauLength := candidates - h.lastImproveAt
	improvRate := 0.0
	if candidates > 0 {
		if plateauLength < candidates {
			improvRate = 10000.0 / float64(plateauLength+1)
		}
	}

	progress := searchdef.SearchProgress{
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

	var rec *searchdef.SearchRecommendation
	if h.mode == "adaptive" && h.adaptiveAssist != nil {
		rec = h.adaptiveAssist.Checkpoint(progress)
	} else {
		rec = h.assist.Checkpoint(progress)
	}

	recommendedAction := searchdef.SearchContinue
	confidence := searchdef.Confidence(0)
	reasons := ""
	if rec != nil {
		recommendedAction = rec.Action
		confidence = rec.Confidence
		reasons = strings.Join(rec.Reasons, ";")
	}

	safe, safetyRule := EvaluateSearchSafety(progress, rec, h.config)

	finalAction := searchdef.SearchContinue
	accepted := false

	if (h.mode == "assist" || h.mode == "adaptive") && rec != nil && safe {
		finalAction = rec.Action
		accepted = true
		if rec.Action == searchdef.SearchAdjustBudget && rec.NewBudget > 0 {
			h.iterationsTotal = rec.NewBudget
		}
	}

	record := searchdef.SearchAssistRecord{
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

// OnRestart resets plateau tracking after a policy-driven restart from best.
func (h *SearchHookRunner) OnRestart(candidates int) {
	if h == nil {
		return
	}
	h.lastImproveAt = candidates
}
