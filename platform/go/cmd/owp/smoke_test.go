package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// smoke_test.go lives in platform/go/cmd/owp
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))
}

func buildOWP(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "owp")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	pkgDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = pkgDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build owp: %v\n%s", err, out)
	}
	return bin
}

func runOWP(t *testing.T, bin string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = filepath.Join(repoRoot(t), "platform", "go")
	out, err := cmd.CombinedOutput()
	s := string(out)
	// owp writes usage/errors to stderr; CombinedOutput merges both.
	return s, s, err
}

func TestCLI_unknownCommand(t *testing.T) {
	bin := buildOWP(t)
	out, _, err := runOWP(t, bin, "not-a-command")
	if err == nil {
		t.Fatal("expected non-zero exit for unknown command")
	}
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("expected usage text, got: %q", out)
	}
}

func TestCLI_solveCVRP_minimal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI smoke test in -short mode")
	}

	bin := buildOWP(t)
	root := repoRoot(t)
	instance := filepath.Join(root, "examples", "cvrp", "A-n32-k5.vrp")
	label := "smoke-cvrp-" + strings.ReplaceAll(t.Name(), "/", "-")

	out, _, err := runOWP(t, bin,
		"solve-cvrp",
		"--instance", instance,
		"--mode", "sa",
		"--iterations", "5000",
		"--seed", "42",
		"--run-label", label,
	)
	if err != nil {
		t.Fatalf("solve-cvrp failed: %v\n%s", err, out)
	}

	runDir := filepath.Join(root, "platform", "web", "pfrs-lab", "data", "runs", label)
	t.Cleanup(func() { os.RemoveAll(runDir) })

	runJSON := filepath.Join(runDir, "run.json")
	if _, err := os.Stat(runJSON); err != nil {
		t.Fatalf("expected run.json at %s: %v", runJSON, err)
	}
}

func TestCLI_tunePFRS_minimal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI smoke test in -short mode")
	}

	bin := buildOWP(t)
	root := repoRoot(t)
	label := "smoke-tune-pfrs-" + strings.ReplaceAll(t.Name(), "/", "-")

	out, _, err := runOWP(t, bin,
		"tune-pfrs",
		"--instance", "n005w4",
		"--pfrs-iterations-per-worker", "5000",
		"--pfrs-max-total-workers", "4",
		"--pfrs-run-label", label,
		"--worker-decision-mode", "shadow",
	)
	if err != nil {
		t.Fatalf("tune-pfrs failed: %v\n%s", err, out)
	}

	runDir := filepath.Join(root, "platform", "web", "pfrs-lab", "data", "runs", label)
	t.Cleanup(func() { os.RemoveAll(runDir) })

	for _, name := range []string{"run.json", "results.csv", "worker_learning.csv"} {
		path := filepath.Join(runDir, name)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s at %s: %v", name, path, err)
		}
	}
}

func TestCLI_validateINRC2(t *testing.T) {
	bin := buildOWP(t)
	root := repoRoot(t)
	base := filepath.Join(root, "examples", "inrc2", "testdatasets_json", "n005w4")

	out, _, err := runOWP(t, bin, "validate-inrc2",
		filepath.Join(base, "Sc-n005w4.json"),
		filepath.Join(base, "WD-n005w4-1.json"),
		filepath.Join(base, "H0-n005w4-0.json"),
		filepath.Join(base, "Solution_H_0-WD_1-2-3-3", "Sol-n005w4-1-0.json"),
	)
	if err != nil {
		t.Fatalf("validate-inrc2 failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Total Objective:") {
		t.Fatalf("expected validation output, got: %q", out)
	}
}
