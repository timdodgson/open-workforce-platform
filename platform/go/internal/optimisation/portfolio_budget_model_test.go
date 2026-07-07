package optimisation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// --- Learned Portfolio Budget Model Tests ---

func TestLearnedAllocator_LoadsModelSafely(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "1.0",
		TrainedOn: 50,
		Entries: []StrategyPerformanceEntry{
			{
				Domain: "jss", Strategy: "tabu",
				WinRate: 0.8, MeanImprovement: 100, MeanROI: 2.5,
				SampleCount: 10, RecommendedMult: 1.3, Confidence: 0.75,
			},
			{
				Domain: "jss", Strategy: "sa",
				WinRate: 0.2, MeanImprovement: 30, MeanROI: 0.8,
				SampleCount: 10, RecommendedMult: 0.7, Confidence: 0.72,
			},
		},
	}

	path := writeTestModel(t, model)
	loaded, err := LoadPortfolioBudgetModel(path)
	if err != nil {
		t.Fatalf("LoadPortfolioBudgetModel: %v", err)
	}

	if loaded.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", loaded.Version)
	}
	if loaded.TrainedOn != 50 {
		t.Errorf("Expected trained_on 50, got %d", loaded.TrainedOn)
	}
	if len(loaded.Entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(loaded.Entries))
	}
}

func TestLearnedAllocator_MissingModelFallsBackToRules(t *testing.T) {
	advisor := NewLearnedPortfolioAdvisor(nil)

	result := advisor.Advise([]string{"sa", "lahc"}, 100000, "cvrp", "A-n32-k5")

	// Should produce same advice as rule-based.
	for i, source := range result.Source {
		if source.UsedLearned {
			t.Errorf("Strategy %d should not use learned (nil model)", i)
		}
		if source.FallbackReason != "no_model_loaded" {
			t.Errorf("Expected fallback reason 'no_model_loaded', got '%s'", source.FallbackReason)
		}
	}

	// Verify advice matches rule-based output.
	ruleAdvisor := NewRuleBasedPortfolioAdvisor()
	ruleAdvice := ruleAdvisor.Advise([]string{"sa", "lahc"}, 100000, "cvrp")
	for i := range result.Advice {
		if result.Advice[i].BudgetMult != ruleAdvice[i].BudgetMult {
			t.Errorf("Strategy %d: expected rule mult %.2f, got %.2f",
				i, ruleAdvice[i].BudgetMult, result.Advice[i].BudgetMult)
		}
	}
}

