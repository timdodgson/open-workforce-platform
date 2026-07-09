package ilp

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

// RunBenchmark executes a complete ILP benchmark: build model, solve, validate, output.
func RunBenchmark(sc inrc2.Scenario, weekDataFiles []string, initialHist inrc2.History, config BenchmarkConfig) (BenchmarkResult, error) {
	weeks := config.Weeks
	if weeks > len(weekDataFiles) {
		weeks = len(weekDataFiles)
	}

	// Always use monolithic solve. For benchmarking, time doesn't matter —
	// we want the true optimal or best possible solution.

	// Create temp directory for model file.
	tmpDir, err := os.MkdirTemp("", "ilp-benchmark-*")
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	modelPath := filepath.Join(tmpDir, "model.lp")

	// Build the LP model.
	info, err := BuildModel(sc, weekDataFiles, initialHist, weeks, modelPath)
	if err != nil {
		return BenchmarkResult{}, fmt.Errorf("failed to build model: %w", err)
	}

	// Select solver.
	solver := selectSolver(config.SolverName, config.Parallel)
	if !solver.Available() {
		return BenchmarkResult{
			Instance: config.Instance,
			Weeks:    weeks,
			Solver:   solver.Name(),
			Status:   "ERROR",
			Notes:    fmt.Sprintf("solver '%s' not found in PATH", solver.Name()),
		}, fmt.Errorf("solver not available: %s", solver.Name())
	}

	// Solve.
	timeLimit := config.TimeLimit
	if timeLimit == 0 {
		timeLimit = 5 * time.Minute
	}

	solverOutput, err := solver.Solve(modelPath, timeLimit)
	if err != nil && solverOutput.Status == "ERROR" {
		return BenchmarkResult{
			Instance:       config.Instance,
			Weeks:          weeks,
			Solver:         solver.Name(),
			Status:         "ERROR",
			RuntimeSeconds: solverOutput.RuntimeSeconds,
			TimeLimit:      int(timeLimit.Seconds()),
			Notes:          err.Error(),
		}, err
	}

	result := BenchmarkResult{
		Instance:       config.Instance,
		Weeks:          weeks,
		Solver:         solver.Name(),
		Status:         solverOutput.Status,
		Objective:      int(math.Round(solverOutput.Objective)),
		LowerBound:     int(math.Round(solverOutput.LowerBound)),
		RuntimeSeconds: solverOutput.RuntimeSeconds,
		TimeLimit:      int(timeLimit.Seconds()),
		Threads:        0, // HiGHS decides
		Parallel:       config.Parallel,
	}

	// Calculate gap.
	if result.LowerBound > 0 && result.Objective > 0 {
		result.GapPercent = float64(result.Objective-result.LowerBound) / float64(result.LowerBound) * 100
	}

	// If we have solution values, validate against official scorer.
	if len(solverOutput.SolutionValues) > 0 {
		solutions := ExtractSolutions(sc, weeks, solverOutput.SolutionValues)

		// Generate roster.json for dashboard schedule viewer.
		roster := SolutionsToRoster(sc, solutions)
		if len(roster) > 0 && config.OutputPath != "" {
			rosterPath := filepath.Join(filepath.Dir(config.OutputPath), "roster.json")
			rosterJSON, _ := json.MarshalIndent(roster, "", "  ")
			os.MkdirAll(filepath.Dir(rosterPath), 0755)
			os.WriteFile(rosterPath, rosterJSON, 0644)
			result.RosterPath = rosterPath
		}

		totalPenalty, hardViolations, perWeek, valErr := ValidateILPSolution(sc, weekDataFiles, initialHist, solutions)
		if valErr == nil {
			result.Objective = totalPenalty
			result.HardViolations = hardViolations
			// Recalculate gap with validated objective.
			if result.LowerBound > 0 && result.Objective > 0 {
				result.GapPercent = float64(result.Objective-result.LowerBound) / float64(result.LowerBound) * 100
			}
			if config.OutputPath != "" {
				breakdownPath := filepath.Join(filepath.Dir(config.OutputPath), "constraint-breakdown.json")
				if breakdownJSON, err := MarshalConstraintBreakdown(perWeek); err == nil {
					os.WriteFile(breakdownPath, breakdownJSON, 0644)
					result.ConstraintBreakdownPath = breakdownPath
				}
			}
			if hardViolations > 0 {
				// Build per-week breakdown for notes.
				var parts []string
				for w, wr := range perWeek {
					if wr.HardViolations > 0 {
						parts = append(parts, fmt.Sprintf("week %d: %d hard", w+1, wr.HardViolations))
					}
				}
				result.Notes = fmt.Sprintf("ILP solution has %d hard violations (%s) — model is partial",
					hardViolations, joinStrings(parts, ", "))
			}
			if totalPenalty > int(math.Round(solverOutput.Objective)) {
				// Validator found additional penalty from unmodelled constraints.
				unmModelledPenalty := totalPenalty - int(math.Round(solverOutput.Objective))
				if result.Notes != "" {
					result.Notes += fmt.Sprintf("; unmodelled constraint penalty: %d", unmModelledPenalty)
				} else {
					result.Notes = fmt.Sprintf("Solver objective: %d, validated (all constraints): %d, unmodelled penalty: %d",
						int(math.Round(solverOutput.Objective)), totalPenalty, unmModelledPenalty)
				}
			}
		} else {
			result.Notes = fmt.Sprintf("validation failed: %v", valErr)
		}
	}

	// Model completeness metadata.
	result.SupportedConstraints = info.SupportedConstraints
	result.UnsupportedConstraints = info.UnsupportedConstraints
	if len(info.UnsupportedConstraints) > 0 {
		result.ModelCompleteness = "partial"
	} else {
		result.ModelCompleteness = "full"
	}

	// Write output JSON if path specified.
	if config.OutputPath != "" {
		if err := writeResult(config.OutputPath, result); err != nil {
			return result, fmt.Errorf("failed to write output: %w", err)
		}
		result.SolutionPath = config.OutputPath
	}

	// Copy progress CSV if solver produced one.
	if solverOutput.ProgressCSVPath != "" {
		if _, err := os.Stat(solverOutput.ProgressCSVPath); err == nil {
			progressDest := config.ProgressPath
			if progressDest == "" && config.OutputPath != "" {
				// Default: same directory as output, named ilp-progress.csv
				progressDest = filepath.Join(filepath.Dir(config.OutputPath), "ilp-progress.csv")
			}
			if progressDest != "" {
				data, err := os.ReadFile(solverOutput.ProgressCSVPath)
				if err == nil {
					os.MkdirAll(filepath.Dir(progressDest), 0755)
					os.WriteFile(progressDest, data, 0644)
					result.ProgressPath = progressDest
				}
			}
		}
		// Clean up solver temp directory.
		os.RemoveAll(filepath.Dir(solverOutput.ProgressCSVPath))
	}

	_ = info // available for future logging

	return result, nil
}

