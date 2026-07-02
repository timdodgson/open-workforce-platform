package ilp_test

import (
	"testing"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/ilp"
)

func TestHighsSolver_Available(t *testing.T) {
	solver := &ilp.HighsSolver{}
	// On the test machine with highspy installed, this should return true.
	// On CI without Python/highspy, it returns false. Either is valid.
	available := solver.Available()
	t.Logf("HiGHS solver available: %v", available)
}

func TestHighsSolver_Name(t *testing.T) {
	solver := &ilp.HighsSolver{}
	if solver.Name() != "HiGHS" {
		t.Errorf("expected name 'HiGHS', got '%s'", solver.Name())
	}
}

func TestHighsSolver_Available_BadPython(t *testing.T) {
	solver := &ilp.HighsSolver{PythonPath: "nonexistent_python_binary"}
	if solver.Available() {
		t.Error("expected Available() to return false with bad python path")
	}
}
