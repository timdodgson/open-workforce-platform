package ilp

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

//go:embed solve_highs.py
var solveScript embed.FS

// HighsSolver uses Python + highspy to solve LP/MIP models.
// Requires: pip install highspy
type HighsSolver struct {
	// PythonPath overrides the Python executable. Defaults to "python".
	PythonPath string
}

// Name returns the solver identifier.
func (h *HighsSolver) Name() string {
	return "HiGHS"
}

// Available checks if Python and highspy are accessible.
func (h *HighsSolver) Available() bool {
	python := h.python()
	cmd := exec.Command(python, "-c", "from highspy import Highs; print('ok')")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return len(out) > 0
}

// Solve runs the HiGHS solver via Python on the given LP model file.
func (h *HighsSolver) Solve(modelPath string, timeLimit time.Duration) (SolverOutput, error) {
	python := h.python()

	// Write the solver script to a temp file.
	scriptContent, err := solveScript.ReadFile("solve_highs.py")
	if err != nil {
		return SolverOutput{}, fmt.Errorf("failed to read embedded solver script: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "ilp-solver-*")
	if err != nil {
		return SolverOutput{}, fmt.Errorf("failed to create temp dir: %w", err)
	}
	// Note: don't defer RemoveAll here — caller needs the progress CSV.
	// Caller is responsible for cleanup via tmpDir in SolverOutput.

	scriptPath := filepath.Join(tmpDir, "solve_highs.py")
	if err := os.WriteFile(scriptPath, scriptContent, 0644); err != nil {
		os.RemoveAll(tmpDir)
		return SolverOutput{}, fmt.Errorf("failed to write solver script: %w", err)
	}

	progressCSVPath := filepath.Join(tmpDir, "progress.csv")

	// Run: python solve_highs.py <model.lp> <time_limit_seconds> <progress_csv_path>
	timeLimitStr := fmt.Sprintf("%.1f", timeLimit.Seconds())
	cmd := exec.Command(python, scriptPath, modelPath, timeLimitStr, progressCSVPath)
	cmd.Stderr = os.Stderr // Forward heartbeat to user.

	start := time.Now()
	output, err := cmd.Output() // Captures stdout only; stderr goes to terminal.
	elapsed := time.Since(start)

	if err != nil {
		return SolverOutput{
			Status:         "ERROR",
			RuntimeSeconds: elapsed.Seconds(),
		}, fmt.Errorf("Python solver failed: %w\nOutput: %s", err, string(output))
	}

	// Parse JSON output from Python script.
	var pyResult struct {
		Status         string             `json:"status"`
		Objective      float64            `json:"objective"`
		LowerBound     float64            `json:"lowerBound"`
		RuntimeSeconds float64            `json:"runtimeSeconds"`
		Variables      map[string]float64 `json:"variables"`
		Error          string             `json:"error,omitempty"`
	}

	if err := json.Unmarshal(output, &pyResult); err != nil {
		return SolverOutput{
			Status:         "ERROR",
			RuntimeSeconds: elapsed.Seconds(),
		}, fmt.Errorf("failed to parse solver output: %w\nRaw output: %s", err, string(output))
	}

	if pyResult.Error != "" {
		return SolverOutput{
			Status:         pyResult.Status,
			RuntimeSeconds: pyResult.RuntimeSeconds,
		}, fmt.Errorf("solver error: %s", pyResult.Error)
	}

	return SolverOutput{
		Status:          pyResult.Status,
		Objective:       pyResult.Objective,
		LowerBound:      pyResult.LowerBound,
		RuntimeSeconds:  pyResult.RuntimeSeconds,
		SolutionValues:  pyResult.Variables,
		ProgressCSVPath: progressCSVPath,
	}, nil
}

func (h *HighsSolver) python() string {
	if h.PythonPath != "" {
		return h.PythonPath
	}
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

// ParseHighsOutput is kept for test compatibility — parses console output format.
func ParseHighsOutput(output string) SolverOutput {
	// No longer used in production (we parse JSON from Python now).
	// Kept for backward compatibility with tests.
	return SolverOutput{Status: "ERROR"}
}

// ParseHighsSolution is kept for test compatibility — parses solution file format.
func ParseHighsSolution(path string) (map[string]float64, error) {
	// No longer used in production (Python returns variables in JSON).
	// Kept for backward compatibility with tests.
	return nil, fmt.Errorf("not used: solver returns variables via JSON")
}
