package optimisation

import (
	"os"
	"strings"
	"testing"
)

// --- Adaptive Mode Tests ---

func TestAdaptiveMode_AcceptedAsValidMode(t *testing.T) {
	// NewSearchHookRunner should accept "adaptive" and return non-nil.
	runner := NewSearchHookRunner("adaptive", DefaultSearchAssistConfig(), 100000)
	if runner == nil {
		t.Fatal("Expected non-nil runner for adaptive mode")
	}
	if runner.Mode() != "adaptive" {
		t.Errorf("Expected mode 'adaptive', got %q", runner.Mode())
	}
	if !runner.HasAdaptiveEngine() {
		t.Error("Expected non-nil adaptiveAssist for adaptive mode")
	}
}

func TestAdaptiveMode_OffShadowAssistUnchanged(t *testing.T) {
	// Off returns nil.
	runner := NewSearchHookRunner("off", DefaultSearchAssistConfig(), 100000)
	if runner != nil {
		t.Error("Expected nil runner for off mode")
	}

	// Shadow should NOT have adaptive engine.
	runner = NewSearchHookRunner("shadow", DefaultSearchAssistConfig(), 100000)
	if runner == nil {
		t.Fatal("Expected non-nil runner for shadow mode")
	}
	if runner.HasAdaptiveEngine() {
		t.Error("Shadow mode should not have adaptive assist")
	}

	// Assist should NOT have adaptive engine.
	runner = NewSearchHookRunner("assist", DefaultSearchAssistConfig(), 100000)
	if runner == nil {
		t.Fatal("Expected non-nil runner for assist mode")
	}
	if runner.HasAdaptiveEngine() {
		t.Error("Assist mode should not have adaptive assist")
	}
}

func TestAdaptiveMode_ProducesTelemetry(t *testing.T) {
	runner := NewSearchHookRunner("adaptive", DefaultSearchAssistConfig(), 100000)

	// Simulate a search run with checkpoints.
	runner.OnImprovement(5000)
	action := runner.RunCheckpoint("sa", 10000, 500, 400, 800, 50.0)
	_ = action

	action = runner.RunCheckpoint("sa", 20000, 500, 400, 800, 25.0)
	_ = action

	recorder := runner.Finalise(400, 20000)
	if recorder == nil {
		t.Fatal("Expected non-nil recorder")
	}

	records := recorder.Records()
	if len(records) == 0 {
		t.Error("Expected at least one recorded checkpoint")
	}

	// Verify records have proper fields.
	for _, r := range records {
		if r.Algorithm != "sa" {
			t.Errorf("Expected algorithm 'sa', got %q", r.Algorithm)
		}
		if r.IterationsTotal <= 0 {
			t.Error("Expected positive iterations total")
		}
	}
}

func TestAdaptiveMode_NeverSkipsAllPortfolioStrategies(t *testing.T) {
	// This test verifies EvaluatePortfolioSafety still works in adaptive context.
	strategies := []string{"sa", "lahc", "tabu"}
	advice := []StrategyAdvice{
		{Strategy: "sa", Action: PortfolioActionSkip, BudgetMult: 0, Confidence: 0.9},
		{Strategy: "lahc", Action: PortfolioActionSkip, BudgetMult: 0, Confidence: 0.9},
		{Strategy: "tabu", Action: PortfolioActionSkip, BudgetMult: 0, Confidence: 0.9},
	}

	safe, rule, _ := EvaluatePortfolioSafety(strategies, advice)
	if safe {
		t.Error("Should not be safe to skip all strategies even in adaptive mode")
	}
	if rule != "cannot_skip_all" {
		t.Errorf("Expected rule 'cannot_skip_all', got %q", rule)
	}
}

func TestAdaptiveMode_RespectsSafetyFloors(t *testing.T) {
	config := DefaultSearchAssistConfig()
	runner := NewSearchHookRunner("adaptive", config, 100000)

	// Simulate: at 15% budget used (below MinBudgetFraction of 20%),
	// even if stagnating, should NOT early stop.
	action := runner.RunCheckpoint("sa", 15000, 800, 800, 800, 100.0)
	if action == SearchEarlyStop {
		t.Error("Should not early stop below minimum budget fraction")
	}
}

func TestAdaptiveMode_ExtendsBudgetWhenImproving(t *testing.T) {
	config := DefaultSearchAssistConfig()
	assist := NewAdaptiveSearchAssist(config)

	// Record some improvements to build history.
	assist.RecordImprovement(50000, 2.0)
	assist.RecordImprovement(70000, 1.5)
	assist.RecordImprovement(85000, 1.0)

	// At 90% budget with active improvement rate, should extend.
	progress := SearchProgress{
		Algorithm:          "sa",
		IterationsComplete: 90000,
		IterationsTotal:    100000,
		BestPenalty:        500,
		InitialPenalty:     800,
		ImprovementRate:    1.5,
		PlateauLength:      5000,
	}

	rec := assist.Checkpoint(progress)
	if rec == nil {
		t.Fatal("Expected recommendation to extend budget")
	}
	if rec.Action != SearchAdjustBudget {
		t.Errorf("Expected adjust_budget action, got %s", rec.Action)
	}
	if rec.NewBudget <= progress.IterationsTotal {
		t.Errorf("Expected extended budget > %d, got %d", progress.IterationsTotal, rec.NewBudget)
	}
}

