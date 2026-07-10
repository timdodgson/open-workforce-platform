package policy

import (
	"path/filepath"
	"testing"
)

func TestResolvePolicyDir(t *testing.T) {
	if got := ResolvePolicyDir("hybrid", ""); got != defaultPolicyDir {
		t.Fatalf("expected default %q, got %q", defaultPolicyDir, got)
	}
	if got := ResolvePolicyDir("", "/custom"); got != "/custom" {
		t.Fatalf("expected custom dir, got %q", got)
	}
}

func TestRunPostRunPolicyPipeline(t *testing.T) {
	dir := t.TempDir()
	report := RunPostRunPolicyPipeline(PostRunPolicyConfig{
		PolicyMode:          "hybrid",
		PolicyDir:           dir,
		OutputDir:           dir,
		Domain:              "cvrp",
		PolicyDecisionCount: 3,
	})
	if report == nil {
		t.Fatal("expected report")
	}
	if report.LearningRecommendation.Action == "" {
		t.Fatal("expected learning action")
	}
	if _, err := filepath.Glob(filepath.Join(dir, "policy_learning_report.json")); err != nil {
		t.Fatalf("glob: %v", err)
	}
}
