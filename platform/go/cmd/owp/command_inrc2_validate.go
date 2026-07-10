package main

import (
	"fmt"
	"os"

	"github.com/timdodgson/open-workforce-platform/platform/go/internal/infrastructure/inrc2"
)

func runValidateINRC2() {
	if len(os.Args) < 6 {
		fmt.Fprintln(os.Stderr, "Usage: owp validate-inrc2 <scenario-file> <week-file> <history-file> <solution-file>")
		os.Exit(1)
	}

	sc, wd, hist, sol, err := loadINRC2WeekInputs(os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	printINRC2ValidationResult(sc, sol, inrc2.Score(sc, wd, hist, sol))
}

func loadINRC2WeekInputs(scenarioPath, weekPath, historyPath, solutionPath string) (inrc2.Scenario, inrc2.WeekData, inrc2.History, inrc2.Solution, error) {
	sc, err := inrc2.LoadScenario(scenarioPath)
	if err != nil {
		return inrc2.Scenario{}, inrc2.WeekData{}, inrc2.History{}, inrc2.Solution{}, fmt.Errorf("error loading scenario: %w", err)
	}
	wd, err := inrc2.LoadWeekData(weekPath)
	if err != nil {
		return inrc2.Scenario{}, inrc2.WeekData{}, inrc2.History{}, inrc2.Solution{}, fmt.Errorf("error loading week data: %w", err)
	}
	hist, err := inrc2.LoadHistory(historyPath)
	if err != nil {
		return inrc2.Scenario{}, inrc2.WeekData{}, inrc2.History{}, inrc2.Solution{}, fmt.Errorf("error loading history: %w", err)
	}
	if solutionPath == "" {
		return sc, wd, hist, inrc2.Solution{}, nil
	}
	sol, err := inrc2.LoadSolution(solutionPath)
	if err != nil {
		return inrc2.Scenario{}, inrc2.WeekData{}, inrc2.History{}, inrc2.Solution{}, fmt.Errorf("error loading solution: %w", err)
	}
	return sc, wd, hist, sol, nil
}