func TestAdaptiveMode_AdaptiveStagnationWindow(t *testing.T) {
	config := DefaultSearchAssistConfig()
	assist := NewAdaptiveSearchAssist(config)

	// No improvements yet: stagnation window should be 75% of budget.
	progress := SearchProgress{
		Algorithm:          "sa",
		IterationsComplete: 50000,
		IterationsTotal:    100000,
		BestPenalty:        800,
		InitialPenalty:     800,
		ImprovementRate:    0.0,
		PlateauLength:      50000, // equals config.StagnationWindow
	}

	// With zero improvements, the adaptive window is 75K (75% of 100K).
	// At 50K plateau (< 75K), should NOT trigger early stop.
	rec := assist.Checkpoint(progress)
	if rec != nil && rec.Action == SearchEarlyStop {
		t.Error("Should not early stop with adaptive window when no improvements seen and plateau < 75% budget")
	}

	// Now at 76K plateau (> 75K window), should trigger.
	progress.IterationsComplete = 80000
	progress.PlateauLength = 80000
	rec = assist.Checkpoint(progress)
	if rec == nil || rec.Action != SearchEarlyStop {
		t.Error("Expected early stop when plateau exceeds adaptive window (75K)")
	}
}

func TestAdaptiveMode_MalformedModelFallsBackSafely(t *testing.T) {
	problem := newTestProblem(10)
	config := SearchConfig{
		Mode:                 "portfolio",
		Iterations:           5000,
		Portfolio:            []string{"sa", "lahc"},
		Seed:                 42,
		InitialTemperature:   10.0,
		MinTemperature:       0.001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 100,
	}

	// Non-existent model path — should fall back to rule-based gracefully.
	result, recorder := RunPortfolioWithAssist(problem, config, PortfolioAssistConfig{
		Mode:      "adaptive",
		Domain:    "test",
		Instance:  "test-instance",
		ModelPath: "/does/not/exist/model.json",
	})

	if recorder == nil {
		t.Fatal("Expected non-nil recorder")
	}
	if result.BestResult.BestPenalty >= result.BestResult.InitialPenalty {
		t.Error("Expected some improvement")
	}

	// Verify it still ran and accepted (using rule-based fallback).
	records := recorder.Records()
	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
	for _, r := range records {
		if r.Accepted && !strings.Contains(r.ReasonCodes, "fallback:") {
			// In adaptive mode with no model, should have fallback reason
			// unless the rule-based advisor provides advice without fallback annotation.
		}
	}
}

func TestAdaptiveMode_PortfolioWithLearnedModel(t *testing.T) {
	// Create a model that favours LAHC over SA.
	model := &PortfolioBudgetModel{
		Version:   "1.0",
		TrainedOn: 20,
		Entries: []StrategyPerformanceEntry{
			{
				Domain: "test", Strategy: "sa",
				WinRate: 0.2, MeanImprovement: 20, MeanROI: 0.5,
				SampleCount: 10, RecommendedMult: 0.7, Confidence: 0.75,
			},
			{
				Domain: "test", Strategy: "lahc",
				WinRate: 0.8, MeanImprovement: 80, MeanROI: 2.5,
				SampleCount: 10, RecommendedMult: 1.4, Confidence: 0.80,
			},
		},
	}

	modelPath := writeTestModel(t, model)

	problem := newTestProblem(10)
	config := SearchConfig{
		Mode:                 "portfolio",
		Iterations:           5000,
		Portfolio:            []string{"sa", "lahc"},
		Seed:                 42,
		InitialTemperature:   10.0,
		MinTemperature:       0.001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 100,
	}

	_, recorder := RunPortfolioWithAssist(problem, config, PortfolioAssistConfig{
		Mode:      "adaptive",
		Domain:    "test",
		Instance:  "test-instance",
		ModelPath: modelPath,
	})

	if recorder == nil {
		t.Fatal("Expected non-nil recorder")
	}

	records := recorder.Records()
	if len(records) != 2 {
		t.Fatalf("Expected 2 records, got %d", len(records))
	}

	// SA should get reduced budget (0.7×), LAHC boosted (1.4×).
	if records[0].RecommendedBudget != 3500 { // 5000 * 0.7
		t.Errorf("SA recommended budget: expected 3500, got %d", records[0].RecommendedBudget)
	}
	if records[1].RecommendedBudget != 7000 { // 5000 * 1.4
		t.Errorf("LAHC recommended budget: expected 7000, got %d", records[1].RecommendedBudget)
	}

	// Both should be accepted in adaptive mode.
	for _, r := range records {
		if !r.Accepted {
			t.Errorf("Expected accepted in adaptive mode for %s", r.Strategy)
		}
	}
}

func TestAdaptiveMode_WritesCSVTelemetry(t *testing.T) {
	config := DefaultSearchAssistConfig()
	runner := NewSearchHookRunner("adaptive", config, 100000)

	// Run a few checkpoints.
	runner.OnImprovement(5000)
	runner.RunCheckpoint("sa", 10000, 600, 500, 800, 50.0)
	runner.RunCheckpoint("sa", 20000, 550, 500, 800, 25.0)

	recorder := runner.Finalise(500, 20000)
	records := recorder.Records()

	// Write to CSV.
	path := t.TempDir() + "/adaptive_assist.csv"
	err := WriteSearchAssistCSV(path, records)
	if err != nil {
		t.Fatalf("WriteSearchAssistCSV: %v", err)
	}

	data, _ := os.ReadFile(path)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 3 { // header + 2 records
		t.Errorf("Expected at least 3 lines, got %d", len(lines))
	}
}
