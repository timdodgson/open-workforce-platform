package main

import (
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/nrp"
	"github.com/timdodgson/open-workforce-platform/platform/go/internal/optimisation"
)

func runValidateINRC2() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "Usage: owp validate-inrc2 <scenario-file> <week-file> <history-file> <solution-file>")
		os.Exit(1)
	}

	scenarioPath := os.Args[2]
	weekPath := os.Args[3]
	historyPath := os.Args[4]
	solutionPath := os.Args[5]

	sc, err := inrc2.LoadScenario(scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading scenario: %v\n", err)
		os.Exit(1)
	}

	wd, err := inrc2.LoadWeekData(weekPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading week data: %v\n", err)
		os.Exit(1)
	}

	hist, err := inrc2.LoadHistory(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
		os.Exit(1)
	}

	sol, err := inrc2.LoadSolution(solutionPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading solution: %v\n", err)
		os.Exit(1)
	}

	result := inrc2.Score(sc, wd, hist, sol)

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

func runSolveINRC2() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "Usage: owp solve-inrc2 <scenario-file> <week-file> <history-file> <solution-output-file> [--algorithm tabu-search] [--profile default]")
		os.Exit(1)
	}

	scenarioPath := os.Args[2]
	weekPath := os.Args[3]
	historyPath := os.Args[4]
	outputPath := os.Args[5]
	algorithm := parseAlgorithm(os.Args[5:])
	profileName := parseProfile(os.Args[5:])

	algProfile, ok := optimisation.GetProfile(profileName)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown profile: %s\n", profileName)
		os.Exit(1)
	}
	algProfile = applyProfileOverrides(os.Args[5:], algProfile)

	sc, err := inrc2.LoadScenario(scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading scenario: %v\n", err)
		os.Exit(1)
	}

	wd, err := inrc2.LoadWeekData(weekPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading week data: %v\n", err)
		os.Exit(1)
	}

	hist, err := inrc2.LoadHistory(historyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
		os.Exit(1)
	}

	var sol inrc2.Solution
	var scoreResult inrc2.ScoreResult

	if algorithm == "parallel-feasible-roster-search" {
		pfrsConfig := parsePFRSConfig(os.Args[5:])
		pfrsSol, pfrsStats, pfrsScore, err := inrc2.SolveWeekPFRS(sc, wd, hist, pfrsConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error solving: %v\n", err)
			os.Exit(1)
		}
		sol = pfrsSol
		scoreResult = pfrsScore

		if err := inrc2.WriteSolution(sol, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing solution: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("=== INRC-II Solve (PFRS) ===\n\n")
		fmt.Printf("Scenario: %s\n", sc.ID)
		fmt.Printf("Week: %d\n", sol.Week)
		fmt.Printf("Algorithm: parallel-feasible-roster-search\n")
		fmt.Printf("Mode: %s\n", pfrsConfig.Mode)
		fmt.Printf("Assignments: %d\n", len(sol.Assignments))
		fmt.Printf("Hard Violations: %d\n", scoreResult.HardViolations)
		fmt.Printf("Soft Penalty: %d\n", scoreResult.SoftPenalty)
		fmt.Printf("Workers Started: %d\n", pfrsStats.WorkersStarted)
		fmt.Printf("Branches Created: %d\n", pfrsStats.BranchesCreated)
		fmt.Printf("Best Updates: %d\n", pfrsStats.BestUpdates)
		fmt.Printf("Invalid Moves Rejected: %d\n", pfrsStats.InvalidMovesRejected)
		fmt.Printf("Iterations: %d\n", pfrsStats.TotalIterations)
		fmt.Printf("Candidates Evaluated: %d\n", pfrsStats.CandidatesEvaluated)
		fmt.Printf("Duration: %dms\n", pfrsStats.DurationMs)
		fmt.Printf("Output: %s\n", outputPath)
		fmt.Println("\nDone.")
		return
	}

	owpSol, _, err := inrc2.SolveWeek(sc, wd, hist, algorithm, algProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error solving: %v\n", err)
		os.Exit(1)
	}
	sol = owpSol

	if err := inrc2.WriteSolution(sol, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing solution: %v\n", err)
		os.Exit(1)
	}

	// Score and display.
	scoreResult = inrc2.Score(sc, wd, hist, sol)

	fmt.Printf("=== INRC-II Solve ===\n\n")
	fmt.Printf("Scenario: %s\n", sc.ID)
	fmt.Printf("Week: %d\n", sol.Week)
	fmt.Printf("Algorithm: %s\n", algorithm)
	fmt.Printf("Assignments: %d\n", len(sol.Assignments))
	fmt.Printf("Hard Violations: %d\n", scoreResult.HardViolations)
	fmt.Printf("Soft Penalty: %d\n", scoreResult.SoftPenalty)
	fmt.Printf("Output: %s\n", outputPath)
	fmt.Println("\nDone.")
}

func runConvertNRP() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: owp convert-nrp <nrp-input> <output-dataset>")
		os.Exit(1)
	}

	inputPath := os.Args[2]
	outputPath := os.Args[3]

	input, err := nrp.LoadNRP(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading NRP file: %v\n", err)
		os.Exit(1)
	}

	dataset := nrp.Convert(input)

	if err := nrp.WriteDataset(dataset, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing dataset: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Converted %d nurses and %d shift demands into OWP dataset.\n", len(input.Nurses), len(dataset.BusinessEvents))
	fmt.Printf("Output: %s\n", outputPath)
}

// parseAlgorithm reads the --algorithm flag from remaining args.
// Defaults to "constructive".
