package optimisation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExperimentResultsFromRunsDir(t *testing.T) {
	dir := t.TempDir()
	runDir := filepath.Join(dir, "regress-cvrp-sa-rules-s42")
	if err := os.MkdirAll(runDir, 0755); err != nil {
		t.Fatal(err)
	}
	meta := `{
		"problemType": "cvrp",
		"instance": "A-n32-k5",
		"mode": "sa",
		"policyMode": "rules",
		"seed": 42,
		"bestObjective": 830,
		"runtimeMs": 8,
		"candidates": 50000,
		"feasible": true
	}`
	if err := os.WriteFile(filepath.Join(runDir, "run.json"), []byte(meta), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := LoadExperimentResultsFromRunsDir(dir, "regress-")
	if err != nil {
		t.Fatalf("LoadExperimentResultsFromRunsDir: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
	if results[0].Objective != 830 || results[0].PolicyMode != "rules" {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}

func TestInferFromLabel(t *testing.T) {
	if got := inferDomainFromLabel("regress-cvrp-sa-rules-s42"); got != "cvrp" {
		t.Errorf("domain = %q", got)
	}
	if got := inferPolicyFromLabel("regress-cvrp-sa-hybrid-s42"); got != "hybrid" {
		t.Errorf("policy = %q", got)
	}
	if got := inferSeedFromLabel("regress-cvrp-sa-rules-s42"); got != 42 {
		t.Errorf("seed = %d", got)
	}
}
