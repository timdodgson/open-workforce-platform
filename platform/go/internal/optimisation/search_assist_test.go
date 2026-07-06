package optimisation

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
)

// --- SearchAssist Hook Tests ---

func TestSearchAssist_OffModeNoOverhead(t *testing.T) {
	// Off mode should produce nil hook runner and zero behaviour change.
	runner := NewSearchHookRunner("off", DefaultSearchAssistConfig(), 100000)
	if runner != nil {
		t.Error("Expected nil runner for off mode")
	}

	runner2 := NewSearchHookRunner("", DefaultSearchAssistConfig(), 100000)
	if runner2 != nil {
		t.Error("Expected nil runner for empty mode")
	}
}

func TestSearchAssist_ShadowModeRecordsWithoutAction(t *testing.T) {
	runner := NewSearchHookRunner("shadow", DefaultSearchAssistConfig(), 100000)
	if runner == nil {
		t.Fatal("Expected non-nil runner for shadow mode")
	}

	// Simulate a search that stagnates.
	// No improvement ever — plateau grows.
	for c := 10000; c <= 100000; c += 10000 {
		if runner.ShouldCheckpoint(c) {
			action := runner.RunCheckpoint("sa", c, 5000, 5000, 10000, 0.01)
			// Shadow mode should NEVER return anything other than continue.
			if action != SearchContinue {
				t.Errorf("Shadow mode returned action %s at checkpoint %d, expected continue", action, c)
			}
		}
	}

	recorder := runner.Finalise(5000, 100000)
	records := recorder.Records()
	if len(records) == 0 {
		t.Error("Expected records to be collected in shadow mode")
	}

	// Verify recommendations were recorded even though not acted upon.
	hasRec := false
	for _, r := range records {
		if r.RecommendedAction != SearchContinue {
			hasRec = true
			break
		}
	}
	if !hasRec {
		t.Error("Expected at least one non-continue recommendation to be recorded")
	}
}

func TestSearchAssist_AssistModeEarlyStop(t *testing.T) {
	config := DefaultSearchAssistConfig()
	config.StagnationWindow = 20000
	config.MinBudgetFraction = 0.2

	runner := NewSearchHookRunner("assist", config, 100000)
	if runner == nil {
		t.Fatal("Expected non-nil runner for assist mode")
	}

	// Simulate stagnation past min budget.
	stopped := false
	for c := 10000; c <= 100000; c += 10000 {
		if runner.ShouldCheckpoint(c) {
			action := runner.RunCheckpoint("sa", c, 5000, 5000, 10000, 0.01)
			if action == SearchEarlyStop {
				stopped = true
				break
			}
		}
	}

	if !stopped {
		t.Error("Expected assist mode to trigger early stop on stagnating search")
	}
}

func TestSearchAssist_SafetyBlocksEarlyStopBeforeMinBudget(t *testing.T) {
	config := DefaultSearchAssistConfig()
	config.StagnationWindow = 5000
	config.MinBudgetFraction = 0.5 // 50% minimum

	runner := NewSearchHookRunner("assist", config, 100000)
	if runner == nil {
		t.Fatal("Expected non-nil runner")
	}

	// At 10K/100K = 10% budget used, safety should block even if stagnating.
	action := runner.RunCheckpoint("sa", 10000, 5000, 5000, 10000, 0.01)
	if action == SearchEarlyStop {
		t.Error("Safety should block early stop before min budget fraction")
	}
}

func TestSearchAssist_SafetyBlocksStopAfterRecentImprovement(t *testing.T) {
	config := DefaultSearchAssistConfig()
	config.StagnationWindow = 5000
	config.RecentImprovWindow = 3000
	config.MinBudgetFraction = 0.1

	runner := NewSearchHookRunner("assist", config, 100000)
	if runner == nil {
		t.Fatal("Expected non-nil runner")
	}

	// Record improvement at candidate 48000.
	runner.OnImprovement(48000)

	// At 50000, only 2000 since improvement — should be protected.
	action := runner.RunCheckpoint("sa", 50000, 4000, 4000, 10000, 0.01)
	if action == SearchEarlyStop {
		t.Error("Safety should block early stop after recent improvement")
	}
}

