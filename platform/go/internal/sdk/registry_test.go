package sdk_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation/searchdef"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk"
	_ "github.com/timdodgson/open-workforce-platform/platform/go/internal/sdk/builtin"
)

func TestRegisterProblem_duplicateFails(t *testing.T) {
	load := func(path string) (searchdef.Problem, sdk.InstanceMeta, error) {
		return nil, sdk.InstanceMeta{}, nil
	}
	if err := sdk.RegisterProblem(sdk.ProblemDescriptor{Name: "test-dup-problem", Load: load}); err != nil {
		t.Fatalf("first register: %v", err)
	}
	if err := sdk.RegisterProblem(sdk.ProblemDescriptor{Name: "test-dup-problem", Load: load}); err == nil {
		t.Fatal("expected duplicate register error")
	}
}

func TestBuiltinProblemsRegistered(t *testing.T) {
	for _, name := range []string{"cvrp", "vrptw", "jobshop"} {
		if _, ok := sdk.GetProblem(name); !ok {
			t.Fatalf("expected builtin problem %q", name)
		}
	}
}

func TestRegisterSearch_customRunner(t *testing.T) {
	mode := "test-custom-search-mode"
	called := false
	runner := func(problem searchdef.Problem, config optimisation.SearchConfig) optimisation.SearchResult {
		called = true
		return optimisation.SearchResult{BestPenalty: 99}
	}
	if err := sdk.RegisterSearch(mode, runner); err != nil {
		t.Fatalf("register search: %v", err)
	}
	result := sdk.RunSearch(nil, optimisation.SearchConfig{Mode: mode})
	if !called {
		t.Fatal("expected custom runner to be called")
	}
	if result.BestPenalty != 99 {
		t.Fatalf("got penalty %d, want 99", result.BestPenalty)
	}
}

func TestRunSearch_builtinFallback(t *testing.T) {
	runner := sdk.ResolveSearchRunner("sa")
	if runner == nil {
		t.Fatal("expected non-nil runner for built-in sa mode")
	}
}
