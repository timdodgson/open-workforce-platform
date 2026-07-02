package ilp

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// RunRollingBenchmark solves the full horizon using rolling horizon decomposition.
// Each iteration solves a window of `windowSize` weeks, fixes the first week, rolls forward.
// This produces high-quality solutions for multi-week problems that monolithic
// solving cannot handle within reasonable time.
func RunRollingBenchmark(sc inrc2.Scenario, weekDataFiles []string, initialHist inrc2.History, config BenchmarkConfig) (BenchmarkResult, error) {
	weeks := config.Weeks
	if weeks > len(weekDataFiles) {
		weeks = len(weekDataFiles)
	}

	solver := selectSolver(config.SolverName)
	if !solver.Available() {
		return BenchmarkResult{
			Instance: config.Instance,
			Weeks:    weeks,
			Solver:   solver.Name(),
			Status:   "ERROR",
			Notes:    "solver not available",
		}, fmt.Errorf("solver not available")
	}

	// Window size: 2-week lookahead for better boundary handling.
	windowSize := 2

	// Time budget per window.
	windowTimeLimit := config.TimeLimit / time.Duration(weeks)
	if windowTimeLimit < 10*time.Second {
		windowTimeLimit = 10 * time.Second
	}

	fmt.Fprintf(os.Stderr, "  Rolling horizon: %d weeks, window=%d, %s per window\n",
		weeks, windowSize, windowTimeLimit)

	start := time.Now()
	currentHist := initialHist

	// Collect all week solutions.
	allSolutions := make([]inrc2.Solution, weeks)
	totalSolverObj := 0.0

	for w := 0; w < weeks; w++ {
		// Determine window: current week + lookahead (if available).
		windowEnd := w + windowSize
		if windowEnd > weeks {
			windowEnd = weeks
		}
		windowWeeks := windowEnd - w
		windowFiles := weekDataFiles[w:windowEnd]

		fmt.Fprintf(os.Stderr, "  Week %d/%d (window=%d)... ", w+1, weeks, windowWeeks)

		// Build and solve the window.
		absDir, _ := os.Getwd()
		modelPath := filepath.Join(absDir, fmt.Sprintf("ilp_rolling_w%d.lp", w))
		_, err := BuildModel(sc, windowFiles, currentHist, windowWeeks, modelPath)
		if err != nil {
			os.Remove(modelPath)
			return BenchmarkResult{}, fmt.Errorf("week %d: failed to build model: %w", w+1, err)
		}

		solverOutput, err := solver.Solve(modelPath, windowTimeLimit)
		// Clean up model file after solve (keep on error for debugging).
		if err == nil || solverOutput.Status != "ERROR" {
			os.Remove(modelPath)
		}

		if err != nil && solverOutput.Status == "ERROR" {
			return BenchmarkResult{}, fmt.Errorf("week %d: solver failed: %w", w+1, err)
		}

		if solverOutput.Status == "INFEASIBLE" {
			return BenchmarkResult{
				Instance: config.Instance,
				Weeks:    weeks,
				Solver:   solver.Name(),
				Status:   "INFEASIBLE",
				Notes:    fmt.Sprintf("infeasible at week %d", w+1),
			}, fmt.Errorf("infeasible at week %d", w+1)
		}

		totalSolverObj += solverOutput.Objective

		// Extract solution for the FIRST week of the window only.
		if len(solverOutput.SolutionValues) == 0 {
			return BenchmarkResult{}, fmt.Errorf("week %d: no solution values", w+1)
		}

		windowSolutions := ExtractSolutions(sc, windowWeeks, solverOutput.SolutionValues)
		allSolutions[w] = windowSolutions[0]
		allSolutions[w].Week = w

		fmt.Fprintf(os.Stderr, "done (status=%s, obj=%.0f)\n",
			solverOutput.Status, solverOutput.Objective)

		// Update history for next iteration.
		currentHist = inrc2.UpdateHistory(sc, currentHist, allSolutions[w])
	}

	elapsed := time.Since(start)

	// Validate the complete solution against official scorer.
	totalPenalty, hardViolations, _, valErr := ValidateILPSolution(
		sc, weekDataFiles[:weeks], initialHist, allSolutions)
	if valErr != nil {
		return BenchmarkResult{}, fmt.Errorf("validation failed: %w", valErr)
	}

	result := BenchmarkResult{
		Instance:               config.Instance,
		Weeks:                  weeks,
		Solver:                 solver.Name(),
		Status:                 "FEASIBLE",
		Objective:              totalPenalty,
		LowerBound:             int(math.Round(totalSolverObj)),
		RuntimeSeconds:         elapsed.Seconds(),
		TimeLimit:              int(config.TimeLimit.Seconds()),
		HardViolations:         hardViolations,
		ModelCompleteness:      "full",
		SupportedConstraints:   []string{"H1", "H2", "H3", "H4", "S1", "S2", "S3", "S4", "S5", "S6", "S7", "S8"},
		UnsupportedConstraints: []string{},
	}

	if hardViolations > 0 {
		result.Notes = fmt.Sprintf("%d hard violations in rolling solution", hardViolations)
	}

	// Write output.
	if config.OutputPath != "" {
		if err := writeResult(config.OutputPath, result); err != nil {
			return result, fmt.Errorf("failed to write output: %w", err)
		}
		result.SolutionPath = config.OutputPath
	}

	return result, nil
}