func TestSearchAssist_CSVOutput(t *testing.T) {
	records := []SearchAssistRecord{
		{
			Algorithm: "sa", Checkpoint: 0, Candidates: 10000, IterationsTotal: 100000,
			CurrentPenalty: 5000, BestPenalty: 5000, InitialPenalty: 10000,
			Temperature: 50.0, PlateauLength: 10000, ImprovementRate: 0.0,
			RecommendedAction: SearchContinue, Confidence: 0.5,
			Accepted: false, FinalAction: SearchContinue,
			FinalBestPenalty: 4500, TotalCandidates: 100000, RuntimeMs: 500,
		},
		{
			Algorithm: "sa", Checkpoint: 5, Candidates: 60000, IterationsTotal: 100000,
			CurrentPenalty: 5000, BestPenalty: 5000, InitialPenalty: 10000,
			Temperature: 0.001, PlateauLength: 50000, ImprovementRate: 0.0,
			RecommendedAction: SearchEarlyStop, Confidence: 0.72, Reasons: "stagnation_50000_cands;budget_60_pct",
			SafetyTriggered: false,
			Accepted: true, FinalAction: SearchEarlyStop,
			FinalBestPenalty: 5000, TotalCandidates: 60000, RuntimeMs: 300,
		},
	}

	path := t.TempDir() + "/generic_search_assist.csv"
	err := WriteSearchAssistCSV(path, records)
	if err != nil {
		t.Fatalf("WriteSearchAssistCSV: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (header + 2 rows), got %d", len(lines))
	}
	if !strings.Contains(lines[2], "early_stop") {
		t.Error("Expected early_stop in second data row")
	}
}

// --- PortfolioAssist Tests ---

func TestPortfolioAssist_OffModeNoChange(t *testing.T) {
	// Off mode should return nil recorder and use RunPortfolio directly.
	problem := newTestProblem(10)
	config := SearchConfig{
		Mode:       "portfolio",
		Iterations: 1000,
		Portfolio:  []string{"sa", "lahc"},
		Seed:       42,
		InitialTemperature: 10.0,
		MinTemperature:     0.001,
		CoolingMode:        "adaptive",
		LateAcceptanceLength: 100,
	}

	result, recorder := RunPortfolioWithAssist(problem, config, PortfolioAssistConfig{Mode: "off"})
	if recorder != nil {
		t.Error("Expected nil recorder for off mode")
	}
	if result.BestResult.BestPenalty >= result.BestResult.InitialPenalty {
		t.Error("Expected some improvement from portfolio")
	}
}

func TestPortfolioAssist_ShadowModeRecords(t *testing.T) {
	problem := newTestProblem(10)
	config := SearchConfig{
		Mode:       "portfolio",
		Iterations: 1000,
		Portfolio:  []string{"sa", "lahc"},
		Seed:       42,
		InitialTemperature: 10.0,
		MinTemperature:     0.001,
		CoolingMode:        "adaptive",
		LateAcceptanceLength: 100,
	}

	result, recorder := RunPortfolioWithAssist(problem, config, PortfolioAssistConfig{
		Mode:     "shadow",
		Domain:   "test",
		Instance: "test-instance",
	})

	if recorder == nil {
		t.Fatal("Expected non-nil recorder for shadow mode")
	}

	records := recorder.Records()
	if len(records) != 2 {
		t.Errorf("Expected 2 records (one per strategy), got %d", len(records))
	}

	// Shadow mode should NOT change budgets.
	for _, r := range records {
		if r.FinalBudget != r.OriginalBudget {
			t.Errorf("Shadow mode changed budget for %s: %d -> %d", r.Strategy, r.OriginalBudget, r.FinalBudget)
		}
		if r.Accepted {
			t.Errorf("Shadow mode should not accept recommendations, strategy=%s", r.Strategy)
		}
	}

	// Result should still be valid.
	if result.BestResult.BestPenalty >= result.BestResult.InitialPenalty {
		t.Error("Expected improvement even in shadow mode")
	}
}

func TestPortfolioAssist_AssistModeAdjustsBudgets(t *testing.T) {
	problem := newTestProblem(10)
	config := SearchConfig{
		Mode:       "portfolio",
		Iterations: 10000,
		Portfolio:  []string{"sa", "lahc"},
		Seed:       42,
		InitialTemperature: 10.0,
		MinTemperature:     0.001,
		CoolingMode:        "adaptive",
		LateAcceptanceLength: 100,
	}

	_, recorder := RunPortfolioWithAssist(problem, config, PortfolioAssistConfig{
		Mode:     "assist",
		Domain:   "test",
		Instance: "test-instance",
	})

	if recorder == nil {
		t.Fatal("Expected non-nil recorder for assist mode")
	}

	records := recorder.Records()
	// At least one strategy should have accepted budget change.
	hasAccepted := false
	for _, r := range records {
		if r.Accepted {
			hasAccepted = true
			break
		}
	}
	if !hasAccepted {
		t.Error("Expected at least one accepted recommendation in assist mode")
	}
}

func TestPortfolioAssist_NeverSkipsAllStrategies(t *testing.T) {
	strategies := []string{"sa", "lahc", "tabu"}
	// Force all to skip via custom advice.
	advice := []StrategyAdvice{
		{Strategy: "sa", Action: PortfolioActionSkip, BudgetMult: 0, Confidence: 0.9},
		{Strategy: "lahc", Action: PortfolioActionSkip, BudgetMult: 0, Confidence: 0.9},
		{Strategy: "tabu", Action: PortfolioActionSkip, BudgetMult: 0, Confidence: 0.9},
	}

	safe, rule, fixed := EvaluatePortfolioSafety(strategies, advice)
	if safe {
		t.Error("Should not be safe to skip all strategies")
	}
	if rule != "cannot_skip_all" {
		t.Errorf("Expected rule 'cannot_skip_all', got '%s'", rule)
	}

	// Fixed advice should have all running.
	runCount := 0
	for _, a := range fixed {
		if a.Action != PortfolioActionSkip {
			runCount++
		}
	}
	if runCount == 0 {
		t.Error("Fixed advice should have at least one running strategy")
	}
}

