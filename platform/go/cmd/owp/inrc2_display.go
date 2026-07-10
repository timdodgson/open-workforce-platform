package main

import (
	"fmt"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func printINRC2ValidationResult(sc inrc2.Scenario, sol inrc2.Solution, result inrc2.ScoreResult) {
	fmt.Println("=== INRC-II Validation ===")
	fmt.Println()
	fmt.Printf("Scenario: %s\n", sc.ID)
	fmt.Printf("Week: %d\n", sol.Week)
	fmt.Printf("Assignments: %d\n", len(sol.Assignments))
	fmt.Println()

	fmt.Printf("Hard Violations: %d\n", result.HardViolations)
	if result.HardViolations > 0 {
		for _, v := range result.HardDetails {
			fmt.Printf("  [%s] %s (nurse=%s, day=%s)\n", v.Code, v.Message, v.Nurse, inrc2.DayName(v.Day))
		}
	}
	fmt.Println()

	fmt.Printf("Soft Penalty: %d\n", result.SoftPenalty)
	if len(result.SoftDetails) > 0 {
		fmt.Println("  Breakdown:")
		for _, d := range result.SoftDetails {
			if d.Nurse != "" {
				fmt.Printf("    [%s] nurse=%s penalty=%d\n", d.Constraint, d.Nurse, d.Penalty)
			} else {
				fmt.Printf("    [%s] penalty=%d\n", d.Constraint, d.Penalty)
			}
		}
	}
	fmt.Println()

	fmt.Printf("Total Objective: %d\n", result.TotalObjective)
	fmt.Println()
	fmt.Println("Done.")
}

func printINRC2SolveResult(sc inrc2.Scenario, algorithm string, out inrc2.WeekSolveResult, outputPath string) {
	if out.PFRSStats != nil {
		stats := out.PFRSStats
		fmt.Printf("=== INRC-II Solve (PFRS) ===\n\n")
		fmt.Printf("Scenario: %s\n", sc.ID)
		fmt.Printf("Week: %d\n", out.Solution.Week)
		fmt.Printf("Algorithm: parallel-feasible-roster-search\n")
		fmt.Printf("Mode: %s\n", out.PFRSMode)
		fmt.Printf("Assignments: %d\n", len(out.Solution.Assignments))
		fmt.Printf("Hard Violations: %d\n", out.Score.HardViolations)
		fmt.Printf("Soft Penalty: %d\n", out.Score.SoftPenalty)
		fmt.Printf("Workers Started: %d\n", stats.WorkersStarted)
		fmt.Printf("Branches Created: %d\n", stats.BranchesCreated)
		fmt.Printf("Best Updates: %d\n", stats.BestUpdates)
		fmt.Printf("Invalid Moves Rejected: %d\n", stats.InvalidMovesRejected)
		fmt.Printf("Iterations: %d\n", stats.TotalIterations)
		fmt.Printf("Candidates Evaluated: %d\n", stats.CandidatesEvaluated)
		fmt.Printf("Duration: %dms\n", stats.DurationMs)
		fmt.Printf("Output: %s\n", outputPath)
		fmt.Println("\nDone.")
		return
	}

	fmt.Printf("=== INRC-II Solve ===\n\n")
	fmt.Printf("Scenario: %s\n", sc.ID)
	fmt.Printf("Week: %d\n", out.Solution.Week)
	fmt.Printf("Algorithm: %s\n", algorithm)
	fmt.Printf("Assignments: %d\n", len(out.Solution.Assignments))
	fmt.Printf("Hard Violations: %d\n", out.Score.HardViolations)
	fmt.Printf("Soft Penalty: %d\n", out.Score.SoftPenalty)
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Println("\nDone.")
}
