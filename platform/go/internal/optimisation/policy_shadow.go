// policy_shadow.go — Shadow Mode for Learned Policies.
//
// Runs learned policies alongside rules without influencing optimisation.
// Both predictions are recorded for every decision point.
//
// Flow:
//
//	Rule → prediction (this controls the optimiser)
//	Learned → prediction (recorded but NOT applied)
//	Compare → agreement/disagreement
//	Record → policy_shadow.csv
//
// Tracks: agreement rate, disagreement reasons, confidence, expected regret.
// This is purely evaluation — zero behaviour change.
package optimisation

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// ───────────────────────────────────────────────────────────────
// Shadow Record
// ───────────────────────────────────────────────────────────────

// ShadowRecord captures one decision point with both rule and learned predictions.
type ShadowRecord struct {
	// Identity
	Timestamp    time.Time
	RunID        string
	DecisionType string
	Domain       string
	Instance     string
	Algorithm    string

	// Rule decision (what actually happens).
	RuleAction     string
	RuleConfidence float64
	RuleReason     string

	// Learned decision (what would have happened).
	LearnedAction     string
	LearnedConfidence float64
	LearnedReason     string
	LearnedPolicyID   string
	LearnedVersion    string

	// Comparison
	Agreement      bool    // rule and learned agree on action
	ExpectedRegret float64 // estimated regret of rule vs learned (positive = learned is better)

	// Outcome (filled after execution with rule decision).
	Outcome float64
}

// ───────────────────────────────────────────────────────────────
// Shadow Metrics
// ───────────────────────────────────────────────────────────────

// ShadowMetrics summarises shadow evaluation performance.
type ShadowMetrics struct {
	TotalDecisions     int
	Agreements         int
	Disagreements      int
	AgreementRate      float64
	MeanLearnedConf    float64
	MeanExpectedRegret float64
	LearnedWouldWin    int // cases where learned prediction was better (post-hoc)
	LearnedWouldLose   int // cases where learned prediction was worse
}

// ───────────────────────────────────────────────────────────────
// Shadow Runner
// ───────────────────────────────────────────────────────────────

// PolicyShadowRunner evaluates learned policies in shadow mode.
// It does NOT change optimiser behaviour. Rules always control execution.
type PolicyShadowRunner struct {
	mu       sync.Mutex
	records  []ShadowRecord
	dir      string
	disabled bool
}

// NewPolicyShadowRunner creates a shadow runner.
// If dir is empty, recording is disabled (no file I/O).
func NewPolicyShadowRunner(dir string) *PolicyShadowRunner {
	if dir == "" {
		return &PolicyShadowRunner{disabled: true}
	}
	return &PolicyShadowRunner{dir: dir}
}

// Evaluate runs both rule and learned policies, records comparison, returns rule decision.
// The learned prediction is never applied — only recorded.
func (r *PolicyShadowRunner) Evaluate(
	ctx PolicyContext,
	rulePolicy Policy,
	learnedPolicy Policy,
) PolicyDecision {
	// Rule always controls execution.
	ruleDecision := rulePolicy.Decide(ctx)

	// Learned prediction (shadow — not applied).
	learnedDecision := learnedPolicy.Decide(ctx)

	// Compare.
	agreement := ruleDecision.Action == learnedDecision.Action

	// Estimate expected regret (positive = learned might be better).
	expectedRegret := 0.0
	if !agreement && learnedDecision.Confidence > ruleDecision.Confidence {
		expectedRegret = learnedDecision.Confidence - ruleDecision.Confidence
	}

	record := ShadowRecord{
		Timestamp:         time.Now(),
		RunID:             ctx.Features.RunID,
		DecisionType:      ctx.DecisionType,
		Domain:            ctx.Domain,
		Instance:          ctx.Instance,
		Algorithm:         ctx.Features.Algorithm,
		RuleAction:        ruleDecision.Action,
		RuleConfidence:    ruleDecision.Confidence,
		RuleReason:        ruleDecision.Reason,
		LearnedAction:     learnedDecision.Action,
		LearnedConfidence: learnedDecision.Confidence,
		LearnedReason:     learnedDecision.Reason,
		LearnedPolicyID:   learnedDecision.PolicyID,
		LearnedVersion:    learnedDecision.PolicyVersion,
		Agreement:         agreement,
		ExpectedRegret:    expectedRegret,
	}

	r.mu.Lock()
	r.records = append(r.records, record)
	r.mu.Unlock()

	// Always return rule decision — learned is shadow only.
	return ruleDecision
}

// RecordOutcome updates the most recent record with the observed outcome.
func (r *PolicyShadowRunner) RecordOutcome(outcome float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.records) > 0 {
		r.records[len(r.records)-1].Outcome = outcome
	}
}

// Metrics computes shadow evaluation metrics.
func (r *PolicyShadowRunner) Metrics() ShadowMetrics {
	r.mu.Lock()
	defer r.mu.Unlock()

	m := ShadowMetrics{TotalDecisions: len(r.records)}
	if m.TotalDecisions == 0 {
		return m
	}

	totalConf := 0.0
	totalRegret := 0.0

	for _, rec := range r.records {
		if rec.Agreement {
			m.Agreements++
		} else {
			m.Disagreements++
		}
		totalConf += rec.LearnedConfidence
		totalRegret += rec.ExpectedRegret
	}

	m.AgreementRate = float64(m.Agreements) / float64(m.TotalDecisions)
	m.MeanLearnedConf = totalConf / float64(m.TotalDecisions)
	m.MeanExpectedRegret = totalRegret / float64(m.TotalDecisions)

	return m
}

// Records returns all shadow records.
func (r *PolicyShadowRunner) Records() []ShadowRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]ShadowRecord{}, r.records...)
}

// ───────────────────────────────────────────────────────────────
// CSV Output
// ───────────────────────────────────────────────────────────────

// WriteCSV writes policy_shadow.csv.
func (r *PolicyShadowRunner) WriteCSV() error {
	if r.disabled {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.records) == 0 {
		return nil
	}

	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return fmt.Errorf("policy_shadow: mkdir: %w", err)
	}

	path := filepath.Join(r.dir, "policy_shadow.csv")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("policy_shadow: create: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	header := []string{
		"timestamp", "run_id", "decision_type", "domain", "instance", "algorithm",
		"rule_action", "rule_confidence", "rule_reason",
		"learned_action", "learned_confidence", "learned_reason",
		"learned_policy_id", "learned_version",
		"agreement", "expected_regret", "outcome",
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, rec := range r.records {
		agree := "0"
		if rec.Agreement {
			agree = "1"
		}
		row := []string{
			rec.Timestamp.Format(time.RFC3339),
			rec.RunID, rec.DecisionType, rec.Domain, rec.Instance, rec.Algorithm,
			rec.RuleAction, strconv.FormatFloat(rec.RuleConfidence, 'f', 4, 64), rec.RuleReason,
			rec.LearnedAction, strconv.FormatFloat(rec.LearnedConfidence, 'f', 4, 64), rec.LearnedReason,
			rec.LearnedPolicyID, rec.LearnedVersion,
			agree, strconv.FormatFloat(rec.ExpectedRegret, 'f', 4, 64),
			strconv.FormatFloat(rec.Outcome, 'f', 2, 64),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}
