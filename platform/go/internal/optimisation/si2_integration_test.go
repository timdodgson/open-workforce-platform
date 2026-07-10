package optimisation

import (
	"os"
	"path/filepath"
	"testing"
)

func testPolicyDir(t *testing.T) string {
	t.Helper()
	candidates := []string{
		filepath.Join("..", "..", "ml", "policies"),
		filepath.Join("..", "..", "..", "ml", "policies"),
	}
	for _, dir := range candidates {
		if _, err := os.Stat(filepath.Join(dir, StagnationPolicyFile)); err == nil {
			return dir
		}
	}
	t.Skip("platform/ml/policies not found (run train_policies.py)")
	return ""
}

func TestRunSearch_HybridProducesPolicyTelemetry(t *testing.T) {
	dir := testPolicyDir(t)
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Iterations = 20000
	config.Seed = 42
	config.AssistMode = "shadow"
	config.AssistConfig = DefaultSearchAssistConfig()
	config.AssistConfig.CheckpointInterval = 2000
	config.PolicyMode = "hybrid"
	config.PolicyDir = dir
	config.PolicyDomain = "cvrp"
	config.PolicyInstance = "test"

	result := RunSearch(problem, config)
	if len(result.PolicyDecisions) == 0 {
		t.Fatal("expected policy decisions with hybrid mode")
	}
	if result.BestPenalty >= result.InitialPenalty {
		t.Errorf("search should improve: best=%d initial=%d", result.BestPenalty, result.InitialPenalty)
	}
}

func TestPortfolio_HybridPolicyAssist(t *testing.T) {
	dir := testPolicyDir(t)
	problem := &mockProblem{}
	config := DefaultSearchConfig()
	config.Mode = "portfolio"
	config.Iterations = 6000
	config.Seed = 42
	config.PolicyMode = "hybrid"
	config.PolicyDir = dir
	config.PolicyDomain = "cvrp"
	config.PolicyInstance = "test"

	_, recorder := RunPortfolioWithAssist(problem, config, PortfolioAssistConfig{
		Domain:   "cvrp",
		Instance: "test",
	})
	if recorder == nil {
		t.Fatal("expected portfolio assist recorder when policy-mode set")
	}
	if len(recorder.Records()) == 0 {
		t.Fatal("expected portfolio assist records")
	}
}

func TestPolicyEvaluation_EndToEnd(t *testing.T) {
	dir := t.TempDir()
	err := WritePolicyEvaluationCSV(dir, PolicyEvaluationInput{
		RunID: "integration", Domain: "cvrp", Instance: "A-n32-k5", Algorithm: "sa",
		InitialPenalty: 1000, BestPenalty: 800,
		Decisions: []PolicySearchDecision{
			{PolicyUsed: "hybrid_learned", Action: "early_stop", Confidence: 0.85},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "policy_evaluation.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 50 {
		t.Fatal("policy_evaluation.csv too short")
	}
}