// Compare compares a PFRS run result against an ILP benchmark result.
func Compare(ilpResult BenchmarkResult, pfrsPenalty int, pfrsRuntime float64) ComparisonResult {
	gap := pfrsPenalty - ilpResult.Objective
	gapPct := 0.0
	if ilpResult.Objective > 0 {
		gapPct = float64(gap) / float64(ilpResult.Objective) * 100
	}

	return ComparisonResult{
		Instance:     ilpResult.Instance,
		Weeks:        ilpResult.Weeks,
		ILPObjective: ilpResult.Objective,
		ILPStatus:    ilpResult.Status,
		PFRSPenalty:  pfrsPenalty,
		AbsoluteGap:  gap,
		GapPercent:   gapPct,
		ILPRuntime:   ilpResult.RuntimeSeconds,
		PFRSRuntime:  pfrsRuntime,
	}
}

// selectSolver returns the appropriate solver implementation.
func selectSolver(name string, parallel bool) Solver {
	switch name {
	case "highs", "HiGHS", "":
		return &HighsSolver{Parallel: parallel}
	default:
		return &HighsSolver{Parallel: parallel}
	}
}

// writeResult writes a BenchmarkResult to JSON.
func writeResult(path string, result BenchmarkResult) error {
	if dir := filepath.Dir(path); dir != "" {
		os.MkdirAll(dir, 0755)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// LoadBenchmarkResult reads a previously written benchmark result JSON.
func LoadBenchmarkResult(path string) (BenchmarkResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return BenchmarkResult{}, err
	}
	var result BenchmarkResult
	if err := json.Unmarshal(data, &result); err != nil {
		return BenchmarkResult{}, err
	}
	return result, nil
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += sep + p
	}
	return result
}