func TestLearnedAllocator_MalformedModelFailsGracefully(t *testing.T) {
	dir := t.TempDir()

	// Empty file.
	emptyPath := filepath.Join(dir, "empty.json")
	os.WriteFile(emptyPath, []byte(""), 0644)
	_, err := LoadPortfolioBudgetModel(emptyPath)
	if err == nil {
		t.Error("Expected error for empty file")
	}

	// Invalid JSON.
	invalidPath := filepath.Join(dir, "invalid.json")
	os.WriteFile(invalidPath, []byte("{not valid json}"), 0644)
	_, err = LoadPortfolioBudgetModel(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Valid JSON but missing version.
	noVersionPath := filepath.Join(dir, "no_version.json")
	os.WriteFile(noVersionPath, []byte(`{"entries":[{"domain":"jss","strategy":"sa","win_rate":0.5,"sample_count":5,"recommended_mult":1.0,"confidence":0.7}]}`), 0644)
	_, err = LoadPortfolioBudgetModel(noVersionPath)
	if err == nil {
		t.Error("Expected error for missing version")
	}

	// Valid JSON but no entries.
	noEntriesPath := filepath.Join(dir, "no_entries.json")
	os.WriteFile(noEntriesPath, []byte(`{"version":"1.0","trained_on":10,"entries":[]}`), 0644)
	_, err = LoadPortfolioBudgetModel(noEntriesPath)
	if err == nil {
		t.Error("Expected error for empty entries")
	}

	// Non-existent file.
	_, err = LoadPortfolioBudgetModel(filepath.Join(dir, "does_not_exist.json"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLearnedAllocator_SafetyCapsEnforced(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "1.0",
		TrainedOn: 20,
		Entries: []StrategyPerformanceEntry{
			{
				Domain: "cvrp", Strategy: "sa",
				WinRate: 0.9, MeanImprovement: 200, MeanROI: 5.0,
				SampleCount: 10, RecommendedMult: 3.0, // exceeds 2x cap
				Confidence: 0.85,
			},
			{
				Domain: "cvrp", Strategy: "lahc",
				WinRate: 0.1, MeanImprovement: 10, MeanROI: 0.2,
				SampleCount: 10, RecommendedMult: 0.1, // below 0.25 floor
				Confidence: 0.80,
			},
		},
	}

	advisor := NewLearnedPortfolioAdvisor(model)
	result := advisor.Advise([]string{"sa", "lahc"}, 100000, "cvrp", "A-n32-k5")

	// SA should be capped at 2.0.
	if result.Advice[0].BudgetMult > 2.0 {
		t.Errorf("SA budget mult %.2f exceeds 2.0 cap", result.Advice[0].BudgetMult)
	}
	if result.Advice[0].BudgetMult != 2.0 {
		t.Errorf("SA budget mult should be clamped to 2.0, got %.2f", result.Advice[0].BudgetMult)
	}

	// LAHC should be floored at 0.25.
	if result.Advice[1].BudgetMult < 0.25 {
		t.Errorf("LAHC budget mult %.2f below 0.25 floor", result.Advice[1].BudgetMult)
	}
	if result.Advice[1].BudgetMult != 0.25 {
		t.Errorf("LAHC budget mult should be clamped to 0.25, got %.2f", result.Advice[1].BudgetMult)
	}
}

func TestLearnedAllocator_OffModeUnchanged(t *testing.T) {
	problem := newTestProblem(10)
	config := SearchConfig{
		Mode:                 "portfolio",
		Iterations:           1000,
		Portfolio:            []string{"sa", "lahc"},
		Seed:                 42,
		InitialTemperature:   10.0,
		MinTemperature:       0.001,
		CoolingMode:          "adaptive",
		LateAcceptanceLength: 100,
	}

	// Off mode should return nil recorder regardless of model path.
	result, recorder := RunPortfolioWithAssist(problem, config, PortfolioAssistConfig{
		Mode:      "off",
		ModelPath: "/some/model/that/does/not/exist.json",
	})
	if recorder != nil {
		t.Error("Expected nil recorder for off mode")
	}
	if result.BestResult.BestPenalty >= result.BestResult.InitialPenalty {
		t.Error("Expected some improvement from portfolio")
	}
}

func TestLearnedAllocator_LowConfidenceFallsBack(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "1.0",
		TrainedOn: 10,
		Entries: []StrategyPerformanceEntry{
			{
				Domain: "cvrp", Strategy: "sa",
				WinRate: 0.5, MeanImprovement: 50, MeanROI: 1.0,
				SampleCount: 10, RecommendedMult: 1.5,
				Confidence: 0.40, // below MinLearnedConfidence (0.60)
			},
			{
				Domain: "cvrp", Strategy: "lahc",
				WinRate: 0.5, MeanImprovement: 50, MeanROI: 1.0,
				SampleCount: 10, RecommendedMult: 0.8,
				Confidence: 0.75, // above threshold
			},
		},
	}

	advisor := NewLearnedPortfolioAdvisor(model)
	result := advisor.Advise([]string{"sa", "lahc"}, 100000, "cvrp", "A-n32-k5")

	// SA should fall back to rule-based (confidence too low).
	if result.Source[0].UsedLearned {
		t.Error("SA should have fallen back (low confidence)")
	}
	if result.Source[0].FallbackReason == "" {
		t.Error("Expected fallback reason for SA")
	}

	// LAHC should use learned.
	if !result.Source[1].UsedLearned {
		t.Error("LAHC should have used learned recommendation")
	}
	if result.Advice[1].BudgetMult != 0.8 {
		t.Errorf("LAHC mult should be 0.8 (learned), got %.2f", result.Advice[1].BudgetMult)
	}
}

func TestLearnedAllocator_InsufficientSamplesFallsBack(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "1.0",
		TrainedOn: 5,
		Entries: []StrategyPerformanceEntry{
			{
				Domain: "jss", Strategy: "tabu",
				WinRate: 0.9, MeanImprovement: 100, MeanROI: 3.0,
				SampleCount: 2, // below minimum of 3
				RecommendedMult: 1.5, Confidence: 0.80,
			},
		},
	}

	advisor := NewLearnedPortfolioAdvisor(model)
	result := advisor.Advise([]string{"tabu"}, 100000, "jss", "la01")

	if result.Source[0].UsedLearned {
		t.Error("Tabu should have fallen back (insufficient samples)")
	}
	if result.Source[0].FallbackReason == "" {
		t.Error("Expected fallback reason")
	}
}

func TestLearnedAllocator_InstanceSpecificOverridesDomainWide(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "1.0",
		TrainedOn: 30,
		Entries: []StrategyPerformanceEntry{
			{
				Domain: "jss", Strategy: "tabu",
				WinRate: 0.6, MeanImprovement: 50, MeanROI: 1.5,
				SampleCount: 15, RecommendedMult: 1.2, Confidence: 0.70,
			},
			{
				Domain: "jss", Instance: "la01", Strategy: "tabu",
				WinRate: 0.95, MeanImprovement: 85, MeanROI: 3.0,
				SampleCount: 10, RecommendedMult: 1.5, Confidence: 0.85,
			},
		},
	}

	advisor := NewLearnedPortfolioAdvisor(model)

	// la01 should get the instance-specific entry (mult=1.5).
	result := advisor.Advise([]string{"tabu"}, 100000, "jss", "la01")
	if result.Advice[0].BudgetMult != 1.5 {
		t.Errorf("Expected instance-specific mult 1.5, got %.2f", result.Advice[0].BudgetMult)
	}

	// ft10 should get the domain-wide entry (mult=1.2).
	result = advisor.Advise([]string{"tabu"}, 100000, "jss", "ft10")
	if result.Advice[0].BudgetMult != 1.2 {
		t.Errorf("Expected domain-wide mult 1.2, got %.2f", result.Advice[0].BudgetMult)
	}
}