func TestPortfolioAssist_MinTwoStrategiesForLargePortfolio(t *testing.T) {
	strategies := []string{"sa", "lahc", "tabu"}
	advice := []StrategyAdvice{
		{Strategy: "sa", Action: PortfolioActionRun, BudgetMult: 1.0, Confidence: 0.5},
		{Strategy: "lahc", Action: PortfolioActionSkip, BudgetMult: 0, Confidence: 0.8},
		{Strategy: "tabu", Action: PortfolioActionSkip, BudgetMult: 0, Confidence: 0.8},
	}

	safe, rule, _ := EvaluatePortfolioSafety(strategies, advice)
	if safe {
		t.Error("Should not be safe with only 1 running strategy in 3-strategy portfolio")
	}
	if rule != "min_two_strategies" {
		t.Errorf("Expected rule 'min_two_strategies', got '%s'", rule)
	}
}

func TestPortfolioAssist_CSVOutput(t *testing.T) {
	records := []PortfolioAssistRecord{
		{
			Domain: "cvrp", Instance: "A-n32-k5", Strategy: "sa", Seed: 42,
			OriginalBudget: 500000, RecommendedBudget: 550000, FinalBudget: 550000,
			Recommendation: PortfolioActionBoostBudget, Confidence: 0.55,
			ReasonCodes: "sa_generally_strong", Accepted: true,
			ResultObjective: 784, StrategyWon: true, RuntimeMs: 200,
		},
		{
			Domain: "cvrp", Instance: "A-n32-k5", Strategy: "lahc", Seed: 7961,
			OriginalBudget: 500000, RecommendedBudget: 450000, FinalBudget: 450000,
			Recommendation: PortfolioActionReduceBudget, Confidence: 0.5,
			ReasonCodes: "lahc_slower_convergence_in_portfolio", Accepted: true,
			ResultObjective: 810, StrategyWon: false, RuntimeMs: 180,
		},
	}

	path := t.TempDir() + "/portfolio_assist.csv"
	err := WritePortfolioAssistCSV(path, records)
	if err != nil {
		t.Fatalf("WritePortfolioAssistCSV: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(lines[1], "sa_generally_strong") {
		t.Error("Expected reason codes in row 1")
	}
	if !strings.Contains(lines[1], "cvrp") {
		t.Error("Expected domain in row 1")
	}
}

// --- Test Problem (simple for unit tests) ---

type testProblem struct {
	size int
}

type testSolution struct {
	values []int
}

type testMove struct {
	i, j int
}

func newTestProblem(size int) *testProblem {
	return &testProblem{size: size}
}

func (p *testProblem) CreateInitialSolution() (Solution, error) {
	vals := make([]int, p.size)
	for i := range vals {
		vals[i] = p.size - i // reverse order = high penalty
	}
	return &testSolution{values: vals}, nil
}

func (p *testProblem) Evaluate(sol Solution) int {
	ts := sol.(*testSolution)
	penalty := 0
	for i := 1; i < len(ts.values); i++ {
		if ts.values[i] < ts.values[i-1] {
			penalty += ts.values[i-1] - ts.values[i]
		}
	}
	return penalty
}

func (p *testProblem) TryMove(sol Solution, rng *rand.Rand) MoveResult {
	ts := sol.(*testSolution)
	n := len(ts.values)
	i := rng.Intn(n)
	j := rng.Intn(n)
	if i == j {
		return MoveResult{Valid: false}
	}
	ts.values[i], ts.values[j] = ts.values[j], ts.values[i]
	return MoveResult{Valid: true, Move: testMove{i, j}}
}

func (p *testProblem) UndoMove(sol Solution, move Move) {
	ts := sol.(*testSolution)
	m := move.(testMove)
	ts.values[m.i], ts.values[m.j] = ts.values[m.j], ts.values[m.i]
}

func (p *testProblem) CloneSolution(sol Solution) Solution {
	ts := sol.(*testSolution)
	clone := make([]int, len(ts.values))
	copy(clone, ts.values)
	return &testSolution{values: clone}
}

func (p *testProblem) SerializeSolution(sol Solution) ([]byte, error) {
	return []byte("{}"), nil
}

func (p *testProblem) SolutionFingerprint(sol Solution) string {
	ts := sol.(*testSolution)
	parts := make([]string, len(ts.values))
	for i, v := range ts.values {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, ",")
}