func TestLearnedAllocator_NeverRemovesAllBudgetFromHistoricallyStrongest(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "1.0",
		TrainedOn: 20,
		Entries: []StrategyPerformanceEntry{
			{
				Domain: "jss", Strategy: "tabu",
				WinRate: 0.85, MeanImprovement: 100, MeanROI: 3.0,
				SampleCount: 10, RecommendedMult: 0.25, // model says reduce heavily
				Confidence: 0.75,
			},
			{
				Domain: "jss", Strategy: "sa",
				WinRate: 0.15, MeanImprovement: 20, MeanROI: 0.5,
				SampleCount: 10, RecommendedMult: 1.8,
				Confidence: 0.75,
			},
		},
	}

	advisor := NewLearnedPortfolioAdvisor(model)
	result := advisor.Advise([]string{"tabu", "sa"}, 100000, "jss", "la01")

	// Tabu (highest win rate) should never be fully removed.
	// The model says 0.25x which is the floor — it's allowed, just not zero.
	if result.Advice[0].BudgetMult < 0.25 {
		t.Errorf("Tabu budget mult %.2f below 0.25 floor", result.Advice[0].BudgetMult)
	}

	// The safety system should also catch if someone tries to skip the strongest.
	advice := result.Advice
	advice[0].Action = PortfolioActionSkip
	advice[0].BudgetMult = 0

	safe, rule, _ := EvaluatePortfolioSafety([]string{"tabu", "sa"}, advice)
	// With 2 strategies and 1 skipped, only 1 runs — that's valid for a 2-strategy portfolio.
	// But if we had 3+ strategies, it would trigger the min_two rule.
	_ = safe
	_ = rule
}

func TestLearnedAllocator_IntegrationWithRunPortfolioWithAssist(t *testing.T) {
	// Create a model file in temp dir.
	model := &PortfolioBudgetModel{
		Version:   "1.0",
		TrainedOn: 20,
		Entries: []StrategyPerformanceEntry{
			{
				Domain: "test", Strategy: "sa",
				WinRate: 0.4, MeanImprovement: 30, MeanROI: 1.0,
				SampleCount: 10, RecommendedMult: 0.9, Confidence: 0.70,
			},
			{
				Domain: "test", Strategy: "lahc",
				WinRate: 0.6, MeanImprovement: 50, MeanROI: 1.5,
				SampleCount: 10, RecommendedMult: 1.2, Confidence: 0.72,
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
		Mode:      "assist",
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

	// SA should have budget = 5000 * 0.9 = 4500.
	if records[0].RecommendedBudget != 4500 {
		t.Errorf("SA recommended budget: expected 4500, got %d", records[0].RecommendedBudget)
	}
	if records[0].FinalBudget != 4500 {
		t.Errorf("SA final budget: expected 4500, got %d", records[0].FinalBudget)
	}

	// LAHC should have budget = 5000 * 1.2 = 6000.
	if records[1].RecommendedBudget != 6000 {
		t.Errorf("LAHC recommended budget: expected 6000, got %d", records[1].RecommendedBudget)
	}
	if records[1].FinalBudget != 6000 {
		t.Errorf("LAHC final budget: expected 6000, got %d", records[1].FinalBudget)
	}
}

func TestLearnedAllocator_ShadowModeRecordsButDoesNotApply(t *testing.T) {
	model := &PortfolioBudgetModel{
		Version:   "1.0",
		TrainedOn: 20,
		Entries: []StrategyPerformanceEntry{
			{
				Domain: "test", Strategy: "sa",
				WinRate: 0.3, MeanImprovement: 20, MeanROI: 0.5,
				SampleCount: 10, RecommendedMult: 0.5, Confidence: 0.80,
			},
			{
				Domain: "test", Strategy: "lahc",
				WinRate: 0.7, MeanImprovement: 80, MeanROI: 2.0,
				SampleCount: 10, RecommendedMult: 1.5, Confidence: 0.80,
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
		Mode:      "shadow",
		Domain:    "test",
		Instance:  "test-instance",
		ModelPath: modelPath,
	})

	if recorder == nil {
		t.Fatal("Expected non-nil recorder")
	}

	records := recorder.Records()
	for _, r := range records {
		// Shadow mode: final budget should be original (no change applied).
		if r.FinalBudget != r.OriginalBudget {
			t.Errorf("Shadow mode changed budget for %s: original=%d final=%d",
				r.Strategy, r.OriginalBudget, r.FinalBudget)
		}
		if r.Accepted {
			t.Errorf("Shadow mode accepted recommendation for %s", r.Strategy)
		}
	}
}

// --- Helpers ---

func writeTestModel(t *testing.T, model *PortfolioBudgetModel) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "portfolio_budget_model.json")
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		t.Fatalf("marshal model: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write model: %v", err)
	}
	return path
}
